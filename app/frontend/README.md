# tasuki frontend

The web app of tasuki (React, TypeScript, Vite). This page gets you from a fresh clone to running the frontend and its tests. The stack, the structure, and the rules are in [`TECH.md`](TECH.md).

## Prerequisites

- Docker
- An editor. VS Code with the Dev Containers extension is the easiest; any editor works with a shell.

Nothing else is installed on your machine. Node and pnpm run inside the Dev Container, with the versions in [`mise.toml`](mise.toml).

**Run `pnpm` only inside the Dev Container, never on your machine.** `node_modules` lives in a Docker volume, and pnpm is set up there to block install scripts and very new releases ([`pnpm-workspace.yaml`](pnpm-workspace.yaml)).

## 1. Check the port

The Vite dev server uses host port `5173` (`FRONTEND_PORT`). If it is already in use, copy [`infra/.env.example`](../../infra/.env.example) to `infra/.env` and change it.
