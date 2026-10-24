# ADR-0015: PostgreSQL access — pgx, sqlc, and golang-migrate

- Status: Proposed
- Type: technical
- Date: 2026-10-24
- Issue: #175

## Context

PostgreSQL is the source of truth (ADR-0002); production runs on Neon (`app/backend/TECH.md`). The backend needs a driver, a way to write queries, and a migration tool. Related: #20.

The forces:

- **The onion architecture** (ADR-0011) puts repositories and query services in Infrastructure. Repositories turn rows into entities with the entity factories' `Reconstruct`; nothing outside Infrastructure knows about tables.
- **The project already generates code from a source and checks it in CI**: the API from the OpenAPI document (ADR-0010, ADR-0017). Database access can follow the same rule.
- **Writes across PostgreSQL and MongoDB** use an outbox in the same PostgreSQL transaction as the use case's changes (ADR-0016), so several repositories must share one transaction.
- **The schema uses PostgreSQL types** such as `uuid`, `timestamptz`, `jsonb` (race descriptions), and partial unique indexes (`docs/design/database.md`).
- **lib/pq**, the older Go driver, is in maintenance mode, and its README recommends pgx.
- **Neon's connection pooler** runs PgBouncer in transaction mode, which does not keep session state between transactions; Neon recommends a direct connection for migrations.

## Decision

**The backend uses pgx v5 through `pgxpool`, writes queries in SQL and generates Go code with sqlc, and manages the schema with golang-migrate.**

**Driver**

- pgx v5 is used directly with `pgxpool`, not through `database/sql`, so PostgreSQL types (`uuid`, `timestamptz`, `jsonb`) map to Go types without conversion layers.
- The backend connects to Neon's **direct endpoint**, not the `-pooler` endpoint. `pgxpool` already pools connections in the process, and a direct connection keeps pgx's prepared statement cache working.
- Pool sizes (`MaxConns`, `MinConns`, `MaxConnLifetime`, `MaxConnIdleTime`) come from configuration. **The number of backend instances × `MaxConns` must stay below the database's `max_connections`** (checked with `SHOW max_connections;`); the instance limit is set with hosting (ADR-0018).

**Queries**

- Queries are written in `.sql` files with clear names (`-- name: GetRosterMember :one`), and sqlc generates the Go code (`sql_package: pgx/v5`). Mistakes in column names or types are found when the code is generated, before tests run.
- Updates and inserts use `RETURNING` to read back what they wrote in one round trip — for example the new `version` in optimistic locking.
- The generated code is committed, and CI regenerates it and fails on a difference, as for the API code (ADR-0017).
- The queries, the sqlc configuration, and the generated code are in Infrastructure (`internal/infrastructure/postgres/`). Generated types do not leave Infrastructure: repositories convert them to entities with `Reconstruct`, and query services convert them to DTOs.

**Transactions**

- A unit of work in Infrastructure begins a pgx transaction and gives every repository the same transaction (sqlc's `WithTx`). Use cases ask for it through an Application interface, which is how the outbox row of ADR-0016 is written together with the use case's changes.

**Migrations**

- golang-migrate manages the schema with separate up and down SQL files in `app/backend/migrations/`.
- **Version numbers are zero-padded** (`000001_create_teams.up.sql`), because sqlc reads the files in lexicographic order and golang-migrate in numeric order; padding makes the two orders the same. sqlc reads the up files as the schema and ignores the down files.
- Migrations are applied by the deployment, over the direct connection. The schema is never changed by hand. How they are applied in production is set up in #28.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| lib/pq | In maintenance mode; its README recommends pgx |
| pgx through `database/sql` | Gives up pgx's own type handling for no gain; tasuki has no library that needs `database/sql` |
| Hand-written queries with pgx | Mistakes in column names, order, or types are found only when the code runs. It does not follow the rule the project uses for the API: generate from the source and check it in CI |
| An ORM (GORM, ent) | Hides the SQL, brings its own conventions into Infrastructure, and makes query counts (N+1) harder to see; the queries tasuki needs are fixed in shape |
| Goose for migrations | Would work as well; sqlc reads both. golang-migrate keeps up and down migrations in separate plain SQL files |
| Neon's pooled endpoint (`-pooler`) | Transaction-mode pooling conflicts with pgx's prepared statement cache unless pgx is configured for it, and `pgxpool` already pools connections. Revisit if the number of instances grows beyond what direct connections allow |

## Consequences

- Query mistakes are found at generation time, and the generated functions are typed.
- Queries whose shape changes with the input (optional filters) are awkward in sqlc. The MVP has none (no search, no pagination); if one appears, it is written by hand with pgx in the same package.
- Changing a query changes the generated code, which is part of the pull request.
- Neon suspends an idle database after 5 minutes on the free plan, so the first query after a pause is slower.
- CI needs PostgreSQL to run integration tests (#23), and the deployment needs the direct connection string for migrations (#28).
- Neon's free compute allows roughly a hundred connections, which limits how many backend instances can run at once with a given `MaxConns`.

---

<details>
<summary>日本語</summary>

# ADR-0015: PostgreSQL へのアクセス — pgx、sqlc、golang-migrate

- Status: Proposed
- Type: technical
- Date: 2026-10-24
- Issue: #175

## Context（背景）

PostgreSQL が正（ADR-0002）で、本番は Neon で動く（`app/backend/TECH.md`）。backend には、ドライバ、クエリの書き方、マイグレーションのツールが要る。関連：#20。

考慮したこと：

- **オニオンアーキテクチャ**（ADR-0011）では、リポジトリと QueryService は Infrastructure にある。リポジトリは、Entity の Factory の `Reconstruct` で行を Entity に変える。Infrastructure の外はテーブルを知らない
- **このプロジェクトはすでに、元からコードを生成して CI で確かめている**：OpenAPI の文書から API のコードを（ADR-0010、ADR-0017）。DB へのアクセスも同じルールに揃えられる
- **PostgreSQL と MongoDB をまたぐ書き込み**では、ユースケースの変更と同じ PostgreSQL のトランザクションで outbox に書く（ADR-0016）。そのため、複数のリポジトリが 1 つのトランザクションを共有する必要がある
- **スキーマは PostgreSQL の型を使う**：`uuid`、`timestamptz`、`jsonb`（大会の説明欄）、部分ユニークインデックス（`docs/design/database.md`）
- 古い Go のドライバの **lib/pq** はメンテナンスモードで、README で pgx を勧めている
- **Neon の接続をまとめる仕組み（pooler）**は、PgBouncer を transaction mode で動かしていて、トランザクションの間でセッションの状態を持たない。Neon はマイグレーションに直接の接続を使うよう勧めている

## Decision（決定）

**backend は pgx v5 を `pgxpool` で使い、クエリは SQL で書いて sqlc で Go のコードを生成し、スキーマは golang-migrate で管理する。**

**ドライバ**

- pgx v5 は `database/sql` を通さず `pgxpool` でそのまま使う。PostgreSQL の型（`uuid`、`timestamptz`、`jsonb`）が、変換の層なしに Go の型になる
- backend は Neon の **直接の接続先**につなぎ、`-pooler` の接続先は使わない。`pgxpool` がすでにプロセスの中で接続をまとめていて、直接の接続なら pgx の prepared statement のキャッシュがそのまま動く
- プールの大きさ（`MaxConns`、`MinConns`、`MaxConnLifetime`、`MaxConnIdleTime`）は設定から読む。**backend のインスタンス数 × `MaxConns` は、DB の `max_connections` を下回る必要がある**（`SHOW max_connections;` で確かめる）。インスタンス数の上限はホスティング（ADR-0018）で決める

**クエリ**

- クエリは分かりやすい名前で `.sql` ファイルに書き（`-- name: GetRosterMember :one`）、sqlc が Go のコードを生成する（`sql_package: pgx/v5`）。列名や型の間違いは、テストより前、コードを生成したときに見つかる
- 更新と追加では `RETURNING` を使い、書いたものを 1 回の往復で読み戻す。例：楽観ロックの新しい `version`
- 生成したコードはコミットし、CI で生成し直して差分が出たら失敗させる。API のコードと同じ扱い（ADR-0017）
- クエリ、sqlc の設定、生成したコードは Infrastructure（`internal/infrastructure/postgres/`）に置く。生成した型は Infrastructure の外に出さない：リポジトリは `Reconstruct` で Entity に、QueryService は DTO に変える

**トランザクション**

- Infrastructure の作業単位（unit of work）が pgx のトランザクションを始め、すべてのリポジトリに同じトランザクションを渡す（sqlc の `WithTx`）。ユースケースは Application の interface を通してこれを使う。ADR-0016 の outbox の行も、こうしてユースケースの変更と一緒に書く

**マイグレーション**

- golang-migrate で、上げると戻すを別々の SQL ファイルにして `app/backend/migrations/` で管理する
- **バージョン番号はゼロで埋める**（`000001_create_teams.up.sql`）。sqlc はファイルを辞書順に、golang-migrate は数の順に読むので、ゼロで埋めて 2 つの順番を同じにする。sqlc は上げるファイルをスキーマとして読み、戻すファイルは無視する
- マイグレーションはデプロイの中で、直接の接続で適用する。スキーマを手で変えない。本番での適用のしかたは #28 で用意する

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| lib/pq | メンテナンスモードで、README で pgx を勧めている |
| pgx を `database/sql` 経由で使う | pgx 自身の型の扱いを手放すだけで得るものがない。tasuki には `database/sql` を必要とするライブラリがない |
| pgx でクエリを手書きする | 列名、順番、型の間違いが、動かすまで見つからない。API で採っている「元から生成し、CI で確かめる」というルールに沿わない |
| ORM（GORM、ent） | SQL が隠れ、ORM 自身の作法が Infrastructure に入り、クエリの回数（N+1）が見えにくくなる。tasuki が必要とするクエリは形が決まっている |
| マイグレーションに Goose | 同じように使え、sqlc はどちらも読める。golang-migrate は、上げると戻すを別々の素の SQL ファイルにする |
| Neon の pooler の接続先（`-pooler`） | transaction mode の pooler は、pgx を合わせて設定しない限り pgx の prepared statement のキャッシュとぶつかる。`pgxpool` がすでに接続をまとめている。インスタンス数が増えて直接の接続で足りなくなったら見直す |

## Consequences（影響）

- クエリの間違いは生成のときに見つかり、生成した関数には型が付く
- 入力によって形が変わるクエリ（任意の絞り込み）は sqlc では書きにくい。MVP にはない（検索もページ分けもない）。出てきたら、同じパッケージで pgx で手書きする
- クエリを変えると生成したコードも変わり、PR に含まれる
- Neon の無料プランは、使われていない DB を 5 分で止めるので、止まった後の最初のクエリは遅くなる
- CI で結合テストを動かすには PostgreSQL が要る（#23）。デプロイには、マイグレーション用の直接の接続文字列が要る（#28）
- Neon の無料の計算資源の接続上限は 100 前後なので、ある `MaxConns` で同時に動かせる backend のインスタンス数に上限がある

</details>
