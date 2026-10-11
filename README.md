# tasuki

**A race page for relay marathon teams.** Collect attendance and expected paces, build the running order by drag and drop, and see the predicted total time instantly.

> **Status: early development (pre-MVP).** The development environment and CI are in place; the features below are not implemented yet.

## Why

Relay marathon teams often coordinate everything in a chat app (e.g. LINE notes). That works for messages, but not for structured information:

- **Information gets buried.** Race details, attendance, running order, and fees are mixed in one timeline.
- **Status is invisible.** Organizers count who has not answered yet by hand.
- **Nothing is structured.** When one runner drops out, the whole running order is rewritten by hand.

tasuki does not replace the chat app. **Messages stay in the chat; race information lives in tasuki.** A team shares one URL, and each race gets its own page.

The name comes from the *tasuki*, the sash that relay runners pass from one to the next.

## Features

**MVP**

1. An organizer creates a race page (venue, meeting time, distance, lap-based or fixed-leg format)
2. The URL is shared in the team chat
3. Members open it without logging in, pick their name, and enter attendance and their expected pace
4. The organizer reorders the running order by drag and drop
5. The predicted total time is recalculated instantly

**Planned after the MVP**

- Race-day lap records, and calibrating paces with past results
- Running order history
- Fee collection, practice scheduling, Garmin / Strava integration

## Architecture

```
Browser (React + Vite)        app/frontend
  └─ Go API (chi)             app/backend
       ├─ PostgreSQL          teams, members, races, attendance, running order (source of truth)
       └─ MongoDB             lap records, running order history (secondary)
```

The predicted total time is calculated in the browser; the backend stores and shares data.
Details: [`app/backend/TECH.md`](app/backend/TECH.md), [`app/frontend/TECH.md`](app/frontend/TECH.md).

## Repository layout

| Path | What it is |
| --- | --- |
| [`app/frontend`](app/frontend) | Vite + React + TypeScript |
| [`app/backend`](app/backend) | Go + chi |
| [`infra`](infra) | Dockerfiles, docker compose files, local port settings |
| [`.github`](.github) | CI workflows and PR labeling |

## Getting started

**You only need Docker and an editor.** Go, Node, and pnpm run inside Dev Containers, so nothing is installed on your machine.

1. Clone the repository
2. If ports `27017`, `8080`, or `5173` are already in use, copy `infra/.env.example` to `infra/.env` and change them
3. Open `app/backend` or `app/frontend` in VS Code and run **Dev Containers: Reopen in Container**
   (without VS Code: `make dev-up` and `make dev-shell` in that directory)

See each app's `TECH.md` for the commands that run inside the container.

## Contributing

Contributions are welcome. Issues and pull requests are written in English, with a Japanese translation where helpful.

- **Branches**: start with `feat/`, `fix/`, `chore/`, or `hotfix/`. Pull requests are labeled automatically from the branch name and the changed paths.
- **Commits**: [Conventional Commits](https://www.conventionalcommits.org/) in English, e.g. `feat(backend): add race endpoint`.
- **Pull requests**: describe the change with *Summary*, *Motivation*, *What I have done*, and *Test Plans*. CI (lint, tests, builds) must pass.
- **Design decisions**: important decisions are recorded as ADRs in [`docs/adr/`](docs/adr).

Open work is tracked in [issues](https://github.com/way-station/tasuki/issues) — start with the Todo list in [#18](https://github.com/way-station/tasuki/issues/18).

## License

[MIT](LICENSE)
