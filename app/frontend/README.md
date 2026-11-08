# tasuki frontend

The web app of tasuki (React, TypeScript, Vite). This page gets you from a fresh clone to running the frontend and its tests. The stack, the structure, and the rules are in [`TECH.md`](TECH.md).

## Prerequisites

- Docker
- An editor. VS Code with the Dev Containers extension is the easiest; any editor works with a shell.

Nothing else is installed on your machine. Node and pnpm run inside the Dev Container, with the versions in [`mise.toml`](mise.toml).

**Run `pnpm` only inside the Dev Container, never on your machine.** `node_modules` lives in a Docker volume, and pnpm is set up there to block install scripts and very new releases ([`pnpm-workspace.yaml`](pnpm-workspace.yaml)).

## 1. Check the port

The Vite dev server uses host port `5173` (`FRONTEND_PORT`). If it is already in use, copy [`infra/.env.example`](../../infra/.env.example) to `infra/.env` and change it.

## 2. Open the Dev Container

**With VS Code:** open `app/frontend` and run **Dev Containers: Reopen in Container**. `pnpm install` runs automatically. Closing the window stops the container.

**With any editor:** run these on your machine, in `app/frontend`:

```sh
make dev-up      # build and start the container
make dev-shell   # open a shell in it, then run pnpm install once
make dev-down    # stop it when you are done
```

## 3. Run the frontend

Inside the Dev Container:

```sh
pnpm dev   # http://localhost:5173
```

Files you change on your machine reach the container, so the page reloads as you edit.

The pages get their data from the backend: the dev server forwards `/api` to the backend's Dev Container. To use the screens, also start the backend (see [`app/backend/README.md`](../backend/README.md)) and run `make run` there. To log in without LINE during development, see [Logging in during development](../backend/TECH.md#logging-in-during-development).

## 4. Run the tests and checks

Inside the Dev Container:

```sh
pnpm test           # Vitest; pnpm test:watch reruns on changes
pnpm build          # type check (tsc -b) and production build
pnpm lint           # oxlint, including the import rules between layers
pnpm format         # oxfmt; pnpm format:check only reports
```

The tests need no browser and no backend: they use jsdom and a fake API client. CI runs `pnpm generate:api` (and fails when the types are out of date), `lint`, `format:check`, `test`, and `build` on every pull request.

## Common commands

| Command | Runs on | What it does |
| --- | --- | --- |
| `pnpm install` | Dev Container | Installs the packages |
| `pnpm dev` | Dev Container | Starts the dev server on port 5173 |
| `pnpm test` / `pnpm test:watch` | Dev Container | Runs the tests once / on every change |
| `pnpm build` / `pnpm preview` | Dev Container | Builds for production / serves the build |
| `pnpm lint` | Dev Container | oxlint |
| `pnpm format` / `pnpm format:check` | Dev Container | oxfmt: formats / only reports |
| `pnpm generate:api` | Dev Container | Regenerates the API types after `app/backend/api/openapi.yaml` changes |
| `make dev-up` / `dev-shell` / `dev-down` | Your machine | Starts / enters / stops the Dev Container without VS Code |

## Next

- [`TECH.md`](TECH.md): the stack, the layers of a feature (ADR-0012), styling, routes, server data, tests, lint, and the generated API types.
- [`docs/design/screens.md`](../../docs/design/screens.md): what each screen shows.
- [`docs/adr/`](../../docs/adr): why things are the way they are.

