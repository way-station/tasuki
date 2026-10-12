# ADR-0002: PostgreSQL as the source of truth, MongoDB as a secondary store

- Status: Accepted
- Type: technical
- Date: 2026-10-11
- Issue: #10

## Context

tasuki stores two kinds of data with very different shapes.

**Relational data with rules the database must guarantee.** Teams, accounts, roster members (ADR-0005, ADR-0008), races, attendance, and the running order are connected to each other, and they come with rules:

- one attendance answer per member per race;
- only members who attend can be placed in the running order;
- a runner may run at most a limited number of laps in lap-based races;
- every record belongs to a team (ADR-0005).

**Append-only, document-shaped data.**

- Race-day lap records: one run produces a list of lap times, and the number of laps differs per runner.
- Running order history: each change is kept as a full snapshot of the order before and after.
- Future external integration data: raw activities from services such as Garmin or Strava, whose shape is defined by each service and differs between them.

This second kind is written once, read as a whole, and rarely joined with other data.

## Decision

We will place data according to its shape.

| Store | Data |
| --- | --- |
| **PostgreSQL** (source of truth) | Teams, accounts, roster members, account links, invitations, races, attendance, running order |
| **MongoDB** (secondary) | Race-day lap records, running order history, raw data from external integrations |

The PostgreSQL schema stays focused on relational data. Data whose shape changes often, or is defined by external services, is not mixed into the core tables.

Rules between the two stores:

1. **PostgreSQL is always the source of truth.** There are no transactions across the two databases.
2. **Reference, never copy.** MongoDB documents refer to PostgreSQL records only by their IDs. PostgreSQL data (names, race details, …) is never duplicated into MongoDB.
3. **Write summaries back.** When a race ends, its result summary (for example, the total time of each run) is written back to PostgreSQL, so results survive even if MongoDB data is lost.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| PostgreSQL only, with JSONB columns for document-shaped data | Technically possible: JSONB can hold lap arrays, snapshots, and external payloads. We chose to keep append-only and externally shaped data out of the relational schema, so that the core tables and their migrations stay focused on the data the rules above apply to, and so that external payloads can be stored as received without schema changes |
| MongoDB only | The rules above (uniqueness, attendance before running order, lap limits, team ownership) could not be enforced by the database and would be scattered across application code |

## Consequences

- Two databases have to be run, backed up, and monitored, locally and in production.
- No transaction spans both databases. A write that touches both must be designed so that PostgreSQL is updated first and MongoDB data can be rebuilt or ignored if it is missing.
- Because MongoDB only holds IDs, reading lap records together with runner names needs a lookup in PostgreSQL.
- The free MongoDB Atlas tier (M0) cannot enable backups, is automatically paused after 30 days without connections, and may be deactivated under the Terms of Service. Races happen only a few times a year, so long idle periods are expected. Writing summaries back to PostgreSQL keeps race results safe in that case.
- Whether the backend must have MongoDB available at startup is decided separately (#26).

---

<details>
<summary>日本語</summary>

# ADR-0002: PostgreSQL を正、MongoDB を従のストアにする

- Status: Accepted
- Type: technical
- Date: 2026-10-11
- Issue: #10

## Context（背景）

tasuki は、形がまったく違う2種類のデータを扱います。

**DB で守るべきルールを持つ、関係のデータ。** チーム、アカウント、名簿のメンバー（ADR-0005、ADR-0008）、大会、出欠、走順は互いにつながっていて、次のようなルールがあります。

- 1人1大会につき、出欠の回答は1件
- 走順に入れられるのは、出席するメンバーだけ
- 周回制の大会では、1人が走れる周回数に上限がある
- すべてのレコードはチームに属する（ADR-0005）

**追記するだけの、ドキュメントの形をしたデータ。**

- 当日のラップ記録：出走1回ごとにラップタイムの一覧ができ、周回数は人によって違う
- 走順の変更履歴：変更のたびに、変更前と変更後の走順をスナップショットとして丸ごと残す
- 将来の外部連携データ：Garmin や Strava などの生のアクティビティ。形はそれぞれのサービスが決めていて、サービスごとに違う

2つ目の種類は、一度書いたら変えず、まとめて読み、ほかのデータと結合することがほとんどありません。

## Decision（決定）

データの形に合わせて置き場所を決める。

| ストア | データ |
| --- | --- |
| **PostgreSQL**（正） | チーム、アカウント、名簿のメンバー、アカウントの紐付け、招待、大会、出欠、走順 |
| **MongoDB**（従） | 当日のラップ記録、走順の変更履歴、外部連携の生データ |

PostgreSQL のスキーマは関係のデータに集中させる。形がよく変わるデータや、外部のサービスが形を決めるデータは、中心のテーブルに混ぜない。

2つのストアの間のルール：

1. **正は常に PostgreSQL。** 2つの DB をまたぐトランザクションはない
2. **参照のみ、複製しない。** MongoDB のドキュメントは PostgreSQL のレコードを ID で参照するだけにする。名前や大会の情報などの PostgreSQL のデータを、MongoDB に複製しない
3. **サマリーを書き戻す。** 大会が終わったら、結果のサマリー（例：出走ごとの合計タイム）を PostgreSQL に書き戻す。MongoDB のデータが失われても、結果は残る

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| PostgreSQL だけにし、ドキュメントの形のデータは JSONB の列に入れる | 技術的には可能。JSONB でもラップの配列、スナップショット、外部のデータを持てる。そのうえで、追記するだけのデータと外部が形を決めるデータは関係のスキーマの外に置くことを選んだ。中心のテーブルとそのマイグレーションを、上のルールが適用されるデータに集中させるため。また、外部のデータをスキーマの変更なしに受け取ったままの形で保存するため |
| MongoDB だけ | 上のルール（一意性、出欠があってからの走順、周回数の上限、チームへの所属）を DB で守れず、アプリのコードのあちこちに散らばる |

## Consequences（影響）

- ローカルでも本番でも、2つの DB を動かし、バックアップし、監視する必要がある
- 2つの DB をまたぐトランザクションはない。両方に書き込む処理は、先に PostgreSQL を更新し、MongoDB のデータが欠けていても作り直せるか無視できるように設計する
- MongoDB は ID しか持たないので、ラップ記録を選手の名前と一緒に読むには PostgreSQL を参照する必要がある
- MongoDB Atlas の無料枠（M0）はバックアップを有効にできず、接続のない状態が30日続くと自動で一時停止され、利用規約に基づいて無効化されることがある。大会は年に数回なので、長いアイドル期間は避けられない。サマリーを PostgreSQL に書き戻すことで、その場合も大会の結果は守られる
- backend の起動時に MongoDB を必須にするかは、別に決める（#26）

</details>
