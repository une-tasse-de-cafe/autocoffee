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
	Milk       string `json:milk`
	Name       string `json:"name"`
	SugarCount string `json:"sugar_count"`
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"title": "Thank you!", "message": response, "status": orderStatus})
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./src/index.html")
}

func main() {
	http.HandleFunc("/order-coffee", handleCoffeeOrder)
	http.HandleFunc("/", handleHome)

	fmt.Println("Starting server on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}

}
