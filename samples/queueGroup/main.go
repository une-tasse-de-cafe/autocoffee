package main

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

func main() {

	nc, err := nats.Connect("192.168.128.51:4222,192.168.128.52:4222,192.168.128.53:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	if _, err := nc.QueueSubscribe("coffee.>", "coffee-maker", func(m *nats.Msg) {
		fmt.Println("New order received")
		fmt.Println(string(m.Data))
	}); err != nil {
		log.Fatal(err)
	}

	for {
	}

}
