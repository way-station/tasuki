# Architecture Decision Records

This directory records the important architecture decisions of tasuki and the reasons behind them.
See [ADR-0001](0001-record-architecture-decisions.md) for what belongs here.

## Index

| ADR | Title | Status |
| --- | --- | --- |
| [0001](0001-record-architecture-decisions.md) | Record architecture decisions | Accepted |

## How to add an ADR

1. Copy [`template.md`](template.md) to `NNNN-kebab-case-title.md`, using the next number (e.g. `0005-use-sqlc-for-queries.md`).
2. Fill in Context, Decision, Alternatives considered, and Consequences. Set `Status: Proposed`.
3. Add a row to the index above.
4. Open a pull request. Link the related issue.
5. Once the reviewers agree, change the status to `Accepted` in the same pull request, then merge.

### Changing a decision

Do not rewrite an accepted ADR. Write a new ADR that supersedes it, and change the old one's status to `Superseded by ADR-NNNN`.

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
2. Context、Decision、Alternatives considered、Consequences を書く。`Status: Proposed` にする
3. 上の一覧に行を追加する
4. PR を作る。関連する Issue を紐付ける
5. レビュアーの合意が取れたら、同じ PR の中で Status を `Accepted` に変えてからマージする

### 判断を変えるとき

Accepted の ADR は書き換えません。置き換える新しい ADR を書き、古い ADR の Status を `Superseded by ADR-NNNN` にします。

### Status の値

| Status | 意味 |
| --- | --- |
| `Proposed` | PR で議論中 |
| `Accepted` | マージ済み。現在の判断 |
| `Deprecated` | 置き換えなしで、もう適用されない |
| `Superseded by ADR-NNNN` | 新しい ADR に置き換えられた |

</details>
