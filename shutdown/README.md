# Demo Calling `Next()` In a Separate Goroutine

## Commands

```sh
# Start NATS server
make server
# Remove NATS server container and volume
make clean
```

## Run the Examples

#### tmux

```sh
# Run the default example (With missing heartbeat errors).
# Press Ctrl+C to interrupt the consumer process.
# This example will block until a ErrNoHeartbeat is returned from Next().
./tmux/default.sh
# Run the example without missing heartbeat errors.
# Press Ctrl+C to interrupt the consumer process.
# This example will block indefinitely.
./tmux/nohberr.sh
# Find the process ID
pgrep main
# Kill the process
kill -9 <PID>
```

#### Manual

```sh
# Start the NATS server
make server
# Run the example (publishes and consumes messages).
# Press Ctrl+C to interrupt the process.
# This example will block until a ErrNoHeartbeat is returned from Next().
go run ./cmd/messages/main.go
# Run the example without missing heartbeat errors.
# Press Ctrl+C to interrupt the process.
# This example will block indefinitely.
go run ./cmd/messages/main.go --hberr=false
```
