# MVP screen design

- Status: Draft
- Issue: #64 (ADR-0006 and the database design were aligned in #65)
- Related: ADR-0006 (access policy), ADR-0007 (LINE Login), ADR-0008 (accounts and roster members), ADR-0010 (spec-first OpenAPI), `docs/design/database.md`

This document lists tasuki's MVP screens: what each screen shows, what it lets the user do, and who can open it. Visual design is out of scope. The API is designed from this document (ADR-0010).

## Who uses the screens

| User | Who they are |
| --- | --- |
| Organizer | An account listed as an organizer in configuration (ADR-0006). Can open organizer screens even before their own account is linked to a roster member |
| Member | An account whose link to a roster member is approved, and who has not left |
| Waiting | An account that logged in through an invite link and is waiting for approval |
| Rejected | An account whose link was rejected |

An organizer who also runs links their account the same way as everyone else (invite link → LINE login → approval) and may approve their own link. Once linked, an organizer also has every member screen (their own attendance, lap times, and profile).

## Screen list

| ID | Screen | Organizer | Member | Waiting / Rejected / Logged out |
| --- | --- | --- | --- | --- |
| A1 | Invite link landing | — | — | Yes |
| A2 | Waiting for approval | — | — | Waiting |
| A3 | Rejected | — | — | Rejected |
| A4 | Cannot use tasuki | — | — | Left members, invalid links |
| A5 | Log in | — | — | Logged out |
| R1 | Race list (home) | Yes | Yes | — |
| R2 | Race detail | Yes | Yes (parts hidden) | — |
| R3 | Register / edit race | Yes | — | — |
| R4 | Race invitations | Yes | — | — |
| R5 | Attendance list | Yes | — | — |
| R6 | Entry | Yes | — | — |
| R7 | Running order editor | Yes | — | — |
| L1 | My lap times | Yes (if linked) | Yes | — |
| P1 | My profile | Yes (if linked) | Yes | — |
| M1 | Roster | Yes | — | — |
| M2 | Member profile | Yes | — | — |
| M3 | Approvals | Yes | — | — |

## Joining and access screens

### A1 Invite link landing

- Opened from a personal invite link.
- Shows: the team name and a "Log in with LINE" button. It does not show the roster member's name.
- Action: log in with LINE. After login, a pending link between the account and the roster member is created, the invite link becomes unusable, and the user goes to A2.
- If the link is expired, already used, or replaced by a newer link, the user goes to A4.

### A2 Waiting for approval

- Shows: "Waiting for an organizer to approve you." No team data.

### A3 Rejected

- Shows: "You were rejected." No team data. The reason is told outside the app.

### A4 Cannot use tasuki

- Shown to left members who log in, and for unusable invite links (expired, used, or replaced).
- Shows: that tasuki cannot be used, and to contact an organizer.

### A5 Log in

- For people who already joined and are logged out, for example after the session expired. Opening any team screen without a session sends the user here (#113).
- Shows: a "Log in with LINE" button.
- Action: log in with LINE. After login, the user returns to the screen they tried to open (ADR-0013).
- If LINE Login fails or is cancelled, this screen shows an error and the button again.

### Who may open which screen

- A screen opened by someone who may not open it, from the table above, sends them on instead of showing an error: logged out → A5; waiting → A2; rejected → A3; no link (non-organizer) → A4; a member on an organizer screen, or an organizer without a link on a member screen (L1, P1) → R1.

## Race screens

### R1 Race list (home)

- The first screen after login. All lists are newest registered first.
- Organizers: every race of the team, with its name, race date, and whether the running order is confirmed.
- Members see only their own races, in two sections:

| Section | Races | Shows |
| --- | --- | --- |
| My races | Races they answered "attending", including past races | Race name, race date, a "running order confirmed" mark once it is confirmed |
| Invitations | Races they are invited to and have not answered, or answered "undecided" | Race name, race date, my answer |

- A member never sees races they are not invited to, or races they answered "not attending".
- An organizer who is linked to a roster member also sees their own invitations in the Invitations section.
- Action: open a race (R2). Organizers: register a race (R3).

### R2 Race detail

Who can open it: organizers, and invited members whose answer is not "not attending".

| Item | Organizer | Member: not answered / undecided | Member: attending |
| --- | --- | --- | --- |
| Race name | Yes | Yes | Yes |
| Description: type (distance / time), date, free text, leg distances, lap distance, time limit, target time | Yes | Yes | Yes |
| Deadline | Yes | No | No |
| My attendance (answer / change) | If invited | Yes | Yes |
| Running order draft | Yes | No | No |
| Confirmed running order | Yes | No | Yes |
| Prediction (total time or laps) | Yes, draft and confirmed | No | After confirmed |
| Lap times of runners in this race | Yes (view only) | No | Own only, via L1 |

- Attendance: the member picks attending, not attending, or undecided.
  - Between "attending" and "undecided" they can switch at any time.
  - "Not attending" is final for the member: the race disappears from their screens and they cannot change the answer. The screen asks for confirmation first.
  - A member in the entry who changes from "attending" is removed from the entry and the running order. If the running order is confirmed, their lap times for this race are deleted, their average time is recalculated, and their leg becomes empty. The screen warns about this first.
- A confirmed running order shows a leg with nobody in it as "not decided". The prediction then leaves those legs out and says that some legs are empty.
- Organizer actions: edit the race (R3), invitations (R4), attendance list (R5), entry (R6), running order (R7).
- Races cannot be deleted or cancelled in the app.

### R3 Register / edit race (organizers)

- Inputs: name, type (distance / time), race date, deadline, target time, free text.
  - Distance-based: the distance of each leg in km (add or remove legs).
  - Time-based: time limit in minutes and the distance of one lap in km.
- Action: save.

### R4 Race invitations (organizers)

- Shows the roster members who have not left, and whether each one is invited.
- Actions:
  - Invite a member.
  - Cancel an invitation. The member is also removed from attendance, the entry, and the running order. If the running order is confirmed, the member's lap times for this race are deleted and their average time is recalculated. The screen warns about this before cancelling.

### R5 Attendance list (organizers)

- Shows each invited member and their answer: attending, not attending, undecided, or not answered yet.
- View only. Organizers cannot answer for a member.

### R6 Entry (organizers)

- An organizer creates the entry with a name (for example "Team A"). The MVP allows one entry per race.
- Shows the entry's members. A member who changes their answer from "attending" leaves the entry automatically (see R2).
- Actions:
  - Add members whose attendance is "attending".
  - Remove a member. The member is also removed from the running order. If the running order is confirmed, their lap times for this race are deleted and their average time is recalculated.

### R7 Running order editor (organizers)

- Shows the legs in order with the runner of each leg, the status (draft / confirmed), and the prediction.
- Legs:
  - Distance-based: one leg per leg distance in the description.
  - Time-based: the organizer adds and removes legs freely.
- Runners are picked from the entry's members. One member can run any number of legs.
- Actions:
  - Assign or clear the runner of a leg, and save.
  - Confirm. Only when every leg has a runner. After confirming, members can see it.
  - Edit after confirming. Changes are saved and shown to members right away; the order stays confirmed. A member removed from the order loses their lap times for this race, and their average time is recalculated.
- Every change is saved as history. The MVP has no screen to view the history.

## Lap times and profile

### L1 My lap times

- Opened from R2, for races where the user is a runner in the confirmed running order.
- Shows the user's lap times for that race (per km, `MM:SS`).
- Actions: add lap times, edit one, delete one. The average time is recalculated each time.
- Only the user can change their own lap times. Organizers can see them but not change them.

### P1 My profile

- Shows: name, average time (`MM:SS`), lap count.
- Members enter nothing here.

## Roster and approvals (organizers)

### M1 Roster

- Shows each roster member: name, left or not, link status (no account / waiting / approved), and the invite link's expiry if one is active.
- Actions:
  - Add a roster member (name only).
  - Create an invite link and copy it to send directly. Not available for members who already have an approved account or who have left. Creating a new link makes the old one unusable.
  - Mark as left. Invitations to races whose date is today or later are cancelled automatically, with the same effects as cancelling on R4. Past races are not changed.
  - Undo left. Cancelled invitations do not come back.
  - Unlink the approved account (for example when someone changes their LINE account).
  - Open a member's profile (M2).
- Names cannot be changed.

### M2 Member profile

- Shows: name, average time, lap count, and lap times per race.
- View only.

### M3 Approvals

- Shows waiting links: LINE display name, the roster member's name, and when they logged in.
- Actions: approve or reject. Rejections are told outside the app.

## Main flows

### Joining the team

1. Organizer adds the person to the roster (M1).
2. Organizer creates an invite link (M1) and sends it directly.
3. The person opens it (A1), logs in with LINE, and waits (A2).
4. Organizer approves (M3). The person can now use tasuki (R1). If rejected, the person sees A3.

### Inviting members and answering attendance

1. Organizer registers the race (R3).
2. Organizer invites members (R4).
3. Invited members find the race in the Invitations section (R1), read the description on R2, and answer attendance. Races they attend move to "My races".
4. Organizer checks the answers (R5).

### Building and confirming the running order

1. Organizer creates the entry and adds attending members (R6).
2. Organizer assigns runners to legs and checks the prediction (R7).
3. Organizer confirms (R7). Members who answered "attending" see the running order and the prediction on R2.

### Submitting lap times

1. After the race, a runner opens the race (R2) and then L1.
2. The runner enters lap times per km. The average time and lap count on their profile (P1) are updated.

---

<details>
<summary>日本語</summary>

# MVP の画面の設計

- Status: Draft
- Issue: #64（ADR-0006 とデータベースの設計は #65 で合わせた）
- 関連：ADR-0006（アクセスの方針）、ADR-0007（LINE ログイン）、ADR-0008（アカウントと名簿のメンバー）、ADR-0010（仕様を先に書く OpenAPI）、`docs/design/database.md`

tasuki の MVP の画面を一覧にする。各画面に出すもの、できる操作、誰が開けるか。見た目のデザインは範囲外。API はこの文書から設計する（ADR-0010）。

## 画面を使う人

| 人 | どんな人か |
| --- | --- |
| 幹事 | 設定で幹事として並べられたアカウント（ADR-0006）。自分のアカウントが名簿のメンバーに紐付く前でも、幹事の画面を開ける |
| メンバー | 名簿のメンバーとの紐付けが承認され、退会していないアカウント |
| 承認待ち | 招待リンクからログインし、承認を待っているアカウント |
| 却下 | 紐付けを却下されたアカウント |

走る幹事も、ほかの人と同じ手順（招待リンク → LINE ログイン → 承認）で自分のアカウントを紐付ける。自分の紐付けを自分で承認してよい。紐付いた後は、幹事もメンバーの画面（自分の出欠、ラップタイム、プロフィール）をすべて使える。

## 画面の一覧

| ID | 画面 | 幹事 | メンバー | 承認待ち／却下／未ログイン |
| --- | --- | --- | --- | --- |
| A1 | 招待リンクの入口 | — | — | 開ける |
| A2 | 承認待ち | — | — | 承認待ち |
| A3 | 却下 | — | — | 却下 |
| A4 | 利用できない | — | — | 退会したメンバー、使えないリンク |
| A5 | ログイン | — | — | 未ログイン |
| R1 | 大会一覧（ホーム） | 開ける | 開ける | — |
| R2 | 大会の詳細 | 開ける | 開ける（一部を隠す） | — |
| R3 | 大会の登録・編集 | 開ける | — | — |
| R4 | 大会への招待 | 開ける | — | — |
| R5 | 出欠の一覧 | 開ける | — | — |
| R6 | エントリー | 開ける | — | — |
| R7 | 走順の編集 | 開ける | — | — |
| L1 | 自分のラップタイム | 開ける（紐付け済みなら） | 開ける | — |
| P1 | 自分のプロフィール | 開ける（紐付け済みなら） | 開ける | — |
| M1 | 名簿 | 開ける | — | — |
| M2 | メンバーのプロフィール | 開ける | — | — |
| M3 | 承認 | 開ける | — | — |

## 参加とアクセスの画面

### A1 招待リンクの入口

- 個人の招待リンクから開く
- 表示：チーム名と「LINE でログイン」のボタン。名簿のメンバーの名前は出さない
- 操作：LINE でログインする。ログイン後、アカウントと名簿のメンバーの紐付けが承認待ちで作られ、招待リンクは使えなくなり、A2 に進む
- リンクが期限切れ、使用済み、新しいリンクに置き換え済みのときは A4 に進む

### A2 承認待ち

- 表示：「幹事の承認を待っています」。チームのデータは出さない

### A3 却下

- 表示：「却下されました」。チームのデータは出さない。理由はアプリの外で伝える

### A4 利用できない

- 退会したメンバーがログインしたとき、使えない招待リンク（期限切れ、使用済み、置き換え済み）のときに出す
- 表示：tasuki を使えないこと、幹事に連絡すること

### A5 ログイン

- すでに参加していて、ログインしていない人のため（セッションが切れたときなど）。セッションなしでチームの画面を開くと、ここに送る（#113）
- 表示：「LINE でログイン」のボタン
- 操作：LINE でログインする。ログイン後は、開こうとしていた画面に戻る（ADR-0013）
- LINE ログインが失敗したか取り消されたら、この画面にエラーとボタンをもう一度出す

### どの画面を誰が開けるか

- 上の表で開けない人が画面を開いたら、エラーではなく次の画面へ送る：未ログイン → A5、承認待ち → A2、却下 → A3、紐付けなし（幹事以外）→ A4、幹事の画面を開いたメンバーと、メンバーの画面（L1、P1）を開いた紐付けのない幹事 → R1

## 大会の画面

### R1 大会一覧（ホーム）

- ログイン後に最初に開く画面。どの一覧も新しく登録した順
- 幹事：チームのすべての大会。大会名、開催日、走順が確定したかどうかを出す
- メンバーは自分の大会だけを、2つの欄に分けて見る

| 欄 | 大会 | 出すもの |
| --- | --- | --- |
| 参加する大会 | 「参加」と答えた大会。過去の大会も含む | 大会名、開催日、確定したら「走順確定」の印 |
| 招待 | 誘われていて、未回答か「保留」の大会 | 大会名、開催日、自分の回答 |

- メンバーには、誘われていない大会と、「不参加」と答えた大会は出ない
- 名簿のメンバーに紐付いた幹事は、自分への招待も「招待」の欄に出る
- 操作：大会を開く（R2）。幹事は大会を登録する（R3）

### R2 大会の詳細

開ける人：幹事と、誘われていて回答が「不参加」ではないメンバー

| 項目 | 幹事 | メンバー：未回答／保留 | メンバー：参加 |
| --- | --- | --- | --- |
| 大会名 | 見られる | 見られる | 見られる |
| 説明欄：種別（距離制／時間制）、開催日、自由な文章、区ごとの距離、1周の距離、制限時間、目標タイム | 見られる | 見られる | 見られる |
| 締め切り | 見られる | 見られない | 見られない |
| 自分の出欠（回答・変更） | 誘われていれば | 見られる | 見られる |
| 走順の下書き | 見られる | 見られない | 見られない |
| 確定した走順 | 見られる | 見られない | 見られる |
| 予想（合計タイムまたは周回数） | 見られる（下書きも確定も） | 見られない | 確定した後 |
| この大会の走者のラップタイム | 見られる（見るだけ） | 見られない | 自分の分だけ、L1 で |

- 出欠：メンバーは参加、不参加、保留のどれかを選ぶ
  - 「参加」と「保留」の間は、いつでも変えられる
  - 「不参加」はメンバーにとって最後の回答。その大会はメンバーの画面から消え、回答を変えられない。先に画面で確認する
  - エントリーに入っているメンバーが「参加」から変えると、エントリーと走順から外れる。走順が確定していれば、この大会のラップタイムは消え、平均タイムは計算し直され、その区は空きになる。先に画面で注意を出す
- 確定した走順で誰も入っていない区は「未定」と出す。予想はその区を除いて計算し、空きの区があることを示す
- 幹事の操作：大会の編集（R3）、招待（R4）、出欠の一覧（R5）、エントリー（R6）、走順（R7）
- 大会の削除や中止はアプリではできない

### R3 大会の登録・編集（幹事）

- 入力：大会名、種別（距離制／時間制）、開催日、締め切り、目標タイム、自由な文章
  - 距離制：区ごとの距離（km）。区は足したり消したりできる
  - 時間制：制限時間（分）と1周の距離（km）
- 操作：保存する

### R4 大会への招待（幹事）

- 退会していない名簿のメンバーと、それぞれが誘われているかを出す
- 操作：
  - メンバーを誘う
  - 招待を取り消す。そのメンバーは出欠、エントリー、走順からも外れる。走順が確定していれば、そのメンバーのこの大会のラップタイムは消え、平均タイムは計算し直される。取り消す前に画面で注意を出す

### R5 出欠の一覧（幹事）

- 誘ったメンバーごとに回答を出す：参加、不参加、保留、未回答
- 見るだけ。幹事がメンバーの代わりに回答することはできない

### R6 エントリー（幹事）

- 幹事が名前（例：「A チーム」）を付けてエントリーを作る。MVP では大会ごとに1つまで
- エントリーのメンバーを出す。「参加」から回答を変えたメンバーは、自動でエントリーから外れる（R2 を参照）
- 操作：
  - 出欠が「参加」のメンバーを入れる
  - メンバーを外す。そのメンバーは走順からも外れる。走順が確定していれば、そのメンバーのこの大会のラップタイムは消え、平均タイムは計算し直される

### R7 走順の編集（幹事）

- 区を順に並べ、区ごとの走者、状態（下書き／確定）、予想を出す
- 区：
  - 距離制：説明欄の区ごとの距離と同じ数の区
  - 時間制：幹事が区を自由に足したり消したりする
- 走者はエントリーのメンバーから選ぶ。1人が何区でも走れる
- 操作：
  - 区の走者を決める、外す、保存する
  - 確定する。すべての区に走者が入っているときだけ。確定するとメンバーが見られる
  - 確定した後に直す。保存するとすぐメンバーに見え、確定のまま。走順から外れたメンバーは、この大会のラップタイムが消え、平均タイムは計算し直される
- 変更はすべて履歴に残す。MVP には履歴を見る画面はない

## ラップタイムとプロフィール

### L1 自分のラップタイム

- R2 から開く。自分が確定した走順の走者になっている大会だけ
- その大会の自分のラップタイム（1km あたり、`MM:SS`）を出す
- 操作：ラップタイムを足す、1つを直す、1つを消す。そのたびに平均タイムを計算し直す
- 自分のラップタイムを変えられるのは本人だけ。幹事は見られるが変えられない

### P1 自分のプロフィール

- 表示：名前、平均タイム（`MM:SS`）、ラップ数
- メンバーがここで入力するものはない

## 名簿と承認（幹事）

### M1 名簿

- 名簿のメンバーごとに出す：名前、退会したかどうか、紐付けの状態（アカウントなし／承認待ち／承認済み）、有効な招待リンクがあればその期限
- 操作：
  - 名簿のメンバーを追加する（名前だけ）
  - 招待リンクを作り、コピーして本人に直接送る。承認済みのアカウントがあるメンバーと、退会したメンバーには作れない。新しいリンクを作ると、古いリンクは使えなくなる
  - 退会にする。開催日が今日以降の大会への招待は自動で取り消され、R4 で取り消したときと同じことが起きる。過去の大会は変えない
  - 退会を取り消す。取り消された招待は戻らない
  - 承認済みのアカウントとの紐付けを解除する（LINE のアカウントを変えた人など）
  - メンバーのプロフィールを開く（M2）
- 名前は変えられない

### M2 メンバーのプロフィール

- 表示：名前、平均タイム、ラップ数、大会ごとのラップタイム
- 見るだけ

### M3 承認

- 承認待ちの紐付けを出す：LINE の表示名、名簿のメンバーの名前、ログインした日時
- 操作：承認する、却下する。却下はアプリの外で伝える

## 主な流れ

### チームへの参加

1. 幹事がその人を名簿に追加する（M1）
2. 幹事が招待リンクを作り（M1）、本人に直接送る
3. 本人がリンクを開き（A1）、LINE でログインして待つ（A2）
4. 幹事が承認する（M3）。本人は tasuki を使えるようになる（R1）。却下されたら A3 が出る

### 大会への招待と出欠の回答

1. 幹事が大会を登録する（R3）
2. 幹事がメンバーを誘う（R4）
3. 誘われたメンバーは「招待」の欄（R1）で大会を見つけ、R2 で説明欄を見て、出欠を答える。参加する大会は「参加する大会」に移る
4. 幹事が回答を確かめる（R5）

### 走順の作成と確定

1. 幹事がエントリーを作り、参加するメンバーを入れる（R6）
2. 幹事が区に走者を入れ、予想を確かめる（R7）
3. 幹事が確定する（R7）。「参加」と答えたメンバーが、R2 で走順と予想を見られる

### ラップタイムの提出

1. 大会の後、走者は大会（R2）から L1 を開く
2. 1km ごとのラップタイムを入れる。プロフィール（P1）の平均タイムとラップ数が更新される

</details>
