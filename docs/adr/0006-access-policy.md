# ADR-0006: Access policy — team-only visibility, roles, and organizer approval

- Status: Accepted
- Type: business
- Date: 2026-10-11 (updated 2026-10-12, #61, #65)
- Issue: #46, #61, #65

## Context

tasuki holds personal data about team members: attendance, lap times, and average times. This data accumulates over years, so who can see and change it matters more over time.

The original draft let members answer without logging in, by picking their name from a list. That is convenient, but anyone with the URL could answer as anyone else, and invite links shared in LINE can be forwarded or screenshotted outside the team.

Organizers decide who runs in each race. They ask members directly, and only the members they invite count as participants. Race details such as the deadline are part of that decision and should not reach members who were not asked.

## Decision

**Visibility**

- Only invited, logged-in members of the team can see anything in tasuki. Nothing is public.
- Members see only the races they are invited to. A member who answers "not attending" no longer sees that race, and cannot change the answer.
- A race's description can be seen only by organizers and the members an organizer invited to that race.
- The deadline inside a race's description can be seen only by organizers.
- Attendance answers can be seen only by organizers. Members see only their own answer.
- A running order and its prediction can be seen by members only after it is confirmed, and only by members who answered "attending"; drafts are visible only to organizers.
- A member's lap times and profile (average time) can be seen by that member and by organizers. Only the member can change their lap times.
- Organizers can mark a member as left. A left member cannot log in, and other members see them as left.

**Roles**

| Role | Can do |
| --- | --- |
| Organizer | Register and edit the races the team enters, invite members to each race (and cancel invitations), build and confirm the running order (and edit it after confirming), manage the roster, create invite links, approve or reject members, mark members as left |
| Member | Answer attendance for the races they are invited to, submit, edit, and delete their own lap times; view what is visible to them above |

- Organizers do not change other people's records (attendance or lap times).

- Organizers are configured outside the app (an environment variable listing their LINE user IDs). There is no in-app screen to grant or remove the organizer role.

**Joining the team**

1. An organizer creates a personal invite link for a roster member and sends it to that person directly. The link is valid for one week and can be used once. Creating a new link for the same person makes the old one unusable.
2. The person opens it and logs in (ADR-0007). The link identifies the roster member, so the person does not pick a name.
3. An organizer approves the link between the account and the roster member (ADR-0008). Until then, the person cannot use the app.
4. If an organizer rejects it, the organizer tells the person outside the app.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| No login; pick a name each time (original draft) | Anyone with the URL can answer as anyone; cannot protect personal data that accumulates over time |
| Every member sees every race's details | Members who were not asked would see details, including the deadline, meant only for the members the organizers chose |
| Inviting members to a race only on LINE, outside the app | The app cannot know who was invited, so it cannot limit who sees the description or who answers attendance |
| One invite link posted to the team's LINE group, then picking a name | The link can be forwarded, and picking the wrong name links an account to someone else's records |
| No organizer approval | A leaked invite link links an account to someone else's records |
| Managing organizers in the app | In the single-team MVP (ADR-0005) the team has two fixed organizers; an in-app role management feature is not worth its cost and risk yet |

## Consequences

- Personal data stays inside the team, and records cannot be edited by someone pretending to be another member.
- Organizers do more work: inviting members to each race and creating one invite link per person, again when a link expires.
- New members wait for an organizer's approval before they can use the app; organizers need an approval screen.
- The API must hide races from members who were not invited or answered "not attending", the description from members who were not invited, the deadline from everyone but organizers, and the running order and prediction from members who did not answer "attending".
- Changing organizers requires changing configuration and redeploying. When tasuki becomes a SaaS (ADR-0005), organizers must be managed per team in the database, and this ADR will be superseded for that part.
- Authorization checks are needed on every API: logged in, approved, not left, invited to the race, the member's attendance, and organizer-only where applicable.
- A member who answers "not attending" by mistake needs an organizer to cancel and re-send the invitation.

---

<details>
<summary>日本語</summary>

# ADR-0006: アクセスの方針（チーム内だけ・役割・幹事の承認）

- Status: Accepted
- Type: business
- Date: 2026-10-11（2026-10-12 更新、#61、#65）
- Issue: #46, #61, #65

## Context（背景）

tasuki はチームのメンバーの個人データを扱います。出欠、ラップタイム、平均タイムです。このデータは何年も積み上がるので、誰が見られて誰が変えられるかは、時間とともに重要になります。

最初の草案では、メンバーはログインせず、一覧から名前を選んで回答する形でした。手軽ですが、URL を知っていれば誰でも他人として回答でき、LINE で共有した招待リンクは転送やスクリーンショットでチームの外に出ることがあります。

大会ごとに誰が走るかは、幹事が決めます。幹事がメンバーに直接声をかけ、誘われたメンバーだけが参加として数えられます。締め切りなどの大会の詳細はその判断の一部なので、誘われていないメンバーには見せません。

## Decision（決定）

**見られる範囲**

- tasuki の内容は、招待されてログインしたチームのメンバーだけが見られる。外部には一切公開しない
- メンバーが見られるのは、誘われた大会だけ。「不参加」と答えたメンバーは、その大会を見られなくなり、回答も変えられない
- 大会の説明欄は、幹事と、幹事がその大会に誘ったメンバーだけが見られる
- 説明欄の中の締め切りは、幹事だけが見られる
- 出欠の回答は、幹事だけが見られる。メンバーが見られるのは自分の回答だけ
- 走順とその予想は、確定した後に、「参加」と答えたメンバーだけが見られる。下書きは幹事だけが見られる
- メンバーのラップタイムとプロフィール（平均タイム）は、本人と幹事が見られる。ラップタイムを変えられるのは本人だけ
- 幹事はメンバーを退会にできる。退会したメンバーはログインできず、ほかのメンバーには退会と表示される

**役割**

| 役割 | できること |
| --- | --- |
| 幹事 | チームが出場する大会の登録と編集、大会ごとのメンバーの招待（と取り消し）、走順の作成と確定（確定後の修正も）、名簿の管理、招待リンクの作成、メンバーの承認と却下、メンバーの退会 |
| メンバー | 誘われた大会の出欠の回答、自分のラップタイムの提出・修正・削除。上の範囲で見られるものの閲覧 |

- 幹事は、ほかの人の記録（出欠、ラップタイム）を変えない

- 幹事はアプリの外（LINE ユーザー ID を並べた環境変数）で決める。アプリの中に、幹事の役割を付けたり外したりする画面はない

**チームへの参加**

1. 幹事が名簿のメンバーごとに個人の招待リンクを作り、本人に直接送る。リンクの期限は1週間で、使えるのは1回だけ。同じ人に新しいリンクを作ると、古いリンクは使えなくなる
2. 本人がリンクを開いてログインする（ADR-0007）。リンクで名簿の誰かが決まっているので、名前は選ばない
3. アカウントと名簿のメンバーの紐付けを、幹事が承認する（ADR-0008）。承認されるまでアプリは使えない
4. 幹事が却下した場合は、幹事がアプリの外で本人に伝える

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| ログインなし、毎回名前を選ぶ（最初の草案） | URL を知っていれば誰でも他人として回答できる。積み上がる個人データを守れない |
| メンバー全員がすべての大会の詳細を見られる | 誘われていないメンバーにも、幹事が選んだメンバーだけに向けた詳細（締め切りを含む）が見えてしまう |
| 大会への招待は LINE だけで行い、アプリの外にとどめる | アプリが誰を誘ったかを知らないので、説明欄を見られる人も、出欠を答えられる人も制限できない |
| チームの LINE グループに1つの招待リンクを流し、名前を選ぶ | リンクが転送されることがあり、名前を選び間違えると、アカウントが他人の記録に紐付く |
| 幹事の承認なし | 招待リンクが漏れると、アカウントが他人の記録に紐付く |
| アプリの中で幹事を管理する | 1チームの MVP（ADR-0005）では幹事は2人で固定されている。アプリ内で役割を管理する機能は、今はコストとリスクに見合わない |

## Consequences（影響）

- 個人データはチームの中にとどまり、他人になりすまして記録を書き換えることはできない
- 幹事の作業が増える。大会ごとのメンバーの招待と、1人ずつの招待リンクの作成（期限が切れたら作り直し）
- 新しいメンバーは、使い始める前に幹事の承認を待つ。幹事には承認画面が必要
- API は、誘われていないか「不参加」と答えたメンバーには大会を、誘われていないメンバーには説明欄を、幹事以外には締め切りを、「参加」と答えていないメンバーには走順と予想を返さないようにする必要がある
- 幹事を変えるには、設定を変えて再デプロイする。tasuki を SaaS にするとき（ADR-0005）は、チームごとの幹事をデータベースで管理する必要があり、その部分はこの ADR を置き換える
- すべての API で認可のチェック（ログイン済みか、承認済みか、退会していないか、その大会に誘われているか、そのメンバーの出欠、必要に応じて幹事か）が必要になる
- 間違えて「不参加」と答えたメンバーは、幹事に招待を取り消して誘い直してもらう必要がある

</details>
