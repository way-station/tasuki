# ADR-0014: Server-side sessions in an HttpOnly cookie

- Status: Accepted
- Type: technical
- Date: 2026-10-19
- Issue: #165

## Context

After LINE Login, the backend issues its own session; ADR-0007 left how the session is stored to the implementation. The API first carried a token in an `Authorization: Bearer` header, which means the frontend has to keep the token somewhere scripts can read it. #162 changed the API to an HttpOnly cookie; this ADR records the reasons and the rules that come with it.

The forces:

- **Sessions must end before they expire.** Logging out makes the session unusable (`DELETE /session`), a member who leaves can no longer use tasuki, and a user whose link is removed stops working as a member (`docs/design/api.md`, ADR-0006).
- **Members open tasuki in the LINE in-app browser.** Invite links arrive in LINE (ADR-0007). If the login is lost on every reload or every time LINE is closed, members have to log in again each time.
- **A script injected into the page (XSS) can read anything that page scripts can read**, including a token kept in `localStorage`, and use it from elsewhere until it expires.
- **There is one client (the browser) and one backend.** No other service or app needs to verify the session.
- The backend uses PostgreSQL as the source of truth (ADR-0002) and the onion architecture (ADR-0011).

## Decision

**Sessions are kept on the server, and the browser holds only a session token in an HttpOnly cookie.**

**Sessions**

- A session is a row in PostgreSQL: the user, a SHA-256 hash of the session token, when it expires, and when it was revoked. The session token itself is not stored, so a leaked database does not give usable tokens.
- The session token is a random value from a cryptographically secure generator. It means nothing by itself; only the row says whose it is.
- A session is revoked by setting its revocation time: on logout (`DELETE /session`), and for all of a user's sessions when its roster member leaves.
- What a user may do (member, organizer) is read from current data on every request, so removing a link takes effect on the next request without touching sessions.
- How long a session lasts is decided in #84. There is no refresh token (api.md).

**Cookie**

- The session token is sent in the cookie `tasuki_session` with `HttpOnly`, `Secure`, `SameSite=Lax`, and `Path=/api`.
- The frontend never reads, stores, or sends the session token; the browser sends the cookie. Nothing about it is kept in TanStack Query (ADR-0013).
- The frontend and the API are served from the same site.

**CSRF**

- Because the browser sends the cookie automatically, requests forged by other sites must be rejected:
  - `SameSite=Lax`: the browser does not send the cookie on requests from other sites, except top-level navigations with safe methods.
  - Requests that change data (`POST`, `PUT`, `PATCH`, `DELETE`) must carry an `Origin` header equal to the frontend's origin; otherwise `403 forbidden`.
- `GET` requests never change data.

**Identifying the user**

- A middleware in Presentation reads the cookie and passes the user ID to the controllers through the request context. It is a closure that captures its dependencies (`func Authenticate(uc) func(http.Handler) http.Handler`).
- The middleware does not read the database itself. Looking up the session goes through an Application use case and a query service (ADR-0011).
- Missing, unknown, expired, or revoked sessions return `401 unauthenticated`.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| Bearer token in `localStorage` | An XSS script can read the token and use it from anywhere until it expires |
| Bearer token in memory only | Lost on every reload and every time the LINE in-app browser is closed, so members log in again each time |
| Stateless tokens (JWT) | Cannot be ended before they expire without keeping state on the server again, which logout and leaving require. With one backend and one client, not reading the database on each request gains little |
| Access token in memory + refresh token in an HttpOnly cookie | The refresh cookie needs the same CSRF protection, and token renewal adds flows to build, for no benefit over a session cookie with one backend |
| LIFF, logging in again silently inside LINE, so the token can stay in memory | Solves storage only inside LINE; tasuki can also be opened in other browsers. Whether to use LIFF for login is still open in ADR-0007, separate from how sessions are stored |
| Sessions in the backend's memory | Lost on restart, and not shared when the backend runs as more than one process |

## Consequences

- An XSS script cannot read the session token or take it elsewhere. It can still send requests while the page is open, so preventing XSS (escaping, no raw HTML) is still needed.
- Logout, leaving, and removing a link take effect on the next request.
- Every authenticated request reads the session from PostgreSQL.
- CSRF protection becomes a backend rule that every route that changes data relies on; the `Origin` check runs for all of them, not per handler.
- Hosting (#7) must serve the frontend and the API from the same site (for example the API under `/api` on the frontend's domain). In local development, the frontend's dev server forwards `/api` to the backend so the cookie stays on one origin. Whether browsers accept a `Secure` cookie on `http://localhost` must be checked when this is implemented.
- How long the cookie and the session survive in the LINE in-app browser must be checked on real devices.
- The Go server stays stateless: any process can serve any request, because the sessions are in PostgreSQL.

---

<details>
<summary>日本語</summary>

# ADR-0014: サーバー側のセッションと HttpOnly Cookie

- Status: Accepted
- Type: technical
- Date: 2026-10-19
- Issue: #165

## Context（背景）

LINE ログインの後、backend は自前のセッションを出す。どう保存するかは ADR-0007 で実装に任せていた。API は最初、`Authorization: Bearer` ヘッダーでトークンを運ぶ形だった。この形では、frontend がスクリプトから読める場所にトークンを置く必要がある。#162 で API を HttpOnly Cookie に変えた。この ADR は、その理由と、一緒に決まるルールを記録する。

考慮したこと：

- **セッションは期限の前に終わらせる必要がある。** ログアウトするとセッションは使えなくなる（`DELETE /session`）。退会したメンバーは tasuki を使えなくなる。紐付けを外されたユーザーはメンバーとして使えなくなる（`docs/design/api.md`、ADR-0006）
- **メンバーは LINE のアプリ内ブラウザで tasuki を開く。** 招待リンクは LINE で届く（ADR-0007）。リロードや LINE を閉じるたびにログインが消えると、そのたびにログインし直すことになる
- **ページに埋め込まれたスクリプト（XSS）は、ページのスクリプトが読めるものを何でも読める。** `localStorage` のトークンも読まれ、期限が来るまでほかの場所から使われる
- **クライアントはブラウザ 1 つ、backend も 1 つ。** セッションを確かめる必要のある、ほかのサービスやアプリはない
- backend は PostgreSQL を正とし（ADR-0002）、オニオンアーキテクチャを使う（ADR-0011）

## Decision（決定）

**セッションはサーバーに持ち、ブラウザは HttpOnly Cookie に入れたセッショントークンだけを持つ。**

**セッション**

- セッションは PostgreSQL の行：ユーザー、セッショントークンの SHA-256 ハッシュ、期限、無効にした時刻。セッショントークンそのものは保存しないので、DB が漏れても使えるトークンは出ない
- セッショントークンは、暗号学的に安全な乱数生成器で作るランダムな値。それ自体は意味を持たず、誰のものかは行だけが知っている
- セッションは、無効にした時刻を入れて無効にする：ログアウトのとき（`DELETE /session`）、名簿のメンバーが退会したときはそのユーザーのすべてのセッション
- ユーザーが何をできるか（メンバー、幹事）は、リクエストのたびに今のデータから読む。紐付けを外すと、セッションに触れなくても次のリクエストから効く
- セッションの期限の長さは #84 で決める。リフレッシュトークンはない（api.md）

**Cookie**

- セッショントークンは Cookie `tasuki_session` で運ぶ。`HttpOnly`、`Secure`、`SameSite=Lax`、`Path=/api`
- frontend はセッショントークンを読まず、保存せず、送らない。Cookie はブラウザが送る。TanStack Query にも何も入れない（ADR-0013）
- frontend と API は同じサイトから出す

**CSRF**

- ブラウザは Cookie を自動で送るので、ほかのサイトが偽って送らせたリクエストを断る必要がある：
  - `SameSite=Lax`：ブラウザは、ほかのサイトからのリクエストに Cookie を付けない。安全なメソッドでのページの移動だけは例外
  - データを変えるリクエスト（`POST`、`PUT`、`PATCH`、`DELETE`）は、frontend のオリジンと同じ `Origin` ヘッダーを持つ必要がある。なければ `403 forbidden`
- `GET` のリクエストはデータを変えない

**ユーザーの判定**

- Presentation のミドルウェアが Cookie を読み、ユーザー ID をリクエストの context でコントローラに渡す。依存を閉じ込めたクロージャ（`func Authenticate(uc) func(http.Handler) http.Handler`）にする
- ミドルウェアは DB を直接読まない。セッションの照合は Application のユースケースと QueryService を通す（ADR-0011）
- セッションがない、知らない、期限切れ、無効のときは `401 unauthenticated`

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| Bearer トークンを `localStorage` に置く | XSS のスクリプトがトークンを読み、期限が来るまでどこからでも使える |
| Bearer トークンをメモリにだけ置く | リロードや LINE のアプリ内ブラウザを閉じるたびに消え、メンバーはそのたびにログインし直す |
| ステートレスなトークン（JWT） | サーバーに状態を持ち直さない限り、期限の前に終わらせられない。ログアウトと退会にはそれが要る。backend とクライアントが 1 つずつなので、リクエストごとに DB を読まずに済む利点は小さい |
| メモリのアクセストークン + HttpOnly Cookie のリフレッシュトークン | リフレッシュ用の Cookie にも同じ CSRF の対策が要り、更新の流れも作ることになる。backend が 1 つなら、セッションの Cookie より良い点がない |
| LIFF（LINE の中で自動でログインし直し、トークンはメモリに置く） | 解決するのは LINE の中だけで、tasuki はほかのブラウザでも開ける。ログインに LIFF を使うかは ADR-0007 でまだ決めておらず、セッションの保存方法とは別の話 |
| セッションを backend のメモリに置く | 再起動で消え、backend を複数のプロセスで動かすと共有されない |

## Consequences（影響）

- XSS のスクリプトは、セッショントークンを読めず、持ち出せない。ただし、ページが開いている間はリクエストを送れるので、XSS を防ぐこと（エスケープ、生の HTML を入れない）は今までどおり必要
- ログアウト、退会、紐付けを外すことは、次のリクエストから効く
- 認証の要るリクエストはすべて、PostgreSQL からセッションを読む
- CSRF の対策は、データを変えるすべてのルートが頼る backend のルールになる。`Origin` の確認は、ハンドラごとではなく全体で行う
- ホスティング（#7）は、frontend と API を同じサイトから出す必要がある（例：frontend のドメインの `/api` で API を出す）。ローカル開発では、frontend の開発サーバーが `/api` を backend に転送し、Cookie が 1 つのオリジンに収まるようにする。`http://localhost` でブラウザが `Secure` の Cookie を受け付けるかは、実装するときに確かめる
- LINE のアプリ内ブラウザで Cookie とセッションがどれだけ残るかは、実機で確かめる必要がある
- Go のサーバーはステートレスのまま。セッションは PostgreSQL にあるので、どのプロセスもどのリクエストを受けてよい

</details>
