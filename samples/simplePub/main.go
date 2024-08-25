package main

import (
	"github.com/nats-io/nats.go"
	"log"
)

func main() {

	nc, err := nats.Connect("nats://192.168.128.51:4222")

	if err != nil {
		log.Fatal(err)
	}

	defer nc.Close()
	subj, msg := "coffee.krups-01", []byte("{\"size\":\"medium\", \"bean_type\":\"Arabica\", \"name\":\"Quentin\", \"milk\": \"free\", \"sugar_count\":2 }")
	nc.Publish(subj, msg)

	nc.Flush()

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Published [%s] : '%s'\n", subj, msg)
	}
}
