package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	topic = "coffee.*"
)

func main() {

	url := os.Getenv("NATS_URL")
	if url == "" {
		url = "192.168.128.51:4222,192.168.128.52:4222"
	}

	nc, _ := nats.Connect(url)
	defer nc.Drain()

	fmt.Printf("Listening to topic %s \n", topic)
	sub, _ := nc.SubscribeSync(topic)

	msg, _ := sub.NextMsg(10 * time.Minute)
	if msg == nil {
		fmt.Println("no message received")
		os.Exit(1)
	}
	fmt.Printf("msg data: %q on subject %q\n", string(msg.Data), msg.Subject)

	time.Sleep(10 * time.Second)
	msg, _ = sub.NextMsg(10 * time.Minute)

	fmt.Printf("msg data: %q on subject %q\n", string(msg.Data), msg.Subject)

}
