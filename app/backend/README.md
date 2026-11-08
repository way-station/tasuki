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

## 2. Open the Dev Container

Either way starts three containers: MongoDB, PostgreSQL, and the backend's Dev Container.

**With VS Code:** open `app/backend` and run **Dev Containers: Reopen in Container**. Port 8080 is forwarded to your machine. Closing the window stops the containers.

**With any editor:** run these on your machine, in `app/backend`:

```sh
make dev-up      # build and start the containers
make dev-shell   # open a shell in the Dev Container
make dev-down    # stop them when you are done
```

Port 8080 is not forwarded this way; check the backend from inside the container with `curl localhost:8080/health`.

## 3. Run the backend

Inside the Dev Container:

```sh
make db-migrate   # create the PostgreSQL tables (once, and after new migrations)
make run          # http://localhost:8080/health
```

`/health` answers `200` when MongoDB and PostgreSQL are reachable.

To log in without LINE during development, see [Logging in during development](TECH.md#logging-in-during-development).
