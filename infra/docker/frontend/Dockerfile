# Build context: app/frontend (see infra/compose.dev.yaml)
#
# Stages:
#   dev - development environment for the Dev Container
# There is no production frontend image: the backend image builds and serves it (ADR-0018).

FROM jdxcode/mise:2026.9.18-debian AS dev

# Install Node and pnpm from app/frontend/mise.toml, used as the global mise config.
COPY mise.toml /mise/config.toml
RUN mise install

# Trust the mise.toml in the mounted repository.
ENV MISE_TRUSTED_CONFIG_PATHS=/workspaces/tasuki
# Keep the pnpm store in a volume (see infra/compose.dev.yaml), never inside the mounted repository.
RUN pnpm config set store-dir /pnpm/store --global

WORKDIR /workspaces/tasuki/app/frontend
