# ADR-0001: Record architecture decisions

- Status: Accepted
- Type: technical
- Date: 2026-10-11
- Issue: #9

## Context

tasuki is an open-source project. Contributors who join later need to understand not only *how* the system is built, but *why* — for example, why the backend uses both PostgreSQL and MongoDB, or why development happens inside Dev Containers.

These reasons currently live in pull request descriptions and conversations, where they are hard to find and easily lost.

## Decision

We will record architecture decisions as ADRs (Architecture Decision Records) in `docs/adr/` in this repository.

- **Location**: in this repository, next to the code. A pull request that changes a decision updates the related ADR in the same pull request.
- **What goes in**: only decisions that constrain the future direction of the project — technology choices, data ownership, repository structure, development and deployment workflow.
- **What stays out**: implementation details (they belong in code, `TECH.md`, or pull requests) and personal notes such as learning logs or review memos.
- **Format**: one decision per file, `NNNN-kebab-case-title.md`, based on [`template.md`](template.md) (Context / Decision / Alternatives considered / Consequences). Written in English, with a Japanese translation in a collapsed section.
- **Immutability**: an accepted ADR is not rewritten when the decision changes. A new ADR supersedes it, and the old one's status becomes `Superseded by ADR-NNNN`.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| A separate repository (e.g. `tasuki-docs`) | Decisions drift from the code; a design change cannot update the ADR in the same pull request |
| Only `TECH.md` | `TECH.md` describes the current state, not the history and reasons behind each decision |
| Pull request descriptions only | Hard to find later; not organized by decision |

## Consequences

- New contributors can learn the reasons behind the design from `docs/adr/`.
- Significant design changes need an extra step: writing or superseding an ADR.
- `TECH.md` and ADRs must stay consistent; `TECH.md` describes *what* is used, ADRs explain *why*.

---

<details>
<summary>日本語</summary>

# ADR-0001: 設計判断を記録する

- Status: Accepted
- Type: technical
- Date: 2026-10-11
- Issue: #9

## Context（背景）

tasuki は OSS のプロジェクトです。後から参加するコントリビューターは、システムが*どう*作られているかだけでなく、*なぜ*そうなっているかを知る必要があります。例えば、backend が PostgreSQL と MongoDB の両方を使う理由や、開発を Dev Container の中で行う理由です。

今はこうした理由が PR の説明文や会話の中にしかなく、探しにくく、失われやすい状態です。

## Decision（決定）

設計判断を ADR（Architecture Decision Record）として、このリポジトリの `docs/adr/` に記録する。

- **置き場所**：このリポジトリの中、コードの隣。判断を変える PR では、関連する ADR も同じ PR で更新する
- **書くもの**：プロジェクトの将来の方向を縛る判断だけ。技術の選定、データの持ち主、リポジトリの構成、開発とデプロイの進め方
- **書かないもの**：実装の詳細（コード、`TECH.md`、PR に書く）と、学習メモやレビューのメモなどの個人的な記録
- **書式**：1ファイルに1つの判断。ファイル名は `NNNN-kebab-case-title.md`。[`template.md`](template.md)（Context / Decision / Alternatives considered / Consequences）を使う。英語で書き、日本語訳を折りたたみで付ける
- **書き換えない**：Accepted になった ADR は、判断が変わっても書き換えない。新しい ADR で置き換え、古い ADR の Status を `Superseded by ADR-NNNN` にする

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| 別のリポジトリ（例：`tasuki-docs`） | コードと内容がずれる。設計を変える PR の中で ADR を更新できない |
| `TECH.md` だけ | `TECH.md` は今の状態を書くもので、判断の経緯や理由は残らない |
| PR の説明文だけ | 後から探しにくく、判断ごとに整理されない |

## Consequences（影響）

- 新しいコントリビューターが、`docs/adr/` から設計の理由を知ることができる
- 大きな設計変更のときに、ADR を書く（または置き換える）手間が増える
- `TECH.md` と ADR の内容を揃える必要がある。`TECH.md` は*何を*使うか、ADR は*なぜ*そうしたかを書く

</details>
