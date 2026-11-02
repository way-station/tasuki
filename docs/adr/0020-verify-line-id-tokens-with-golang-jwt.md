# ADR-0020: Verify LINE ID tokens with golang-jwt, HS256 only

- Status: Accepted
- Type: technical
- Date: 2026-11-02
- Issue: #84

## Context

ADR-0007 decided that the backend verifies LINE's ID token and identifies the user by its `sub` claim, but not how. Every login depends on this check: a token that passes it logs in as the user named in `sub`, so a mistake here lets anyone log in as anyone.

Facts, confirmed in the LINE Developers documentation (Verify ID token, November 2026):

- For web login, which tasuki uses, LINE signs the ID token with **HS256** and the channel secret. For native apps, the LINE SDK, and LIFF, it signs with **ES256** and publishes the public keys as a JWK set.
- The token can be verified on the server, or by sending it to `POST https://api.line.me/oauth2/v2.1/verify`.

What a verification has to check comes from OpenID Connect Core 1.0 section 3.1.3.7: the signature, `iss`, `aud`, `exp`, and `nonce`. RFC 8725 (JWT Best Current Practices) adds that the verifier must decide which algorithms it accepts, instead of using the `alg` the token names (sections 2.1 and 3.1). Libraries that trusted `alg` have accepted unsigned tokens (`alg: none`) and tokens signed with a public key used as an HMAC secret.

## Decision

The backend verifies the ID token itself with **`github.com/golang-jwt/jwt/v5`**, in `internal/infrastructure/line`.

- Only HS256 is accepted (`jwt.WithValidMethods`). The library checks this before it looks up the key or verifies the signature.
- The library also checks `iss` (`https://access.line.me`), `aud` (the channel ID), and `exp`, which is required.
- tasuki's code checks `nonce` against the one stored for the login, and that `sub` is present.
- Each of these checks has a test that sends a token breaking it (`client_test.go`).

If tasuki moves to LIFF (left open by ADR-0007), tokens will be ES256 and need LINE's JWK set. That change gets its own ADR.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| Verify HS256 by hand with `crypto/hmac` | Every check — splitting and decoding the token, constant-time comparison, `alg`, the claims and their types — would be ours to write and keep correct, in the code every login depends on. A widely used library already does these checks; we only configure them |
| LINE's verify endpoint | One more request to LINE on every login, and logins fail when that endpoint is slow or down. The token already comes straight from LINE's token endpoint, so a local check is enough |
| A larger JOSE library (for example `lestrrat-go/jwx`) | Its JWK set support is not needed while only HS256 is accepted; golang-jwt covers what web login needs with a smaller API |

## Consequences

- The checks that decide who logs in are done by a library many projects use and review, and tasuki's part is a few lines of configuration.
- tasuki depends on one more module and must take its security updates.
- The `jwt.WithValidMethods` line must stay. Without it, the token chooses its own algorithm. The test with an HS512 token fails if it is removed.
- Moving to LIFF needs ES256 and fetching LINE's JWK set; golang-jwt verifies ES256, but fetching and caching the keys is new work.

---

<details>
<summary>日本語</summary>

# ADR-0020: LINE の ID トークンを golang-jwt で検証し、HS256 だけを受け付ける

- Status: Accepted
- Type: technical
- Date: 2026-11-02
- Issue: #84

## Context（背景）

ADR-0007 で、backend が LINE の ID トークンを検証し、`sub` クレームでユーザーを識別することにしました。ただし、どう検証するかは決めていませんでした。すべてのログインがこの確認に頼っています。確認を通ったトークンは `sub` のユーザーとしてログインできるので、ここを間違えると、誰でも誰にでもなりすませます。

LINE Developers のドキュメント（Verify ID token、2026年11月）で確認した事実：

- tasuki が使う Web ログインでは、LINE は ID トークンをチャネルシークレットを使って **HS256** で署名する。ネイティブアプリ、LINE SDK、LIFF では **ES256** で署名し、公開鍵を JWK セットとして公開している
- 検証はサーバーで行うか、`POST https://api.line.me/oauth2/v2.1/verify` に送って行える

検証で確かめることは、OpenID Connect Core 1.0 の 3.1.3.7 節で決まっています：署名、`iss`、`aud`、`exp`、`nonce`。さらに RFC 8725（JWT のベストプラクティス）は、トークンが名乗る `alg` を使うのではなく、受け付けるアルゴリズムを検証する側が決めるべきだとしています（2.1 節と 3.1 節）。`alg` を信じたライブラリは、署名のないトークン（`alg: none`）や、公開鍵を HMAC のシークレットとして署名したトークンを通してしまったことがあります。

## Decision（決定）

backend は、ID トークンを **`github.com/golang-jwt/jwt/v5`** で自分で検証する。場所は `internal/infrastructure/line`。

- 受け付けるのは HS256 だけ（`jwt.WithValidMethods`）。ライブラリは、鍵を探したり署名を確かめたりする前にこれを確かめる
- `iss`（`https://access.line.me`）、`aud`（チャネル ID）、`exp`（必須）もライブラリが確かめる
- `nonce` がログインのときに保存した値と同じであることと、`sub` があることは、tasuki のコードで確かめる
- それぞれの確認に、それを破ったトークンを送るテストがある（`client_test.go`）

tasuki が LIFF に移る場合（ADR-0007 で未定）、トークンは ES256 になり、LINE の JWK セットが必要になる。その変更は別の ADR にする。

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| `crypto/hmac` で HS256 を自分で検証する | トークンの分割とデコード、処理時間が一定の比較、`alg`、クレームとその型の確認を、すべてのログインが頼るコードで、すべて自分たちで書いて正しく保つことになる。広く使われているライブラリがこれらを確かめるので、tasuki は設定だけを書く |
| LINE の verify エンドポイント | ログインのたびに LINE へのリクエストが1回増え、そのエンドポイントが遅いときや止まっているときにログインできなくなる。トークンは LINE のトークンエンドポイントから直接受け取るので、手元の確認で足りる |
| もっと大きな JOSE のライブラリ（`lestrrat-go/jwx` など） | HS256 だけを受け付ける間は、JWK セットへの対応は要らない。Web ログインに必要なことは、golang-jwt がもっと小さな API でまかなえる |

## Consequences（影響）

- 誰がログインできるかを決める確認を、多くのプロジェクトが使い、見ているライブラリが行う。tasuki が書くのは数行の設定になる
- 依存するモジュールが1つ増え、そのセキュリティ更新を取り込む必要がある
- `jwt.WithValidMethods` の行は消してはいけない。消すと、トークンの側がアルゴリズムを選べる。消すと HS512 のトークンのテストが落ちる
- LIFF に移るには、ES256 と、LINE の JWK セットの取得が必要になる。golang-jwt は ES256 を検証できるが、鍵の取得とキャッシュは新しい作業になる

</details>
