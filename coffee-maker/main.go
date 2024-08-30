package main

import (
	"context"
	"encoding/json"
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

type CoffeeOrder struct {
	Size       string `json:"size"`
	BeanType   string `json:"bean_type"`
	Milk       string `json:"milk"`
	Name       string `json:"name"`
	SugarCount string `json:"sugar_count"`
}

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

		var coffee CoffeeOrder
		err := json.Unmarshal(msg.Data(), &coffee)
		if err != nil {
			fmt.Println("Error unmarshalling order: ", err)
			// If the order is invalid, delete it
			msg.Term()
			return
		}

		msg.InProgress()

		number := rand.Intn(100)
		// 5% of chance to fail the coffee
		if number <= 5 {
			fmt.Print("--- failed !")
			msg.Nak()
			return
		} else {

			subjectStock := fmt.Sprintf("coffee.stock.%s.dec.%s", coffee.BeanType, coffee.Size)
			msg.Ack()
			_, err := nc.Request(subjectStock, []byte(""), 2*time.Second)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
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
