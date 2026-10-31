# ADR-0013: Frontend state and server data

- Status: Accepted
- Type: technical
- Date: 2026-10-18
- Issue: #160
- Updated: 2026-10-30, the router changed from React Router to TanStack Router (#31)

## Context

Almost all data on tasuki's screens comes from the backend: races, invitations, attendance, entries, running orders, lap times, and the roster (`docs/design/api.md`). There is little state that belongs only to the frontend.

Some of that state follows from earlier decisions, and it shapes how the frontend should hold data:

- **One change affects several screens.** Changing an attendance answer can remove the member from the entry and the running order; submitting lap times recalculates the member's average time (`docs/design/screens.md`, R2 and L1). After a change, every screen that shows the affected data must show the new data.
- **Concurrent edits are detected with `version`.** Races and running orders return `409 version_conflict` when someone else saved first (api.md).
- **What a caller may see depends on who they are** (ADR-0006). The same request returns different data to an organizer and to a member.
- **The token expires and there is no refresh token** (api.md). When it expires, the user logs in with LINE again, which leaves the page.
- **Editing can take a long time.** An organizer may leave the running order editor (R7) open and stop in the middle, for example by falling asleep, until the token expires.
- **Predictions are calculated by the backend** (ADR-0011). The frontend only shows the total time or laps it returns.

Without a rule, each screen would decide for itself where to keep data and when to fetch it again, and screens would show old data after a change.

## Decision

**Each kind of state has one place.** We do not use a global store (such as Redux or Zustand).

| State | Where it lives | Examples |
| --- | --- | --- |
| Which screen and which data is shown | **URL**, through **TanStack Router without loaders** | `/races/$raceId`, a selected tab |
| Data from the backend | **TanStack Query**, for all communication with the backend, including public APIs | Race list, running order, `GET /me`, `GET /invite-links/{token}`, `POST /session` |
| UI state inside a screen | `useState` / `useReducer` | A form's current input, whether a dialog is open |
| UI state shared across screens | **React Context**, for this only; never data from the backend | Notifications after saving |
| Values that can be calculated from the above | **Not stored**; calculated when rendering | Whether every leg of a draft running order has a runner |
| Drafts of edits not yet saved | **`localStorage`**, saved on every edit | R3 race form, R7 running order, L1 lap times |

**Rules for data from the backend**

- Data is fetched only through TanStack Query. Components do not fetch in `useEffect`.
- **No optimistic updates.** After a mutation succeeds, the queries it affects are invalidated and fetched again. The screen shows what the backend returned, not what the frontend expected.
- On `409 version_conflict`, the latest data is fetched again and the user is told that someone else saved first.
- Query keys are defined in one place, so a mutation can invalidate every query it affects.
- When the signed-in user changes (log in, log out, `401`), the whole query cache is cleared, so data visible to one user is never shown to another. `401` is handled once, in the query and mutation caches' `onError`, not in each screen.
- Requests that fail with a 4xx status are not retried.
- The session token is not handled by the frontend at all: it is an HttpOnly cookie that the browser sends (#162). Nothing about it is kept in TanStack Query.

**Rules for drafts**

- On screens that edit data and save with a button, the draft is written to `localStorage` on every edit. Saving to the backend still happens only when the user presses save.
- A draft is kept when the token expires or the page is reloaded or closed. After logging in again, the user returns to the screen and the draft is restored.
- A draft is stored per user and per edited resource, so one user never sees another's draft. It is deleted when it is saved successfully and when the user logs out.
- A draft remembers the `version` it was based on. If someone else saved in the meantime, saving it returns `409 version_conflict` and the user is told.
- A draft is copied from the query data once, when editing starts. Data fetched again in the background does not overwrite it.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| `useEffect` + `useState` for fetching | Each screen keeps its own copy of the data. After a change, every screen that shows the affected data has to be found and fetched again by hand |
| Route loaders (TanStack Router loaders, React Router data mode) | Data would be fetched in two places, route loaders and components. Invalidating after a mutation would work differently depending on where the data was fetched |
| React Router without data mode | Route params are plain strings (`useParams` returns `string \| undefined`), and a link to a wrong path or without a required param is not caught by the compiler. Its type-safe routes need framework mode, which brings its own build plugin and loaders. TanStack Router checks paths, params, and search params with types, without loaders |
| SWR | Mutations and invalidating related data, which tasuki needs after every change, are weaker than in TanStack Query |
| RTK Query | Requires a Redux store, while tasuki has almost no state that belongs only to the frontend |
| A global store (Redux, Zustand) for UI state | Apart from data from the backend, the only state shared across screens is small (notifications); Context is enough |
| TanStack Query as a store for UI state | Query is built for data from the backend: entries are removed when unused and fetched again on focus, so every use would need these turned off. Clearing the cache on logout would also clear UI state. It would also break the rule that Query holds data from the backend |
| Optimistic updates | Many changes are decided by the backend: the average time is recalculated, removal from the entry follows backend rules, and `version` can reject the change. The frontend would show values that may be wrong |
| Drafts in memory only | A draft would be lost when the token expires, the page is reloaded, or the LINE in-app browser is closed |
| Saving every edit to the backend (autosave) | `version` would change on every edit, so concurrent edits would conflict more often, and the backend would have to accept incomplete running orders. Edits not yet sent would still be lost when the token expires |

## Consequences

- After any change, every screen shows the backend's data, with one rule (invalidate after success) instead of one per screen.
- There are no rollbacks to write, because nothing is shown before the backend accepts it. In exchange, the screen waits for the backend before showing the result of a change.
- Query keys become a shared contract: a mutation that does not invalidate a key leaves that screen with old data. Keeping the keys in one place reduces this risk but does not remove it.
- TanStack Query's defaults must be set for tasuki: no retries on 4xx, and how long data is considered fresh.
- An organizer who stops in the middle of editing does not lose the draft, even when the token expires. Logging in again must return the user to the screen they were on.
- Drafts in `localStorage` are visible to scripts on the page. They contain only what the user typed, not credentials.
- In the frontend structure (ADR-0012, #107), queries and mutations are in the Hooks layer, and draft storage is part of the Container that owns the editing screen.

---

<details>
<summary>日本語</summary>

# ADR-0013: frontend の状態とサーバーのデータ

- Status: Accepted
- Type: technical
- Date: 2026-10-18
- Issue: #160
- 更新：2026-10-30、ルーターを React Router から TanStack Router に変えた（#31）

## Context（背景）

tasuki の画面に出るデータは、ほとんどが backend から来る：大会、招待、出欠、エントリー、走順、ラップタイム、名簿（`docs/design/api.md`）。frontend だけが持つ状態は少ない。

そのデータには、これまでの決定から来る性質があり、frontend でのデータの持ち方に関わる：

- **1 つの変更が、いくつもの画面に効く。** 出欠の回答を変えると、エントリーと走順から外れることがある。ラップタイムを登録すると、平均タイムが計算し直される（`docs/design/screens.md` の R2、L1）。変更の後は、影響するデータを出しているすべての画面が新しいデータを出す必要がある
- **同時の編集は `version` で見つける。** 大会と走順は、ほかの人が先に保存していると `409 version_conflict` を返す（api.md）
- **見られる範囲は人によって違う**（ADR-0006）。同じリクエストでも、幹事とメンバーでは返るデータが違う
- **トークンには期限があり、リフレッシュトークンはない**（api.md）。切れたら LINE でログインし直し、そのときページを離れる
- **編集には時間がかかることがある。** 幹事が走順の編集（R7）を開いたまま、寝落ちなどで途中で止まり、トークンが切れることがある
- **予想は backend が計算する**（ADR-0011）。frontend は、返ってきた合計タイムや周回数を表示するだけ

ルールがないと、データをどこに置き、いつ取り直すかを画面ごとに決めることになり、変更の後に古いデータを出す画面が出る。

## Decision（決定）

**状態は種類ごとに置き場所を 1 つにする。** 全体ストア（Redux や Zustand など）は使わない。

| 状態 | 置き場所 | 例 |
| --- | --- | --- |
| どの画面で、どのデータを見ているか | **URL**。**loader を使わない TanStack Router** | `/races/$raceId`、選んでいるタブ |
| backend のデータ | **TanStack Query**。公開 API も含め、backend とのやり取りはすべてこれを通す | 大会一覧、走順、`GET /me`、`GET /invite-links/{token}`、`POST /session` |
| 画面の中の UI の状態 | `useState` / `useReducer` | フォームの入力中の値、ダイアログが開いているか |
| 画面をまたぐ UI の状態 | **React Context**。これだけに使い、backend のデータは入れない | 保存した後の通知 |
| 上から計算できる値 | **保存しない**。描画のときに計算する | 走順の下書きで、すべての区に走者が入っているか |
| まだ保存していない編集の下書き | **`localStorage`**。編集のたびに保存する | R3 の大会のフォーム、R7 の走順、L1 のラップタイム |

**backend のデータのルール**

- データの取得は TanStack Query だけを通す。コンポーネントの `useEffect` で fetch しない
- **楽観的更新は使わない。** mutation が成功したら、影響する query を invalidate して取り直す。画面には、frontend が予想した値ではなく、backend が返した値を出す
- `409 version_conflict` のときは、最新のデータを取り直し、ほかの人が先に保存したことを伝える
- query key は 1 か所で定義する。mutation が影響するすべての query を invalidate できるようにするため
- ログイン中のユーザーが変わったとき（ログイン、ログアウト、`401`）は、query のキャッシュをすべて消す。あるユーザーに見えるデータを、ほかのユーザーに出さないため。`401` は各画面ではなく、query と mutation のキャッシュの `onError` で 1 か所だけで扱う
- 4xx で失敗したリクエストは再試行しない
- セッションのトークンは frontend では一切扱わない。ブラウザが送る HttpOnly Cookie にする（#162）。TanStack Query にも何も入れない

**下書きのルール**

- データを編集してボタンで保存する画面では、編集のたびに下書きを `localStorage` に書く。backend への保存は、今までどおり保存を押したときだけ
- 下書きは、トークンが切れても、ページをリロードしたり閉じたりしても残る。ログインし直した後は元の画面に戻り、下書きが戻る
- 下書きはユーザーごと、編集しているリソースごとに保存する。あるユーザーの下書きを、ほかのユーザーに見せないため。保存に成功したときと、ログアウトしたときに消す
- 下書きは、元にした `version` を覚えておく。その間にほかの人が保存していたら、保存したときに `409 version_conflict` が返り、そのことを伝える
- 下書きは、編集を始めたときに query のデータから 1 回だけ写す。裏で取り直したデータで上書きしない

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| `useEffect` + `useState` で取得する | 画面ごとにデータのコピーを持つことになる。変更の後、影響するデータを出している画面を自分で探して取り直す必要がある |
| route の loader（TanStack Router の loader、React Router の data mode） | データを取る場所が、route の loader とコンポーネントの 2 つになる。mutation の後の invalidate が、どこで取ったデータかによって違う動きになる |
| data mode を使わない React Router | route の引数がただの文字列になり（`useParams` が `string \| undefined` を返す）、間違ったパスや、必要な引数のないリンクをコンパイラが見つけられない。型の付いたルートには framework mode が要り、専用のビルドプラグインと loader が付いてくる。TanStack Router は loader なしで、パス、引数、クエリ文字列を型で確かめられる |
| SWR | tasuki では変更のたびに関係するデータの invalidate が要るが、mutation と invalidate が TanStack Query より弱い |
| RTK Query | Redux のストアが前提になる。tasuki には frontend だけが持つ状態がほとんどない |
| UI の状態に全体ストア（Redux、Zustand） | backend のデータを除くと、画面をまたいで共有する状態は小さい（通知）。Context で足りる |
| TanStack Query を UI の状態のストアにする | Query は backend のデータのために作られていて、使われなくなったデータは消え、フォーカスが戻ると取り直す。使うたびにそれを止める必要がある。ログアウトでキャッシュを消すと UI の状態も消える。Query は backend のデータを持つ、というルールも崩れる |
| 楽観的更新 | 多くの変更は backend が決める：平均タイムは計算し直され、エントリーから外れるかは backend のルールで決まり、`version` で変更が断られることがある。frontend は間違っているかもしれない値を出すことになる |
| 下書きをメモリにだけ置く | トークンが切れたとき、リロードしたとき、LINE のアプリ内ブラウザを閉じたときに消える |
| 編集のたびに backend へ保存する（オートセーブ） | 編集のたびに `version` が変わり、同時の編集がぶつかりやすくなる。backend は途中までの走順を受け付ける必要がある。まだ送っていない編集は、トークンが切れたらやはり消える |

## Consequences（影響）

- どの変更の後も、すべての画面が backend のデータを出す。ルールは画面ごとではなく 1 つ（成功したら invalidate）
- backend が受け付けるまで何も出さないので、取り消しの処理を書かなくてよい。その代わり、変更の結果は backend の返事を待ってから出る
- query key は共有の約束になる。mutation がある key を invalidate しないと、その画面は古いデータのままになる。key を 1 か所にまとめるとこの危険は減るが、なくなりはしない
- TanStack Query のデフォルトを tasuki に合わせて設定する必要がある：4xx で再試行しない、データを新しいとみなす期間
- 幹事が編集の途中で止まっても、トークンが切れても、下書きは消えない。ログインし直したら、元の画面に戻れる必要がある
- `localStorage` の下書きは、ページ上のスクリプトから読める。中身はユーザーが入力したものだけで、認証情報は入らない
- frontend の構成（ADR-0012、#107）では、query と mutation は Hooks 層に置き、下書きの保存は、その編集画面を持つ Container の役割にする

</details>
