package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
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

	// Parallel consume.
	for i := 0; i < 5; i++ {
		cc, err := consumer.Consume(
			func(consumeID int) jetstream.MessageHandler {
				return func(msg jetstream.Msg) {
					fmt.Printf("Received msg on consume %d: %s\n", consumeID, msg.Data())

					// Simulate work
					if isSlow {
						time.Sleep(30 * time.Second)
					} else {
						time.Sleep(2 * time.Second)
					}

					msg.Ack()
				}
			}(i),
			jetstream.PullMaxMessages(1),
			jetstream.ConsumeErrHandler(func(consumeCtx jetstream.ConsumeContext, err error) {
				fmt.Printf("consumer error: %v\n", err)
			}),
		)

		if err != nil {
			log.Fatal(err)
		}

		defer cc.Stop()
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
