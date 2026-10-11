# Build context: app/backend (see infra/compose.yaml)
#
# Stages:
#   dev   - development environment for the Dev Container (infra/compose.dev.yaml)
#   build - compiles the production binary
#   (last) production image; the default target

FROM golang:1.26.1 AS dev
COPY --from=golangci/golangci-lint:v2.14.0 /usr/bin/golangci-lint /usr/local/bin/golangci-lint
RUN go install golang.org/x/tools/gopls@v0.23.0 \
 && go install github.com/go-delve/delve/cmd/dlv@v1.27.2

FROM golang:1.26.1 AS build
WORKDIR /src

# Download dependencies first so this layer is cached until go.mod / go.sum change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /server
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
