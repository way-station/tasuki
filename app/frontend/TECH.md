# frontend — Tech Stack

The tasuki frontend is a single-page app built with Vite + React + TypeScript. It calls the Go backend (`app/backend`).

## Principles

1. **Predictions are calculated by the backend.** The frontend only shows the total time or laps it returns. Members may not see other members' average times (ADR-0006), so the frontend never receives the data to calculate them (`app/backend/TECH.md`, ADR-0011).
2. **No Node toolchain on the host.** Development happens inside the Dev Container; the host only needs Docker and an editor.

## Stack

| Area | Technology | Version |
| --- | --- | --- |
| Language | TypeScript | 6.0 |
| UI | React | 19 |
| Styling | Tailwind CSS, with its Vite plugin (#33) | 4.3.3 (exact) |
| Build / dev server | Vite | 8 |
| Runtime / package manager | Node / pnpm | see `mise.toml` (24 LTS / 10) |
| Toolchain versions | mise (`mise.toml`) | — |
| Lint / format | oxlint and oxfmt (ADR-0019) | oxlint 1.x / oxfmt 0.73.0 (exact, pre-1.0) |
| Routing | TanStack Router, without loaders (ADR-0013) | 1.170.45 |
| Server data | TanStack Query (ADR-0013) | 5 |
| API client | openapi-fetch, typed by the generated API types | 0.17.0 (exact, pre-1.0) |
| Tests | Vitest, Testing Library, jsdom | Vitest 5 |
| API types | openapi-typescript, generated from `app/backend/api/openapi.yaml` (ADR-0017) | 7.13.0 (exact) |
| Dev environment | Dev Container (`infra/docker/frontend/Dockerfile`, `infra/compose.dev.yaml`) | — |

### Why these choices

- **Node + pnpm instead of Bun**: the backend is Go, so Bun is no longer needed; the Vite / React ecosystem assumes Node.
- **pnpm**: strict `node_modules` (no undeclared "phantom" dependencies), and safer defaults: dependency install scripts are blocked unless allowed (`pnpm-workspace.yaml`).
- **Tailwind CSS instead of CSS Modules**:
  - UI components only display their props (ADR-0012). With the styles written in the JSX, each component needs no CSS file of its own.
  - The design tokens from the UI design (#146) go in Tailwind's `@theme` as variables, so every component uses only those values.
  - Classes without a prefix apply to small screens, and `md:` and other prefixes add larger screens. This is mobile-first, which suits an app opened mostly from LINE on phones.
  - The official Vite plugin (`@tailwindcss/vite`) is the whole setup.
- **mise**: `mise.toml` is the single source of truth for the Node and pnpm versions. The Dev Container image installs exactly these versions. `packageManager` in `package.json` is intentionally not used, to avoid a second place for the pnpm version.

### Not decided yet (#5)

| Topic | Candidates |
| --- | --- |
| Production image | depends on hosting (#7) |

## Structure (ADR-0012)

Each feature is split into Types, Mapper, Hooks, Container, and UI. Imports go Container → UI and Container → Hooks → Mapper → Types; Hooks call the backend through the `ApiClient` interface; the UI imports only view types.

```
src/
  features/<feature>/     # auth, races, participation, running-orders, laps, roster
    types/  mapper/  hooks/  container/  ui/
  app/                    # routes, the common layout, the not-found page
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
pnpm dev           # http://localhost:5173
pnpm build
pnpm test         # pnpm test:watch reruns on changes
pnpm lint
pnpm format        # pnpm format:check only reports
pnpm generate:api  # after app/backend/api/openapi.yaml changes
```

- `node_modules` and the pnpm store live in Docker volumes, so nothing is installed on the host.
- File changes on the host reach the container, so HMR works.

| Command | Runs on |
| --- | --- |
| `pnpm install` / `dev` / `build` / `test` / `test:watch` / `lint` / `format` / `format:check` / `preview` / `generate:api` | Inside the Dev Container |
| `make dev-up` / `dev-shell` / `dev-down` | Host (Docker only) |

### Styling

- Styles are Tailwind classes in the JSX. `src/index.css` imports Tailwind; the design tokens (#146) go in its `@theme` block.
- Write for small screens first, then add `md:` and larger prefixes where needed.
- `pnpm format` sorts the classes; CI fails on unsorted classes through `pnpm format:check`.

### Routes

- `src/app/router.tsx` defines the routes in code (no generated route tree): one route per screen in `docs/design/screens.md`, all inside the common layout (`src/app/Layout.tsx`). Unknown URLs show the not-found page.
- The router is registered for types, so a link to a path that does not exist, or without a required param, is a compile error.
- Until a feature's container exists, its routes show a placeholder page with the screen ID.
- Routes only pick the screen: no loaders (ADR-0013). Containers read the URL with `useParams` / `useSearch` and fetch through hooks.
- Reloading on any URL works because the server returns `index.html` for every path. Vite does this in `pnpm dev` and `pnpm preview`; in production the backend does it for every path outside `/api` (ADR-0018, #197).

### API client and server data

- Hooks call the backend only through the `ApiClient` interface (`src/api/client.ts`), which they get with `useApiClient()` (ADR-0012). It has one method per endpoint, such as `getHealth()`; each feature adds the methods its hooks need. The implementation uses openapi-fetch, so paths, parameters, and bodies are checked against the OpenAPI document.
- Error responses become an `ApiError` with the `status` and the problem+json `code` (`docs/design/api.md`).
- Query keys are all in `src/api/queryKeys.ts` (ADR-0013).
- The `QueryClient` (`src/api/queryClient.ts`) does not retry 4xx responses, clears the whole cache on any `401`, and keeps the default `staleTime` of `0`, so screens fetch again when they open or the window regains focus (ADR-0013).
- Tests pass `createFakeApiClient({ ... })` (`src/api/fakeApiClient.ts`) with only the methods they need, and report errors with `problem(status, code)`, the same `ApiError` the real client throws.
- `pnpm dev` and `pnpm preview` proxy `/api` and `/health` to the backend (`BACKEND_URL`, set in `infra/compose.dev.yaml`), so the browser sees one origin as in production (ADR-0018). Start the backend with `make run` in its container.

### Tests

- Vitest runs the tests, with Testing Library and jsdom for components, so no browser is needed. Settings are in the `test` block of `vite.config.ts`; `src/test/setup.ts` adds the DOM matchers (`toBeInTheDocument()` and others).
- Test files sit next to what they test, as `*.test.ts` / `*.test.tsx`.
- Predictions are calculated by the backend, so frontend tests cover mappers, hooks with a fake `ApiClient` (#32), containers, and UI components.
- CI runs `pnpm test --bail=1`, which stops at the first failing test.

### Lint and format

- oxfmt uses its defaults, the same as Prettier: double quotes and semicolons. `.oxfmtrc.json` lists what it skips (the generated `src/api/schema.gen.ts` and Markdown) and turns on Tailwind class sorting, which orders classes the same way as the official prettier-plugin-tailwindcss.
- CI runs `pnpm lint` and `pnpm format:check`.
- oxlint enforces the import directions of ADR-0012 per layer folder (`overrides` in `.oxlintrc.json`):

| Files in | May not import |
| --- | --- |
| `src/features/*/ui/` | `container/`, `hooks/`, `mapper/`, anything in `api/` (including the generated types), `@tanstack/react-router`, `@tanstack/react-query` |
| `src/features/*/container/` | `mapper/`, anything in `api/` |
| `src/features/*/hooks/` | `container/`, `ui/` |
| `src/features/*/mapper/` | `hooks/`, `container/`, `ui/`, `api/` except the generated types |
| `src/features/*/types/` | `mapper/`, `hooks/`, `container/`, `ui/`, `api/` except the generated types |

### Generated API types

The TypeScript types of the API are generated from `app/backend/api/openapi.yaml` with openapi-typescript into `src/api/schema.gen.ts` (ADR-0010, ADR-0017).

- After the document changes, run `pnpm generate:api` and commit the result. Never edit `schema.gen.ts` by hand.
- openapi-typescript is pinned exactly in `package.json`. It lists TypeScript 5 as its peer; it has been checked with TypeScript 6.0.
- CI runs `pnpm generate:api` and fails when anything changes. The frontend workflow also runs when only `openapi.yaml` changes.
- Only the API client and the hooks use these types; UI components use view types (ADR-0012).

---

<details>
<summary>日本語</summary>

# frontend — 技術構成

tasuki のフロントエンドは、Vite + React + TypeScript で作る SPA です。Go の backend（`app/backend`）を呼び出します。

## 原則

1. **予想は backend が計算する。** frontend は、返ってきた合計タイムや周回数を表示するだけです。メンバーはほかのメンバーの平均タイムを見られない（ADR-0006）ので、frontend には計算に使うデータが届きません（`app/backend/TECH.md`、ADR-0011）
2. **ホストに Node のツールチェーンを入れない。** 開発は Dev Container の中で行い、ホストに必要なのは Docker とエディタだけです

## 技術スタック

| 対象 | 技術 | バージョン |
| --- | --- | --- |
| 言語 | TypeScript | 6.0 |
| UI | React | 19 |
| スタイリング | Tailwind CSS。Vite のプラグインで使う（#33） | 4.3.3（固定） |
| ビルド / 開発サーバー | Vite | 8 |
| ランタイム / パッケージマネージャー | Node / pnpm | `mise.toml` 参照（24 LTS / 10） |
| ツールのバージョン管理 | mise（`mise.toml`） | — |
| Lint / Format | oxlint と oxfmt（ADR-0019） | oxlint 1.x / oxfmt 0.73.0（固定、1.0 より前） |
| ルーティング | TanStack Router。loader は使わない（ADR-0013） | 1.170.45 |
| サーバーのデータ | TanStack Query（ADR-0013） | 5 |
| API クライアント | openapi-fetch。生成した API の型で型が付く | 0.17.0（固定、1.0 前） |
| テスト | Vitest、Testing Library、jsdom | Vitest 5 |
| API の型 | openapi-typescript。`app/backend/api/openapi.yaml` から生成（ADR-0017） | 7.13.0（固定） |
| 開発環境 | Dev Container（`infra/docker/frontend/Dockerfile`、`infra/compose.dev.yaml`） | — |

### 選定理由

- **Bun ではなく Node + pnpm**：backend が Go になったので Bun は不要です。Vite / React のエコシステムは Node を前提にしています
- **pnpm**：`node_modules` が厳密で、`package.json` に書いていない依存（幽霊依存）を使えません。依存のインストール時スクリプトは、許可しない限り実行されません（`pnpm-workspace.yaml`）
- **CSS Modules ではなく Tailwind CSS**：
  - UI の部品は props を表示するだけ（ADR-0012）。見た目を JSX に書くので、部品ごとの CSS ファイルが要らない
  - UI デザインの基本（#146）の値は、Tailwind の `@theme` に変数として書く。すべての部品がその値だけを使う
  - 接頭辞のないクラスは小さい画面に効き、`md:` などの接頭辞で大きい画面を足す。モバイルファーストなので、ほとんど LINE からスマホで開く tasuki に合う
  - 設定は Vite の公式プラグイン（`@tailwindcss/vite`）を足すだけ
- **mise**：Node と pnpm のバージョンは `mise.toml` だけに書きます。Dev Container のイメージは、ここに書いたバージョンをそのまま入れます。pnpm のバージョンを書く場所が2つにならないよう、`package.json` の `packageManager` は使いません

### 未決定のこと（#5）

| 項目 | 候補 |
| --- | --- |
| 本番用イメージ | ホスティング（#7）次第 |

## 構成（ADR-0012）

機能ごとに Types、Mapper、Hooks、Container、UI に分ける。import は Container → UI と Container → Hooks → Mapper → Types の向き。Hooks は `ApiClient` の interface を通して backend を呼ぶ。UI は画面用の型だけを import する。

```
src/
  features/<feature>/     # auth, races, participation, running-orders, laps, roster
    types/  mapper/  hooks/  container/  ui/
  app/                    # ルート、共通のレイアウト、not found ページ
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
pnpm dev           # http://localhost:5173
pnpm build
pnpm test         # pnpm test:watch は変更のたびに実行し直す
pnpm lint
pnpm format        # pnpm format:check は確かめるだけ
pnpm generate:api  # app/backend/api/openapi.yaml を変えたら
```

- `node_modules` と pnpm のストアは Docker の volume にあるので、ホストには何もインストールされません
- ホストでのファイル変更はコンテナに届くので、HMR が動きます

| コマンド | 実行する場所 |
| --- | --- |
| `pnpm install` / `dev` / `build` / `test` / `test:watch` / `lint` / `format` / `format:check` / `preview` / `generate:api` | Dev Container の中 |
| `make dev-up` / `dev-shell` / `dev-down` | ホスト（Docker だけ） |

### スタイリング

- 見た目は JSX に Tailwind のクラスで書く。`src/index.css` で Tailwind を読み込み、デザインの値（#146）はその `@theme` に書く
- まず小さい画面向けに書き、必要なところだけ `md:` 以上の接頭辞で足す
- `pnpm format` がクラスを並べ替える。並んでいないクラスは、CI の `pnpm format:check` で落ちる

### ルート

- `src/app/router.tsx` に、ルートをコードで定義する（ルートの木は生成しない）。`docs/design/screens.md` の画面ごとに 1 つのルートを、すべて共通のレイアウト（`src/app/Layout.tsx`）の中に置く。知らない URL は not found ページを出す
- ルーターを型に登録しているので、存在しないパスへのリンクや、必要な引数のないリンクはコンパイルエラーになる
- 機能の Container ができるまでは、画面 ID を出す仮のページを出す
- ルートは画面を選ぶだけで、loader は使わない（ADR-0013）。Container が `useParams` / `useSearch` で URL を読み、Hooks を通してデータを取る
- どの URL で再読み込みしても画面が出るのは、サーバーがどのパスにも `index.html` を返すから。`pnpm dev` と `pnpm preview` では Vite がそうする。本番では backend が `/api` 以外のパスでそうする（ADR-0018、#197）

### API クライアントとサーバーのデータ

- Hooks は `ApiClient` の interface（`src/api/client.ts`）だけを通して backend を呼ぶ。`useApiClient()` で受け取る（ADR-0012）。エンドポイントごとに `getHealth()` のようなメソッドを 1 つ持ち、各機能が自分の Hooks に要るメソッドを足す。実装は openapi-fetch で、パス、パラメータ、body は OpenAPI と照らして型で確かめられる
- エラーの応答は、`status` と problem+json の `code` を持つ `ApiError` になる（`docs/design/api.md`）
- query key はすべて `src/api/queryKeys.ts` に置く（ADR-0013）
- `QueryClient`（`src/api/queryClient.ts`）は 4xx を再試行せず、`401` でキャッシュをすべて消し、`staleTime` は既定の `0` のまま。画面を開いたときとウィンドウに戻ったときに取り直す（ADR-0013）
- テストでは、要るメソッドだけを渡した `createFakeApiClient({ ... })`（`src/api/fakeApiClient.ts`）を使い、エラーは `problem(status, code)` で返す。本物のクライアントと同じ `ApiError` になる
- `pnpm dev` と `pnpm preview` は `/api` と `/health` を backend に中継する（`BACKEND_URL`、`infra/compose.dev.yaml` で設定）。本番と同じく、ブラウザからは 1 つのオリジンに見える（ADR-0018）。backend はそのコンテナで `make run` で起動する

### テスト

- テストは Vitest で実行する。部品は Testing Library と jsdom で描画するので、ブラウザは要らない。設定は `vite.config.ts` の `test`。`src/test/setup.ts` で DOM 用の matcher（`toBeInTheDocument()` など）を足している
- テストファイルは、テストする対象の隣に `*.test.ts` / `*.test.tsx` として置く
- 予想は backend が計算するので、frontend のテストは Mapper、偽の `ApiClient`（#32）を使う Hooks、Container、UI の部品を対象にする
- CI では `pnpm test --bail=1` を実行し、最初に落ちたテストで止める

### Lint と format

- oxfmt は既定の設定を使う。Prettier と同じで、ダブルクォートとセミコロンあり。`.oxfmtrc.json` には、対象外のもの（生成した `src/api/schema.gen.ts` と Markdown）と、Tailwind のクラスの並べ替えを書く。並べ方は公式の prettier-plugin-tailwindcss と同じ
- CI で `pnpm lint` と `pnpm format:check` を実行する
- oxlint で、ADR-0012 の import の向きを層のフォルダごとに強制する（`.oxlintrc.json` の `overrides`）：

| 対象のファイル | import してはいけないもの |
| --- | --- |
| `src/features/*/ui/` | `container/`、`hooks/`、`mapper/`、`api/` のすべて（生成した型も含む）、`@tanstack/react-router`、`@tanstack/react-query` |
| `src/features/*/container/` | `mapper/`、`api/` のすべて |
| `src/features/*/hooks/` | `container/`、`ui/` |
| `src/features/*/mapper/` | `hooks/`、`container/`、`ui/`、生成した型以外の `api/` |
| `src/features/*/types/` | `mapper/`、`hooks/`、`container/`、`ui/`、生成した型以外の `api/` |

### 生成する API の型

API の TypeScript の型は、openapi-typescript で `app/backend/api/openapi.yaml` から `src/api/schema.gen.ts` に生成する（ADR-0010、ADR-0017）。

- 文書が変わったら `pnpm generate:api` を実行し、結果をコミットする。`schema.gen.ts` は手で編集しない
- openapi-typescript は `package.json` で厳密にバージョンを固定している。peer は TypeScript 5 だが、TypeScript 6.0 で動くことを確かめた
- CI で `pnpm generate:api` を実行し、何か変わったら失敗させる。frontend の workflow は `openapi.yaml` だけが変わったときも動く
- この型を使うのは API クライアントと Hooks だけ。UI の部品は画面用の型を使う（ADR-0012）

</details>
