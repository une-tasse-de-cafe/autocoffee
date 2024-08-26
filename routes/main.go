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

func sendOrderToController(order CoffeeOrder) error {
	url := os.Getenv("NATS_URL")
	if url == "" {
		return errors.New("Please provide nats url in NATS_URL env")
	}

	nc, _ := nats.Connect(url)
	defer nc.Drain()

	jsonData, err := json.Marshal(order)
	if err != nil {
		return errors.New("Error converting to JSON:" + err.Error())

	}
	rep, err := nc.Request(subject, jsonData, 2*time.Second)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	var response controllerResponse
	err = json.Unmarshal(rep.Data, &response)
	if err != nil {
		return errors.New("can't unmarshall response from controller")
	}

	return nil
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

	var orderStatus string
	responseTitle := "Thank you!"
	orderStatus = "success"

	response := fmt.Sprintf("Order received from %s : %s size coffee, %s beans, with %s and  %s sugar(s).",
		order.Name, order.Size, order.BeanType, order.Milk, order.SugarCount)
	fmt.Println(response)

	if order.Size == "" || order.BeanType == "" || order.Name == "" || order.SugarCount == "" || order.Milk == "" {
		response = "Something's missing 🤔"
		orderStatus = "error"
	}

	if orderStatus != "failed" {
		err = sendOrderToController(order)
		if err != nil {
			orderStatus = "error"
			responseTitle = "Oh..."
			response = "Sadly, we can't transfer your order to our backend 😢"
		}
	}

	json.NewEncoder(w).Encode(map[string]string{"title": responseTitle, "message": response, "status": orderStatus})

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
