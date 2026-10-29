COMPOSE     := docker compose -f ../../infra/compose.yaml
COMPOSE_DEV := $(COMPOSE) -f ../../infra/compose.dev.yaml

.PHONY: run build test test-integration lint fmt generate \
	db-migrate db-rollback db-reset db-new \
	dev-up dev-shell dev-down up down

# ---- Run inside the Dev Container ------------------------------------------
# Environment variables (MONGODB_URI, ...) are provided by infra/compose.dev.yaml.

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

# Extra go test flags; CI passes -failfast to stop at the first failing test.
TESTFLAGS ?=

test:
	go test -race $(TESTFLAGS) ./...

test-integration:
	go test -race -tags integration $(TESTFLAGS) ./...

lint:
	golangci-lint run

fmt:
	golangci-lint fmt

# Regenerates committed code, such as the server code from api/openapi.yaml (ADR-0017).
generate:
	go generate ./...

# Migrations (golang-migrate, ADR-0015). They use DATABASE_URL; pass another one to target a
# different database, for example: make db-migrate DATABASE_URL=postgres://.../tasuki_test
MIGRATE := migrate -path migrations -database "$(DATABASE_URL)"

db-migrate:
	$(MIGRATE) up

db-rollback:
	$(MIGRATE) down 1

# Drops everything in the database, then applies every migration again.
db-reset:
	$(MIGRATE) drop -f
	$(MIGRATE) up

# Creates the next pair of migration files: make db-new NAME=create_accounts
db-new:
	@test -n "$(NAME)" || (echo "usage: make db-new NAME=create_something" && exit 1)
	migrate create -ext sql -dir migrations -seq -digits 6 $(NAME)

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
