package nats_test

import (
	"context"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	mainStream   = "tasks"
	mirrorStream = "mirror"
	consumerName = "TaskConsumer"
)

var cases = map[string]struct {
	streamSubjects        []string
	mirrorFilterSubject   string
	consumerFilterSubject string
}{
	"multiple non-filtered consumers not allowed": {
		streamSubjects:        []string{"tasks.>"},
		mirrorFilterSubject:   "",
		consumerFilterSubject: "",
	},
	"filtered consumer": {
		streamSubjects:        []string{"tasks.>"},
		mirrorFilterSubject:   "",
		consumerFilterSubject: "tasks.one",
	},
	"mirror and consumer conflicting filters": {
		streamSubjects:        []string{"tasks.>"},
		mirrorFilterSubject:   "tasks.*",
		consumerFilterSubject: "tasks.one",
	},
	// Only successful test case.
	"non conflicting filters": {
		streamSubjects:        []string{"tasks.>"},
		mirrorFilterSubject:   "tasks.example",
		consumerFilterSubject: "tasks.one",
	},
}

func TestMirrorStream(t *testing.T) {
	for name, tc := range cases {
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

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Stream
			_, err = js.CreateStream(ctx, jetstream.StreamConfig{
				Name:      mainStream,
				Subjects:  tc.streamSubjects,
				Retention: jetstream.WorkQueuePolicy, // Required for the failure
				Storage:   jetstream.MemoryStorage,
			})

			if err != nil {
				t.Fatal(err)
			}

			_, err = js.CreateStream(ctx, jetstream.StreamConfig{
				Name:      mirrorStream,
				Retention: jetstream.InterestPolicy,
				Storage:   jetstream.MemoryStorage,
				Mirror: &jetstream.StreamSource{
					Name:          mainStream, // The origin stream to source messages from.
					FilterSubject: tc.mirrorFilterSubject,
				},
			})

			if err != nil {
				t.Fatal(err)
			}

			_, err = js.CreateOrUpdateConsumer(ctx, mainStream, jetstream.ConsumerConfig{
				Durable:       consumerName,
				AckPolicy:     jetstream.AckExplicitPolicy,
				MemoryStorage: true,
				FilterSubject: tc.consumerFilterSubject,
			})

			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestMirrorStreamWithLegacyAPI(t *testing.T) {
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			srv, shutdown := RunJetStreamServer(t)
			defer shutdown()

			nc, err := nats.Connect(srv.ClientURL())

			if err != nil {
				t.Fatal(err)
			}

			defer nc.Close()

			js, err := nc.JetStream()

			if err != nil {
				t.Fatal(err)
			}

			// Stream
			_, err = js.AddStream(&nats.StreamConfig{
				Name:      mainStream,
				Subjects:  tc.streamSubjects,
				Retention: nats.WorkQueuePolicy, // Required for the failure
				Storage:   nats.MemoryStorage,
			})

			if err != nil {
				t.Fatal(err)
			}

			// Mirror stream
			_, err = js.AddStream(&nats.StreamConfig{
				Name: mirrorStream,
				Mirror: &nats.StreamSource{
					Name:          mainStream, // The origin stream to source messages from.
					FilterSubject: tc.mirrorFilterSubject,
				},
				Retention: nats.InterestPolicy,
				Storage:   nats.MemoryStorage,
			})

			if err != nil {
				t.Fatal(err)
			}

			_, err = js.AddConsumer(mainStream, &nats.ConsumerConfig{
				Durable:       consumerName,
				AckPolicy:     nats.AckExplicitPolicy,
				FilterSubject: tc.consumerFilterSubject,
			})

			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
