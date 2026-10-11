# ADR-0005: One team per deployment

- Status: Proposed
- Type: business
- Date: 2026-10-11
- Issue: #45

## Context

tasuki starts from one real relay marathon team that coordinates on LINE. Its MVP is about that team's races: attendance, expected paces, running order, and the predicted total time.

Supporting many teams in one deployment (multi-tenancy) would require team ownership, organizers per team, invitations per team, and strict data isolation between teams in every query. That is roughly double the work of the MVP, for users we do not have yet.

## Decision

One deployment of tasuki serves exactly one team.

- Other teams use tasuki by deploying their own instance (self-hosting).
- The team's identity (name, etc.) is configuration or a single record, not something users create in the app.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| Multi-tenant from the start | Doubles the MVP scope (team ownership, per-team organizers, isolation) before a single team uses the app |
| Single team now, multi-tenant designed in (e.g. `team_id` everywhere but unused) | Carries the cost of multi-tenancy in every table and query without delivering it; easy to get isolation wrong while it is untested |

## Consequences

- Tables and queries can assume a single team, which keeps the data model and authorization simple.
- Moving to multi-tenancy later means adding a team reference to the main tables, migrating existing data, and adding team isolation to every query. This decision should be superseded by a new ADR if that happens.
- Each self-hosting team needs its own hosting and its own LINE Login channel (see ADR-0007).

---

<details>
<summary>日本語</summary>

# ADR-0005: 1つのデプロイに1チーム

- Status: Proposed
- Type: business
- Date: 2026-10-11
- Issue: #45

## Context（背景）

tasuki は、LINE で連絡を取り合っている実在の1つのリレーマラソンチームから始まります。MVP はそのチームの大会のためのもので、出欠、想定ペース、走順、予想合計タイムを扱います。

1つのデプロイで多くのチームを扱う（マルチテナント）には、チームの持ち主、チームごとの幹事、チームごとの招待、すべてのクエリでのチーム間のデータの分離が必要です。まだいないユーザーのために、MVP のおよそ倍の作業が増えます。

## Decision（決定）

tasuki の1つのデプロイは、ちょうど1つのチームだけを扱う。

- ほかのチームは、自分たちでデプロイして（セルフホストして）使う
- チームの情報（名前など）は設定か1件のレコードにし、アプリの中でユーザーが作るものにはしない

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| 最初からマルチテナントにする | 1つのチームも使っていない段階で、MVP の範囲が倍になる（チームの持ち主、チームごとの幹事、分離） |
| 今は1チームだが、マルチテナントを見越して作る（例：全テーブルに `team_id` を持たせるが使わない） | マルチテナントの機能を提供しないまま、そのコストだけを全テーブルとクエリが負う。テストされない分離は誤りやすい |

## Consequences（影響）

- テーブルとクエリは1チームを前提にでき、データモデルと認可が単純になる
- 後でマルチテナントにする場合は、主要なテーブルにチームへの参照を足し、既存データを移行し、すべてのクエリにチームの分離を加える。そのときは新しい ADR でこの ADR を置き換える
- セルフホストするチームは、それぞれホスティングと LINE ログインのチャネルが必要になる（ADR-0007）

</details>
