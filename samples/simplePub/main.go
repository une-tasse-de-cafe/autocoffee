package main

import (
	"log"

	"github.com/nats-io/nats.go"
)

func main() {

	nc, err := nats.Connect("nats://192.168.128.51:4222")

	if err != nil {

		log.Fatal(err)

	}

	defer nc.Close()

	subj, msg := "coffee.order", []byte("{\"size\":\"medium\", \"bean_type\":\"Arabica\", \"origin\":\"Colombia\", \"sugar_count\", 2 }")
	nc.Publish(subj, msg)

	nc.Flush()

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Published [%s] : '%s'\n", subj, msg)
	}

}
