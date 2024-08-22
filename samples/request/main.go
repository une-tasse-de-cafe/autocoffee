package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {

	url := os.Getenv("NATS_URL")
	if url == "" {
		url = "192.168.128.51:4222"
	}

	nc, _ := nats.Connect(url)
	defer nc.Drain()

	rep, err := nc.Request("greet.joe", nil, time.Second)
	if err != nil {
		fmt.Println("Err " + err.Error())
		return
	}
	fmt.Println(rep)

	fmt.Println(string(rep.Data))

	rep, _ = nc.Request("greet.sue", nil, time.Second)
	fmt.Println(string(rep.Data))

	rep, _ = nc.Request("greet.bob", nil, time.Second)
	fmt.Println(string(rep.Data))

	//sub.Unsubscribe()

	_, err = nc.Request("greet.joe", nil, time.Second)
	fmt.Println(err)
}
