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

---

<details>
<summary>日本語</summary>

# tasuki frontend

tasuki の Web アプリ（React、TypeScript、Vite）。このページは、clone したばかりの状態から、frontend とそのテストを動かすまでを説明する。技術構成、構成、決まりは [`TECH.md`](TECH.md) にある。

## 必要なもの

- Docker
- エディタ。Dev Containers の拡張を入れた VS Code が一番かんたん。シェルが使えれば、どのエディタでもよい

手元のマシンには、ほかに何も入れない。Node と pnpm は、[`mise.toml`](mise.toml) の版で Dev Container の中で動く。

**`pnpm` は Dev Container の中でだけ動かし、手元のマシンでは動かさない。** `node_modules` は Docker の volume にあり、pnpm はそこで、インストールのスクリプトと出たばかりの版を止めるように設定してある（[`pnpm-workspace.yaml`](pnpm-workspace.yaml)）。

## 1. ポートを確かめる

Vite の開発サーバーは、ホストのポート `5173`（`FRONTEND_PORT`）を使う。すでに使われていれば、[`infra/.env.example`](../../infra/.env.example) を `infra/.env` にコピーして変える。

## 2. Dev Container を開く

**VS Code を使う場合：** `app/frontend` を開き、**Dev Containers: Reopen in Container** を実行する。`pnpm install` は自動で動く。ウィンドウを閉じると、コンテナは止まる。

**ほかのエディタを使う場合：** 手元のマシンの `app/frontend` で、次を実行する：

```sh
make dev-up      # コンテナを作って起動する
make dev-shell   # 中のシェルを開く。最初に 1 回 pnpm install を動かす
make dev-down    # 終わったら止める
```

## 3. frontend を動かす

Dev Container の中で：

```sh
pnpm dev   # http://localhost:5173
```

手元のマシンで変えたファイルはコンテナに届くので、編集するとページが読み込み直される。

ページのデータは backend から取る。開発サーバーは `/api` を backend の Dev Container に転送する。画面を使うには、backend も起動し（[`app/backend/README.md`](../backend/README.md) を参照）、そこで `make run` を動かす。開発中に LINE を使わずにログインする方法は、[開発中のログイン](../backend/TECH.md#開発中のログイン)を参照。

## 4. テストと確認を動かす

Dev Container の中で：

```sh
pnpm test           # Vitest。pnpm test:watch は変更のたびに動かし直す
pnpm build          # 型の確認（tsc -b）と本番用のビルド
pnpm lint           # oxlint。層どうしの import の決まりも確かめる
pnpm format         # oxfmt。pnpm format:check は報告だけ
```

テストにはブラウザも backend も要らない。jsdom と偽の API クライアントを使う。CI は、PR ごとに `pnpm generate:api`（型が古いと失敗する）、`lint`、`format:check`、`test`、`build` を動かす。

## よく使うコマンド

| コマンド | 実行する場所 | すること |
| --- | --- | --- |
| `pnpm install` | Dev Container | パッケージを入れる |
| `pnpm dev` | Dev Container | 開発サーバーをポート 5173 で起動する |
| `pnpm test` / `pnpm test:watch` | Dev Container | テストを 1 回／変更のたびに動かす |
| `pnpm build` / `pnpm preview` | Dev Container | 本番用にビルドする／ビルドしたものを出す |
| `pnpm lint` | Dev Container | oxlint |
| `pnpm format` / `pnpm format:check` | Dev Container | oxfmt：整形する／報告だけ |
| `pnpm generate:api` | Dev Container | `app/backend/api/openapi.yaml` を変えた後、API の型を作り直す |
| `make dev-up` / `dev-shell` / `dev-down` | 手元のマシン | VS Code を使わずに Dev Container を起動する／入る／止める |

## 次に読むもの

- [`TECH.md`](TECH.md)：技術構成、機能の層（ADR-0012）、スタイリング、ルート、サーバーのデータ、テスト、lint、生成する API の型
- [`docs/design/screens.md`](../../docs/design/screens.md)：各画面に出すもの
- [`docs/adr/`](../../docs/adr)：なぜそうなっているか

</details>
