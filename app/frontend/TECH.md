# frontend — Tech Stack

The tasuki frontend is a single-page app built with Vite + React + TypeScript. It calls the Go backend (`app/backend`).

## Principles

1. **Total time is calculated on the frontend only.** The backend stores and shares data.
2. **No Node toolchain on the host.** Development happens inside the Dev Container; the host only needs Docker and an editor.

## Stack

| Area | Technology | Version |
| --- | --- | --- |
| Language | TypeScript | 6.0 |
| UI | React | 19 |
| Build / dev server | Vite | 8 |
| Runtime / package manager | Node / pnpm | see `mise.toml` (24 LTS / 10) |
| Toolchain versions | mise (`mise.toml`) | — |
| Lint | oxlint (Vite template default; to be confirmed in #5) | 1.x |
| Routing | React Router, without data mode (ADR-0013) | 7 |
| Server data | TanStack Query (ADR-0013) | 5 |
| Dev environment | Dev Container (`infra/docker/frontend/Dockerfile`, `infra/compose.dev.yaml`) | — |

### Why these choices

- **Node + pnpm instead of Bun**: the backend is Go, so Bun is no longer needed; the Vite / React ecosystem assumes Node.
- **pnpm**: strict `node_modules` (no undeclared "phantom" dependencies), and safer defaults: dependency install scripts are blocked unless allowed (`pnpm-workspace.yaml`).
- **mise**: `mise.toml` is the single source of truth for the Node and pnpm versions. The Dev Container image installs exactly these versions. `packageManager` in `package.json` is intentionally not used, to avoid a second place for the pnpm version.

### Not decided yet (#5)

| Topic | Candidates |
| --- | --- |
| API types shared with the backend | OpenAPI → generated TypeScript types |
| Styling | Tailwind CSS / CSS Modules |
| Lint / format | oxlint (current) / Biome / ESLint + Prettier |
| Tests | Vitest + Testing Library |
| Production image | depends on hosting (#7) |

## Structure (ADR-0012)

Each feature is split into Types, Mapper, Hooks, Container, and UI. Imports go Container → UI and Container → Hooks → Mapper → Types; Hooks call the backend through the `ApiClient` interface; the UI imports only view types.

```
src/
  features/<feature>/     # auth, races, participation, running-orders, laps, roster
    types/  mapper/  hooks/  container/  ui/
  api/                    # ApiClient interface and implementation, query keys
  shared/                 # only what more than one feature needs
```

## Supply-chain safety (`pnpm-workspace.yaml`)

- `onlyBuiltDependencies: []` — no dependency may run install scripts. Add a package only after reviewing it.
- `minimumReleaseAge: 1440` — only versions published at least one day ago are installed.

## Local development

1. If host port 5173 is already in use, set `FRONTEND_PORT` in `infra/.env` (copy `infra/.env.example`).
2. Open `app/frontend` in VS Code → **Dev Containers: Reopen in Container** (`pnpm install` runs automatically).
   Without VS Code: `make dev-up` then `make dev-shell` (run on the host; Docker only).
3. Inside the container:

```sh
pnpm dev        # http://localhost:5173
pnpm build
pnpm lint
```

- `node_modules` and the pnpm store live in Docker volumes, so nothing is installed on the host.
- File changes on the host reach the container, so HMR works.

| Command | Runs on |
| --- | --- |
| `pnpm install` / `dev` / `build` / `lint` / `preview` | Inside the Dev Container |
| `make dev-up` / `dev-shell` / `dev-down` | Host (Docker only) |

---

<details>
<summary>日本語</summary>

# frontend — 技術構成

tasuki のフロントエンドは、Vite + React + TypeScript で作る SPA です。Go の backend（`app/backend`）を呼び出します。

## 原則

1. **合計タイムの計算はフロントエンドだけで行う。** backend は保存と共有だけを担当します
2. **ホストに Node のツールチェーンを入れない。** 開発は Dev Container の中で行い、ホストに必要なのは Docker とエディタだけです

## 技術スタック

| 対象 | 技術 | バージョン |
| --- | --- | --- |
| 言語 | TypeScript | 6.0 |
| UI | React | 19 |
| ビルド / 開発サーバー | Vite | 8 |
| ランタイム / パッケージマネージャー | Node / pnpm | `mise.toml` 参照（24 LTS / 10） |
| ツールのバージョン管理 | mise（`mise.toml`） | — |
| Lint | oxlint（Vite テンプレートのデフォルト。#5 で確定） | 1.x |
| ルーティング | React Router。data mode は使わない（ADR-0013） | 7 |
| サーバーのデータ | TanStack Query（ADR-0013） | 5 |
| 開発環境 | Dev Container（`infra/docker/frontend/Dockerfile`、`infra/compose.dev.yaml`） | — |

### 選定理由

- **Bun ではなく Node + pnpm**：backend が Go になったので Bun は不要です。Vite / React のエコシステムは Node を前提にしています
- **pnpm**：`node_modules` が厳密で、`package.json` に書いていない依存（幽霊依存）を使えません。依存のインストール時スクリプトは、許可しない限り実行されません（`pnpm-workspace.yaml`）
- **mise**：Node と pnpm のバージョンは `mise.toml` だけに書きます。Dev Container のイメージは、ここに書いたバージョンをそのまま入れます。pnpm のバージョンを書く場所が2つにならないよう、`package.json` の `packageManager` は使いません

### 未決定のこと（#5）

| 項目 | 候補 |
| --- | --- |
| backend との API の型共有 | OpenAPI → TypeScript の型を生成 |
| スタイリング | Tailwind CSS / CSS Modules |
| Lint / Format | oxlint（現状）/ Biome / ESLint + Prettier |
| テスト | Vitest + Testing Library |
| 本番用イメージ | ホスティング（#7）次第 |

## 構成（ADR-0012）

機能ごとに Types、Mapper、Hooks、Container、UI に分ける。import は Container → UI と Container → Hooks → Mapper → Types の向き。Hooks は `ApiClient` の interface を通して backend を呼ぶ。UI は画面用の型だけを import する。

```
src/
  features/<feature>/     # auth, races, participation, running-orders, laps, roster
    types/  mapper/  hooks/  container/  ui/
  api/                    # ApiClient の interface と実装、query key
  shared/                 # 複数の機能が必要とするものだけ
```

## サプライチェーン対策（`pnpm-workspace.yaml`）

- `onlyBuiltDependencies: []`：どの依存にもインストール時スクリプトを実行させません。中身を確認したパッケージだけ追加します
- `minimumReleaseAge: 1440`：公開から1日以上経ったバージョンだけをインストールします

## ローカル開発

1. ホストの 5173 番ポートが使用中なら、`infra/.env` で `FRONTEND_PORT` を設定する（`infra/.env.example` をコピー）
2. VS Code で `app/frontend` を開き、**Dev Containers: Reopen in Container** を実行する（`pnpm install` は自動で実行される）
   VS Code を使わない場合：ホストで `make dev-up` → `make dev-shell`（Docker だけで動く）
3. コンテナの中で：

```sh
pnpm dev        # http://localhost:5173
pnpm build
pnpm lint
```

- `node_modules` と pnpm のストアは Docker の volume にあるので、ホストには何もインストールされません
- ホストでのファイル変更はコンテナに届くので、HMR が動きます

| コマンド | 実行する場所 |
| --- | --- |
| `pnpm install` / `dev` / `build` / `lint` / `preview` | Dev Container の中 |
| `make dev-up` / `dev-shell` / `dev-down` | ホスト（Docker だけ） |

</details>
