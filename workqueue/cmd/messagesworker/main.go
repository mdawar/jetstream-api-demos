package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	var isSlow bool
	flag.BoolVar(&isSlow, "slow", false, "Start a slow worker")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nc, err := nats.Connect(nats.DefaultURL, nats.Name("Worker"))
	if err != nil {
		log.Fatalf("cannot connect to NATS server: %v", err)
	}

	defer nc.Drain()

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatalf("cannot create JetStream instance: %v", err)
	}

	stream, err := js.CreateStream(ctx, jetstream.StreamConfig{
		Name:      "jobs",
		Subjects:  []string{"jobs.>"},
		Retention: jetstream.WorkQueuePolicy,
	})

	if err != nil {
		log.Fatalf("cannot create stream: %v", err)
	}

	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "JobsConsumer",
		ReplayPolicy:  jetstream.ReplayInstantPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       5 * time.Minute,
		MaxDeliver:    1,
		MaxAckPending: -1,
	})

	if err != nil {
		log.Fatalf("cannot create consumer: %v", err)
	}

	fmt.Println("Consumer ID:", consumer.CachedInfo().Name)

	iter, err := consumer.Messages(
		jetstream.PullMaxMessages(1),
		jetstream.WithMessagesErrOnMissingHeartbeat(true),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer iter.Stop()

	parallelConsumers := 5
	sem := make(chan struct{}, parallelConsumers)

	for {
		sem <- struct{}{}

		go func() {
			defer func() {
				<-sem
			}()

			msg, err := iter.Next()
			if err != nil {
				fmt.Println("next err: ", err)
				return
			}

			fmt.Printf("Received msg: %s\n", msg.Data())

			// Simulate work
			if isSlow {
				time.Sleep(30 * time.Second)
			} else {
				time.Sleep(2 * time.Second)
			}

			msg.Ack()
		}()
	}
}
