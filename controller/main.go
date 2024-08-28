package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nats-io/nats.go"
)

const (
	subject = "coffee.web.requests"
)

type controllerResponse struct {
	Status  string `json:"status"` // success or error
	Message string `json:"message"`
}

func main() {

	url := os.Getenv("NATS_URL")
	if url == "" {
		url = "192.168.128.51:4222"
	}

	nc, _ := nats.Connect(url)
	defer nc.Drain()

	sub, _ := nc.Subscribe(subject, func(msg *nats.Msg) {

		fmt.Println("Received a new order: " + string(msg.Data))

		var response controllerResponse
		response.Status = "success"
		response.Message = "The coffee has been successfully scheduled"

		// Handle message

		jsonData, _ := json.Marshal(response)
		fmt.Printf("Status: %s, Message: %s\n", response.Status, response.Message)
		msg.Respond(jsonData)
	})

	fmt.Println("Waiting for orders")
	// wait forever
	for {
	}
	defer sub.Unsubscribe()

}
