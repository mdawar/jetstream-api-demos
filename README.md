# NATS JetStream Simplified API Demo

## Commands

```sh
# Start NATS server
make server
# Remove NATS server container and volume
make clean
# Start producer
make producer
# Start worker using Consume() (2s delay)
make consumeworker
# Start a slow worker using Consume() (30s delay)
make slowconsumeworker
# Start worker using Messages() (2s delay)
make messagesworker
# Start a slow worker using Messages() (30s delay)
make slowmessagesworker
```

## tmux Examples

Demos using `Consume()`:

```sh
# Demo showing how messages are pulled by workers
./tmux/consumeworker.sh
# Demo with slow workers showing the heartbeat errors
# Run `make produce` a second time to see how slow workers pull
# more messages that they should.
./tmux/slowconsumeworker.sh
```

Demos using `Messages()`:

```sh
# Demo showing how messages are pulled by workers with fair distribution
./tmux/messagesworker.sh
# Demo with slow workers showing no heartbeat errors and with expected distribution
# Run `make produce` a second time to see the same expected behavior
./tmux/slowmessagesworker.sh
```
