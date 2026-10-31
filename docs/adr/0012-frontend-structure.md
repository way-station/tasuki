# ADR-0012: Frontend structure

- Status: Accepted
- Type: technical
- Date: 2026-10-24
- Issue: #107

## Context

Each screen of the frontend has to do several different things: call the backend, convert what the backend returns into what the screen shows, hold UI state and drafts (ADR-0013), and render.

The data the backend returns is not in the shape the screens use. The API uses `snake_case`, times as strings such as `"3:28"`, and codes for states (`docs/design/api.md`), while the screens show labels and take input in forms. The API types are generated from the OpenAPI document (ADR-0010), so any change to the API changes these types.

If one component does all of these things, a change to any one of them — the API, the state, or the look — touches the same file, and the component can only be checked with a running backend.

The frontend implementation issues are already split per feature (#109, #115, #121, #127, #133, #139), with one sub-issue per layer.

## Decision

**Each feature is split into five layers: Types, Mapper, Hooks, Container, and UI.** Each layer has one responsibility (single responsibility principle).

| Layer | Contains | Does not |
| --- | --- | --- |
| **Types** | Types generated from the OpenAPI document, and view types (what the screens use) | — |
| **Mapper** | Pure functions that convert API types into view types, and form input into request bodies | Call the backend, hold state |
| **Hooks** | TanStack Query queries and mutations (ADR-0013). They call the backend through the API client and convert with the Mapper (for example in `select`) | Render, hold UI state |
| **Container** | Calls hooks, reads the URL (TanStack Router), holds UI state (`useState` / `useReducer`) and drafts (ADR-0013), and passes data and callbacks to the UI as props | Call the backend directly, convert API types |
| **UI** | Components that only display what they receive as props, and report user actions through callback props | Fetch data, use the router or TanStack Query, hold state, import API types |

**Dependency direction (imports)**

```
Container ──→ UI
    │
    └──→ Hooks ──→ Mapper ──→ Types
           │
           └──→ API client (interface)
UI ──→ Types (view types only)
```

- Imports go only in the directions above. In particular, the UI does not import the Container, Hooks, Mapper, or the generated API types.
- At run time, data flows UI → Container → Hooks → backend → Hooks (Mapper) → Container → UI. A user action in the UI calls a callback the Container passed; the UI does not import the Container.

**API client behind an interface**

- Hooks call the backend through an `ApiClient` interface, not through `fetch` or a concrete client (dependency inversion principle).
- The real implementation uses the generated API types. Tests pass a fake implementation instead.
- A fake must keep the contract of the real client — the same results and the same errors (`401`, `409 version_conflict`, `problem+json`) — so hooks behave the same with either (Liskov substitution principle).
- The client is given to the hooks with a React Context provider. This passes a dependency, not state, so it does not conflict with ADR-0013's rule that Context is used only for UI state shared across screens.

**Folders**

```
src/
  features/
    auth/                 # #109
      types/  mapper/  hooks/  container/  ui/
    races/                # #115
    participation/        # #121
    running-orders/       # #127
    laps/                 # #133
    roster/               # #139
  app/                    # what belongs to no feature: routes, the common layout, the not-found page (#31)
  api/                    # ApiClient interface and implementation, query keys (ADR-0013)
  shared/                 # only what more than one feature needs (see below)
```

- Folders are per feature first, then per layer, matching the issues.

**Shared code**

- Code is moved to `shared/` only when a second feature needs the same thing, and then it is not duplicated (DRY). It is not moved before that happens, because what will actually be shared is not known yet (YAGNI).
- Exceptions decided elsewhere: the API client and query keys are in `api/` from the start (ADR-0013: query keys in one place), and the common UI components from the UI design (#147) are in `shared/ui/`.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| No split (one component fetches, converts, holds state, and renders) | API changes, state, and the look are mixed in one file, against the single responsibility principle |
| Hooks only, without Containers | Where state is held would differ from screen to screen, and where drafts are saved (ADR-0013) would not be decided |
| Feature-Sliced Design or a similar larger structure | More layers than tasuki needs at its size (YAGNI) |
| Folders per layer first (`hooks/races`, `ui/races`) | One feature's change would be spread across folders, while the issues and the work are per feature |
| Hooks call `fetch` or a concrete client directly | The backend could not be replaced in tests without network mocking; hooks would depend on the implementation instead of an interface |

## Consequences

- When the API changes, the changes stay in Types, Mapper, and Hooks; the UI does not change as long as the view types stay the same.
- The UI can be checked without the backend, by passing props. Mappers are pure functions and can be tested alone. Hooks can be tested with a fake API client.
- Each feature has more files, and data is passed from the Container to the UI as props.
- Splitting Containers from UI components is the container/presentational pattern. Its author no longer recommends it in general now that hooks exist; tasuki keeps it for the reasons above (one place for state and drafts, UI without the backend) and accepts the extra props.
- The import rules are only conventions until a lint rule enforces them; the rule is chosen with the lint tool (#34).
- The frontend implementation issues stay split by these layers.

---

<details>
<summary>日本語</summary>

# ADR-0012: frontend の構成

- Status: Accepted
- Type: technical
- Date: 2026-10-24
- Issue: #107

## Context（背景）

frontend の画面は、いくつもの違うことをする必要がある：backend を呼ぶ、backend が返したものを画面に出す形に変える、UI の状態と下書きを持つ（ADR-0013）、描画する。

backend が返すデータは、画面が使う形ではない。API は `snake_case` で、タイムは `"3:28"` のような文字列、状態はコード（`docs/design/api.md`）。一方で画面は、表示用の言葉を出し、フォームで入力を受け取る。API の型は OpenAPI の文書から生成する（ADR-0010）ので、API が変わればこの型も変わる。

1 つのコンポーネントがこれを全部やると、API、状態、見た目のどれを変えても同じファイルに手が入り、backend を動かさないとそのコンポーネントを確かめられない。

frontend の実装の Issue は、すでに機能ごとに分かれていて（#109、#115、#121、#127、#133、#139）、層ごとに sub-issue がある。

## Decision（決定）

**機能ごとに、Types、Mapper、Hooks、Container、UI の 5 つの層に分ける。** 各層の責任は 1 つ（単一責任の原則）。

| 層 | 入るもの | しないこと |
| --- | --- | --- |
| **Types** | OpenAPI の文書から生成した型と、画面用の型（画面が使うもの） | — |
| **Mapper** | API の型を画面用の型に、フォームの入力をリクエストの本文に変える純粋な関数 | backend を呼ぶ、状態を持つ |
| **Hooks** | TanStack Query の query と mutation（ADR-0013）。API クライアントを通して backend を呼び、Mapper で変換する（例：`select` で） | 描画する、UI の状態を持つ |
| **Container** | Hooks を呼び、URL を読み（TanStack Router）、UI の状態（`useState` / `useReducer`）と下書き（ADR-0013）を持ち、データとコールバックを props で UI に渡す | backend を直接呼ぶ、API の型を変換する |
| **UI** | props で受け取ったものを表示するだけで、ユーザーの操作はコールバックの props で知らせる部品 | データを取る、ルーターや TanStack Query を使う、状態を持つ、API の型を import する |

**依存の向き（import）**

```
Container ──→ UI
    │
    └──→ Hooks ──→ Mapper ──→ Types
           │
           └──→ API クライアント（interface）
UI ──→ Types（画面用の型だけ）
```

- import は上の向きだけにする。特に、UI は Container、Hooks、Mapper、生成した API の型を import しない
- 実行するときのデータの流れは UI → Container → Hooks → backend → Hooks（Mapper）→ Container → UI。UI での操作は、Container が渡したコールバックを呼ぶ。UI が Container を import するわけではない

**interface の後ろに置く API クライアント**

- Hooks は、`fetch` や具体的なクライアントではなく、`ApiClient` の interface を通して backend を呼ぶ（依存性逆転の原則）
- 本物の実装は生成した API の型を使う。テストでは偽物の実装を渡す
- 偽物は本物の約束を守る必要がある。同じ結果と同じエラー（`401`、`409 version_conflict`、`problem+json`）を返し、どちらでも Hooks が同じように動くようにする（リスコフの置換原則）
- クライアントは React Context の provider で Hooks に渡す。これは状態ではなく依存を渡すものなので、「Context は画面をまたぐ UI の状態だけに使う」という ADR-0013 のルールとはぶつからない

**フォルダ**

```
src/
  features/
    auth/                 # #109
      types/  mapper/  hooks/  container/  ui/
    races/                # #115
    participation/        # #121
    running-orders/       # #127
    laps/                 # #133
    roster/               # #139
  app/                    # どの機能にも属さないもの：ルート、共通のレイアウト、not found ページ（#31）
  api/                    # ApiClient の interface と実装、query key（ADR-0013）
  shared/                 # 複数の機能が必要とするものだけ（下を参照）
```

- フォルダは、まず機能ごと、その中を層ごとに分ける。Issue と同じ形

**共通の処理**

- 2 つ目の機能が同じものを必要としたときにだけ `shared/` に移し、重複させない（DRY）。それより前には移さない。何が本当に共通になるかは、まだ分からないため（YAGNI）
- ほかで決まっている例外：API クライアントと query key は最初から `api/` に置く（ADR-0013：query key は 1 か所）。UI のデザイン（#147）の共通の部品は `shared/ui/` に置く

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| 分けない（1 つのコンポーネントが取得、変換、状態、描画をすべてやる） | API の変更、状態、見た目が 1 つのファイルに混ざる。単一責任の原則に反する |
| Hooks だけで分け、Container を置かない | 状態を持つ場所が画面ごとにばらつき、下書きをどこで保存するか（ADR-0013）が決まらない |
| Feature-Sliced Design のような、より大きな構成 | tasuki の規模に対して層が多すぎる（YAGNI） |
| フォルダをまず層ごとに分ける（`hooks/races`、`ui/races`） | 1 つの機能の変更が複数のフォルダに散らばる。Issue と作業は機能ごと |
| Hooks が `fetch` や具体的なクライアントを直接呼ぶ | ネットワークをモックしない限り、テストで backend を差し替えられない。Hooks が interface ではなく実装に依存する |

## Consequences（影響）

- API が変わっても、変更は Types、Mapper、Hooks に収まる。画面用の型が変わらなければ UI は変わらない
- UI は props を渡せば backend なしで確かめられる。Mapper は純粋な関数なので単体でテストできる。Hooks は偽物の API クライアントでテストできる
- 機能ごとのファイルが増え、データは Container から UI へ props で渡すことになる
- Container と UI の部品を分けるのは、container / presentational パターン。考案者は、hooks がある今は一般には勧めていない。tasuki は上の理由（状態と下書きの置き場所を 1 つにする、backend なしで UI を確かめる）で採用し、props が増えることを受け入れる
- import のルールは、lint で強制するまでは約束にすぎない。ルールは lint のツールと一緒に決める（#34）
- frontend の実装の Issue は、引き続きこの層で分ける

</details>
