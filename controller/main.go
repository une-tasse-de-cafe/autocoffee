package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	subjectSub                = "coffee.web.requests"
	subjectStockManagerPrefix = "coffee.stock"
	coffeeMakersSubjectPrefix = "coffee.orders"
)

type controllerResponse struct {
	Status  string `json:"status"` // success or error
	Message string `json:"message"`
}

type CoffeeOrder struct {
	Size       string `json:"size"`
	BeanType   string `json:"bean_type"`
	Milk       string `json:"milk"`
	Name       string `json:"name"`
	SugarCount string `json:"sugar_count"`
}

func main() {

	url := os.Getenv("NATS_URL")
	if url == "" {
		url = "192.168.128.51:4222"
	}

	coffeeQuantityMap := make(map[string]int)
	coffeeQuantityMap["small"] = 9
	coffeeQuantityMap["medium"] = 17
	coffeeQuantityMap["large"] = 30

	nc, _ := nats.Connect(url)
	defer nc.Drain()

	sub, _ := nc.Subscribe(subjectSub, func(msg *nats.Msg) {

		var order CoffeeOrder

		fmt.Println("Received a new order: " + string(msg.Data))

		err := json.Unmarshal(msg.Data, &order)
		if err != nil {
			fmt.Println("Error unmarshalling order: ", err)
			return
		}

		var response controllerResponse

		resp, err := nc.Request(fmt.Sprintf("%s.%s.%s", subjectStockManagerPrefix, order.BeanType, "get"), []byte(""), 2*time.Second)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		stockQuantity, err := strconv.Atoi(string(resp.Data))
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Println("Received stock quantity: ", stockQuantity)

		fmt.Println(order)
		fmt.Println("Needed :", coffeeQuantityMap[order.Size])

		// If the quantity of coffee is greater than the amount of coffee needed for the order
		// then schedule the order
		// else return an error response
		fmt.Printf("Order ask for %d (size %s), we have currently %d", coffeeQuantityMap[order.Size], order.Size, stockQuantity)
		if coffeeQuantityMap[order.Size] <= stockQuantity {
			response.Status = "success"
			response.Message = "The coffee has been successfully scheduled"
		} else {
			response.Status = "error"
			response.Message = fmt.Sprintf("Insufficient coffee for type %s", order.BeanType)
		}

		// Notify coffee-makers that the order is scheduled

		if response.Status == "success" {
			err = nc.Publish(fmt.Sprintf("%s.%s", coffeeMakersSubjectPrefix, order.BeanType), msg.Data)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}

		jsonData, _ := json.Marshal(response)
		fmt.Printf("Status: %s, Message: %s\n", response.Status, response.Message)
		msg.Respond(jsonData)
	})

	defer sub.Unsubscribe()

	fmt.Println("Waiting for orders")
	// wait forever
	select {}

}
