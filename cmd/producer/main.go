package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("cannot connect to NATS server: %v", err)
	}

	defer nc.Drain()

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatalf("cannot create JetStream instance: %v", err)
	}

	_, err = js.CreateStream(ctx, jetstream.StreamConfig{
		Name:      "jobs",
		Subjects:  []string{"jobs.>"},
		Retention: jetstream.WorkQueuePolicy,
	})

	if err != nil {
		log.Fatalf("cannot create stream: %v", err)
	}

	for i := 0; i < 100; i++ {
		time.Sleep(50 * time.Millisecond)
		fmt.Println("Publishing job", i)

		_, err := js.Publish(ctx, "jobs.info", []byte(fmt.Sprintf("Job payload: %d", i)))

		if err != nil {
			log.Printf("cannot publish job #%d: %v", i, err)
		}
	}
}
