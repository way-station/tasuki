# ADR-0009: Dev-only login for local development

- Status: Proposed
- Type: technical
- Date: 2026-10-11
- Issue: #49

## Context

With LINE Login (ADR-0007) and the access policy (ADR-0006), almost every screen and API requires a logged-in, approved user, and many require an organizer.

During local development, real LINE Login is impractical:

- Developers need to try the app as different people: an organizer, an approved member, a member waiting for approval, someone not yet linked.
- A published production LINE Login channel would give developers real accounts with real permissions. The developer of tasuki is one of the production organizers only because they build the app, and should not need production organizer rights to develop.
- A separate development channel would only allow its Admins and Testers to log in, and still gives one fixed identity per LINE account.

## Decision

Local development uses a **dev-only login** that signs in as any user without LINE.

- It is enabled only by a dedicated flag (for example `DEV_LOGIN_ENABLED=true`). The default is off.
- The flag is set only in `infra/compose.dev.yaml` (the Dev Container). Production configuration never contains it.
- When it is on, the backend logs a clear warning at startup.
- A test asserts that the dev login is off when the flag is not set, and CI runs it.
- After a dev login, the session and authorization checks are the same as after a LINE login. Only the step that proves the identity is replaced.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| A separate LINE Login channel for development | Only its Admins / Testers can log in; one identity per LINE account; every contributor needs a LINE Developers setup |
| Using production accounts | Developers would need production organizer rights; risk of changing real data |
| Mocking authentication only in automated tests | Does not help when trying the app by hand in the browser |

## Consequences

- Developers and contributors can work on every screen and permission case without a LINE account or production rights.
- If the dev login were ever enabled in production, anyone could log in as an organizer. This risk is contained by the default-off flag, keeping the flag out of production configuration, the startup warning, and the test in CI.
- The real LINE Login flow still needs to be checked separately (for example on a staging deployment) before release.

---

<details>
<summary>日本語</summary>

# ADR-0009: ローカル開発専用の仮ログイン

- Status: Proposed
- Type: technical
- Date: 2026-10-11
- Issue: #49

## Context（背景）

LINE ログイン（ADR-0007）とアクセスの方針（ADR-0006）により、ほとんどの画面と API はログイン済みで承認済みのユーザーを必要とし、その多くは幹事であることを必要とします。

ローカル開発で本物の LINE ログインを使うのは現実的ではありません。

- 開発者は、さまざまな立場でアプリを試す必要がある。幹事、承認済みのメンバー、承認待ちのメンバー、まだ紐付いていない人など
- 公開済みの本番の LINE ログインチャネルを使うと、開発者が本物の権限を持つ本物のアカウントで動くことになる。tasuki の開発者が本番の幹事の1人なのは、アプリを作っているからにすぎず、開発のために本番の幹事権限が必要であってはならない
- 開発用のチャネルを別に作っても、ログインできるのはその Admin と Tester だけで、LINE アカウント1つにつき1つの立場しか試せない

## Decision（決定）

ローカル開発では、LINE を使わずに任意のユーザーとしてログインできる**開発専用の仮ログイン**を使う。

- 有効にするには専用のフラグ（例：`DEV_LOGIN_ENABLED=true`）が必要で、デフォルトは無効
- フラグを設定するのは `infra/compose.dev.yaml`（Dev Container）だけ。本番の設定には書かない
- 有効なときは、backend が起動時に目立つ警告をログに出す
- フラグがないときに仮ログインが無効であることをテストで確かめ、CI で実行する
- 仮ログインの後のセッションと認可のチェックは、LINE ログインの後と同じ。置き換えるのは本人確認の部分だけ

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| 開発用の LINE ログインチャネルを別に作る | ログインできるのはその Admin / Tester だけ。LINE アカウント1つにつき1つの立場しか試せない。コントリビューター全員に LINE Developers の準備が必要 |
| 本番のアカウントを使う | 開発者に本番の幹事権限が必要になる。本物のデータを変えてしまう危険がある |
| 自動テストの中だけ認証をモックする | ブラウザで手動で試すときの役に立たない |

## Consequences（影響）

- 開発者もコントリビューターも、LINE アカウントや本番の権限なしで、すべての画面と権限のパターンを試せる
- 仮ログインが本番で有効になれば、誰でも幹事としてログインできてしまう。この危険は、デフォルトで無効なフラグ、本番の設定に書かないこと、起動時の警告、CI のテストで抑える
- 本物の LINE ログインの流れは、リリース前に別途（例えばステージング環境で）確認する必要がある

</details>
