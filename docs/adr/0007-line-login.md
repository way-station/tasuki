# ADR-0007: Authenticate all users with LINE Login

- Status: Accepted
- Type: technical
- Date: 2026-10-11
- Issue: #47

## Context

ADR-0006 requires every member and organizer to log in, so tasuki needs an authentication method that relay runners will actually use. The team already coordinates in a LINE group, and tasuki's URL is shared there. Most people will open tasuki from inside the LINE app.

Facts about LINE Login, confirmed in the LINE Developers documentation (October 2026):

- LINE Login v2.1 supports OpenID Connect and returns an ID token when the `openid` scope is requested. PKCE is supported (`S256` only).
- For web login, the ID token is signed with HS256 using the channel secret. It can be verified on the server, or with the `POST https://api.line.me/oauth2/v2.1/verify` endpoint.
- User IDs are different for each provider, even for the same person.
- While a channel is in "Developing" status, only users with the Admin or Tester role can log in. Publishing the channel cannot be undone.
- Getting a user's email address requires applying for permission.

## Decision

Organizers and members all log in with **LINE Login (OpenID Connect)**.

- The backend runs the authorization code flow with PKCE, verifies the ID token, and identifies the user by the `sub` claim (LINE user ID).
- After login, the backend issues its own session. The session storage method is decided when it is implemented.
- Only `openid` and `profile` scopes are requested. No email address is needed.
- Authentication is kept behind one interface in the backend, so another login method can be added later without changing authorization or the data model.
- Whether to use LIFF instead of plain web login is decided when the login screen is implemented.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| Email magic link | Needs an email delivery service; runners are not used to checking email for this team |
| Email + password | Password storage, reset flows, and user recovery to build and secure ourselves |
| One shared passphrase | Cannot tell members apart, so it cannot meet ADR-0006 (own data only, approval per person) |
| Google login | Not every runner uses a Google user; LINE is what the team already uses |
| No login | Rejected by ADR-0006 |

## Consequences

- Logging in from the LINE app is close to one tap, and no passwords are stored by tasuki.
- Teams that do not use LINE (for example, in English-speaking regions) cannot use tasuki as is. Adding another provider later is possible because authentication sits behind one interface.
- One LINE Login channel, operated by us, serves every team (ADR-0005). It must be published before anyone other than its Admins and Testers can log in, and publishing cannot be undone.
- Because user IDs differ per provider, changing the LINE provider changes every user's ID. The provider must stay the same for the service, or users must be re-linked.
- Local development cannot rely on LINE Login for arbitrary test users; see ADR-0009.

---

<details>
<summary>日本語</summary>

# ADR-0007: 全員を LINE ログインで認証する

- Status: Accepted
- Type: technical
- Date: 2026-10-11
- Issue: #47

## Context（背景）

ADR-0006 で、メンバーと幹事の全員がログインすることにしました。そのため、ランナーが実際に使う認証方法が必要です。チームはもう LINE のグループで連絡を取り合っていて、tasuki の URL もそこで共有します。ほとんどの人は LINE のアプリの中から tasuki を開きます。

LINE Developers のドキュメントで確認した LINE ログインの事実（2026年10月）：

- LINE ログイン v2.1 は OpenID Connect に対応していて、`openid` スコープを要求すると ID トークンが返る。PKCE に対応している（`S256` のみ）
- Web ログインの ID トークンは、チャネルシークレットを使って HS256 で署名される。サーバーで検証するか、`POST https://api.line.me/oauth2/v2.1/verify` のエンドポイントで検証できる
- ユーザー ID は、同じ人でもプロバイダーごとに違う値になる
- チャネルが「開発中」の間は、Admin か Tester の役割を持つユーザーしかログインできない。公開は元に戻せない
- メールアドレスの取得には、別途申請が必要

## Decision（決定）

幹事もメンバーも、**LINE ログイン（OpenID Connect）**でログインする。

- backend は PKCE つきの認可コードフローを実行し、ID トークンを検証して、`sub` クレーム（LINE のユーザー ID）でユーザーを識別する
- ログイン後は、backend が自前のセッションを発行する。セッションの保存方法は実装するときに決める
- 要求するスコープは `openid` と `profile` だけ。メールアドレスは使わない
- 認証の部分は backend の1つのインターフェースの裏にまとめ、認可やデータモデルを変えずに、後からほかのログイン方法を足せるようにする
- 通常の Web ログインではなく LIFF を使うかは、ログイン画面を実装するときに決める

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| メールのマジックリンク | メール送信の仕組みが必要。ランナーはこのチームの用事でメールを見る習慣がない |
| メール + パスワード | パスワードの保存、リセット、ユーザー復旧を自分たちで作り、守る必要がある |
| 共通の合言葉1つ | メンバーを区別できず、ADR-0006（自分のデータだけ、1人ずつの承認）を満たせない |
| Google ログイン | ランナー全員が Google ユーザーを使うとは限らない。チームが使っているのは LINE |
| ログインなし | ADR-0006 で採用しないと決めた |

## Consequences（影響）

- LINE のアプリからならログインはほぼワンタップで、tasuki はパスワードを保存しない
- LINE を使わないチーム（英語圏など）は、そのままでは tasuki を使えない。認証を1つのインターフェースの裏にまとめているので、後からほかのログイン方法を足すことはできる
- 私たちが運用する1つの LINE ログインのチャネルで、すべてのチームに対応する（ADR-0005）。Admin と Tester 以外がログインするにはチャネルの公開が必要で、公開は元に戻せない
- ユーザー ID はプロバイダーごとに違うので、LINE のプロバイダーを変えると全員のユーザー ID が変わる。サービスとしてプロバイダーを変えないか、ユーザーを紐付け直す必要がある
- ローカル開発では、任意のテストユーザーで LINE ログインを使えない。ADR-0009 を参照

</details>
