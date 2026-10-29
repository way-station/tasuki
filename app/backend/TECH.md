# backend — Tech Stack

The tasuki backend is a single Go service. It serves the API for the frontend (`app/frontend`) and owns both databases:

```
React (Vite)  ── app/frontend
  └─ Go (chi) ── app/backend
       ├─ PostgreSQL  teams, members, races, attendance, running order  ← source of truth
       └─ MongoDB     lap records, running order history, external data ← secondary
```

| Data | Store | Why |
| --- | --- | --- |
| Teams, members, races, attendance, entries, running order | PostgreSQL | Relational, with constraints the DB must guarantee (one attendance per member per race, only entry members in the running order) |
| Lap records | MongoDB | Per race and roster member, a variable-length list of lap times (per km, `MM:SS`) that members add to, edit, and delete |
| Running order history | MongoDB | "Who changed the order, when, from what to what." A full snapshot per change — a typical event log |
| External integration data (future) | MongoDB | Raw activities from Garmin / Strava have a different shape per service |

## Principles (do not break these)

1. **Postgres is always the source of truth.** There are no transactions across the two databases.
2. **Reference, never copy.** MongoDB documents refer to Postgres rows only by their IDs (the ID type is decided in #68). Do not duplicate Postgres data (member names, race info, …) into MongoDB — duplicated data always drifts.
3. **Write summaries back to Postgres.** When laps are submitted, edited, or deleted, write each member's average time and lap count back to `roster_members` in Postgres (`docs/design/database.md`). Atlas M0 (free tier) has no automatic backups and idle clusters may be terminated, so actual results must survive even if MongoDB data is lost.
4. **The MVP uses both databases.** Lap records (`lap_records`) and running order history (`running_order_history`) are in MongoDB; everything else is in Postgres (`docs/design/database.md`).
5. **Predictions are calculated by the backend.** The total time and the number of laps are calculated in Application (ADR-0011) and returned as results. Members may not see other members' average times (ADR-0006), so the frontend never receives the data to calculate them.

## Stack

| Area | Technology | Version |
| --- | --- | --- |
| Language | Go | 1.26 (see `go.mod`) |
| HTTP router | `github.com/go-chi/chi/v5` | v5.3.x |
| MongoDB driver | `go.mongodb.org/mongo-driver/v2` | v2.9.x |
| PostgreSQL driver | `github.com/jackc/pgx/v5` with `pgxpool`; Neon direct endpoint (ADR-0015) | v5.x |
| SQL | sqlc (Go generated from SQL, committed) (ADR-0015) | — |
| Migrations | golang-migrate, zero-padded versions (ADR-0015) | — |
| Logging | `log/slog` (JSON) | stdlib |
| Lint / format | golangci-lint v2 (`.golangci.yml`) | v2.x |
| API code generation | oapi-codegen; generated code is committed (ADR-0017) | v2.8.0 |
| Container | Multi-stage build → `distroless/static` (`infra/docker/backend/Dockerfile`) | — |
| Local DB | MongoDB in Docker (`infra/compose.yaml`) | 8.0 |
| Production DB | Neon (PostgreSQL), MongoDB Atlas — both in Singapore (ADR-0018) | — |
| Hosting | Cloud Run, `asia-southeast1`; one container serving the API and the built frontend (ADR-0018) | — |

### Why these choices

- **Go for the whole backend**: one language and one service for the API and both databases.
- **chi**: needed for routing — route groups, sub-routers, and URL parameters such as `/races/{raceID}/laps`. It is built on `net/http` (`http.Handler`), so handlers and middleware stay compatible with the standard library.
  - Do not use `middleware.RealIP`: it is deprecated because it trusts `X-Forwarded-For` and similar headers and allows IP spoofing.
- **Standard library otherwise**: logging (`log/slog`) and JSON use the standard library.

## Directory structure

```
app/backend/
├─ cmd/server/            # entry point: config → DB connections → HTTP server, graceful shutdown
├─ internal/
│  ├─ config/             # loads settings from environment variables
│  ├─ presentation/       # routing (chi), controllers (`controller/`), and the code generated from `api/openapi.yaml` (`api/`)
│  ├─ application/        # use cases, commands (input validation), DTOs / view models, their factories, query service interfaces
│  ├─ domain/             # entities, value objects, domain services, entity factories, repository interfaces
│  └─ infrastructure/     # DB connections, repositories, and query services, per store (`mongodb/`, `postgres/`)
├─ migrations/           # PostgreSQL schema: golang-migrate up / down SQL files (ADR-0015)
├─ .devcontainer/         # Dev Container definition (development happens inside the container)
└─ Makefile
```

- Onion architecture with CQRS ([ADR-0011](../../docs/adr/0011-onion-architecture-cqrs.md)). Dependencies point inward: `presentation` → `application` → `domain`; `infrastructure` → `application`, `domain`. `domain` imports no other layer.
- Writes go through use cases and repositories; reads go through query services.
- Entities, DTOs, and commands are created by factories, not by themselves. Entities and value objects keep their fields unexported.
- Everything is under `internal/`; nothing is meant to be imported from outside this module.
- Inside a layer, code is grouped per feature (for example `application/health/`).
- `/health` follows the layers too: `controller.HealthController` → `health.CheckUseCase` → `health.QueryService` (`health.MongoQueryService` and `health.PostgresQueryService`, implemented by `mongodb.HealthQueryService` and `postgres.HealthQueryService`).

Module path: `github.com/way-station/tasuki/app/backend`

## Environment variables

The application reads **only environment variables**; it never loads `.env` itself. The same binary runs everywhere, and only the values differ.

| Name | Required | Default | Description |
| --- | --- | --- | --- |
| `MONGODB_URI` | yes | — | MongoDB connection string |
| `MONGODB_DATABASE` | no | `tasuki` | MongoDB database name |
| `DATABASE_URL` | yes | — | PostgreSQL connection string. Pool sizes go in its parameters (`pool_max_conns`, `pool_min_conns`, `pool_max_conn_lifetime`, `pool_max_conn_idle_time`; ADR-0015) |
| `PORT` | no | `8080` | HTTP port |
| `BACKEND_PORT` | no | `8080` | Host port of the backend container (`infra/.env`, compose only) |
| `MONGO_PORT` | no | `27017` | Host port of the MongoDB container (`infra/.env`, compose only) |
| `POSTGRES_PORT` | no | `5432` | Host port of the PostgreSQL container (`infra/.env`, compose only) |

Who provides the values:

| Environment | Provided by |
| --- | --- |
| Dev Container | `environment` in `infra/compose.dev.yaml` |
| Production image run locally (`make up`) | `environment` in `infra/compose.yaml` |
| Production | The hosting platform's environment variables / secrets |

Host port overrides go in `infra/.env` (copy `infra/.env.example`); docker compose reads it automatically. Never commit `.env`.

## Local development

**No Go toolchain is needed on the host.** The host only needs Docker (and an editor). All development work happens inside the Dev Container.

### Option A: Dev Container (recommended)

1. If host ports 27017 / 5432 / 8080 are already in use, copy `infra/.env.example` to `infra/.env` and change them.
2. Open `app/backend` in VS Code and run **Dev Containers: Reopen in Container** (JetBrains: open `.devcontainer/devcontainer.json` via Gateway).
3. In the container terminal:

```sh
make db-migrate         # creates the PostgreSQL tables
make run                # http://localhost:8080/health (port is forwarded to the host)
make test
make test-integration   # uses the MongoDB and PostgreSQL containers
make lint
```

The editor tools (gopls, golangci-lint, Delve) are installed in the container with pinned versions.

### Option B: Shell only (any editor)

```sh
cd app/backend
make dev-up             # build and start the dev container and MongoDB
make dev-shell          # open a shell in it, then run make test etc.
make dev-down
```

### Where each command runs

| Command | Runs on |
| --- | --- |
| `make run` / `build` / `test` / `test-integration` / `lint` / `fmt` / `generate` | Inside the Dev Container |
| `make db-migrate` / `db-rollback` / `db-reset` / `db-new` | Inside the Dev Container |
| `make dev-up` / `dev-shell` / `dev-down` / `up` / `down` | Host (Docker only) |

### Generated code

The Go types and the chi server interface are generated from `api/openapi.yaml` with oapi-codegen into `internal/presentation/api/openapi.gen.go` (ADR-0010, ADR-0017).

- After changing `api/openapi.yaml`, run `make generate` and commit the result. Never edit `openapi.gen.go` by hand.
- oapi-codegen is pinned as a Go tool in `go.mod`; the settings are in `internal/presentation/api/oapi-codegen.yaml`.
- Controllers implement `StrictServerInterface` (strict-server): they take a typed request object and return a typed response object, so a status code or a body the document does not define fails to compile.
- When something does not fit, check in this order: (1) can the document express it, (2) does oapi-codegen generate types for it, (3) if not, handle it in hand-written code that wraps the generated code — a middleware, a strict middleware, or the response error handler — without editing `openapi.gen.go`.
- CI runs `make generate` and fails when anything changes.
- The generated interface names the request `r`; controllers name it `req` (see Naming). Parameter names do not have to match the interface.

### Migrations

The PostgreSQL schema is changed only by migrations in `migrations/` (golang-migrate, ADR-0015).

| Command | What it does |
| --- | --- |
| `make db-migrate` | Applies every migration not applied yet |
| `make db-rollback` | Undoes the last migration |
| `make db-reset` | Drops everything in the database, then applies every migration (local only) |
| `make db-new NAME=create_accounts` | Creates the next pair of files, `00000N_create_accounts.up.sql` and `.down.sql` |

- Versions are zero-padded to six digits, so sqlc and golang-migrate read the files in the same order (ADR-0015).
- The commands use `DATABASE_URL`. To migrate the test database, pass it: `make db-migrate DATABASE_URL=postgres://app:tasuki@postgres:5432/tasuki_test?sslmode=disable`.

### Queries (sqlc)

PostgreSQL queries are written in SQL and sqlc generates the Go code (ADR-0015). Everything is in `internal/infrastructure/postgres/`:

| Path | What it is |
| --- | --- |
| `queries/*.sql` | Queries with a name, for example `-- name: GetTeam :one` |
| `sqlc.yaml` | Settings: the schema is the up files in `migrations/`, the generated code uses pgx v5 |
| `sqlcgen/*.gen.go` | Generated code. Never edit it by hand |

- After changing a query or a migration, run `make generate` and commit the result. A wrong column or type stops the generation with an error.
- sqlc is a prebuilt binary copied from the official image in the dev stage of `infra/docker/backend/Dockerfile`, and installed with `sqlc-dev/setup-sqlc` in CI. Building it with `go tool` would compile its PostgreSQL parser in C, which takes minutes. Change the version in both places.
- `uuid` columns are `uuid.UUID` from `github.com/google/uuid`, and a nullable one is `*uuid.UUID` (`overrides` in `sqlc.yaml`). Other types are pgx's defaults, such as `pgtype.Timestamptz`.
- Generated types stay in Infrastructure: repositories convert them to entities, and query services to DTOs (ADR-0015).

### Running the production image

```sh
make up                 # build the production image and start backend + MongoDB
make down
```

## Docker

Container files live in `infra/`, separated from the application code.

| File | Role |
| --- | --- |
| `infra/docker/backend/Dockerfile` | Builds the backend images. Stages: `dev` (Dev Container), `build`, and the production image (default target). The build context is `app/backend`. |
| `infra/docker/backend/Dockerfile.dockerignore` | Keeps `.env`, `bin/`, and tests out of the image |
| `infra/compose.yaml` | Local only: runs the databases and the production backend image |
| `infra/compose.dev.yaml` | Local only: adds the `backend-dev` service used by the Dev Container (repository mounted, Go caches in volumes) |
| `infra/.env.example` | Host port settings for compose |

- **One image for local and production.** The image never contains configuration; differences between environments are passed only as environment variables at runtime.
- **Multi-stage build.** `golang` builds a static binary (`CGO_ENABLED=0`); only that binary is copied into `distroless/static` (about 12 MB), which runs as a non-root user.
- `go.mod` / `go.sum` are copied and downloaded before the source, so dependencies are re-downloaded only when they change.
- distroless has no shell or curl, so there is no `HEALTHCHECK` in the image. Health checks are left to the hosting platform via `GET /health`.
- **The `dev` stage is for development only.** It adds golangci-lint, gopls, and Delve (pinned versions) on top of the same `golang` image used for the production build, so the Go version is the same in development and production.

## Naming

Names are descriptive, including method receivers (`func (client *Client) Ping`) and table-test variables (`testCase`). Only these may be shorter:

| Name | Only for |
| --- | --- |
| `t` | `*testing.T` |
| `w` | `http.ResponseWriter` |
| `i` | the index of a loop of a few lines |
| `ok`, `id` | they are words |

- `r` is not used: `req` for `*http.Request`, `router` for a router.
- Common abbreviations of three letters or more (`ctx`, `err`, `cfg`, `req`) are fine.
- varnamelen in golangci-lint checks this. It does not see struct fields, or a short name used only on the line after its declaration (such as the receiver of a one-line method), so reviews still check those.

## Testing

| Command | What it runs |
| --- | --- |
| `make test` | Unit tests (no database needed) |
| `make test-integration` | Also runs tests with the `integration` build tag against real databases |
| `make lint` | golangci-lint |

- Unit tests replace databases with fakes through interfaces.
- Integration tests are files named `*_integration_test.go` with `//go:build integration`, and use separate test databases (`tasuki_test`).
- Apply the migrations to the test database before integration tests that need tables: `make db-migrate DATABASE_URL=postgres://app:tasuki@postgres:5432/tasuki_test?sslmode=disable`.
- CI starts MongoDB and PostgreSQL as service containers, applies the migrations to `tasuki_test`, then runs `make test-integration`.

## Endpoints

| Method | Path | Description |
| --- | --- | --- |
| GET | `/health` | Pings MongoDB and PostgreSQL at the same time. `200` only when both are reachable, otherwise `503`; the body says which store is unreachable (`{"status":"error","mongo":"ok","postgres":"unreachable"}`) |

---

<details>
<summary>日本語</summary>

# backend — 技術構成

tasuki のバックエンドは Go の1つのサービスです。フロントエンド（`app/frontend`）向けの API を提供し、2つのデータベースの両方を扱います。

```
React (Vite)  ── app/frontend
  └─ Go (chi) ── app/backend
       ├─ PostgreSQL  チーム・メンバー・大会・出欠・走順      ← 正
       └─ MongoDB     ラップ記録・走順履歴・外部連携データ    ← 従
```

| データ | 保存先 | 理由 |
| --- | --- | --- |
| チーム、メンバー、大会、出欠、エントリー、走順 | PostgreSQL | 関係型で、DB で保証したい制約がある（1人1大会につき出欠は1件、走順に入れるのはエントリーのメンバーだけ） |
| ラップの記録 | MongoDB | 大会と名簿のメンバーごとに、ラップタイム（1km あたり、`MM:SS`）の可変長の一覧。メンバーが足したり、直したり、消したりする |
| 走順の変更履歴 | MongoDB | 「誰がいつ、どの並びからどの並びへ変えたか」。変更ごとにスナップショットを丸ごと保存する、典型的なイベントログ |
| 外部連携データ（将来） | MongoDB | Garmin / Strava の生アクティビティは、サービスごとに形が違う |

## 原則（崩さないこと）

1. **正は常に Postgres。** DB をまたぐトランザクションはありません
2. **参照のみ、複製しない。** MongoDB のドキュメントは Postgres の行を ID で参照するだけにします（ID の型は #68 で決めます）。メンバー名や大会情報などを MongoDB に複製しないでください。二重管理は必ずずれます
3. **サマリーを Postgres に書き戻す。** ラップが提出、修正、削除されたら、メンバーごとの平均タイムとラップ数を Postgres の `roster_members` に書き戻します（`docs/design/database.md`）。Atlas M0（無料枠）は自動バックアップがなく、アイドル状態のクラスタは終了される可能性があるため、MongoDB のデータが消えても実績だけは残るようにします
4. **MVP は2つのデータベースの両方を使う。** ラップの記録（`lap_records`）と走順の変更履歴（`running_order_history`）は MongoDB、それ以外は Postgres に置きます（`docs/design/database.md`）
5. **予想はバックエンドで計算する。** 合計タイムと周回数は Application（ADR-0011）で計算し、結果を返します。メンバーはほかのメンバーの平均タイムを見られない（ADR-0006）ので、フロントエンドには計算に使うデータを渡しません

## 技術スタック

| 対象 | 技術 | バージョン |
| --- | --- | --- |
| 言語 | Go | 1.26（`go.mod` 参照） |
| HTTP ルーター | `github.com/go-chi/chi/v5` | v5.3.x |
| MongoDB ドライバ | `go.mongodb.org/mongo-driver/v2` | v2.9.x |
| PostgreSQL ドライバ | `github.com/jackc/pgx/v5` を `pgxpool` で。Neon には直接の接続（ADR-0015） | v5.x |
| SQL | sqlc（SQL から Go を生成し、コミットする）（ADR-0015） | — |
| マイグレーション | golang-migrate。バージョンはゼロで埋める（ADR-0015） | — |
| ログ | `log/slog`（JSON） | 標準ライブラリ |
| Lint / Format | golangci-lint v2（`.golangci.yml`） | v2.x |
| API のコード生成 | oapi-codegen。生成したコードはコミットする（ADR-0017） | v2.8.0 |
| コンテナ | マルチステージビルド → `distroless/static`（`infra/docker/backend/Dockerfile`） | — |
| ローカル DB | Docker の MongoDB（`infra/compose.yaml`） | 8.0 |
| 本番 DB | Neon（PostgreSQL）、MongoDB Atlas。どちらもシンガポール（ADR-0018） | — |
| ホスティング | Cloud Run、`asia-southeast1`。API とビルドした frontend を 1 つのコンテナで出す（ADR-0018） | — |

### 選定理由

- **バックエンドはすべて Go**：API と2つの DB を、1つの言語・1つのサービスで扱います。
- **chi**：ルーティングに必要なため採用します。ルートのグループ化、サブルーター、`/races/{raceID}/laps` のような URL パラメータを使います。`net/http`（`http.Handler`）の上に作られているので、ハンドラやミドルウェアは標準ライブラリと互換のままです
  - `middleware.RealIP` は使わないこと。`X-Forwarded-For` などのヘッダーを信用して IP の偽装を許すため、非推奨になっています
- **それ以外は標準ライブラリ**：ログ（`log/slog`）や JSON は標準ライブラリを使います

## ディレクトリ構成

```
app/backend/
├─ cmd/server/         # エントリポイント：設定 → DB 接続 → HTTP サーバー、グレースフルシャットダウン
├─ internal/
│  ├─ config/          # 環境変数から設定を読む
│  ├─ presentation/    # ルーティング（chi）、コントローラー（`controller/`）、`api/openapi.yaml` から生成したコード（`api/`）
│  ├─ application/     # ユースケース、コマンド（入力の検証）、DTO／ビューモデル、それらのファクトリー、クエリサービスのインターフェース
│  ├─ domain/          # エンティティ、値オブジェクト、ドメインサービス、エンティティのファクトリー、リポジトリのインターフェース
│  └─ infrastructure/  # DB の接続、リポジトリ、クエリサービス。保存先ごとに分ける（`mongodb/`、`postgres/`）
├─ migrations/        # PostgreSQL のスキーマ：golang-migrate の上げる／戻す SQL ファイル（ADR-0015）
├─ .devcontainer/      # Dev Container の定義（開発はコンテナの中で行う）
└─ Makefile
```

- CQRS を取り入れたオニオンアーキテクチャ（[ADR-0011](../../docs/adr/0011-onion-architecture-cqrs.md)）。依存は内側に向ける：`presentation` → `application` → `domain`、`infrastructure` → `application`、`domain`。`domain` はほかの層を import しない
- 更新はユースケースとリポジトリを通り、参照はユースケースとクエリサービスを通る
- エンティティ、DTO、コマンドは、自分自身ではなくファクトリーが作る。エンティティと値オブジェクトの項目は公開しない
- 層の中は機能ごとにまとめる（例：`application/health/`）
- `/health` も層に従う：`controller.HealthController` → `health.CheckUseCase` → `health.QueryService`（`health.MongoQueryService` と `health.PostgresQueryService`。`mongodb.HealthQueryService` と `postgres.HealthQueryService` が実装する）
- すべて `internal/` 配下です。このモジュールの外から import されることは想定していません

モジュールパス：`github.com/way-station/tasuki/app/backend`

## 環境変数

アプリは**環境変数だけ**を読み、`.env` を自分では読み込みません。どの環境でも同じバイナリが動き、違うのは値だけです。

| 名前 | 必須 | デフォルト | 説明 |
| --- | --- | --- | --- |
| `MONGODB_URI` | ○ | — | MongoDB の接続文字列 |
| `MONGODB_DATABASE` | — | `tasuki` | MongoDB のデータベース名 |
| `DATABASE_URL` | ○ | — | PostgreSQL の接続文字列。プールの大きさはそのパラメータで指定する（`pool_max_conns`、`pool_min_conns`、`pool_max_conn_lifetime`、`pool_max_conn_idle_time`、ADR-0015） |
| `PORT` | — | `8080` | HTTP ポート |
| `BACKEND_PORT` | — | `8080` | backend コンテナのホスト側ポート（`infra/.env`、compose のみ） |
| `MONGO_PORT` | — | `27017` | MongoDB コンテナのホスト側ポート（`infra/.env`、compose のみ） |
| `POSTGRES_PORT` | — | `5432` | PostgreSQL コンテナのホスト側ポート（`infra/.env`、compose のみ） |

値を渡すもの：

| 環境 | 渡すもの |
| --- | --- |
| Dev Container | `infra/compose.dev.yaml` の `environment` |
| 本番用イメージをローカルで実行（`make up`） | `infra/compose.yaml` の `environment` |
| 本番 | ホスティング先の環境変数・シークレットの設定 |

ホスト側のポートを変える場合は `infra/.env` に書きます（`infra/.env.example` をコピー）。docker compose が自動で読み込みます。`.env` はコミットしないでください。

## ローカル開発

**ホストに Go は不要です。** ホストに必要なのは Docker（とエディタ）だけで、開発作業はすべて Dev Container の中で行います。

### 方法 A：Dev Container（おすすめ）

1. ホストの 27017 / 5432 / 8080 番ポートが使用中なら、`infra/.env.example` を `infra/.env` にコピーして変更する
2. VS Code で `app/backend` を開き、**Dev Containers: Reopen in Container** を実行する（JetBrains の場合は Gateway から `.devcontainer/devcontainer.json` を開く）
3. コンテナのターミナルで：

```sh
make db-migrate         # PostgreSQL のテーブルを作る
make run                # http://localhost:8080/health（ポートはホストに転送される）
make test
make test-integration   # MongoDB と PostgreSQL のコンテナを使う
make lint
```

エディタ用のツール（gopls、golangci-lint、Delve）は、バージョンを固定してコンテナに入っています。

### 方法 B：シェルだけ使う（エディタを問わない）

```sh
cd app/backend
make dev-up             # 開発用コンテナと MongoDB をビルドして起動
make dev-shell          # コンテナのシェルに入り、make test などを実行
make dev-down
```

### 各コマンドの実行場所

| コマンド | 実行する場所 |
| --- | --- |
| `make run` / `build` / `test` / `test-integration` / `lint` / `fmt` / `generate` | Dev Container の中 |
| `make db-migrate` / `db-rollback` / `db-reset` / `db-new` | Dev Container の中 |
| `make dev-up` / `dev-shell` / `dev-down` / `up` / `down` | ホスト（Docker だけ） |

### 生成するコード

Go の型と chi のサーバーのインターフェースは、oapi-codegen で `api/openapi.yaml` から `internal/presentation/api/openapi.gen.go` に生成する（ADR-0010、ADR-0017）。

- `api/openapi.yaml` を変えたら `make generate` を実行し、結果をコミットする。`openapi.gen.go` は手で編集しない
- oapi-codegen は Go のツールとして `go.mod` でバージョンを固定している。設定は `internal/presentation/api/oapi-codegen.yaml`
- コントローラーは `StrictServerInterface`（strict-server）を実装する。型の付いたリクエストを受け取り、型の付いたレスポンスを返すので、文書にないステータスコードや本文はコンパイルエラーになる
- 当てはまらないものがあれば、この順で確かめる：(1) 文書に書けるか、(2) oapi-codegen がそれを型として生成するか、(3) できなければ、`openapi.gen.go` を編集せず、生成したコードを包む手書きのコード（ミドルウェア、strict 用のミドルウェア、レスポンスのエラー処理）で扱う
- CI で `make generate` を実行し、何か変わったら失敗させる
- 生成したインターフェースではリクエストが `r` だが、コントローラーでは `req` と書く（命名を参照）。引数の名前はインターフェースと同じでなくてよい

### マイグレーション

PostgreSQL のスキーマは、`migrations/` のマイグレーションでだけ変える（golang-migrate、ADR-0015）。

| コマンド | すること |
| --- | --- |
| `make db-migrate` | まだ適用していないマイグレーションをすべて適用する |
| `make db-rollback` | 最後のマイグレーションを 1 つ戻す |
| `make db-reset` | DB の中身をすべて消してから、すべてのマイグレーションを適用し直す（ローカル専用） |
| `make db-new NAME=create_accounts` | 次の番号のファイルの組 `00000N_create_accounts.up.sql` と `.down.sql` を作る |

- バージョン番号は 6 桁でゼロ埋めする。sqlc と golang-migrate がファイルを同じ順番で読むようにするため（ADR-0015）
- コマンドは `DATABASE_URL` を使う。テスト用の DB に適用するときは指定する：`make db-migrate DATABASE_URL=postgres://app:tasuki@postgres:5432/tasuki_test?sslmode=disable`

### クエリ（sqlc）

PostgreSQL のクエリは SQL で書き、sqlc が Go のコードを生成する（ADR-0015）。すべて `internal/infrastructure/postgres/` に置く：

| 場所 | 中身 |
| --- | --- |
| `queries/*.sql` | 名前付きのクエリ。例：`-- name: GetTeam :one` |
| `sqlc.yaml` | 設定：スキーマは `migrations/` の上げるファイル、生成するコードは pgx v5 を使う |
| `sqlcgen/*.gen.go` | 生成したコード。手で編集しない |

- クエリかマイグレーションを変えたら `make generate` を実行し、結果をコミットする。列や型を間違えると、生成がエラーで止まる
- sqlc は、`infra/docker/backend/Dockerfile` の dev ステージで公式イメージからビルド済みのバイナリをコピーし、CI では `sqlc-dev/setup-sqlc` で入れる。`go tool` でビルドすると PostgreSQL の構文解析を C からビルドするので数分かかるため。バージョンは両方で変える
- `uuid` の列は `github.com/google/uuid` の `uuid.UUID`、NULL を許す列は `*uuid.UUID` にする（`sqlc.yaml` の `overrides`）。ほかの型は pgx の既定（`pgtype.Timestamptz` など）
- 生成した型は Infrastructure の外に出さない：リポジトリは Entity に、QueryService は DTO に変える（ADR-0015）

### 本番用イメージを動かす

```sh
make up                 # 本番用イメージをビルドし、backend と MongoDB を起動
make down
```

## Docker

コンテナ関連のファイルは、アプリのコードと分けて `infra/` に置きます。

| ファイル | 役割 |
| --- | --- |
| `infra/docker/backend/Dockerfile` | backend のイメージを作る。ステージは `dev`（Dev Container）、`build`、本番用イメージ（デフォルトのターゲット）。ビルドコンテキストは `app/backend` |
| `infra/docker/backend/Dockerfile.dockerignore` | `.env`、`bin/`、テストをイメージに入れない |
| `infra/compose.yaml` | ローカル専用。DB と本番用の backend イメージを起動する |
| `infra/compose.dev.yaml` | ローカル専用。Dev Container が使う `backend-dev` サービスを追加する（リポジトリをマウント、Go のキャッシュはボリューム） |
| `infra/.env.example` | compose のホスト側ポートの設定 |

- **local と本番で同じイメージを使う。** イメージに設定は入れません。環境ごとの違いは、実行時の環境変数だけで渡します
- **マルチステージビルド。** `golang` で静的バイナリ（`CGO_ENABLED=0`）をビルドし、そのバイナリだけを `distroless/static`（約 12MB）にコピーします。実行ユーザーは root ではありません
- ソースより先に `go.mod` / `go.sum` をコピーして依存を取得するので、依存が変わったときだけ再ダウンロードされます
- distroless にはシェルも curl もないため、イメージに `HEALTHCHECK` は入れていません。ヘルスチェックは `GET /health` を使い、ホスティング先に任せます
- **`dev` ステージは開発専用です。** 本番のビルドと同じ `golang` イメージに、golangci-lint、gopls、Delve（バージョン固定）を足したものです。開発と本番で Go のバージョンが揃います

## 命名

名前は説明的にする。メソッドのレシーバ（`func (client *Client) Ping`）とテーブルテストの変数（`testCase`）も同じ。短くてよいのは次だけです。

| 名前 | 使ってよいところ |
| --- | --- |
| `t` | `*testing.T` |
| `w` | `http.ResponseWriter` |
| `i` | 数行のループの添字 |
| `ok`、`id` | 単語なので |

- `r` は使わない：`*http.Request` は `req`、ルーターは `router`
- 3 文字以上のよくある略語（`ctx`、`err`、`cfg`、`req`）はよい
- golangci-lint の varnamelen で確かめる。構造体のフィールドと、宣言の次の行でしか使わない短い名前（1 行のメソッドのレシーバなど）は検出できないので、レビューで確かめる

## テスト

| コマンド | 内容 |
| --- | --- |
| `make test` | ユニットテスト（DB 不要） |
| `make test-integration` | `integration` ビルドタグ付きのテストも含め、実際の DB に対して実行 |
| `make lint` | golangci-lint |

- ユニットテストでは、インターフェース経由で DB をフェイクに置き換えます
- 結合テストは `*_integration_test.go` に `//go:build integration` を付けて書き、テスト用のデータベース（`tasuki_test`）を使います
- テーブルが必要な結合テストの前に、テスト用の DB にマイグレーションを適用します：`make db-migrate DATABASE_URL=postgres://app:tasuki@postgres:5432/tasuki_test?sslmode=disable`
- CI では MongoDB と PostgreSQL を service container で起動し、`tasuki_test` にマイグレーションを適用してから `make test-integration` を実行します

## エンドポイント

| メソッド | パス | 説明 |
| --- | --- | --- |
| GET | `/health` | MongoDB と PostgreSQL に同時に ping します。両方に届くときだけ `200`、それ以外は `503`。どちらに届かないかを本文で返します（`{"status":"error","mongo":"ok","postgres":"unreachable"}`） |

</details>
