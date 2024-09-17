package main

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	subject     = "coffee.web.requests"
	concurrency = 5
)

type CoffeeOrder struct {
	Size       string `json:"size"`
	BeanType   string `json:"bean_type"`
	Milk       string `json:"milk"`
	Name       string `json:"name"`
	SugarCount string `json:"sugar_count"`
}

type controllerResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

var (
	nc *nats.Conn
)

func sendOrderToController(order CoffeeOrder) (controllerResponse, error) {
	var response controllerResponse

	log.Printf("☕ New order received for %s", order.Name)

	jsonData, err := json.Marshal(order)
	if err != nil {
		return response, errors.New("Error converting to JSON:" + err.Error())

	}

	rep, err := nc.Request(subject, jsonData, 2*time.Second)
	if err != nil {
		log.Printf("Error sending order: %v", err)
		return response, err
	}

	err = json.Unmarshal(rep.Data, &response)
	if err != nil {
		return response, errors.New("can't unmarshall response from controller")
	}

	return response, nil
}

func handleCoffeeOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var order CoffeeOrder
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		log.Printf("Error unmarshalling order: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	responseTitle := "Thank you!"
	var controllerResponse controllerResponse

	if order.Size == "" || order.BeanType == "" || order.Name == "" || order.SugarCount == "" || order.Milk == "" {
		controllerResponse.Message = "Something's missing 🤔"
		controllerResponse.Status = "error"
		json.NewEncoder(w).Encode(map[string]string{"title": "Oh...", "message": controllerResponse.Message, "status": controllerResponse.Status})
		return
	}

	controllerResponse, err = sendOrderToController(order)
	if err != nil {

		log.Println("💢 Coffee cannot be scheduled !", err)
		controllerResponse.Status = "error"
		responseTitle = "Oh..."
		controllerResponse.Message = "Sadly, we were not able to discuss with our backend system..."
	}

	log.Println("🗣️ Response : ", controllerResponse.Message)

	json.NewEncoder(w).Encode(map[string]string{"title": responseTitle, "message": controllerResponse.Message, "status": controllerResponse.Status})

}

func handleHome(w http.ResponseWriter, numberOfPendingOrders int) {

	tmpl, err := template.ParseFiles("./src/index.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var waitingTime string
	if numberOfPendingOrders != -1 {
		// Each coffee take about 1 minute

		waitingTime = strconv.Itoa(numberOfPendingOrders/concurrency) + "min"
	} else {
		waitingTime = "Unknown"
	}

	data := struct {
		WaitingTime string
	}{
		WaitingTime: waitingTime,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func main() {

	opt, err := nats.NkeyOptionFromSeed("seed.txt")
	if err != nil {
		log.Fatal(err)
	}

	nc, err = nats.Connect(os.Getenv("NATS_URL"), opt)
	if err != nil {
		log.Fatal("connect to nats: ", err)
	}

	defer nc.Drain()

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal(err)
	}

	kv, err := js.CreateKeyValue(context.TODO(), jetstream.KeyValueConfig{
		Bucket: "orders-values",
	})
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/order-coffee", handleCoffeeOrder)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/index", http.StatusMovedPermanently)
	})

	http.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		var numberOfPendingOrders int

		kvValuePendingOrders, err := kv.Get(context.Background(), "orders.pending")
		if err != nil {
			numberOfPendingOrders = -1
		} else {
			numberOfPendingOrders, err = strconv.Atoi(string(kvValuePendingOrders.Value()))
			if err != nil {
				numberOfPendingOrders = -1
			}
		}

		handleHome(w, numberOfPendingOrders)
	})

	log.Println("Starting server on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}

}
