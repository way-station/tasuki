# ADR-0006: Access policy — team-only visibility, roles, and organizer approval

- Status: Accepted
- Type: business
- Date: 2026-10-11
- Issue: #46

## Context

tasuki holds personal data about team members: attendance, expected paces, and — after the MVP — race results, past records, and each runner's strengths. This data accumulates over years, so who can see and change it matters more over time.

The original draft let members answer without logging in, by picking their name from a list. That is convenient, but anyone with the URL could answer as anyone else, and invite links shared in LINE can be forwarded or screenshotted outside the team.

Planned features are team-wide, not per race: searching past races, analysing the team's records, and comparing runners' strengths all need access to the whole team's history.

## Decision

**Visibility**

- Only invited, logged-in members of the team can see anything in tasuki. Nothing is public.
- Visibility is team-wide: every member can see all races, attendance, running orders, and past data of the team.

**Roles**

| Role | Can do |
| --- | --- |
| Organizer | Register and edit the races the team enters, build and confirm the running order, correct anyone's records, approve members |
| Member | Answer attendance and enter their own pace and profile for themselves only; view team data |

- Organizers are configured outside the app (an environment variable listing their LINE user IDs). There is no in-app screen to grant or remove the organizer role.

**Joining the team**

1. An organizer shares an invite link in the team's LINE group.
2. A person opens it and logs in (ADR-0007).
3. On first login, they pick their own name from the team roster.
4. An organizer approves that link between the account and the roster member (ADR-0008). Until then, the person cannot use the app.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| No login; pick a name each time (original draft) | Anyone with the URL can answer as anyone; cannot protect personal data that accumulates over time |
| Per-race visibility (only that race's runners see it) | Members must see a race before answering attendance; past-data search and team analysis need team-wide access |
| Per-member invite links instead of picking a name | Organizers must create and send a link to each person individually |
| No organizer approval | A leaked invite link or a wrong name choice links an account to someone else's records |
| Managing organizers in the app | In the single-team MVP (ADR-0005) the team has two fixed organizers; an in-app role management feature is not worth its cost and risk yet |

## Consequences

- Personal data stays inside the team, and records cannot be edited by someone pretending to be another member.
- New members wait for an organizer's approval before they can use the app; organizers need an approval screen.
- Changing organizers requires changing configuration and redeploying. When tasuki becomes a SaaS (ADR-0005), organizers must be managed per team in the database, and this ADR will be superseded for that part.
- Authorization checks are needed on every API: logged in, approved, and organizer-only where applicable.

---

<details>
<summary>日本語</summary>

# ADR-0006: アクセスの方針（チーム内だけ・役割・幹事の承認）

- Status: Accepted
- Type: business
- Date: 2026-10-11
- Issue: #46

## Context（背景）

tasuki はチームのメンバーの個人データを扱います。出欠、想定ペース、そして MVP の後には大会の結果、過去の記録、選手ごとの得意な距離です。このデータは何年も積み上がるので、誰が見られて誰が変えられるかは、時間とともに重要になります。

最初の草案では、メンバーはログインせず、一覧から名前を選んで回答する形でした。手軽ですが、URL を知っていれば誰でも他人として回答でき、LINE で共有した招待リンクは転送やスクリーンショットでチームの外に出ることがあります。

予定している機能は大会単位ではなくチーム単位です。過去の大会の検索、チームの記録の分析、選手の得意な距離の比較は、どれもチーム全体の履歴を見られる必要があります。

## Decision（決定）

**見られる範囲**

- tasuki の内容は、招待されてログインしたチームのメンバーだけが見られる。外部には一切公開しない
- 見られる範囲はチーム全体。メンバーは全員、チームのすべての大会、出欠、走順、過去データを見られる

**役割**

| 役割 | できること |
| --- | --- |
| 幹事 | チームが出場する大会の登録と編集、走順の作成と確定、全員の記録の修正、メンバーの承認 |
| メンバー | 自分の出欠、ペース、プロフィールの入力（自分の分だけ）。チームのデータの閲覧 |

- 幹事はアプリの外（LINE ユーザー ID を並べた環境変数）で決める。アプリの中に、幹事の役割を付けたり外したりする画面はない

**チームへの参加**

1. 幹事が、チームの LINE グループに招待リンクを共有する
2. 招待された人がリンクを開いてログインする（ADR-0007）
3. 初回ログインのときに、チームの名簿から自分の名前を選ぶ
4. アカウントと名簿のメンバーの紐付けを、幹事が承認する（ADR-0008）。承認されるまでアプリは使えない

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| ログインなし、毎回名前を選ぶ（最初の草案） | URL を知っていれば誰でも他人として回答できる。積み上がる個人データを守れない |
| 大会ごとに見られる人を制限する（その大会の走者だけ） | メンバーは出欠を答える前に大会を見る必要がある。過去データの検索やチームの分析には、チーム全体を見られる必要がある |
| 名前を選ぶ代わりに、メンバーごとの専用招待リンク | 幹事が1人ずつリンクを作って送る必要がある |
| 幹事の承認なし | 招待リンクが漏れたり名前を選び間違えたりすると、アカウントが他人の記録に紐付く |
| アプリの中で幹事を管理する | 1チームの MVP（ADR-0005）では幹事は2人で固定されている。アプリ内で役割を管理する機能は、今はコストとリスクに見合わない |

## Consequences（影響）

- 個人データはチームの中にとどまり、他人になりすまして記録を書き換えることはできない
- 新しいメンバーは、使い始める前に幹事の承認を待つ。幹事には承認画面が必要
- 幹事を変えるには、設定を変えて再デプロイする。tasuki を SaaS にするとき（ADR-0005）は、チームごとの幹事をデータベースで管理する必要があり、その部分はこの ADR を置き換える
- すべての API で認可のチェック（ログイン済みか、承認済みか、必要に応じて幹事か）が必要になる

</details>
