package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nats-io/nats.go"
)

func main() {

	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	fmt.Println("waiting for new tasks...")
	if _, err = nc.Subscribe("coffee.*", func(m *nats.Msg) {
		fmt.Println("New order received")
		fmt.Println(string(m.Data))

	}); err != nil {
		log.Fatal(err)
	}

	for {
	}
}
