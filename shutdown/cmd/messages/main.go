package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Example taken from: https://github.com/nats-io/nats.go/blob/main/examples/jetstream/js-messages/main.go
// Changed to run Next() in a separate goroutine.
func main() {
	var errOnMissingHeartbeat bool
	flag.BoolVar(&errOnMissingHeartbeat, "hberr", true, "Return error for missing heartbeat from Next()")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal(err)
	}
	s, err := js.CreateStream(ctx, jetstream.StreamConfig{
		Name:     "TEST_STREAM",
		Subjects: []string{"FOO.*"},
	})
	if err != nil {
		log.Fatal(err)
	}

	cons, err := s.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:   "TestConsumerMessages",
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		log.Fatal(err)
	}
	go endlessPublish(ctx, nc, js)

	iter, err := cons.Messages(
		jetstream.PullMaxMessages(1),
		jetstream.WithMessagesErrOnMissingHeartbeat(errOnMissingHeartbeat), // When false, Next() will block indefinitely
	)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			// Next() blocks until a message is available.
			// After calling Stop() keeps blocking until we get an ErrNoHeartbeat,
			// then next call will return ErrMsgIteratorClosed because of the
			// atomic.LoadUint32(&s.closed) == 1 check in Next().
			// If ReportMissingHeartbeats is false, then it will block indefinitely.
			msg, err := iter.Next()

			if err != nil {
				if errors.Is(err, jetstream.ErrMsgIteratorClosed) {
					fmt.Println("Iterator closed")
					break
				}

				fmt.Println("Next err: ", err)
				// If Stop() was called then the next Next() call will return ErrMsgIteratorClosed.
				continue
			}

			fmt.Println(string(msg.Data()))
			msg.Ack()
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	c := <-sig

	fmt.Println("Received signal:", c)
	iter.Stop() // Should make Next() return ErrMsgIteratorClosed

	fmt.Println("Waiting for consumer goroutine...")
	wg.Wait()

	fmt.Println("Done")
}

func endlessPublish(ctx context.Context, nc *nats.Conn, js jetstream.JetStream) {
	var i int
	for {
		time.Sleep(500 * time.Millisecond)
		if nc.Status() != nats.CONNECTED {
			continue
		}
		if _, err := js.Publish(ctx, "FOO.TEST1", []byte(fmt.Sprintf("msg %d", i))); err != nil {
			fmt.Println("pub error: ", err)
		}
		i++
	}
}
