package main

import (
	"os"

	"github.com/nats-io/nats.go"
)

func main() {

	url := os.Getenv("NATS_URL")
	if url == "" {
		url = "192.168.128.51:4222"
	}

	nc, _ := nats.Connect(url)
	defer nc.Drain()

	sub, _ := nc.Subscribe("greet.*", func(msg *nats.Msg) {

		name := msg.Subject[6:]
		msg.Respond([]byte("hello, " + name))
	})

	// wait forever
	for {
	}
	sub.Unsubscribe()

}
