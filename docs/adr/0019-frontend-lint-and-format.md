# ADR-0019: Frontend lint and format — oxlint and oxfmt

- Status: Accepted
- Type: technical
- Date: 2026-10-24
- Issue: #179

## Context

The frontend needs a linter and a formatter so that code written at different times and by different people looks and behaves the same. ADR-0012 also sets import directions between the layers (Container → UI, Container → Hooks → Mapper → Types; the UI imports only view types), which stay conventions until a tool checks them.

oxlint is already in the frontend: it came with the Vite template (`app/frontend/package.json`, `pnpm lint`). No formatter is set up yet. The backend already uses golangci-lint (`app/backend/TECH.md`); this ADR covers only the frontend.

## Decision

**The frontend uses oxlint for linting and oxfmt for formatting.** Both are from the Oxc project.

- `pnpm lint` runs oxlint; a format script runs oxfmt. Both run in the Dev Container (no host toolchain) and in CI.
- **The import directions of ADR-0012 are enforced with oxlint's `no-restricted-imports`**, configured per layer folder with `overrides`. For example, files under `src/features/*/ui/` may not import from `hooks/`, `container/`, `mapper/`, or the generated API types.
- The oxfmt version is pinned exactly, because it is not 1.0 yet (see Consequences).
- Lint and format settings are kept in configuration files in `app/frontend/`; the exact rules are set up in #34.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| ESLint + Prettier | Two tools, each with its own configuration and plugins, to keep in step. oxlint is already set up, and oxfmt formats the same way as Prettier |
| Biome (lint and format in one tool) | Would replace oxlint, which is already set up, for a similar result |
| oxlint without a formatter | Formatting would depend on each editor's settings, so diffs would include formatting changes |
| `eslint-plugin-boundaries` (ESLint) for the layer rules | Requires ESLint. `no-restricted-imports` with per-folder `overrides` covers the directions ADR-0012 sets |

## Consequences

- One project (Oxc) provides both tools, so there is one family of configuration to learn and update.
- The ADR-0012 import directions are checked by the linter instead of by review. The check is on import paths only: it catches a UI component importing a hook, not a UI component that holds state.
- **oxfmt is in beta (pre-1.0).** Upgrades may change formatting. The version is pinned and upgraded on purpose, with the formatting changes in their own commit. Because oxfmt follows Prettier's formatting, Prettier is the fallback if oxfmt stops being maintained.
- oxlint's `no-restricted-imports` options must be checked against the oxlint documentation when #34 sets them up; if a needed option is missing, the layer rule is revisited in #34.
- Type errors are still caught by the TypeScript compiler (`tsc -b` in `pnpm build`), not by the linter.

---

<details>
<summary>日本語</summary>

# ADR-0019: frontend の lint と format — oxlint と oxfmt

- Status: Accepted
- Type: technical
- Date: 2026-10-24
- Issue: #179

## Context（背景）

frontend には、時期や人が違っても同じ見た目と振る舞いのコードになるよう、linter と formatter が要る。ADR-0012 は層の間の import の向き（Container → UI、Container → Hooks → Mapper → Types。UI は画面用の型だけを import する）も決めているが、ツールで確かめるまでは約束にすぎない。

oxlint はすでに frontend に入っている。Vite のテンプレートに付いてきたもの（`app/frontend/package.json`、`pnpm lint`）。formatter はまだない。backend はすでに golangci-lint を使っている（`app/backend/TECH.md`）。この ADR は frontend だけを扱う。

## Decision（決定）

**frontend の lint は oxlint、format は oxfmt を使う。** どちらも Oxc プロジェクトのもの。

- `pnpm lint` で oxlint を、format 用のスクリプトで oxfmt を動かす。どちらも Dev Container の中（ホストにツールを入れない）と CI で動かす
- **ADR-0012 の import の向きは、oxlint の `no-restricted-imports` で強制する。** 層のフォルダごとに `overrides` で設定する。例えば `src/features/*/ui/` の下のファイルは、`hooks/`、`container/`、`mapper/`、生成した API の型を import できない
- oxfmt はまだ 1.0 ではないので、バージョンを厳密に固定する（影響を参照）
- lint と format の設定は `app/frontend/` の設定ファイルに置く。具体的なルールは #34 で設定する

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| ESLint + Prettier | 2 つのツールそれぞれに設定とプラグインがあり、揃え続ける必要がある。oxlint はすでに入っていて、oxfmt は Prettier と同じように整形する |
| Biome（lint と format を 1 つのツールで） | すでに入っている oxlint を置き換えることになり、得られる結果は同程度 |
| formatter を使わず oxlint だけ | 整形が各自のエディタの設定次第になり、差分に整形の変更が混ざる |
| 層のルールに `eslint-plugin-boundaries`（ESLint） | ESLint が必要になる。ADR-0012 が決めた向きは、フォルダごとの `overrides` と `no-restricted-imports` で足りる |

## Consequences（影響）

- 1 つのプロジェクト（Oxc）が両方のツールを出しているので、覚えて更新する設定が 1 系統で済む
- ADR-0012 の import の向きは、レビューではなく linter が確かめる。確かめるのは import のパスだけ。UI が Hooks を import したことは見つけるが、UI が状態を持っていることは見つけない
- **oxfmt は beta（1.0 より前）。** 更新で整形が変わることがある。バージョンを固定し、意図して更新し、整形の変更はそれだけのコミットにする。oxfmt は Prettier の整形に合わせているので、oxfmt のメンテナンスが止まったら Prettier に戻せる
- oxlint の `no-restricted-imports` のオプションは、#34 で設定するときに oxlint のドキュメントで確かめる。必要なオプションがなければ、層のルールは #34 で見直す
- 型のエラーは、linter ではなく TypeScript のコンパイラ（`pnpm build` の `tsc -b`）が見つける

</details>
