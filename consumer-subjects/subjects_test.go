package consumersubjects_test

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// runJetStreamServer starts a NATS server on a random port with JetStream enabled.
func runJetStreamServer(t testing.TB) (srv *server.Server, shutdown func()) {
	opts := natsserver.DefaultTestOptions
	opts.Port = -1
	opts.JetStream = true
	opts.StoreDir = t.TempDir()

	srv = natsserver.RunServer(&opts)

	shutdown = func() {
		srv.Shutdown()
		srv.WaitForShutdown()
	}

	return srv, shutdown
}

func TestConsumerErrorForNonUniqueConsumers(t *testing.T) {
	t.Parallel()

	srv, shutdown := runJetStreamServer(t)
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

	stream, err := js.CreateStream(ctx, jetstream.StreamConfig{
		Name:      "TEST",
		Subjects:  []string{"one", "two"},
		Retention: jetstream.WorkQueuePolicy,
	})
	if err != nil {
		t.Fatal(err)
	}

	cons1, err := stream.CreateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "c1",
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: "one",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("created consumer: %s", cons1.CachedInfo().Name)

	cons2, err := stream.CreateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "c2",
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: "two",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("created consumer: %s", cons2.CachedInfo().Name)

	// Creating a consumer with overlapping subjects returns an error.
	_, err = stream.CreateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "c3",
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: "two", // Overlapping with "c2".
	})
	if err == nil {
		t.Error("want error when creating a consumer with overlapping subjects, got nil")
	}

	// Trying to update the consumer to a non a unique consumer should return an error.
	_, err = stream.UpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "c2",
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: "one", // Overlapping with "c1".
	})
	if err == nil {
		t.Errorf("want error when updating a consumer with overlapping subjects, got nil")
	}
}

func TestConsumerErrorForNonUniqueConsumersFilterSubjects(t *testing.T) {
	t.Parallel()

	srv, shutdown := runJetStreamServer(t)
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

	stream, err := js.CreateStream(ctx, jetstream.StreamConfig{
		Name:      "TEST",
		Subjects:  []string{"one", "two"},
		Retention: jetstream.WorkQueuePolicy,
	})
	if err != nil {
		t.Fatal(err)
	}

	cons1, err := stream.CreateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:        "c1",
		AckPolicy:      jetstream.AckExplicitPolicy,
		FilterSubjects: []string{"one"},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("created consumer: %s", cons1.CachedInfo().Name)

	cons2, err := stream.CreateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:        "c2",
		AckPolicy:      jetstream.AckExplicitPolicy,
		FilterSubjects: []string{"two"},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("created consumer: %s", cons2.CachedInfo().Name)

	// Creating a consumer with overlapping subjects returns an error.
	_, err = stream.CreateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:        "c3",
		AckPolicy:      jetstream.AckExplicitPolicy,
		FilterSubjects: []string{"two"}, // Overlapping with "c2".
	})
	if err == nil {
		t.Error("want error when creating a consumer with overlapping subjects, got nil")
	}

	// Trying to update the consumer to a non a unique consumer should return an error.
	_, err = stream.UpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:        "c2",
		AckPolicy:      jetstream.AckExplicitPolicy,
		FilterSubjects: []string{"one"}, // Overlapping with "c1".
	})
	if err == nil {
		t.Errorf("want error when updating a consumer with overlapping subjects, got nil")
	}
}
