# How tasuki verifies LINE's ID token

- Decision: [ADR-0020](../adr/0020-verify-line-id-tokens-with-golang-jwt.md)
- Code: `app/backend/internal/infrastructure/line/client.go` (`verify`), tests in `client_test.go`

This guide explains the mechanisms behind ADR-0020: what an ID token is, what its signature proves, why web login uses HS256 and not RS or ES, and what each check in `verify` stops. Read it before changing how tokens are verified, or when moving to LIFF.

## 1. What an ID token is

After LINE Login, the backend receives an ID token from LINE's token endpoint. It is a JWT (RFC 7519): three base64url parts joined by dots.

```
header.payload.signature
```

| Part | Content | Example |
| --- | --- | --- |
| Header | How the signature was made | `{"alg": "HS256", "typ": "JWT"}` |
| Payload | The claims | `{"iss": "https://access.line.me", "sub": "U1234", "aud": "<channel ID>", "exp": ..., "nonce": "...", "name": "Taro"}` |
| Signature | Computed from the header and payload with a key | bytes |

## 2. What the signature proves, and what the payload says

- The **signature** proves who wrote the token: only someone with the key can produce it. For tasuki, that is LINE.
- The **payload** says who the token is for. `sub` is the LINE user ID; every login gets its own token with its own `sub`, `nonce`, and `exp`.

The key is not per user. One key covers every token of the channel, and tokens differ by their payload.

## 3. HS (shared secret) and RS / ES (key pair)

| | HS256 | RS256 / ES256 |
| --- | --- | --- |
| Keys | One secret, used to sign and to verify | A private key signs; a public key verifies |
| Who can verify | Whoever has the secret | Anyone with the public key |
| Who can forge | Whoever has the secret — every verifier | Only the holder of the private key |
| Fits when | Only one party verifies, and it can keep a secret | Many parties verify, or a verifier cannot keep a secret |

HS's weakness is that every verifier can also sign. That does not matter when only one party verifies and keeps the secret safe.

**Why LINE web login uses HS256.** The channel secret is shared by two parties only: LINE and tasuki's backend. It never reaches the browser. The backend is the only verifier, and it can keep the secret, so a shared secret is enough.

**Why LINE uses ES256 for native apps and LIFF.** A secret inside an app or a page can be extracted, so it cannot be kept. LINE signs with a private key and publishes the public keys (`https://api.line.me/oauth2/v2.1/certs`), which can verify but not forge.

tasuki does not choose the algorithm: LINE does, by login type.

## 4. `alg: none`

RFC 7518 section 3.6 defines `none`: a JWT with no signature (the third part is empty), for cases where something else, such as TLS, already protects it. Anyone can make such a token, so a verifier that accepts it lets anyone write any `sub`. In 2015, several JWT libraries accepted `alg: none` as verified. RFC 7518 says unsecured JWTs must not be accepted by default.

golang-jwt implements `none` like any other algorithm (`none.go`), but its `Verify` passes only when the key is the constant `jwt.UnsafeAllowNoneSignatureType`. tasuki passes the channel secret as `[]byte`, so `none` is always refused. The test `alg none` signs with that constant to play the attacker; production code never uses it.

## 5. Why tasuki fixes the algorithm (`jwt.WithValidMethods`)

In golang-jwt v5.3.1:

- `parser.go` picks the verification method from the token's own header: `token.Method = GetSigningMethod(token.Header["alg"])`.
- `hmac.go` implements HS256, HS384, and HS512 as one type, `SigningMethodHMAC`; its `Verify` only requires the key to be `[]byte`.

So, without `WithValidMethods`, a token that says `HS512` is verified with HS512 and the channel secret, and passes if signed with it. The test `alg HS512 with the channel secret` shows this: remove the option and only that case fails. `none` stays refused because of section 4.

With HS only, the practical risk of HS512 is small: whoever can sign HS512 with the secret can also sign HS256. The option matters most when algorithm families mix. If a verifier that holds an RS or ES public key let the token choose `HS256`, an attacker could sign with the public key — which is public — as the HMAC secret, and the token would pass. Fixing the accepted algorithms rules this out (RFC 8725 section 2.1).

`parser.go` checks `validMethods` before calling the key function and before verifying the signature.

## 6. The checks in `verify`

| Check | Done by | What it stops | Test |
| --- | --- | --- | --- |
| Signature with the channel secret | golang-jwt | Tokens not made by LINE | `signed with another secret` |
| `alg` is HS256 | golang-jwt, `WithValidMethods` | A token choosing its own algorithm | `alg none`, `alg HS512 with the channel secret` |
| `iss` is `https://access.line.me` | golang-jwt, `WithIssuer` | Tokens from another issuer | `another issuer`, `no issuer` |
| `aud` is the channel ID | golang-jwt, `WithAudience` | A token LINE issued to another app | `another channel` |
| `exp` present and not passed | golang-jwt, `WithExpirationRequired` | Old tokens, tokens that never expire | `expired`, `no expiry` |
| `nonce` equals the stored one | tasuki | Replaying a token from another login | `another nonce`, `no nonce` |
| `sub` present | tasuki | A token that names no user | `no subject` |

## 7. tasuki's own sessions are not JWTs

The ID token is used once, at login. After it, tasuki issues its own session: a random token in an HttpOnly cookie, with its SHA-256 hash in PostgreSQL (ADR-0014). JWTs appear only in what LINE sends.

## Sources

- LINE Developers, "Verify ID token" and "Integrating LINE Login with your web app"
- RFC 7519 (JWT) section 6, "Unsecured JWTs"
- RFC 7518 (JWA) section 3.6, "Using the Algorithm 'none'"
- RFC 8725 (JWT Best Current Practices) sections 2.1 and 3.1
- OpenID Connect Core 1.0 section 3.1.3.7, "ID Token Validation"
- golang-jwt v5.3.1: `parser.go`, `hmac.go`, `none.go`

---

<details>
<summary>日本語</summary>

# tasuki が LINE の ID トークンを確かめる仕組み

- 決定：[ADR-0020](../adr/0020-verify-line-id-tokens-with-golang-jwt.md)
- コード：`app/backend/internal/infrastructure/line/client.go`（`verify`）、テストは `client_test.go`

ADR-0020 の裏にある仕組みの解説。ID トークンとは何か、署名が何を保証するか、Web ログインがなぜ RS や ES ではなく HS256 なのか、`verify` のそれぞれの確認が何を防いでいるかを説明する。トークンの確かめ方を変える前や、LIFF に移るときに読むこと。

## 1. ID トークンとは

LINE ログインの後、backend は LINE のトークンエンドポイントから ID トークンを受け取る。これは JWT（RFC 7519）で、base64url の3つの部分を点でつないだもの。

```
ヘッダー.中身.署名
```

| 部分 | 内容 | 例 |
| --- | --- | --- |
| ヘッダー | 署名をどう作ったか | `{"alg": "HS256", "typ": "JWT"}` |
| 中身 | クレーム | `{"iss": "https://access.line.me", "sub": "U1234", "aud": "<チャネル ID>", "exp": ..., "nonce": "...", "name": "Taro"}` |
| 署名 | ヘッダーと中身から鍵で計算したもの | バイト列 |

## 2. 署名が保証すること、中身が表すこと

- **署名**は、誰がトークンを書いたかを保証する。鍵を持つ人にしか作れない。tasuki にとっては LINE
- **中身**は、誰のためのトークンかを表す。`sub` は LINE のユーザー ID。ログインのたびに、自分の `sub`、`nonce`、`exp` を持つ別のトークンができる

鍵はユーザーごとではない。1つの鍵がチャネルのすべてのトークンに使われ、トークンの違いは中身にある。

## 3. HS（共通の秘密）と RS／ES（鍵のペア）

| | HS256 | RS256／ES256 |
| --- | --- | --- |
| 鍵 | 1つの秘密で、署名も確認もする | 秘密鍵で署名し、公開鍵で確かめる |
| 確かめられる人 | 秘密を持つ人 | 公開鍵を持つ人なら誰でも |
| 偽物を作れる人 | 秘密を持つ人。つまり確かめる人全員 | 秘密鍵を持つ人だけ |
| 向いている場面 | 確かめるのが1者だけで、そこが秘密を守れる | 確かめる人が多い、または確かめる側が秘密を守れない |

HS の弱点は、確かめる人は誰でも署名もできること。確かめるのが1者だけで、そこが秘密を安全に持てるなら問題にならない。

**LINE の Web ログインが HS256 である理由。** チャネルシークレットを持つのは、LINE と tasuki の backend の2者だけ。ブラウザには渡らない。確かめるのは backend だけで、backend は秘密を守れるので、共通の秘密で足りる。

**LINE がネイティブアプリと LIFF で ES256 を使う理由。** アプリやページの中の秘密は取り出せるので、守れない。LINE は秘密鍵で署名し、公開鍵（`https://api.line.me/oauth2/v2.1/certs`）を公開する。公開鍵では確かめられるが、偽物は作れない。

アルゴリズムを選ぶのは tasuki ではない。ログインの種類ごとに LINE が決めている。

## 4. `alg: none`

RFC 7518 の 3.6 節は `none` を定義している。署名のない JWT（3つ目が空）で、TLS など別のものがすでに守っている場面のためのもの。こうしたトークンは誰でも作れるので、受け付けてしまうと、誰でも好きな `sub` を書ける。2015年には、いくつもの JWT ライブラリが `alg: none` を検証済みとして通していた。RFC 7518 は、署名のない JWT を既定で受け付けてはいけないとしている。

golang-jwt は `none` をほかのアルゴリズムと同じように実装している（`none.go`）が、その `Verify` は鍵が定数 `jwt.UnsafeAllowNoneSignatureType` のときだけ通す。tasuki はチャネルシークレットを `[]byte` で渡すので、`none` は必ず断られる。テストの `alg none` は、攻撃者の役としてこの定数で署名している。本番のコードでは使わない。

## 5. tasuki がアルゴリズムを決めておく理由（`jwt.WithValidMethods`）

golang-jwt v5.3.1 では：

- `parser.go` は、確かめる方法をトークン自身のヘッダーから選ぶ：`token.Method = GetSigningMethod(token.Header["alg"])`
- `hmac.go` は、HS256、HS384、HS512 を1つの型 `SigningMethodHMAC` で実装していて、その `Verify` が鍵に求めるのは `[]byte` であることだけ

そのため `WithValidMethods` がないと、`HS512` を名乗るトークンは HS512 とチャネルシークレットで確かめられ、それで署名してあれば通る。テストの `alg HS512 with the channel secret` がこれを示している。このオプションを消すと、落ちるのはこのケースだけ。`none` は 4 の理由で断られたまま。

HS だけなら、HS512 の実際の危険は小さい。シークレットで HS512 の署名ができる人は、HS256 の署名もできるから。このオプションが一番効くのは、アルゴリズムの種類が混ざる場面。RS や ES の公開鍵を持つ検証側がトークンに `HS256` を選ばせると、攻撃者は公開鍵（公開されている）を HMAC のシークレットにして署名でき、そのトークンが通ってしまう。受け付けるアルゴリズムを決めておけば、これは起きない（RFC 8725 の 2.1 節）。

`parser.go` は、鍵の関数を呼ぶ前、署名を確かめる前に `validMethods` を確かめている。

## 6. `verify` の確認

| 確認 | 誰がするか | 防ぐもの | テスト |
| --- | --- | --- | --- |
| チャネルシークレットでの署名 | golang-jwt | LINE が作っていないトークン | `signed with another secret` |
| `alg` が HS256 | golang-jwt、`WithValidMethods` | アルゴリズムを自分で選ぶトークン | `alg none`、`alg HS512 with the channel secret` |
| `iss` が `https://access.line.me` | golang-jwt、`WithIssuer` | ほかの発行者のトークン | `another issuer`、`no issuer` |
| `aud` がチャネル ID | golang-jwt、`WithAudience` | LINE がほかのアプリに出したトークン | `another channel` |
| `exp` があり、過ぎていない | golang-jwt、`WithExpirationRequired` | 古いトークン、期限のないトークン | `expired`、`no expiry` |
| `nonce` が保存した値と同じ | tasuki | ほかのログインのトークンの使い回し | `another nonce`、`no nonce` |
| `sub` がある | tasuki | ユーザーを名乗らないトークン | `no subject` |

## 7. tasuki 自身のセッションは JWT ではない

ID トークンを使うのはログインのときの1回だけ。その後は tasuki が自分のセッションを発行する：HttpOnly の Cookie に入れたランダムなトークンと、PostgreSQL に置いたその SHA-256 のハッシュ（ADR-0014）。JWT が出てくるのは、LINE から届くものだけ。

## 一次資料

- LINE Developers「Verify ID token」「Integrating LINE Login with your web app」
- RFC 7519（JWT）の 6 節「Unsecured JWTs」
- RFC 7518（JWA）の 3.6 節「Using the Algorithm 'none'」
- RFC 8725（JWT のベストプラクティス）の 2.1 節と 3.1 節
- OpenID Connect Core 1.0 の 3.1.3.7 節「ID Token Validation」
- golang-jwt v5.3.1：`parser.go`、`hmac.go`、`none.go`

</details>
