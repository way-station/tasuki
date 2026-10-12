# ADR-0010: Describe the API spec-first with OpenAPI

- Status: Accepted
- Type: technical
- Date: 2026-10-12
- Issue: #58

## Context

ADR-0004 moved the backend from Hono to Go, which lost Hono RPC: the frontend can no longer import the API's types directly. Without a replacement, the frontend's types for requests and responses are written by hand and drift from the backend without anyone noticing (#30).

The API also needs to be designed before it is built. The access policy (ADR-0006) and the data model (ADR-0005, ADR-0008) shape every endpoint, and the endpoints, request and response shapes, and errors should be reviewable before any handler is written.

The OpenAPI tools for Go and TypeScript were checked against each project's current release (October 2026):

| Tool | Approach | OpenAPI | chi |
| --- | --- | --- | --- |
| oapi-codegen v2.8.0 | Generates Go code from the document | 3.0, and 3.1 (released in v2.8.0, described as initial support; needs Go 1.25+) | Generates a chi server |
| huma v2.39.1 | Generates the document from Go code | 3.1 | Through an adapter (`humachi`) |
| swaggo/swag v1.16.6 | Generates the document from code comments | 2.0; OpenAPI 3 only in v2, a release candidate since 2023 | Yes |
| go-swagger v0.36.6 | Both | 2.0 only | Not directly |
| ogen v1.24.0 | Generates Go code from the document | 3.x | No; uses its own generated router |
| openapi-typescript v7.13.0 | Generates TypeScript types from the document | 3.0 and 3.1, not 2.0 | — |

## Decision

The API is described **spec-first** in one OpenAPI document, which is the contract between the backend and the frontend.

- Any change to the API starts with a change to the document, reviewed in the pull request together with the code.
- The document is written in **OpenAPI 3.0**. Moving to 3.1 is reconsidered when oapi-codegen's 3.1 support matures.
- **Backend**: **oapi-codegen** generates the Go types and the server interface for chi. Handlers implement the generated interface, so a handler that does not match the document does not compile. Routing and middleware stay on chi (ADR-0004).
- **Frontend**: **openapi-typescript** generates the TypeScript types from the same document.
- **API docs UI** (Swagger UI or Redoc) is served only in development. Production does not serve it.
- Left to the implementation (#30): where the document lives, whether generated code is committed, the CI check that generated code is up to date, the frontend HTTP client, and whether to use oapi-codegen's strict server.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| huma (code-first) | The document is an output of Go code, so the API is designed in Go and cannot be reviewed before the handlers exist; handlers take huma's own signatures |
| swaggo/swag (comment annotations) | Comments are not checked against the code and drift; OpenAPI 3, which openapi-typescript needs, exists only in a release candidate |
| go-swagger | OpenAPI 2.0 only, which openapi-typescript does not read |
| ogen | Uses its own generated router instead of chi (ADR-0004) |
| OpenAPI 3.1 now | oapi-codegen describes its 3.1 support as initial; 3.0 covers what tasuki needs |
| Hand-written types on both sides | Nothing keeps the frontend and the backend in sync |

## Consequences

- The API can be designed and reviewed as a document before any code, and the frontend can start from the document without waiting for the backend.
- A response type changed in the document becomes a compile error in Go and a type error in TypeScript after generation.
- The document is written by hand in YAML and is verbose; every API change touches it first.
- A code generation step is added to the backend and the frontend, and CI must fail when generated code is out of date.
- oapi-codegen registers all operations on one router, so the per-group middleware from ADR-0004 and ADR-0006 (login, approved member, organizer) has to be mapped onto generated routes. How is decided in the implementation, for example with the document's `security` field or by mounting the generated handlers on chi sub-routers.
- The backend depends on oapi-codegen and the libraries it uses (kin-openapi). The generated code is plain `net/http` and chi, so the handlers are not tied to a framework.

---

<details>
<summary>日本語</summary>

# ADR-0010: API を OpenAPI で仕様優先に記述する

- Status: Accepted
- Type: technical
- Date: 2026-10-12
- Issue: #58

## Context（背景）

ADR-0004 で backend を Hono から Go に移したので、Hono RPC を失いました。frontend は API の型を直接 import できません。代わりがなければ、frontend のリクエストとレスポンスの型を手で書くことになり、誰も気づかないうちに backend とずれていきます（#30）。

また、API は作る前に設計する必要があります。アクセスの方針（ADR-0006）とデータモデル（ADR-0005、ADR-0008）がすべてのエンドポイントの形を決めます。エンドポイント、リクエストとレスポンスの形、エラーは、ハンドラを1つも書く前にレビューできるべきです。

Go と TypeScript の OpenAPI のツールを、各プロジェクトの今のリリースで確認しました（2026年10月）。

| ツール | 進め方 | OpenAPI | chi |
| --- | --- | --- | --- |
| oapi-codegen v2.8.0 | ドキュメントから Go のコードを生成する | 3.0 と 3.1（3.1 は v2.8.0 で入った初期対応。Go 1.25 以上が必要） | chi のサーバーを生成できる |
| huma v2.39.1 | Go のコードからドキュメントを生成する | 3.1 | アダプタ（`humachi`）経由 |
| swaggo/swag v1.16.6 | コードのコメントからドキュメントを生成する | 2.0。OpenAPI 3 は v2 だけで、v2 は 2023 年からリリース候補のまま | 可 |
| go-swagger v0.36.6 | どちらも | 2.0 のみ | 直接は不可 |
| ogen v1.24.0 | ドキュメントから Go のコードを生成する | 3.x | 不可。自前で生成するルーターを使う |
| openapi-typescript v7.13.0 | ドキュメントから TypeScript の型を生成する | 3.0 と 3.1。2.0 は不可 | — |

## Decision（決定）

API は、1つの OpenAPI のドキュメントで**仕様優先**に記述する。このドキュメントが backend と frontend の間の契約になる。

- API の変更は、必ずドキュメントの変更から始め、コードと一緒に PR でレビューする
- ドキュメントは **OpenAPI 3.0** で書く。oapi-codegen の 3.1 対応が成熟したら、3.1 への移行を考え直す
- **backend**：**oapi-codegen** で Go の型と chi 用のサーバーのインターフェースを生成する。ハンドラは生成されたインターフェースを実装するので、ドキュメントと合わないハンドラはコンパイルが通らない。ルーティングとミドルウェアは chi のまま（ADR-0004）
- **frontend**：同じドキュメントから **openapi-typescript** で TypeScript の型を生成する
- **API ドキュメントの UI**（Swagger UI か Redoc）は開発環境だけで出す。本番では出さない
- 実装（#30）に任せること：ドキュメントの置き場所、生成コードをコミットするか、生成コードが最新かを確かめる CI のチェック、frontend の HTTP クライアント、oapi-codegen の strict server を使うか

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| huma（コード優先） | ドキュメントが Go のコードから出てくるので、API の設計が Go の中で行われ、ハンドラができる前にレビューできない。ハンドラが huma 独自の形になる |
| swaggo/swag（コメントの注釈） | コメントはコードと照合されないので、ずれていく。openapi-typescript に必要な OpenAPI 3 はリリース候補にしかない |
| go-swagger | OpenAPI 2.0 だけで、openapi-typescript が読めない |
| ogen | chi ではなく、自前で生成するルーターを使う（ADR-0004） |
| 今から OpenAPI 3.1 | oapi-codegen 自身が 3.1 対応を初期対応としている。tasuki に必要なことは 3.0 で足りる |
| 両側で型を手で書く | frontend と backend を揃えておく仕組みがない |

## Consequences（影響）

- コードを書く前に、API をドキュメントとして設計し、レビューできる。frontend は backend を待たずにドキュメントから作り始められる
- ドキュメントでレスポンスの型を変えると、生成し直した後、Go ではコンパイルエラー、TypeScript では型エラーになる
- ドキュメントは YAML で手で書き、量が多い。API を変えるときは、必ず先にドキュメントに手を入れる
- backend と frontend にコード生成の手順が増える。生成コードが古いときは CI を落とす必要がある
- oapi-codegen はすべてのオペレーションを1つのルーターに登録するので、ADR-0004 と ADR-0006 のグループごとのミドルウェア（ログイン、承認済みのメンバー、幹事）を、生成されたルートに対応させる必要がある。方法は実装で決める。例えば、ドキュメントの `security` を使う、生成したハンドラを chi のサブルーターに載せる、など
- backend は oapi-codegen と、それが使うライブラリ（kin-openapi）に依存する。生成されるコードは普通の `net/http` と chi なので、ハンドラがフレームワークに縛られることはない

</details>
