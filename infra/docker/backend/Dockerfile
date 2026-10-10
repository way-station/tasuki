# Build context: app/backend (see infra/compose.yaml)

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
