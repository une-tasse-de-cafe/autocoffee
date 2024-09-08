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
	kvBucket     = "orders-values"
	kvVar        = "orders.pending"
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
		log.Println("Please, provide the NATS URL in NATS_URL")
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

	kv, err := js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket: kvBucket,
	})

	if err != nil {
		log.Fatal(err)
	}

	cc, err := cons.Consume(func(msg jetstream.Msg) {

		consInfo, err := cons.Info(context.TODO())
		if err != nil {
			log.Println("Error getting consInfo: ", err)
			// If the order is invalid, delete it
			return
		}

		numPendingOrders := consInfo.NumAckPending - 1

		_, err = kv.Put(context.TODO(), kvVar, []byte(fmt.Sprint(numPendingOrders)))
		if err != nil {
			log.Println("can't save orders pending to kv : " + err.Error())
		}

		log.Printf("☕ New order : %s", string(msg.Data()))

		msg.InProgress()

		time.Sleep(1 * time.Second)
		msg.Ack()

		var coffee CoffeeOrder
		err = json.Unmarshal(msg.Data(), &coffee)
		if err != nil {
			log.Println("Error unmarshalling order: ", err)
			// If the order is invalid, delete it
			msg.Term()
			return
		}

		number := rand.Intn(100)
		// 5% of chance to fail the coffee
		if number <= 5 {
			log.Println("💢 I failed the coffee order, skipping...")
			msg.Nak()
			return
		} else {

			subjectStock := fmt.Sprintf("coffee.stock.%s.dec.%s", coffee.BeanType, coffee.Size)
			msg.Ack()
			log.Println("📦 Updating the stock...")
			_, err = nc.Request(subjectStock, []byte(""), 2*time.Second)
			if err != nil {
				log.Printf("can't update stock: %s \n", err.Error())
				return
			}
		}

		log.Println("✅ Coffee served")

	})

	if err != nil {
		log.Fatal(err)
	}
	defer cc.Drain()

	log.Println("⌛ Waiting for orders...")
	select {}

}
