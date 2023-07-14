package nats_test

import (
	"testing"

	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
)

func RunJetStreamServer(t testing.TB) (srv *server.Server, shutdown func()) {
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
