COMPOSE_DEV := docker compose -f ../../infra/compose.yaml -f ../../infra/compose.dev.yaml

.PHONY: dev-up dev-shell dev-down

# ---- Run on the host (Docker only, no Node needed) -------------------------
# Inside the container, use pnpm directly (pnpm install, pnpm dev, ...).

# Start the development container without an editor integration, then open a shell in it.
dev-up:
	$(COMPOSE_DEV) up -d --build --wait frontend-dev

dev-shell:
	$(COMPOSE_DEV) exec frontend-dev bash

dev-down:
	$(COMPOSE_DEV) down
