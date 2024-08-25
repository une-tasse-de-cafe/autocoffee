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

const (
	consumerName = "coffee-maker"
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

	cfgStream := jetstream.StreamConfig{
		Replicas:    3,
		Name:        "coffeeorders",
		Subjects:    []string{"coffee.orders"},
		Storage:     jetstream.FileStorage,
		Retention:   jetstream.InterestPolicy,
		AllowDirect: true,
	}

	_, err = js.CreateOrUpdateStream(ctx, cfgStream)
	if err != nil {
		log.Fatal(err)
	}

	cfgConsu := jetstream.ConsumerConfig{
		Name:          consumerName,
		FilterSubject: "coffee.orders",
		Durable:       consumerName,
	}

	cons, err := js.CreateConsumer(ctx, cfgStream.Name, cfgConsu)
	if err != nil {
		log.Fatal(err)
	}

	cc, err := cons.Consume(func(msg jetstream.Msg) {
		fmt.Printf("New message from %s : %s - ", msg.Subject(), string(msg.Data()))
		msg.Ack()
		time.Sleep(500 * time.Millisecond)
		fmt.Printf("\n")
	})

	if err != nil {
		log.Fatal(err)
	}
	defer cc.Drain()

	fmt.Println("wait forever")
	for {
	}
}
