package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	subject = "coffee.web.requests"
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

func sendOrderToController(order CoffeeOrder) (controllerResponse, error) {
	var response controllerResponse
	url := os.Getenv("NATS_URL")
	if url == "" {
		return response, errors.New("Please provide nats url in NATS_URL env")
	}

	nc, _ := nats.Connect(url)
	defer nc.Drain()

	jsonData, err := json.Marshal(order)
	if err != nil {
		return response, errors.New("Error converting to JSON:" + err.Error())

	}

	rep, err := nc.Request(subject, jsonData, 2*time.Second)
	if err != nil {
		fmt.Println(err.Error())
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
		fmt.Println(err.Error())
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
		fmt.Println(err.Error())
		controllerResponse.Status = "error"
		responseTitle = "Oh..."
		controllerResponse.Message = "Sadly, we were not able to discuss with our backend system..."
	}

	fmt.Println(controllerResponse)

	json.NewEncoder(w).Encode(map[string]string{"title": responseTitle, "message": controllerResponse.Message, "status": controllerResponse.Status})

	return
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./src/index.html")
}

func main() {
	http.HandleFunc("/order-coffee", handleCoffeeOrder)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/index", http.StatusMovedPermanently)
	})

	http.HandleFunc("/index", handleHome)

	fmt.Println("Starting server on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}

}
