# ADR-0008: Separate accounts from roster members

- Status: Proposed
- Type: technical
- Date: 2026-10-11
- Issue: #48

## Context

tasuki has two different ideas of a "person":

- **Someone who logs in** (ADR-0007): identified by a LINE user ID, with a role (ADR-0006).
- **A runner on the team roster**: has a name, expected pace, race results, and strengths. This is the data that accumulates over the years.

These do not always match one-to-one in time or in role:

- A runner can be on the roster before they ever log in (an organizer registers the roster first).
- One person can be both an organizer and a runner (both current organizers also run).
- The login method may change in the future (ADR-0007), but a runner's history must not be lost when it does.
- Linking a login to a roster member needs organizer approval (ADR-0006), so there is a state where a login exists but is not yet linked.

## Decision

Keep **accounts** and **roster members** as separate entities, linked to each other.

| Entity | Holds |
| --- | --- |
| Account | The login identity (LINE user ID), display name from LINE, organizer flag (from configuration) |
| Roster member | Name, expected pace, profile, and everything related to races (attendance, running order, results) |
| Link | Which account belongs to which roster member, and whether an organizer has approved it |

- Race data (attendance, running order, results) always refers to the **roster member**, never to the account.
- A roster member has at most one approved account, and an account belongs to at most one roster member.

## Alternatives considered

| Option | Why it was not chosen |
| --- | --- |
| A single users table holding login and runner data | Runners cannot exist before logging in; changing the login method would touch runner history; the pending-approval state has no natural place |
| Roster members only, with a LINE user ID column | Mixes authentication data into the runner record; an organizer who is not on the roster has nowhere to live |

## Consequences

- Runner history is independent of how people log in.
- Organizers can prepare the roster before anyone logs in.
- A linking step with an approval state is needed, and screens that show "who am I" must join the account and the roster member.
- Uniqueness of the link (one account ↔ one roster member) should be enforced by database constraints.

---

<details>
<summary>日本語</summary>

# ADR-0008: アカウントと名簿のメンバーを分ける

- Status: Proposed
- Type: technical
- Date: 2026-10-11
- Issue: #48

## Context（背景）

tasuki には、「人」に当たるものが2種類あります。

- **ログインする人**（ADR-0007）：LINE のユーザー ID で識別され、役割を持つ（ADR-0006）
- **チームの名簿の選手**：名前、想定ペース、大会の結果、得意な距離を持つ。何年も積み上がるのはこちらのデータ

この2つは、時期の面でも役割の面でも、いつも1対1で対応するとは限りません。

- ログインする前から、名簿に選手がいることがある（幹事が先に名簿を登録する）
- 1人が幹事と選手を兼ねることがある（今の幹事2人はどちらも走る）
- ログインの方法は将来変わるかもしれない（ADR-0007）が、そのときに選手の履歴を失ってはいけない
- ログインと名簿のメンバーの紐付けには幹事の承認が必要（ADR-0006）なので、ログインはしているがまだ紐付いていない状態がある

## Decision（決定）

**アカウント**と**名簿のメンバー**を別のものとして持ち、互いに紐付ける。

| もの | 持つ情報 |
| --- | --- |
| アカウント | ログインの識別子（LINE のユーザー ID）、LINE の表示名、幹事かどうか（設定から） |
| 名簿のメンバー | 名前、想定ペース、プロフィール、大会に関わるすべて（出欠、走順、結果） |
| 紐付け | どのアカウントがどの名簿のメンバーか、幹事が承認したかどうか |

- 大会のデータ（出欠、走順、結果）は、常に**名簿のメンバー**を参照し、アカウントは参照しない
- 名簿のメンバーには、承認済みのアカウントが最大1つ。アカウントが属する名簿のメンバーも最大1人

## Alternatives considered（検討した他の案）

| 案 | 採用しなかった理由 |
| --- | --- |
| ログインと選手のデータを1つのユーザーテーブルに持つ | ログイン前の選手を登録できない。ログイン方法を変えると選手の履歴に影響する。承認待ちの状態を置く自然な場所がない |
| 名簿のメンバーだけにして、LINE のユーザー ID の列を持たせる | 認証のデータが選手の記録に混ざる。名簿にいない幹事を置く場所がない |

## Consequences（影響）

- 選手の履歴が、ログインの方法から独立する
- 誰もログインしていないうちから、幹事が名簿を用意できる
- 承認の状態を持つ紐付けの手順が必要になる。「自分は誰か」を表示する画面では、アカウントと名簿のメンバーを結合する必要がある
- 紐付けの一意性（1つのアカウント ↔ 1人の名簿のメンバー）は、データベースの制約で守る

</details>
