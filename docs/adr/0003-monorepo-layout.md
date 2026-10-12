# ADR-0003: Monorepo layout

- Status: Accepted
- Type: technical
- Date: 2026-10-11
- Issue: #11

## Context

tasuki consists of a frontend (Vite + React + TypeScript), a backend (Go), the infrastructure to run them (Dockerfiles, docker compose), and project-wide files (CI, ADRs).

The frontend and the backend change together often: a new feature usually needs an API change and the screen that uses it. The project is maintained by a very small team, so keeping several repositories, versions, and pull requests in sync would be a large share of the work.

The two apps use different toolchains: Go for the backend, Node + pnpm for the frontend.

## Decision

We will keep everything in **one repository** with this layout:

```
tasuki/
├─ app/
│  ├─ frontend/   # Vite + React + TypeScript (Node + pnpm)
│  └─ backend/    # Go
├─ infra/         # Dockerfiles, docker compose files, local port settings
├─ docs/adr/      # Architecture decision records
└─ .github/       # CI workflows, PR labeling
```

- **Each app is self-contained and owns its toolchain.** The backend's Go version lives in `app/backend/go.mod`; the frontend's Node and pnpm versions live in `app/frontend/mise.toml`. There is no root-level workspace or build tool tying them together.
- **Infrastructure is separated from application code.** Dockerfiles live in `infra/docker/<app>/` and use the app directory as their build context; compose files live in `infra/`.
- **Database migrations** will live in a top-level `db/migrations/` when PostgreSQL is added (#21).
- The Go module path is `github.com/way-station/tasuki/app/backend`.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| Separate repositories for frontend and backend | API changes and the screens that use them would need coordinated pull requests and versioning across repositories |
| A monorepo build tool (Nx, Turborepo, Bun / pnpm workspaces at the root) | Only one JavaScript project exists, and Go has its own toolchain; a root tool would add configuration without benefit |
| Dockerfiles and compose files inside each app directory | Mixes how the app is built and run with its source code; infrastructure for several apps would be scattered |
| One toolchain file for everything (e.g. Go also managed by mise at the root) | Go already pins its version in `go.mod`; a second place for the same version would drift |

## Consequences

- A single pull request can change the API and the screen that uses it, and review sees both.
- CI must handle two toolchains. Workflows are split per app and filtered by path, so a change to one app does not run the other's checks.
- Contributors need to know which directory owns which tool: `make` / Go in `app/backend`, `pnpm` in `app/frontend`, Docker files in `infra/`.
- Dockerfiles refer to build contexts outside their own directory (`../app/<app>`), which must be kept in sync when directories move.

---

<details>
<summary>日本語</summary>

# ADR-0003: モノレポの構成

- Status: Accepted
- Type: technical
- Date: 2026-10-11
- Issue: #11

## Context（背景）

tasuki は、frontend（Vite + React + TypeScript）、backend（Go）、それらを動かすインフラ（Dockerfile、docker compose）、プロジェクト全体のファイル（CI、ADR）からなります。

frontend と backend は一緒に変わることが多いです。新しい機能では、たいてい API の変更と、それを使う画面の両方が必要になります。プロジェクトはとても少人数で保守するので、複数のリポジトリ、バージョン、PR の同期を取るだけで、作業の大きな割合を占めてしまいます。

2つのアプリはツールチェーンが違います。backend は Go、frontend は Node + pnpm です。

## Decision（決定）

すべてを**1つのリポジトリ**に置き、次の構成にする。

```
tasuki/
├─ app/
│  ├─ frontend/   # Vite + React + TypeScript（Node + pnpm）
│  └─ backend/    # Go
├─ infra/         # Dockerfile、docker compose、ローカルのポート設定
├─ docs/adr/      # 設計判断の記録
└─ .github/       # CI のワークフロー、PR のラベル付け
```

- **各アプリは自分で完結し、自分のツールチェーンを持つ。** backend の Go のバージョンは `app/backend/go.mod`、frontend の Node と pnpm のバージョンは `app/frontend/mise.toml` に書く。ルートに、両方をまとめるワークスペースやビルドツールは置かない
- **インフラはアプリのコードと分ける。** Dockerfile は `infra/docker/<アプリ>/` に置き、ビルドコンテキストはアプリのディレクトリにする。compose のファイルは `infra/` に置く
- **データベースのマイグレーション**は、PostgreSQL を追加するとき（#21）にルートの `db/migrations/` に置く
- Go のモジュールパスは `github.com/way-station/tasuki/app/backend`

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| frontend と backend を別のリポジトリにする | API の変更とそれを使う画面に、リポジトリをまたいだ PR とバージョンの調整が必要になる |
| モノレポ用のビルドツール（Nx、Turborepo、ルートでの Bun / pnpm のワークスペース） | JavaScript のプロジェクトは1つだけで、Go には独自のツールチェーンがある。ルートのツールは、得るものなく設定だけを増やす |
| Dockerfile と compose のファイルを各アプリのディレクトリに置く | アプリの作り方・動かし方がソースコードに混ざる。複数のアプリのインフラがあちこちに散らばる |
| すべてのツールを1つのファイルで管理する（例：ルートの mise で Go も管理する） | Go はすでに `go.mod` でバージョンを固定している。同じバージョンを書く場所が2つになると、ずれる |

## Consequences（影響）

- 1つの PR で、API とそれを使う画面を一緒に変えられ、レビューでも両方を見られる
- CI で2つのツールチェーンを扱う必要がある。ワークフローはアプリごとに分け、パスで絞り込むので、片方のアプリの変更でもう片方のチェックは走らない
- コントリビューターは、どのディレクトリがどのツールを持つかを知る必要がある。`app/backend` では `make` と Go、`app/frontend` では `pnpm`、Docker のファイルは `infra/`
- Dockerfile は自分のディレクトリの外のビルドコンテキスト（`../app/<アプリ>`）を参照するので、ディレクトリを移すときは合わせて直す必要がある

</details>
