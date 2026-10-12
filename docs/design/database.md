# MVP database design

- Status: Draft
- Issue: #60
- Related: ADR-0002 (PostgreSQL and MongoDB), ADR-0005 (multi-tenant data), ADR-0006 (access policy), ADR-0008 (accounts and roster members)

This document describes the data tasuki's MVP stores and the rules it follows. Column types, keys, and indexes are decided by the maintainer before implementation.

## Business rules

### Races

- An organizer registers a race with a **name** and a **description**.
- The description is one field that holds text, lists, and numbers: free text, the distance of each leg (km), the distance of one lap (km), the time limit (minutes), the deadline, and the target time.
  - Units are fixed: distances in km (e.g. `30.5`), time limits in minutes (e.g. `180`).
- Races are distance-based or time-based.
- Races are not deleted or cancelled in the app. Cancellations are announced on LINE.

### Invitations to a race and attendance

- An organizer invites members to each race in the app.
- Only invited members answer attendance: **attending**, **not attending**, or **undecided**. They can change it at any time.
- Attendance is per race. Later, an organizer splits the attending members into entries (for example team A and team B). The MVP has one entry per race, but the tables allow several.

### Running order

- A running order is a list of legs: which roster member runs which leg. There are no constraints on who runs or how often.
- A running order is a draft until an organizer confirms it. Members see it only after it is confirmed.
- After it is confirmed, only organizers can change it.
- Every change is kept as history.

### Lap times and average time

- Members submit their own lap times per race, as time per km (`MM:SS`).
- Each roster member has an **average time** (per km) and a **lap count**.

| Laps submitted | Average time |
| --- | --- |
| 0 | `3:40` (default; not counted as a lap) |
| 1 | That lap |
| 2 or more | The mean of all laps |

- When new laps are submitted, the average is recalculated from the stored average and lap count:
  `new average = (average × lap count + sum of new laps) ÷ (lap count + number of new laps)`
  - Example: 5 laps averaging `3:30`, then `3:20`, `3:25`, `3:25` → (3:30 × 5 + 3:20 + 3:25 + 3:25) ÷ 8 = 3:27.5 → `3:28`
- Times are stored and shown as `MM:SS`. Fractions of a second are rounded half up.
- The profile shows the average time. Members enter nothing in the profile.

### Predictions

| Race | Prediction | Calculation |
| --- | --- | --- |
| Distance-based | Total time | For each runner: average time × the total distance of the legs they run. Sum over all runners |
| Time-based | Number of laps | Time limit ÷ the mean of the average times of the runners in the running order = distance (km). Distance ÷ lap distance, rounded down to whole laps |

- Example (time-based): 180 min ÷ `3:30`/km ≈ 51.4 km; with 2 km laps → 25 laps.
- Predictions are estimates and may differ from the result.

### Roster and members

- Only organizers add people to the roster. A roster member has a name only (plus the average time and lap count above).
- Organizers can mark a roster member as left. Left members cannot log in and are shown as left. Their data is kept.

### Invite links (joining the team)

- An organizer creates a personal invite link for a roster member. It is valid for one week.
- Creating a new link for the same roster member makes the old one unusable.
- The person logs in with LINE through the link, and an organizer approves or rejects the link between the account and the roster member. Rejections are told outside the app. Applying again is low priority.

### Who sees what

| Data | Organizers | Invited members | Other members |
| --- | --- | --- | --- |
| Race name | Yes | Yes | Yes |
| Race description (except the deadline) | Yes | Yes | No |
| Deadline (inside the description) | Yes | No | No |
| Running order draft | Yes | No | No |
| Confirmed running order | Yes | Yes | Yes |

The API removes the parts a user may not see. See ADR-0006.

## PostgreSQL tables

Every team-owned table belongs to a team (ADR-0005).

| Table | Main fields |
| --- | --- |
| `teams` | The team (tenant) |
| `accounts` | LINE user ID, display name |
| `roster_members` | Team, name, left or not, average time (`MM:SS`), lap count |
| `invitations` | Roster member, token, expiry (one week), when it was made unusable |
| `member_links` | Account, roster member, status (pending / approved / rejected) |
| `races` | Team, name, description (JSON: text, leg distances in km, lap distance in km, time limit in minutes, deadline, target time) |
| `race_invitations` | Race, roster member (invited) |
| `attendances` | Race, roster member, status (attending / not attending / undecided) |
| `entries` | Race, entry name (one per race in the MVP) |
| `entry_members` | Entry, roster member (how attending members are split into entries) |
| `running_orders` | Entry, status (draft / confirmed) |
| `running_order_legs` | Running order, leg number, roster member |

## MongoDB collections

MongoDB refers to PostgreSQL IDs and does not copy data (ADR-0002).

| Collection | Contents |
| --- | --- |
| `lap_records` | Race, roster member, lap times per km (`MM:SS`) |
| `running_order_history` | A snapshot of the running order at each change, who changed it, when |

The average time and lap count are written back to `roster_members` in PostgreSQL when laps are submitted (ADR-0002).

## Left to the maintainer

- Column types, primary and foreign keys, unique constraints, and indexes
- ID type
- How `MM:SS` and the JSON description are stored
- Whether `entry_members` is needed in the MVP (the MVP has one entry per race)

---

<details>
<summary>日本語</summary>

# MVP のデータベースの設計

- Status: Draft
- Issue: #60
- 関連：ADR-0002（PostgreSQL と MongoDB）、ADR-0005（マルチテナントのデータ）、ADR-0006（アクセスの方針）、ADR-0008（アカウントと名簿のメンバー）

tasuki の MVP が保存するデータと、そのルールを書く。列の型、キー、インデックスは、実装の前にメンテナーが決める。

## 業務のルール

### 大会

- 幹事が、**大会名**と**説明欄**で大会を登録する
- 説明欄は1つの欄で、文章、配列、数値が入る。自由な文章、区ごとの距離（km）、1周の距離（km）、制限時間（分）、締め切り、目標タイム
  - 単位は固定。距離は km（例：`30.5`）、制限時間は分（例：`180`）
- 大会には距離制と時間制がある
- 大会の削除や中止はアプリでは行わない。中止は LINE で連絡する

### 大会への招待と出欠

- 幹事が、アプリの中で大会ごとにメンバーを誘う
- 出欠を答えられるのは誘われたメンバーだけ。**参加**、**不参加**、**保留**のどれか。いつでも変えられる
- 出欠は大会ごと。その後、幹事が参加するメンバーをエントリー（例：A チームと B チーム）に分ける。MVP ではエントリーは大会ごとに1つだが、テーブルは複数に対応する

### 走順

- 走順は区の並びで、どの区を名簿の誰が走るか。誰が走るか、何回走るかに制約はない
- 幹事が確定するまで、走順は下書き。メンバーは確定した後に見られる
- 確定した後に直せるのは幹事だけ
- 変更はすべて履歴に残す

### ラップタイムと平均タイム

- メンバーは、大会ごとに自分のラップタイムを 1km あたりのタイム（`MM:SS`）で提出する
- 名簿のメンバーは、**平均タイム**（1km あたり）と**ラップ数**を持つ

| 提出したラップの数 | 平均タイム |
| --- | --- |
| 0回 | `3:40`（デフォルト値。ラップとしては数えない） |
| 1回 | そのラップ |
| 2回以上 | すべてのラップの平均 |

- 新しいラップが提出されたら、保存している平均タイムとラップ数から計算し直す
  `新しい平均 = (平均 × ラップ数 + 新しいラップの合計) ÷ (ラップ数 + 新しいラップの数)`
  - 例：5回で平均 `3:30` の人が `3:20`、`3:25`、`3:25` を出した → (3:30 × 5 + 3:20 + 3:25 + 3:25) ÷ 8 = 3:27.5 → `3:28`
- タイムは `MM:SS` で保存し、表示する。秒の端数は四捨五入する
- プロフィールには平均タイムを表示する。メンバーがプロフィールに入力するものはない

### 予想

| 大会 | 予想するもの | 計算 |
| --- | --- | --- |
| 距離制 | 合計タイム | 走る人ごとに、平均タイム × その人が走る区の距離の合計。全員分を足す |
| 時間制 | 周回数 | 制限時間 ÷ 走順に入っている人の平均タイムの平均 = 距離（km）。距離 ÷ 1周の距離を、切り捨てて整数の周回数にする |

- 例（時間制）：180分 ÷ `3:30`/km ≈ 51.4km。1周 2km なら 25周
- 予想なので、結果と前後することがある

### 名簿とメンバー

- 名簿に人を追加するのは幹事だけ。名簿のメンバーが持つのは名前だけ（と、上の平均タイムとラップ数）
- 幹事は名簿のメンバーを退会にできる。退会したメンバーはログインできず、退会と表示される。データは残す

### 招待リンク（チームへの参加）

- 幹事が、名簿のメンバーごとに個人の招待リンクを作る。期限は1週間
- 同じメンバーに新しいリンクを作ると、古いリンクは使えなくなる
- 本人がリンクから LINE でログインし、アカウントと名簿のメンバーの紐付けを幹事が承認または却下する。却下はアプリの外で伝える。再申請は優先度が低い

### 誰が何を見られるか

| データ | 幹事 | 誘われたメンバー | ほかのメンバー |
| --- | --- | --- | --- |
| 大会名 | 見られる | 見られる | 見られる |
| 大会の説明欄（締め切りを除く） | 見られる | 見られる | 見られない |
| 締め切り（説明欄の中） | 見られる | 見られない | 見られない |
| 走順の下書き | 見られる | 見られない | 見られない |
| 確定した走順 | 見られる | 見られる | 見られる |

見てはいけない部分は API が取り除く。ADR-0006 を参照。

## PostgreSQL のテーブル

チームが持つテーブルは、すべてチームに属する（ADR-0005）。

| テーブル | 主な項目 |
| --- | --- |
| `teams` | チーム（テナント） |
| `accounts` | LINE のユーザー ID、表示名 |
| `roster_members` | チーム、名前、退会したかどうか、平均タイム（`MM:SS`）、ラップ数 |
| `invitations` | 名簿のメンバー、トークン、期限（1週間）、使えなくした日時 |
| `member_links` | アカウント、名簿のメンバー、状態（承認待ち／承認／却下） |
| `races` | チーム、大会名、説明欄（JSON：文章、区ごとの距離 km、1周の距離 km、制限時間 分、締め切り、目標タイム） |
| `race_invitations` | 大会、名簿のメンバー（誘われた人） |
| `attendances` | 大会、名簿のメンバー、状態（参加／不参加／保留） |
| `entries` | 大会、エントリー名（MVP では大会ごとに1つ） |
| `entry_members` | エントリー、名簿のメンバー（参加するメンバーをエントリーに分ける） |
| `running_orders` | エントリー、状態（下書き／確定） |
| `running_order_legs` | 走順、区の番号、名簿のメンバー |

## MongoDB のコレクション

MongoDB は PostgreSQL の ID を参照するだけで、データを複製しない（ADR-0002）。

| コレクション | 中身 |
| --- | --- |
| `lap_records` | 大会、名簿のメンバー、1km あたりのラップタイム（`MM:SS`） |
| `running_order_history` | 変更のたびの走順のスナップショット、変更した人、日時 |

ラップが提出されたら、平均タイムとラップ数を PostgreSQL の `roster_members` に書き戻す（ADR-0002）。

## メンテナーが決めること

- 列の型、主キーと外部キー、一意制約、インデックス
- ID の型
- `MM:SS` と JSON の説明欄の保存方法
- MVP で `entry_members` が必要か（MVP ではエントリーは大会ごとに1つ）

</details>
