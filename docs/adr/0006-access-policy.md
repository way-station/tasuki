# ADR-0006: Access policy — team-only visibility, roles, and organizer approval

- Status: Accepted
- Type: business
- Date: 2026-10-11 (updated 2026-10-12, #61)
- Issue: #46, #61

## Context

tasuki holds personal data about team members: attendance, lap times, and average times. This data accumulates over years, so who can see and change it matters more over time.

The original draft let members answer without logging in, by picking their name from a list. That is convenient, but anyone with the URL could answer as anyone else, and invite links shared in LINE can be forwarded or screenshotted outside the team.

Organizers decide who runs in each race. They ask members directly, and only the members they invite count as participants. Race details such as the deadline are part of that decision and should not reach members who were not asked.

## Decision

**Visibility**

- Only invited, logged-in members of the team can see anything in tasuki. Nothing is public.
- Every member can see the names of all the team's races.
- A race's description can be seen only by organizers and the members an organizer invited to that race.
- The deadline inside a race's description can be seen only by organizers.
- A running order can be seen by members only after it is confirmed; drafts are visible only to organizers.
- Organizers can mark a member as left. A left member cannot log in, and other members see them as left.

**Roles**

| Role | Can do |
| --- | --- |
| Organizer | Register and edit the races the team enters, invite members to each race, build and confirm the running order (and edit it after confirming), correct anyone's records, manage the roster, create invite links, approve or reject members, mark members as left |
| Member | Answer attendance for the races they are invited to, submit their own lap times; view what is visible to them above |

- Organizers are configured outside the app (an environment variable listing their LINE user IDs). There is no in-app screen to grant or remove the organizer role.

**Joining the team**

1. An organizer creates a personal invite link for a roster member and sends it to that person directly. The link is valid for one week. Creating a new link for the same person makes the old one unusable.
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
- The API must hide the description from members who were not invited, and the deadline from everyone but organizers.
- Changing organizers requires changing configuration and redeploying. When tasuki becomes a SaaS (ADR-0005), organizers must be managed per team in the database, and this ADR will be superseded for that part.
- Authorization checks are needed on every API: logged in, approved, not left, invited to the race, and organizer-only where applicable.

---

<details>
<summary>日本語</summary>

# ADR-0006: アクセスの方針（チーム内だけ・役割・幹事の承認）

- Status: Accepted
- Type: business
- Date: 2026-10-11（2026-10-12 更新、#61）
- Issue: #46, #61

## Context（背景）

tasuki はチームのメンバーの個人データを扱います。出欠、ラップタイム、平均タイムです。このデータは何年も積み上がるので、誰が見られて誰が変えられるかは、時間とともに重要になります。

最初の草案では、メンバーはログインせず、一覧から名前を選んで回答する形でした。手軽ですが、URL を知っていれば誰でも他人として回答でき、LINE で共有した招待リンクは転送やスクリーンショットでチームの外に出ることがあります。

大会ごとに誰が走るかは、幹事が決めます。幹事がメンバーに直接声をかけ、誘われたメンバーだけが参加として数えられます。締め切りなどの大会の詳細はその判断の一部なので、誘われていないメンバーには見せません。

## Decision（決定）

**見られる範囲**

- tasuki の内容は、招待されてログインしたチームのメンバーだけが見られる。外部には一切公開しない
- チームのすべての大会の名前は、メンバー全員が見られる
- 大会の説明欄は、幹事と、幹事がその大会に誘ったメンバーだけが見られる
- 説明欄の中の締め切りは、幹事だけが見られる
- 走順は、確定した後にメンバーが見られる。下書きは幹事だけが見られる
- 幹事はメンバーを退会にできる。退会したメンバーはログインできず、ほかのメンバーには退会と表示される

**役割**

| 役割 | できること |
| --- | --- |
| 幹事 | チームが出場する大会の登録と編集、大会ごとのメンバーの招待、走順の作成と確定（確定後の修正も）、全員の記録の修正、名簿の管理、招待リンクの作成、メンバーの承認と却下、メンバーの退会 |
| メンバー | 誘われた大会の出欠の回答、自分のラップタイムの提出。上の範囲で見られるものの閲覧 |

- 幹事はアプリの外（LINE ユーザー ID を並べた環境変数）で決める。アプリの中に、幹事の役割を付けたり外したりする画面はない

**チームへの参加**

1. 幹事が名簿のメンバーごとに個人の招待リンクを作り、本人に直接送る。リンクの期限は1週間。同じ人に新しいリンクを作ると、古いリンクは使えなくなる
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
- API は、誘われていないメンバーには説明欄を、幹事以外には締め切りを返さないようにする必要がある
- 幹事を変えるには、設定を変えて再デプロイする。tasuki を SaaS にするとき（ADR-0005）は、チームごとの幹事をデータベースで管理する必要があり、その部分はこの ADR を置き換える
- すべての API で認可のチェック（ログイン済みか、承認済みか、退会していないか、その大会に誘われているか、必要に応じて幹事か）が必要になる

</details>
