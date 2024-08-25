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
	defer nc.Close()

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := jetstream.ConsumerConfig{Name: "krups-01"}
	cons, _ := js.CreateConsumer(ctx, "orders", cfg)

	fmt.Println("# Consume messages using Consume()")

	cc, err := cons.Consume(func(msg jetstream.Msg) {
		fmt.Printf("New message from %s : %s - ", msg.Subject(), string(msg.Data()))
		fmt.Print("ack")
		msg.InProgress()
		//		fmt.Print("  nack\n")
		time.Sleep(500 * time.Millisecond)
		fmt.Printf("\n")

	})
	defer cc.Drain()

	fmt.Println("wait forever")
	for {
	}
}
