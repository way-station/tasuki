# Architecture Decision Records

This directory records the important architecture decisions of tasuki and the reasons behind them.
See [ADR-0001](0001-record-architecture-decisions.md) for what belongs here.

## Index

| ADR | Title | Type | Status |
| --- | --- | --- | --- |
| [0001](0001-record-architecture-decisions.md) | Record architecture decisions | technical | Accepted |
| [0002](0002-postgresql-and-mongodb.md) | PostgreSQL as the source of truth, MongoDB as a secondary store | technical | Accepted |
| [0003](0003-monorepo-layout.md) | Monorepo layout | technical | Accepted |
| [0004](0004-backend-stack.md) | Backend stack — Go with chi | technical | Accepted |
| [0005](0005-multi-tenant-data-single-team-mvp.md) | Multi-tenant data model, single-team MVP | business | Accepted |
| [0006](0006-access-policy.md) | Access policy — team-only visibility, roles, and organizer approval | business | Accepted |
| [0007](0007-line-login.md) | Authenticate all users with LINE Login | technical | Accepted |
| [0008](0008-separate-accounts-and-roster-members.md) | Separate accounts from roster members | technical | Accepted |
| [0009](0009-dev-only-login.md) | Dev-only login for local development | technical | Accepted |
| [0010](0010-spec-first-openapi.md) | Describe the API spec-first with OpenAPI | technical | Accepted |
| [0011](0011-onion-architecture-cqrs.md) | Backend onion architecture with CQRS | technical | Accepted |
| [0012](0012-frontend-structure.md) | Frontend structure | technical | Accepted |
| [0013](0013-frontend-state-and-server-data.md) | Frontend state and server data | technical | Accepted |
| [0014](0014-server-side-sessions-in-httponly-cookie.md) | Server-side sessions in an HttpOnly cookie | technical | Accepted |
| [0017](0017-commit-generated-api-code.md) | Commit generated API code | technical | Proposed |

## How to add an ADR

1. Copy [`template.md`](template.md) to `NNNN-kebab-case-title.md`, using the next number (e.g. `0005-use-sqlc-for-queries.md`).
2. Fill in Context, Decision, Alternatives considered, and Consequences. Set `Status: Proposed` and the `Type`.
3. Add a row to the index above.
4. Open a pull request. Link the related issue.
5. Once the reviewers agree, change the status to `Accepted` in the same pull request, then merge.

### Changing a decision

Do not rewrite an accepted ADR. Write a new ADR that supersedes it, and change the old one's status to `Superseded by ADR-NNNN`.

### Type values

| Type | Meaning |
| --- | --- |
| `business` | *What* tasuki does and for whom — product decisions that shape the system's structure (e.g. who can see what) |
| `technical` | *How* it is built — technology and implementation decisions. Refer to the business ADR behind it in the Context |

### Status values

| Status | Meaning |
| --- | --- |
| `Proposed` | Under discussion in a pull request |
| `Accepted` | Merged; the current decision |
| `Deprecated` | No longer applies, with no replacement |
| `Superseded by ADR-NNNN` | Replaced by a newer ADR |

---

<details>
<summary>日本語</summary>

# Architecture Decision Records（設計判断の記録）

このディレクトリには、tasuki の重要な設計判断とその理由を記録します。
何を書くかは [ADR-0001](0001-record-architecture-decisions.md) を参照してください。

## 一覧

上の Index の表を参照してください。

## ADR の追加方法

1. [`template.md`](template.md) を `NNNN-kebab-case-title.md` としてコピーする。番号は次の連番にする（例：`0005-use-sqlc-for-queries.md`）
2. Context、Decision、Alternatives considered、Consequences を書く。`Status: Proposed` にし、`Type` を選ぶ
3. 上の一覧に行を追加する
4. PR を作る。関連する Issue を紐付ける
5. レビュアーの合意が取れたら、同じ PR の中で Status を `Accepted` に変えてからマージする

### 判断を変えるとき

Accepted の ADR は書き換えません。置き換える新しい ADR を書き、古い ADR の Status を `Superseded by ADR-NNNN` にします。

### Type の値

| Type | 意味 |
| --- | --- |
| `business` | tasuki が*何を*、誰のためにするか。システムの構造に響くプロダクトの判断（例：誰が何を見られるか） |
| `technical` | *どう*作るか。技術と実装の判断。元になった business の ADR を Context で参照する |

### Status の値

| Status | 意味 |
| --- | --- |
| `Proposed` | PR で議論中 |
| `Accepted` | マージ済み。現在の判断 |
| `Deprecated` | 置き換えなしで、もう適用されない |
| `Superseded by ADR-NNNN` | 新しい ADR に置き換えられた |

</details>
