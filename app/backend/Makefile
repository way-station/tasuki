include $(wildcard .env)
export

COMPOSE := docker compose -f ../../infra/compose.yaml

.PHONY: run build test test-integration lint fmt mongo-up mongo-down up down

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test -race ./...

test-integration:
	go test -race -tags integration ./...

lint:
	golangci-lint run

fmt:
	golangci-lint fmt

# Start only MongoDB; run the backend on the host with `make run`.
mongo-up:
	$(COMPOSE) up -d --wait mongo

mongo-down:
	$(COMPOSE) stop mongo

# Build the backend image and start it together with MongoDB.
up:
	$(COMPOSE) up -d --build --wait

down:
	$(COMPOSE) down
