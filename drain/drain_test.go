package drain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestJetStreamPullConsumerMessages(t *testing.T) {
	// Test methods that unsubscribe.
	cases := map[string]func(jetstream.MessagesContext){
		"stop":  func(mc jetstream.MessagesContext) { mc.Stop() },
		"drain": func(mc jetstream.MessagesContext) { mc.Drain() },
	}

	for name, unsubscribe := range cases {
		t.Run(name, func(t *testing.T) {
			srv, shutdown := RunJetStreamServer(t)
			defer shutdown()

			nc, err := nats.Connect(srv.ClientURL())
			if err != nil {
				t.Fatal(err)
			}
			defer nc.Close()

			js, err := jetstream.New(nc)
			if err != nil {
				t.Fatal(err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			s, err := js.CreateStream(ctx, jetstream.StreamConfig{
				Name:     "TEST_STREAM",
				Subjects: []string{"FOO.*"},
			})
			if err != nil {
				t.Fatal(err)
			}

			cons, err := s.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
				Durable:   "TestConsumer",
				AckPolicy: jetstream.AckExplicitPolicy,
			})
			if err != nil {
				t.Fatal(err)
			}

			iter, err := cons.Messages()
			if err != nil {
				t.Fatal(err)
			}

			errs := make(chan error)
			wait := make(chan struct{})

			go func() {
				close(wait)
				// Next blocks waiting for new messages.
				// It returns ErrMsgIteratorClosed when Stop() or Drain() are called.
				_, err := iter.Next()
				errs <- err
			}()

			// Make sure the goroutine calling Next() is running.
			<-wait

			// Unsubscribing using Stop() or Drain() should make Next() return ErrMsgIteratorClosed.
			unsubscribe(iter)

			select {
			case <-time.After(5 * time.Second):
				t.Errorf("timed out waiting for Next() to return")
			case err := <-errs:
				if !errors.Is(err, jetstream.ErrMsgIteratorClosed) {
					t.Errorf("want error %q, got %q", jetstream.ErrMsgIteratorClosed, err)
				}
			}
		})
	}
}
