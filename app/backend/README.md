# tasuki backend

The Go API of tasuki (chi, PostgreSQL, MongoDB). This page gets you from a fresh clone to running the backend and its tests. The design, the stack, and the rules are in [`TECH.md`](TECH.md).

## Prerequisites

- Docker
- An editor. VS Code with the Dev Containers extension is the easiest; any editor works with a shell.

Nothing else is installed on your machine. Go, golangci-lint, sqlc, and golang-migrate run inside the Dev Container.

## 1. Check the ports

The containers use these host ports:

| Variable | Default | Used by |
| --- | --- | --- |
| `MONGO_PORT` | `27017` | MongoDB |
| `POSTGRES_PORT` | `5432` | PostgreSQL |
| `BACKEND_PORT` | `8080` | The production image (`make up`) |

If one is already in use, copy [`infra/.env.example`](../../infra/.env.example) to `infra/.env` and change it.
