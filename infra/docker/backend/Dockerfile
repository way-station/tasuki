# Build context: the repository root for the production image (see infra/compose.yaml),
# which needs both app/backend and app/frontend. The dev stage copies nothing from the
# context, so infra/compose.dev.yaml builds it from app/backend.
#
# Stages:
#   dev            - development environment for the Dev Container (infra/compose.dev.yaml)
#   build          - compiles the production binary
#   frontend-build - builds the frontend (ADR-0018)
#   (last)         production image with both; the default target

FROM golang:1.26.1 AS dev
COPY --from=golangci/golangci-lint:v2.14.0 /usr/bin/golangci-lint /usr/local/bin/golangci-lint
# A prebuilt binary: building sqlc compiles its PostgreSQL parser in C, which takes minutes.
COPY --from=sqlc/sqlc:1.31.1 /workspace/sqlc /usr/local/bin/sqlc
RUN go install golang.org/x/tools/gopls@v0.23.0 \
 && go install github.com/go-delve/delve/cmd/dlv@v1.27.2 \
 && go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.20.1

FROM golang:1.26.1 AS build
WORKDIR /src

# Download dependencies first so this layer is cached until go.mod / go.sum change.
COPY app/backend/go.mod app/backend/go.sum ./
RUN go mod download

COPY app/backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# The same image as the frontend Dev Container, so Node and pnpm come from app/frontend/mise.toml.
FROM jdxcode/mise:2026.9.18-debian AS frontend-build
COPY app/frontend/mise.toml /mise/config.toml
RUN mise install
WORKDIR /src

# Install dependencies first so this layer is cached until the lockfile changes.
COPY app/frontend/package.json app/frontend/pnpm-lock.yaml app/frontend/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

COPY app/frontend/ ./
RUN pnpm build

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /server
COPY --from=frontend-build /src/dist /frontend
ENV FRONTEND_DIR=/frontend
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
