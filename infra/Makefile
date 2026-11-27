COMPOSE := docker compose -f compose.yaml -f compose.dev.yaml --profile infra
RUN     := $(COMPOSE) run --rm

.PHONY: gcloud gcloud-login gcloud-forget

# ---- Run on the host (Docker only, no gcloud needed) ----------------------
# Production tools for ADR-0021. Each target starts a container for one command.

# Any gcloud command: make gcloud ARGS="run services list"
gcloud:
	$(RUN) gcloud gcloud $(ARGS)

# Signs gcloud in. Open the printed URL in the host's browser, then paste the code back.
gcloud-login:
	$(RUN) gcloud gcloud auth login --no-launch-browser

# Deletes the sign-in volume, so no credentials stay on this machine. Only this volume:
# the databases' volumes belong to the same Compose project and are kept.
gcloud-forget:
	docker volume rm tasuki_gcloud-config
