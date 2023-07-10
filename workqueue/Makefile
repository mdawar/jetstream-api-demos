server:
	docker compose up --no-log-prefix

clean:
	docker compose down -v

producer:
	go run ./cmd/producer/main.go

consumeworker:
	go run ./cmd/consumeworker/main.go

slowconsumeworker:
	go run ./cmd/consumeworker/main.go -slow

messagesworker:
	go run ./cmd/messagesworker/main.go

slowmessagesworker:
	go run ./cmd/messagesworker/main.go -slow
