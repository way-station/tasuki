COMPOSE     := docker compose -f ../../infra/compose.yaml
COMPOSE_DEV := $(COMPOSE) -f ../../infra/compose.dev.yaml

.PHONY: run build test test-integration lint fmt generate \
	dev-up dev-shell dev-down up down

# ---- Run inside the Dev Container ------------------------------------------
# Environment variables (MONGODB_URI, ...) are provided by infra/compose.dev.yaml.

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

# Regenerates committed code, such as the server code from api/openapi.yaml (ADR-0017).
generate:
	go generate ./...

# ---- Run on the host (Docker only, no Go needed) ---------------------------

# Start the development container without an editor integration, then open a shell in it.
dev-up:
	$(COMPOSE_DEV) up -d --build --wait mongo postgres backend-dev

dev-shell:
	$(COMPOSE_DEV) exec backend-dev bash

dev-down:
	$(COMPOSE_DEV) down

# Build the production image and start it together with MongoDB and PostgreSQL.
up:
	$(COMPOSE) up -d --build --wait mongo postgres backend

down:
	$(COMPOSE) down
