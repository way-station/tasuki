# ADR-0005: Multi-tenant data model, single-team MVP

- Status: Accepted
- Type: business
- Date: 2026-10-11
- Issue: #45

## Context

tasuki starts from one real relay marathon team that coordinates on LINE. Its MVP is about that team's races: attendance, expected paces, running order, and the predicted total time.

If other teams want to use tasuki later, it will be offered as a **SaaS**: one service that we operate, shared by many teams. It will not be distributed as software that each team deploys on its own.

A SaaS that serves many teams from one environment must keep each team's data strictly separate. Members' paces, results, and strengths are personal data; one query that forgets to filter by team would show them to another team. Adding team separation to an existing single-team schema later means adding a column to every main table, migrating data, and reviewing every query — the most expensive and riskiest point to change.

## Decision

**The data model is multi-tenant from the start; the MVP's features serve a single team.**

- A **team** (tenant) exists in the data model. Every team-owned record belongs to a team, and every query is scoped to the current team.
- The MVP has exactly one team. There is no screen to create teams, switch teams, or sign up as a new team.
- When tasuki becomes a SaaS, the missing pieces are team sign-up and per-team organizer management — not changes to the existing data.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| Single-team data model now, multi-tenant later | Retrofitting tenant separation touches every table and query; one missed filter leaks another team's personal data |
| Full multi-tenant SaaS in the MVP (sign-up, team management) | Doubles the MVP scope before a single team uses the app |
| One deployment per team (self-hosting, or us deploying a separate instance per team) | Every team needs its own hosting and LINE Login channel, and operating many instances does not scale; it does not match the SaaS direction |

## Consequences

- Every team-owned table carries a team reference, and every query must filter by team, even while there is only one team. Tests should cover that one team cannot see another team's data.
- The MVP behaves as a single-team app; the team can be created by a setup step or migration.
- Organizers are configured by environment variable for now (ADR-0006). Per-team organizer management in the database is needed when tasuki becomes a SaaS.
- A single LINE Login channel serves all teams (ADR-0007).
- Whether one user can belong to several teams is decided when tasuki becomes a SaaS (ADR-0008).
- Running a SaaS later also brings operational and legal responsibilities (availability, backups, terms of service, privacy policy) that need their own decisions.

---

<details>
<summary>日本語</summary>

# ADR-0005: データはマルチテナント、MVP の機能は1チーム

- Status: Accepted
- Type: business
- Date: 2026-10-11
- Issue: #45

## Context（背景）

tasuki は、LINE で連絡を取り合っている実在の1つのリレーマラソンチームから始まります。MVP はそのチームの大会のためのもので、出欠、想定ペース、走順、予想合計タイムを扱います。

将来ほかのチームが tasuki を使う場合は、**SaaS** として提供します。私たちが運用する1つのサービスを、多くのチームで共有する形です。各チームが自分でデプロイするソフトウェアとしては配りません。

1つの環境で多くのチームを扱う SaaS では、チームごとのデータを厳密に分ける必要があります。メンバーのペース、結果、得意な距離は個人データです。チームでの絞り込みを1か所でも忘れると、他のチームに見えてしまいます。1チーム前提のスキーマに後からチームの区別を入れるには、主要なテーブルすべてへの列の追加、データの移行、すべてのクエリの見直しが必要で、最も高くつき、最も危険な変更になります。

## Decision（決定）

**データモデルは最初からマルチテナントにし、MVP の機能は1チームだけを扱う。**

- データモデルに**チーム**（テナント）を持つ。チームが持つレコードはすべてチームに属し、すべてのクエリを現在のチームで絞り込む
- MVP のチームはちょうど1つ。チームを作る画面、チームを切り替える画面、新しいチームとして登録する画面は作らない
- tasuki を SaaS にするときに足りないのは、チームの登録とチームごとの幹事の管理だけで、既存のデータは変えずに済む

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| 今は1チーム前提のデータモデルにし、後でマルチテナントにする | 後からチームの区別を入れると、すべてのテーブルとクエリに手が入る。絞り込みを1か所忘れると、他のチームの個人データが漏れる |
| MVP から完全なマルチテナントの SaaS にする（登録、チームの管理） | 1つのチームも使っていない段階で、MVP の範囲が倍になる |
| チームごとに1つのデプロイ（セルフホスト、または私たちがチームごとに別の環境を用意する） | チームごとにホスティングと LINE ログインのチャネルが必要になり、多くの環境の運用は広がらない。SaaS の方向に合わない |

## Consequences（影響）

- チームが持つテーブルはすべてチームへの参照を持ち、チームが1つしかない間も、すべてのクエリでチームを絞り込む必要がある。あるチームから別のチームのデータが見えないことをテストで確かめる
- MVP は1チームのアプリとして動く。チームは初期設定の手順かマイグレーションで作る
- 幹事は今は環境変数で決める（ADR-0006）。SaaS にするときは、チームごとの幹事をデータベースで管理する必要がある
- LINE ログインのチャネルは1つで全チームに対応できる（ADR-0007）
- 1人のユーザーが複数のチームに所属できるかは、SaaS にするときに決める（ADR-0008）
- SaaS を運用するときには、運用上と法律上の責任（可用性、バックアップ、利用規約、プライバシーポリシー）も生じ、それぞれ判断が必要になる

</details>
