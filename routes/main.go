package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type CoffeeOrder struct {
	Size       string `json:"size"`
	BeanType   string `json:"bean_type"`
	Origin     string `json:"origin"`
	SugarCount int    `json:"sugar_count"`
}

func handleCoffeeOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order CoffeeOrder
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if order.Size == "" || order.BeanType == "" || order.Origin == "" || order.SugarCount < 0 {
		http.Error(w, "Missing or invalid parameters", http.StatusBadRequest)
		return
	}

	response := fmt.Sprintf("Order received: %s size coffee, %s beans from %s, with %d sugar(s).",
		order.Size, order.BeanType, order.Origin, order.SugarCount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": response})
}

func main() {
	http.HandleFunc("/order-coffee", handleCoffeeOrder)

	fmt.Println("Starting server on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
