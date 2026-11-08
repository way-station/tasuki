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

## 4. Run the tests

Inside the Dev Container:

```sh
make test                # unit tests, no database needed
make lint                # golangci-lint
make fmt                 # format the code

# integration tests use a separate database; create its tables first
make db-migrate DATABASE_URL=postgres://app:tasuki@postgres:5432/tasuki_test?sslmode=disable
make test-integration
```

CI runs the tests, golangci-lint, and `make generate` on every pull request, and fails when the generated code is out of date.

## Common commands

| Command | Runs on | What it does |
| --- | --- | --- |
| `make run` | Dev Container | Starts the API on port 8080 |
| `make test` / `make test-integration` | Dev Container | Unit tests / also the integration tests |
| `make lint` / `make fmt` | Dev Container | golangci-lint |
| `make generate` | Dev Container | Regenerates the committed code after `api/openapi.yaml` or the queries change |
| `make db-migrate` / `db-rollback` / `db-reset` | Dev Container | Applies / undoes the last / reapplies every migration |
| `make db-new NAME=create_users` | Dev Container | Creates the next pair of migration files |
| `make dev-up` / `dev-shell` / `dev-down` | Your machine | Starts / enters / stops the Dev Container without VS Code |
| `make up` / `down` | Your machine | Builds and runs the production image |

## Next

- [`TECH.md`](TECH.md): the stack, the directory structure, environment variables, migrations, sqlc, testing, and naming.
- [`docs/design/api.md`](../../docs/design/api.md): the API design.
- [`docs/adr/`](../../docs/adr): why things are the way they are.

---

<details>
<summary>日本語</summary>

# tasuki backend

tasuki の Go の API（chi、PostgreSQL、MongoDB）。このページは、clone したばかりの状態から、backend とそのテストを動かすまでを説明する。設計、技術構成、決まりは [`TECH.md`](TECH.md) にある。

## 必要なもの

- Docker
- エディタ。Dev Containers の拡張を入れた VS Code が一番かんたん。シェルが使えれば、どのエディタでもよい

手元のマシンには、ほかに何も入れない。Go、golangci-lint、sqlc、golang-migrate は Dev Container の中で動く。

## 1. ポートを確かめる

コンテナは、次のホストのポートを使う：

| 変数 | 既定 | 使うもの |
| --- | --- | --- |
| `MONGO_PORT` | `27017` | MongoDB |
| `POSTGRES_PORT` | `5432` | PostgreSQL |
| `BACKEND_PORT` | `8080` | 本番用のイメージ（`make up`） |

すでに使われているポートがあれば、[`infra/.env.example`](../../infra/.env.example) を `infra/.env` にコピーして変える。

## 2. Dev Container を開く

どちらの方法でも、MongoDB、PostgreSQL、backend の Dev Container の 3 つのコンテナが起動する。

**VS Code を使う場合：** `app/backend` を開き、**Dev Containers: Reopen in Container** を実行する。ポート 8080 は手元のマシンに転送される。ウィンドウを閉じると、コンテナは止まる。

**ほかのエディタを使う場合：** 手元のマシンの `app/backend` で、次を実行する：

```sh
make dev-up      # コンテナを作って起動する
make dev-shell   # Dev Container の中のシェルを開く
make dev-down    # 終わったら止める
```

この方法では、ポート 8080 は転送されない。backend はコンテナの中から `curl localhost:8080/health` で確かめる。

## 3. backend を動かす

Dev Container の中で：

```sh
make db-migrate   # PostgreSQL のテーブルを作る（最初に 1 回と、新しいマイグレーションの後）
make run          # http://localhost:8080/health
```

`/health` は、MongoDB と PostgreSQL につながると `200` を返す。

開発中に LINE を使わずにログインする方法は、[開発中のログイン](TECH.md#開発中のログイン)を参照。

## 4. テストを動かす

Dev Container の中で：

```sh
make test                # 単体テスト。データベースは要らない
make lint                # golangci-lint
make fmt                 # コードを整形する

# 結合テストは別のデータベースを使う。先にそのテーブルを作る
make db-migrate DATABASE_URL=postgres://app:tasuki@postgres:5432/tasuki_test?sslmode=disable
make test-integration
```

CI は、PR ごとにテスト、golangci-lint、`make generate` を動かし、生成したコードが古いと失敗する。

## よく使うコマンド

| コマンド | 実行する場所 | すること |
| --- | --- | --- |
| `make run` | Dev Container | API をポート 8080 で起動する |
| `make test` / `make test-integration` | Dev Container | 単体テスト／結合テストも |
| `make lint` / `make fmt` | Dev Container | golangci-lint |
| `make generate` | Dev Container | `api/openapi.yaml` やクエリを変えた後、コミットしているコードを作り直す |
| `make db-migrate` / `db-rollback` / `db-reset` | Dev Container | マイグレーションを適用する／最後の 1 つを戻す／すべて適用し直す |
| `make db-new NAME=create_users` | Dev Container | 次のマイグレーションのファイルの組を作る |
| `make dev-up` / `dev-shell` / `dev-down` | 手元のマシン | VS Code を使わずに Dev Container を起動する／入る／止める |
| `make up` / `down` | 手元のマシン | 本番用のイメージを作って動かす |

## 次に読むもの

- [`TECH.md`](TECH.md)：技術構成、ディレクトリ構成、環境変数、マイグレーション、sqlc、テスト、命名
- [`docs/design/api.md`](../../docs/design/api.md)：API の設計
- [`docs/adr/`](../../docs/adr)：なぜそうなっているか

</details>
