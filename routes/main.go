package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/nats-io/nats.go"
)

const (
	streamName = "ORDERS-WEB"
	subject    = "coffee.web.requests"
)

type CoffeeOrder struct {
	Size       string `json:"size"`
	BeanType   string `json:"bean_type"`
	Milk       string `json:milk`
	Name       string `json:"name"`
	SugarCount string `json:"sugar_count"`
}

func sendOrderToController(order CoffeeOrder) error {
	url := os.Getenv("NATS_URL")
	if url == "" {
		return errors.New("Please provide nats url in NATS_URL env")
	}

	nc, _ := nats.Connect(url)
	defer nc.Drain()

	js, _ := nc.JetStream()

	js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{subject},
	})

	jsonData, err := json.Marshal(order)
	if err != nil {
		return errors.New("Error converting to JSON:" + err.Error())

	}
	ack, err := js.Publish(subject, jsonData)

	if err != nil {
		return err
	}

	fmt.Println(ack)
	return nil
}

func handleCoffeeOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order CoffeeOrder
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if order.Size == "" || order.BeanType == "" || order.Name == "" || order.SugarCount == "" || order.Milk == "" {
		http.Error(w, "Missing or invalid parameters", http.StatusBadRequest)
		return
	}
	// orderStatus must be "succeed" or failed
	var orderStatus string
	orderStatus = "success"

	response := fmt.Sprintf("Order received from %s : %s size coffee, %s beans, with %s and  %s sugar(s).",
		order.Name, order.Size, order.BeanType, order.Milk, order.SugarCount)
	fmt.Println(response)
	err = sendOrderToController(order)

	if err != nil {
		fmt.Println("ERR")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"title": "Thank you!", "message": response, "status": orderStatus})
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
