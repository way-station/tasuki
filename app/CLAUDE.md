# Rules for app/

Every change under `app/` (the backend and the frontend) follows these rules, in addition to each app's `TECH.md`. Issue: #379.

## 1. Environment-dependent values come from environment variables

- A value that changes between environments is read from an environment variable and never hard-coded. This applies to both apps. Examples are URLs, origins, ports, hosts, credentials, IDs, and switches.
  - Backend: read it in `internal/config` (`app/backend/TECH.md`).
  - Frontend: the app calls the API at the relative path `/api` (ADR-0014, ADR-0018), so it needs no URL. If a value is needed, use a `VITE_` variable. Vite writes it into the built files, where anyone can read it, so a secret never goes there.
- If a hard-coded value is really needed, write an English comment next to it that says why it is needed, and why no alternative could be found.
- These are not covered: tests, and the local-only values in `infra/compose*.yaml`.

## 2. Constants say why they are constants

A constant is allowed. Each one has an English comment that says three things:

1. Why it is a constant.
2. What the alternative is, such as an environment variable or a setting.
3. Why the alternative was not chosen.

This covers values that could be changed: timeouts, limits, sizes, retry counts, lifetimes, URLs, names of external resources, and defaults. It covers them as Go `const` or package-level `var`, as TypeScript module-level `const`, and also when written inline; an inline one becomes a named constant with this comment.

These are not covered, because they have no alternative to explain:

- Sets of values, such as `race.Attending` or `roster.LinkPending`.
- Text, colors, and class names on screens, such as `labels.ts` and `shared/ui/classes.ts`.
- Unit conversions, such as `60` and `3600`.

```go
// SessionLifetime is how long a session lasts (ADR-0014).
// Why a constant: it is a rule of the login design, the same in every environment.
// Alternative: a SESSION_LIFETIME environment variable.
// Not chosen: no environment needs a different value, and a wrong one would log users out early.
const SessionLifetime = 30 * 24 * time.Hour
```

## 3. Go names

- One-letter names are not allowed. The only exceptions are those in `app/backend/TECH.md` (Naming): `t` for `*testing.T`, `w` for `http.ResponseWriter`, and `i` for the index of a loop of a few lines.
- This includes parameters in interface methods and type parameters, which golangci-lint's `varnamelen` does not check.
- Every other name is meaningful and in lowerCamelCase. Exported names start with a capital letter, because Go exports only those. Initialisms keep one case, as in `userID` and `ServeHTTP` (Go Code Review Comments).

## 4. Tests run in Docker

- Tests run only in the Dev Containers, never with Go, Node, or pnpm installed on the host.
  - Backend: `make test` and `make test-integration` in `backend-dev`.
  - Frontend: `pnpm test` in `frontend-dev`.
- Each PR records the commands that were run in Docker and their results, under how it was checked.

## 5. Static analysis passes

- Backend: `make lint` (golangci-lint) and `go vet ./...` pass.
- Frontend: `pnpm lint` (oxlint), `pnpm format:check` (oxfmt), and `tsc -b` pass.
- Each app's `TECH.md` is followed.

