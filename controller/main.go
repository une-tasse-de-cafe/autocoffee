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
	Status  string `json:"status"`
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
		response := controllerResponse{
			Status:  "success",
			Message: "The coffee has been successfully scheduled",
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			fmt.Println("Error marshaling to JSON:", err)
			msg.Respond([]byte("fail"))
		}
		fmt.Printf("Status: %s, Message: %s\n", response.Status, response.Message)

		msg.Respond(jsonData)
	})

	fmt.Println("Waiting for orders")
	// wait forever
	for {
	}
	defer sub.Unsubscribe()

}
