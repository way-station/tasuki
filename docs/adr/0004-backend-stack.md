# ADR-0004: Backend stack — Go with chi

- Status: Proposed
- Type: technical
- Date: 2026-10-11
- Issue: #12

## Context

The original plan had two backends: an API written with Hono on Bun (TypeScript) for PostgreSQL, and a separate Go service for MongoDB (ADR-0002). Hono's RPC feature would also have shared API types with the frontend directly.

Two backends meant two languages, two services to deploy and monitor, and a boundary between them for data that is closely related (races and their lap records). The backend has to talk to both PostgreSQL and MongoDB, run as a container (ADR-0003), and stay easy for a small team to maintain.

## Decision

The backend is **one Go service** (`app/backend`) that serves the API and talks to both PostgreSQL and MongoDB.

- **Routing: chi** (`github.com/go-chi/chi/v5`). It is built on `net/http` (`http.Handler`), so handlers and middleware stay compatible with the standard library.
  - chi is chosen for route groups, sub-routers, and composing middleware per group — for example, routes that require login, routes that require an approved member, and organizer-only routes (ADR-0006).
  - Path parameters alone would not justify chi: since Go 1.22, the standard `ServeMux` supports methods and path patterns such as `/races/{raceID}`.
- **Everything else uses the standard library** where it is enough: `log/slog` for logging, `encoding/json` for JSON.
- **Middleware policy**: `middleware.RealIP` is not used. It is deprecated because it trusts headers such as `X-Forwarded-For`, which lets clients spoof their IP address. If the real client IP is needed later, only headers from a trusted proxy will be used.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| Hono on Bun for PostgreSQL + a separate Go service for MongoDB (original plan) | Two languages and two services for closely related data; more to deploy, monitor, and keep consistent |
| Hono on Bun for everything | Would keep Hono RPC type sharing. We preferred the advantages of Go listed below for a backend that talks to two databases and ships as a small container |
| Go with only `net/http` `ServeMux` | Enough for path patterns, but route groups and per-group middleware would have to be built by hand |
| Echo or Gin | Use their own context types and handler signatures instead of `http.Handler`, which ties handlers and middleware to the framework |

Reasons for Go itself: a single statically linked binary (the production image is about 12 MB on `distroless/static`), static typing, a strong standard library for HTTP and logging, and mature drivers for both PostgreSQL and MongoDB.

## Consequences

- One service, one language, and one container image cover the whole backend.
- **Type sharing with the frontend is lost.** Hono RPC is no longer available, so API types must be shared another way (for example, generating TypeScript types from an OpenAPI description). This is decided in #30.
- The backend depends on chi; replacing it later is limited to the routing layer because handlers are plain `http.Handler`s.
- Developers need Go in addition to TypeScript, but both run inside Dev Containers, so nothing needs to be installed on the host.

---

<details>
<summary>日本語</summary>

# ADR-0004: backend の技術スタック — Go と chi

- Status: Proposed
- Type: technical
- Date: 2026-10-11
- Issue: #12

## Context（背景）

当初の計画では、backend は2つでした。PostgreSQL を扱う Hono on Bun（TypeScript）の API と、MongoDB を扱う別の Go のサービスです（ADR-0002）。Hono の RPC の機能を使えば、API の型を frontend と直接共有することもできました。

backend が2つあると、言語が2つ、デプロイと監視をするサービスが2つになり、密接に関係するデータ（大会と、そのラップ記録）の間にサービスの境目ができます。backend は PostgreSQL と MongoDB の両方を扱い、コンテナとして動き（ADR-0003）、少人数でも保守しやすい必要があります。

## Decision（決定）

backend は、API を提供し、PostgreSQL と MongoDB の両方を扱う **1つの Go のサービス**（`app/backend`）にする。

- **ルーティングは chi**（`github.com/go-chi/chi/v5`）。`net/http`（`http.Handler`）の上に作られているので、ハンドラとミドルウェアは標準ライブラリと互換のまま
  - chi を選ぶ理由は、ルートのグループ化、サブルーター、グループごとのミドルウェアの組み立て。例えば、ログインが必要なルート、承認済みのメンバーが必要なルート、幹事だけのルート（ADR-0006）
  - パスのパラメータだけなら chi は不要。Go 1.22 以降、標準の `ServeMux` でもメソッドと `/races/{raceID}` のようなパスのパターンを書ける
- **それ以外は、足りる限り標準ライブラリを使う。** ログは `log/slog`、JSON は `encoding/json`
- **ミドルウェアの方針**：`middleware.RealIP` は使わない。`X-Forwarded-For` などのヘッダーを信用するため、クライアントが IP アドレスを偽装できるという理由で非推奨になっている。後で本当のクライアントの IP が必要になったら、信頼できるプロキシからのヘッダーだけを使う

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| PostgreSQL は Hono on Bun、MongoDB は別の Go のサービス（当初の計画） | 密接に関係するデータに、言語とサービスが2つずつ。デプロイ、監視、整合性の維持が増える |
| すべて Hono on Bun | Hono RPC による型共有は残せる。2つの DB を扱い、小さなコンテナとして配る backend には、下に挙げる Go の利点のほうを選んだ |
| Go の `net/http` の `ServeMux` だけ | パスのパターンには足りるが、ルートのグループ化とグループごとのミドルウェアを自分で作る必要がある |
| Echo や Gin | `http.Handler` ではなく独自のコンテキストの型とハンドラの形を使うので、ハンドラとミドルウェアがフレームワークに縛られる |

Go そのものを選ぶ理由：静的にリンクされた1つのバイナリ（本番のイメージは `distroless/static` で約 12MB）、静的な型、HTTP とログに強い標準ライブラリ、PostgreSQL と MongoDB の両方に成熟したドライバがあること。

## Consequences（影響）

- backend 全体を、1つのサービス、1つの言語、1つのコンテナイメージで扱える
- **frontend との型共有を失う。** Hono RPC が使えなくなるので、API の型は別の方法で共有する必要がある（例：OpenAPI の記述から TypeScript の型を生成する）。#30 で決める
- backend は chi に依存する。ハンドラは普通の `http.Handler` なので、後で置き換えるときもルーティングの層だけで済む
- 開発者は TypeScript に加えて Go も必要になるが、どちらも Dev Container の中で動くので、ホストには何も入れなくてよい

</details>
