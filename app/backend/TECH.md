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
| PostgreSQL driver | not decided yet (see below) | — |
| Logging | `log/slog` (JSON) | stdlib |
| Lint / format | golangci-lint v2 (`.golangci.yml`) | v2.x |
| Container | Multi-stage build → `distroless/static` (`infra/docker/backend/Dockerfile`) | — |
| Local DB | MongoDB in Docker (`infra/compose.yaml`) | 8.0 |
| Production DB | Neon (PostgreSQL), MongoDB Atlas — both in Singapore (ADR-0018) | — |
| Hosting | Cloud Run, `asia-southeast1`; one container serving the API and the built frontend (ADR-0018) | — |

### Why these choices

- **Go for the whole backend**: one language and one service for the API and both databases.
- **chi**: needed for routing — route groups, sub-routers, and URL parameters such as `/races/{raceID}/laps`. It is built on `net/http` (`http.Handler`), so handlers and middleware stay compatible with the standard library.
  - Do not use `middleware.RealIP`: it is deprecated because it trusts `X-Forwarded-For` and similar headers and allows IP spoofing.
- **Standard library otherwise**: logging (`log/slog`) and JSON use the standard library.

### Not decided yet

| Topic | Candidates | Decided in |
| --- | --- | --- |
| PostgreSQL driver | pgx v5 (the de facto standard) | Postgres setup issue |
| How SQL is written | sqlc (generate Go from plain SQL) / pgx with hand-written queries | Postgres setup issue |
| Migration tool | goose / golang-migrate / dbmate | Postgres setup issue |
| Whether generated API code is committed | The OpenAPI document is `app/backend/api/openapi.yaml`; Go code with oapi-codegen, TypeScript types with openapi-typescript (ADR-0010) | #30 |
| Whether MongoDB is required at startup | Required now; the MVP uses MongoDB for lap records and running order history | #26 |

## Directory structure

```
app/backend/
├─ cmd/server/            # entry point: config → DB connections → HTTP server, graceful shutdown
├─ internal/
│  ├─ config/             # loads settings from environment variables
│  ├─ presentation/       # routing (chi) and controllers (`controller/`; they implement the OpenAPI-generated interface)
│  ├─ application/        # use cases, commands (input validation), DTOs / view models, their factories, query service interfaces
│  ├─ domain/             # entities, value objects, domain services, entity factories, repository interfaces
│  └─ infrastructure/     # DB connections, repositories, and query services, per store (`mongodb/`, PostgreSQL later)
├─ .devcontainer/         # Dev Container definition (development happens inside the container)
└─ Makefile
```

- Onion architecture with CQRS ([ADR-0011](../../docs/adr/0011-onion-architecture-cqrs.md)). Dependencies point inward: `presentation` → `application` → `domain`; `infrastructure` → `application`, `domain`. `domain` imports no other layer.
- Writes go through use cases and repositories; reads go through query services.
- Entities, DTOs, and commands are created by factories, not by themselves. Entities and value objects keep their fields unexported.
- Everything is under `internal/`; nothing is meant to be imported from outside this module.
- Inside a layer, code is grouped per feature (for example `application/health/`).
- `/health` follows the layers too: `controller.HealthController` → `health.CheckUseCase` → `health.QueryService` (implemented by `mongodb.HealthQueryService`).

Module path: `github.com/way-station/tasuki/app/backend`

## Environment variables

The application reads **only environment variables**; it never loads `.env` itself. The same binary runs everywhere, and only the values differ.

| Name | Required | Default | Description |
| --- | --- | --- | --- |
| `MONGODB_URI` | yes | — | MongoDB connection string |
| `MONGODB_DATABASE` | no | `tasuki` | MongoDB database name |
| `PORT` | no | `8080` | HTTP port |
| `BACKEND_PORT` | no | `8080` | Host port of the backend container (`infra/.env`, compose only) |
| `MONGO_PORT` | no | `27017` | Host port of the MongoDB container (`infra/.env`, compose only) |

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

1. If host ports 27017 / 8080 are already in use, copy `infra/.env.example` to `infra/.env` and change them.
2. Open `app/backend` in VS Code and run **Dev Containers: Reopen in Container** (JetBrains: open `.devcontainer/devcontainer.json` via Gateway).
3. In the container terminal:

```sh
make run                # http://localhost:8080/health (port is forwarded to the host)
make test
make test-integration   # uses the MongoDB container
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
| `make run` / `build` / `test` / `test-integration` / `lint` / `fmt` | Inside the Dev Container |
| `make dev-up` / `dev-shell` / `dev-down` / `up` / `down` | Host (Docker only) |

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

## Testing

| Command | What it runs |
| --- | --- |
| `make test` | Unit tests (no database needed) |
| `make test-integration` | Also runs tests with the `integration` build tag against real databases |
| `make lint` | golangci-lint |

- Unit tests replace databases with fakes through interfaces.
- Integration tests are files named `*_integration_test.go` with `//go:build integration`, and use separate test databases (`tasuki_test`).

## Endpoints

| Method | Path | Description |
| --- | --- | --- |
| GET | `/health` | Pings the databases. `200 {"status":"ok"}` or `503` |

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
| PostgreSQL ドライバ | 未決定（下記参照） | — |
| ログ | `log/slog`（JSON） | 標準ライブラリ |
| Lint / Format | golangci-lint v2（`.golangci.yml`） | v2.x |
| コンテナ | マルチステージビルド → `distroless/static`（`infra/docker/backend/Dockerfile`） | — |
| ローカル DB | Docker の MongoDB（`infra/compose.yaml`） | 8.0 |
| 本番 DB | Neon（PostgreSQL）、MongoDB Atlas。どちらもシンガポール（ADR-0018） | — |
| ホスティング | Cloud Run、`asia-southeast1`。API とビルドした frontend を 1 つのコンテナで出す（ADR-0018） | — |

### 選定理由

- **バックエンドはすべて Go**：API と2つの DB を、1つの言語・1つのサービスで扱います。
- **chi**：ルーティングに必要なため採用します。ルートのグループ化、サブルーター、`/races/{raceID}/laps` のような URL パラメータを使います。`net/http`（`http.Handler`）の上に作られているので、ハンドラやミドルウェアは標準ライブラリと互換のままです
  - `middleware.RealIP` は使わないこと。`X-Forwarded-For` などのヘッダーを信用して IP の偽装を許すため、非推奨になっています
- **それ以外は標準ライブラリ**：ログ（`log/slog`）や JSON は標準ライブラリを使います

### 未決定のこと

| 項目 | 候補 | 決める場所 |
| --- | --- | --- |
| PostgreSQL ドライバ | pgx v5（事実上の標準） | Postgres の環境構築 Issue |
| SQL の書き方 | sqlc（素の SQL から Go を生成）／ pgx で手書き | Postgres の環境構築 Issue |
| マイグレーションツール | goose / golang-migrate / dbmate | Postgres の環境構築 Issue |
| 生成した API のコードをコミットするか | OpenAPI の文書は `app/backend/api/openapi.yaml`。Go は oapi-codegen、TypeScript の型は openapi-typescript で生成する（ADR-0010） | #30 |
| 起動時に MongoDB を必須にするか | 今は必須。MVP はラップの記録と走順の変更履歴に MongoDB を使う | #26 |

## ディレクトリ構成

```
app/backend/
├─ cmd/server/         # エントリポイント：設定 → DB 接続 → HTTP サーバー、グレースフルシャットダウン
├─ internal/
│  ├─ config/          # 環境変数から設定を読む
│  ├─ presentation/    # ルーティング（chi）とコントローラー（`controller/`。OpenAPI から生成したインターフェースを実装する）
│  ├─ application/     # ユースケース、コマンド（入力の検証）、DTO／ビューモデル、それらのファクトリー、クエリサービスのインターフェース
│  ├─ domain/          # エンティティ、値オブジェクト、ドメインサービス、エンティティのファクトリー、リポジトリのインターフェース
│  └─ infrastructure/  # DB の接続、リポジトリ、クエリサービス。保存先ごとに分ける（`mongodb/`、PostgreSQL は後で）
├─ .devcontainer/      # Dev Container の定義（開発はコンテナの中で行う）
└─ Makefile
```

- CQRS を取り入れたオニオンアーキテクチャ（[ADR-0011](../../docs/adr/0011-onion-architecture-cqrs.md)）。依存は内側に向ける：`presentation` → `application` → `domain`、`infrastructure` → `application`、`domain`。`domain` はほかの層を import しない
- 更新はユースケースとリポジトリを通り、参照はユースケースとクエリサービスを通る
- エンティティ、DTO、コマンドは、自分自身ではなくファクトリーが作る。エンティティと値オブジェクトの項目は公開しない
- 層の中は機能ごとにまとめる（例：`application/health/`）
- `/health` も層に従う：`controller.HealthController` → `health.CheckUseCase` → `health.QueryService`（`mongodb.HealthQueryService` が実装する）
- すべて `internal/` 配下です。このモジュールの外から import されることは想定していません

モジュールパス：`github.com/way-station/tasuki/app/backend`

## 環境変数

アプリは**環境変数だけ**を読み、`.env` を自分では読み込みません。どの環境でも同じバイナリが動き、違うのは値だけです。

| 名前 | 必須 | デフォルト | 説明 |
| --- | --- | --- | --- |
| `MONGODB_URI` | ○ | — | MongoDB の接続文字列 |
| `MONGODB_DATABASE` | — | `tasuki` | MongoDB のデータベース名 |
| `PORT` | — | `8080` | HTTP ポート |
| `BACKEND_PORT` | — | `8080` | backend コンテナのホスト側ポート（`infra/.env`、compose のみ） |
| `MONGO_PORT` | — | `27017` | MongoDB コンテナのホスト側ポート（`infra/.env`、compose のみ） |

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

1. ホストの 27017 / 8080 番ポートが使用中なら、`infra/.env.example` を `infra/.env` にコピーして変更する
2. VS Code で `app/backend` を開き、**Dev Containers: Reopen in Container** を実行する（JetBrains の場合は Gateway から `.devcontainer/devcontainer.json` を開く）
3. コンテナのターミナルで：

```sh
make run                # http://localhost:8080/health（ポートはホストに転送される）
make test
make test-integration   # MongoDB コンテナを使う
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
| `make run` / `build` / `test` / `test-integration` / `lint` / `fmt` | Dev Container の中 |
| `make dev-up` / `dev-shell` / `dev-down` / `up` / `down` | ホスト（Docker だけ） |

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

## テスト

| コマンド | 内容 |
| --- | --- |
| `make test` | ユニットテスト（DB 不要） |
| `make test-integration` | `integration` ビルドタグ付きのテストも含め、実際の DB に対して実行 |
| `make lint` | golangci-lint |

- ユニットテストでは、インターフェース経由で DB をフェイクに置き換えます
- 結合テストは `*_integration_test.go` に `//go:build integration` を付けて書き、テスト用のデータベース（`tasuki_test`）を使います

## エンドポイント

| メソッド | パス | 説明 |
| --- | --- | --- |
| GET | `/health` | DB に ping します。`200 {"status":"ok"}` または `503` |

</details>
