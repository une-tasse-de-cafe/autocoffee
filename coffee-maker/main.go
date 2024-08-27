package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"golang.org/x/exp/rand"
)

const (
	consumerName = "coffeeMakers"
	subjects     = "coffee.orders.*"
	streamName   = "coffee-orders"
)

func main() {

	natsUrl := os.Getenv("NATS_URL")
	if natsUrl == "" {
		fmt.Println("Please, provide the NATS URL in NATS_URL")
		os.Exit(1)
	}

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
		Name:        streamName,
		Subjects:    []string{subjects},
		Storage:     jetstream.FileStorage,
		Retention:   jetstream.WorkQueuePolicy,
		AllowDirect: true,
	}

	_, err = js.CreateOrUpdateStream(ctx, cfgStream)
	if err != nil {
		log.Fatal(err)
	}

	cfgConsu := jetstream.ConsumerConfig{
		Name:          consumerName,
		FilterSubject: subjects,
		Durable:       consumerName,
	}

	cons, err := js.CreateConsumer(ctx, cfgStream.Name, cfgConsu)
	if err != nil {
		log.Fatal(err)
	}

	cc, err := cons.Consume(func(msg jetstream.Msg) {
		fmt.Printf("New message from %s : %s ", msg.Subject(), string(msg.Data()))
		msg.InProgress()

		number := rand.Intn(10)
		if number == 0 {
			fmt.Print("--- failed !")
			msg.Nak()
		} else {
			//			fmt.Print("- succeed")
			time.Sleep(200 * time.Millisecond)
			msg.Ack()
		}

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
