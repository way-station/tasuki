# ADR-0011: Backend onion architecture with CQRS

- Status: Proposed
- Type: technical
- Date: 2026-10-13
- Issue: #80

## Context

The backend so far has two packages: `server` (routes and handlers) and `store` (database connections). That was enough for `/health`, but the MVP API (`docs/design/api.md`) brings much more:

- **Business rules that chain together.** Cancelling an invitation removes the member from attendance, the entry, and the running order, and can delete their lap times and recalculate their average time. A changed attendance answer, a member marked as left, and an edited race trigger the same kind of chain (`docs/design/database.md`).
- **Rules that must hold no matter which endpoint changes the data.** Only attending members can be in an entry, only entry members can run, a running order can be confirmed only when every leg has a runner, and times are `MM:SS` rounded half up.
- **Reads that look different for each caller.** The same race returns different fields to organizers, attending members, and invited members (ADR-0006). Predictions are calculated for the reader, and some screens read from both PostgreSQL and MongoDB (for example a member's laps by race).
- **Two databases** (ADR-0002): PostgreSQL as the source of truth and MongoDB for lap records and running order history.
- **Generated code** (ADR-0010): oapi-codegen generates the server interface that handlers implement.

If these rules live in handlers or in SQL, each endpoint that touches the same data has to repeat them, and the rules cannot be tested without databases.

## Decision

The backend uses an **onion architecture** with four layers. Dependencies point inward only: Presentation → Application → Domain, and Infrastructure → Application and Domain. **Domain depends on nothing.**

| Layer | Contains |
| --- | --- |
| Presentation | **Controllers** that implement the interface generated from the OpenAPI document (ADR-0010), and **routing** (chi, ADR-0004) |
| Application | **Use cases**, the **commands** they take as input, **DTOs / view models** they return, and **query service interfaces** |
| Domain | **Entities**, **value objects**, **factories**, and **repository interfaces** |
| Infrastructure | **Repositories** (implement the Domain repository interfaces) and **query services** (implement the Application query service interfaces), for PostgreSQL and MongoDB |

**CQRS**

- **Commands (writes)**: Controller → use case with a command → entities loaded and saved through repository interfaces. Business rules are enforced by the entities and value objects.
- **Queries (reads)**: Controller → query service interface → DTO / view model. Query services read what each screen needs directly, including across PostgreSQL and MongoDB, without loading entities. What a caller may see (ADR-0006) is applied in the query.

**Factories and encapsulation**

- Entities and value objects keep their fields unexported. They are created only through factories, which reject invalid values (for example a lap time that is not `MM:SS`, or a race description that does not match its type), so an invalid object cannot exist.
- State changes go through methods on the entity, not by setting fields.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| Keep `server` and `store` (handlers call the database directly) | The chained rules would be repeated in every handler that touches the same data, and could only be tested with databases |
| Layered architecture where Domain depends on Infrastructure | Business rules would depend on PostgreSQL and MongoDB details, so they could not be tested alone and a storage change would reach the rules |
| Onion without CQRS (reads also go through repositories and entities) | Reads differ per caller and combine both databases; building them from entities means loading more than needed and mapping entities to many response shapes |
| Plain struct literals instead of factories | Any package could build an invalid object; validation would be spread across callers |

## Consequences

- Business rules can be tested in Domain and Application with fake repositories, without databases.
- Each endpoint touches more files (controller, use case or query service, command or DTO, repository), and data is mapped between layers. This is more code than calling the database from handlers.
- Writes and reads can drift: a rule enforced on the write side (for example who can see a race) must also be applied in the query services. Tests for both sides are needed.
- Repositories that write to both PostgreSQL and MongoDB cannot share a transaction (ADR-0002). Use cases that change both (for example deleting lap times when a runner is removed) must define the order and what happens when the second write fails; this is decided in the implementation issues.
- The API implementation issues (#73–#79) are split by these layers.

---

<details>
<summary>日本語</summary>

# ADR-0011: backend のオニオンアーキテクチャと CQRS

- Status: Proposed
- Type: technical
- Date: 2026-10-13
- Issue: #80

## Context（背景）

backend にはこれまで、`server`（ルートとハンドラー）と `store`（データベースの接続）の2つのパッケージしかありません。`/health` にはそれで足りましたが、MVP の API（`docs/design/api.md`）ではずっと多くのことが必要になります。

- **連鎖する業務のルール**：招待を取り消すと、そのメンバーは出欠、エントリー、走順から外れ、ラップタイムが消えて平均タイムが計算し直されることがある。出欠の回答の変更、退会、大会の編集でも、同じような連鎖が起きる（`docs/design/database.md`）
- **どのエンドポイントからデータを変えても守るべきルール**：エントリーに入れるのは参加のメンバーだけ、走れるのはエントリーのメンバーだけ、すべての区に走者がいるときだけ走順を確定できる、タイムは `MM:SS` で四捨五入
- **呼ぶ人ごとに違う形の読み取り**：同じ大会でも、幹事、参加のメンバー、誘われたメンバーで返す項目が違う（ADR-0006）。予想は読む人のために計算する。PostgreSQL と MongoDB の両方から読む画面もある（例：メンバーの大会ごとのラップ）
- **2つのデータベース**（ADR-0002）：正は PostgreSQL、ラップと走順の履歴は MongoDB
- **生成するコード**（ADR-0010）：oapi-codegen が、ハンドラーが実装するサーバーのインターフェースを生成する

これらのルールをハンドラーや SQL に置くと、同じデータに触るエンドポイントごとに繰り返すことになり、データベースなしではテストできません。

## Decision（決定）

backend は、4つの層の**オニオンアーキテクチャ**にする。依存は内側にだけ向ける：Presentation → Application → Domain、Infrastructure → Application と Domain。**Domain は何にも依存しない。**

| 層 | 入るもの |
| --- | --- |
| Presentation | OpenAPI の文書から生成したインターフェース（ADR-0010）を実装する**コントローラー**と、**ルーティング**（chi、ADR-0004） |
| Application | **ユースケース**、その入力の**コマンド**、返す **DTO／ビューモデル**、**クエリサービスのインターフェース** |
| Domain | **エンティティ**、**値オブジェクト**、**ファクトリー**、**リポジトリのインターフェース** |
| Infrastructure | **リポジトリ**（Domain のリポジトリのインターフェースを実装）と**クエリサービス**（Application のクエリサービスのインターフェースを実装）。PostgreSQL と MongoDB |

**CQRS**

- **コマンド（更新）**：コントローラー → コマンドを受け取るユースケース → リポジトリのインターフェースでエンティティを読み書きする。業務のルールはエンティティと値オブジェクトが守る
- **クエリ（参照）**：コントローラー → クエリサービスのインターフェース → DTO／ビューモデル。クエリサービスは、エンティティを読み込まずに、画面が必要とするものを直接読む。PostgreSQL と MongoDB をまたいでもよい。呼ぶ人が見られる範囲（ADR-0006）はクエリの中で絞る

**ファクトリーとカプセル化**

- エンティティと値オブジェクトの項目は外に公開しない。作るのはファクトリーだけで、正しくない値は断る（例：`MM:SS` でないラップタイム、種別に合わない大会の説明欄）。正しくないオブジェクトは存在できない
- 状態の変更は、項目を直接書き換えるのではなく、エンティティのメソッドで行う

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| `server` と `store` のまま（ハンドラーがデータベースを直接呼ぶ） | 連鎖するルールを、同じデータに触るハンドラーごとに繰り返すことになり、データベースなしではテストできない |
| Domain が Infrastructure に依存するレイヤードアーキテクチャ | 業務のルールが PostgreSQL と MongoDB の詳細に依存し、単独でテストできず、保存方法の変更がルールにまで及ぶ |
| CQRS なしのオニオン（参照もリポジトリとエンティティを通す） | 読み取りは呼ぶ人ごとに違い、2つのデータベースを組み合わせる。エンティティから組み立てると、必要以上に読み込み、エンティティを多くのレスポンスの形に変換することになる |
| ファクトリーを使わず構造体リテラルで作る | どのパッケージからでも正しくないオブジェクトを作れる。検証が呼ぶ側に散らばる |

## Consequences（影響）

- 業務のルールを、偽のリポジトリを使って Domain と Application でテストでき、データベースは要らない
- エンドポイントごとに触るファイルが増え（コントローラー、ユースケースかクエリサービス、コマンドか DTO、リポジトリ）、層の間でデータを変換する。ハンドラーからデータベースを呼ぶより、コードが多くなる
- 更新と参照がずれることがある。更新側で守るルール（例：誰が大会を見られるか）は、クエリサービスでも当てはめる必要がある。両方のテストが必要
- PostgreSQL と MongoDB の両方に書くリポジトリは、トランザクションを共有できない（ADR-0002）。両方を変えるユースケース（例：走者を外したときにラップタイムを消す）は、書く順番と、2つ目の書き込みが失敗したときにどうするかを決める必要がある。これは実装の Issue で決める
- API の実装の Issue（#73〜#79）は、この層で分ける

</details>
