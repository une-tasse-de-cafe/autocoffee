package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	nc, _ := nats.Connect(os.Getenv("NATS_URL"))

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	cfg := jetstream.ConsumerConfig{Name: "krups-01"}
	cons, _ := js.CreateConsumer(ctx, "orders", cfg)

	/* cc, err := cons.Consume(func(msg jetstream.Msg) {
		fmt.Printf("New message from %s : %s \n ", msg.Subject(), string(msg.Data()))
		// fmt.Print("ack")
		// msg.Ack()
		//  msg.InProgress()
		time.Sleep(200 * time.Millisecond)
		//  fmt.Print("  nack\n")

	}) */

	msgs, _ := cons.Fetch(3)
	for msg := range msgs.Messages() {
		fmt.Printf("New message from %s : %s \n", msg.Subject(), string(msg.Data()))
		//msg.DoubleAck(ctx)
	}

	// defer cc.Drain()
	defer cancel()
	defer nc.Drain()

	fmt.Println("wait forever")
	for {
	}
}
