# MVP UI design

- Status: Draft
- Issue: #145 (design basics #146, common components #147, wireframes #148)
- Related: `docs/design/screens.md`, `docs/design/ui-references.md`, ADR-0012 (frontend structure)

`docs/design/screens.md` decides what each screen shows and lets the user do. This document decides how the screens look. It is written in three parts, one per sub-issue of #145:

| Part | Issue | State |
| --- | --- | --- |
| Design basics: colors, typography, spacing | #146 | Not decided yet |
| Common components | #147 | The list of components, their variants, and their states. How they look waits for #146 |
| Wireframes | #148 | Not started |

## Design basics (#146)

Not decided yet. Colors, typography, and spacing go here, as tokens in the `@theme` block of `app/frontend/src/index.css` (`app/frontend/TECH.md`).

## Common components (#147)

The components every MVP screen is built from. For each: what it is for, its variants, its states, and where it is used. **How each one looks is not decided here**; it follows #146. The screens today render plain HTML versions of them (ADR-0012), named in "In the code".

Rules for every component:

- A state is never shown by color alone: it also has text or an icon, so it reads the same without color.
- Text and background colors meet WCAG AA contrast (#146).
- Anything the user can press is large enough to press with a thumb on a phone; the size comes with #146.

### Buttons

| Variant | For | Examples |
| --- | --- | --- |
| Primary | The main action of a screen or a form. At most one per form | Save (R3, R7), Submit (L1), Add (M1), Confirm (R7) |
| Secondary | Other actions | Cancel, Add leg (R7), Add another lap (L1), Copy link and Close (M1), Load the latest (R3, R7), opening another screen from R2 |
| Danger | Actions the user cannot undo, or that delete data. Each one asks first (see Confirmation dialog), except Delete lap (L1): a deleted lap can be typed again | Answer not attending (R2), Cancel invitation (R4), Remove from the entry (R6), Mark as left and Unlink (M1), Reject (M3), Delete lap (L1) |

| State | When |
| --- | --- |
| Default | |
| Disabled | The action is not possible now, for example Confirm while a leg is empty (R7). The reason is written next to it |
| Busy | A request for this action is running. Other actions of the screen are disabled too |

**LINE Login button** (A1, A5) is not one of these variants: its colors, shape, and wording follow LINE's guidelines (`docs/design/ui-references.md`).

In the code: plain `<button>` elements in each feature's `ui/`.

### Form inputs

| Input | For | Used on |
| --- | --- | --- |
| Text | One line: race name, entry name, roster member name | R3, R6, M1 |
| Multi-line text | The race's free text | R3 |
| Number | km and minutes: leg distances, lap distance, time limit | R3 |
| Date | Race date | R3 |
| Date and time | Deadline, in the browser's time zone | R3 |
| `MM:SS` time | Lap times per km | L1 |
| `H:MM:SS` time | Target total time of a distance race | R3 |
| Select | The runner of a leg, from the entry's members | R7 |
| Choice (radio) | Race type; attendance answer | R3; R2 |
| Read-only text | The created invite link, to select and copy | M1 |

Every input has a visible label. Some have a help text below the label, for example "Enter the time per km as M:SS, for example 3:25." (L1) and "Names cannot be changed later." (M1).

| State | When |
| --- | --- |
| Default | |
| Invalid | The value breaks a rule, found before sending or returned by the backend. The message is shown below the input and tied to it for screen readers (`aria-invalid`, `aria-describedby`) |
| Disabled | While the form is being sent |

Repeating inputs: a list of the same input with add and remove buttons, for leg distances (R3), legs (R7, time races), and new lap times (L1).

In the code: `Field` in `features/races/ui/RaceFormView.tsx`, and plain inputs in the other views.

### Confirmation dialog

Asks before an action with effects the user cannot undo. It has a title (the question), one or more messages (what will happen), a confirm button, and Cancel.

| Where | Title (as in the code) | Confirm button |
| --- | --- | --- |
| R2 | "Answer not attending?" / "Change your answer?" | Danger / Primary |
| R4 | "Cancel {name}'s invitation?" | Danger |
| R6 | "Remove {name} from the entry?" | Danger |
| R7 | "Save and delete lap times?", naming the runners who lose laps | Danger |
| M1 | "Mark {name} as left?", "Unlink {name}'s user?" | Danger |
| M3 | "Reject {name}?" | Danger |

| State | When |
| --- | --- |
| Open | Focus moves into the dialog; the screen behind cannot be used |
| Busy | The request is running; both buttons are disabled |

In the code: `shared/ui/Confirmation.tsx` (`role="alertdialog"`).

### Lists and tables

| List | Shows | Used on |
| --- | --- | --- |
| Race list | Name, race date, running order confirmed. Members see two sections: My races and Invitations | R1 |
| Running order | Leg number, distance, runner ("not decided" when empty), the status, and the prediction | R2, R7 |
| Invitations | Each roster member who has not left, and whether they are invited | R4 |
| Attendance | Each invited member and their answer | R5 |
| Entry members | Members of the entry | R6 |
| Lap times | Laps per runner (R2) or per race (M2); the user's own laps with edit and delete (L1) | R2, M2, L1 |
| Roster | Name, left or not, link status, invite link expiry, actions | M1 |
| Waiting links | LINE display name, roster member, when they logged in, approve and reject | M3 |

Whether each one is a table or a list of cards on a phone is decided with the wireframes (#148).

| State | When |
| --- | --- |
| Empty | Nothing to show yet, with a sentence saying so, for example "No lap times yet." |
| With actions | Rows that have their own buttons (R4, R6, L1, M1, M3). Each button names its row for screen readers, for example "Unlink Taro" |

### Status labels

Short labels for the state of a thing, always as text.

| Thing | Labels |
| --- | --- |
| Running order | Draft, Confirmed |
| Leg | Not decided |
| Attendance | Attending, Not attending, Undecided, Not answered |
| User link | No user, Waiting for approval, Approved |
| Roster member | Left |
| Joining | Waiting (A2), Rejected (A3) |

### Loading, empty, and error states

| State | Shown as | Examples |
| --- | --- | --- |
| Loading | The screen's title area says "Loading…" until its data arrives | Every screen |
| Empty | A sentence in place of the list | "No roster members yet.", "Nobody is waiting for approval." |
| Not found | A screen-level message | "Race not found", "Roster member not found" |
| Something went wrong | A screen-level message with what to do | "Something went wrong. Reload the page." |
| Field errors | Below each input (see Form inputs) | R3, L1, M1 |
| Someone else saved first | A message above the form that keeps the user's input, with "Load the latest (discard my changes)" | R3, R7 (`409 version_conflict`, ADR-0013) |

### Notifications

Short messages after an action, for example "Race saved." or "The roster changed in the meantime. The latest is shown."

| Variant | Read by screen readers |
| --- | --- |
| Success | When they are free (`aria-live="polite"`) |
| Error | At once (`role="alert"`) |

In the code: `shared/ui/NotificationList.tsx`, fed by `shared/notification/`.

### Page frame

The layout every screen shares: a header with "tasuki" and the screen below it. R2 links to the screens of the race (R3–R7, L1) with buttons.

In the code: `app/Layout.tsx`.

## Wireframes (#148)

Not started.

---

<details>
<summary>日本語</summary>

# MVP の UI の設計

- Status: Draft
- Issue: #145（デザインの基本 #146、共通の部品 #147、ワイヤーフレーム #148）
- Related: `docs/design/screens.md`、`docs/design/ui-references.md`、ADR-0012（frontend の構成）

`docs/design/screens.md` は、各画面に何を出し、何ができるかを決める。この文書は、画面をどう見せるかを決める。#145 の sub-issue ごとに 3 つに分けて書く：

| 部分 | Issue | 状態 |
| --- | --- | --- |
| デザインの基本：色、文字、余白 | #146 | まだ決めていない |
| 共通の部品 | #147 | 部品の一覧、その種類と状態。見た目は #146 を待つ |
| ワイヤーフレーム | #148 | まだ始めていない |

## デザインの基本（#146）

まだ決めていない。色、文字、余白はここに書き、`app/frontend/src/index.css` の `@theme` に token として置く（`app/frontend/TECH.md`）。

## 共通の部品（#147）

MVP のすべての画面を作る部品。それぞれについて、何のためか、種類、状態、使う場所を書く。**見た目はここでは決めない。** #146 に従う。今の画面は、これらの素の HTML 版を出している（ADR-0012）。それを「コード」に書く。

すべての部品の決まり：

- 状態を色だけで示さない。文字かアイコンも付け、色がなくても同じに読めるようにする
- 文字と背景の色は、WCAG AA のコントラストを満たす（#146）
- 押せるものは、スマホで親指で押せる大きさにする。大きさは #146 で決める

### ボタン

| 種類 | 何のためか | 例 |
| --- | --- | --- |
| 主 | 画面やフォームの主な操作。1 つのフォームに 1 つまで | 保存（R3、R7）、送る（L1）、追加（M1）、確定（R7） |
| 副 | ほかの操作 | キャンセル、区を足す（R7）、ラップを足す（L1）、リンクをコピー・閉じる（M1）、最新を読む（R3、R7）、R2 からほかの画面を開く |
| 危険 | 取り消せない、またはデータを消す操作。どれも先に確認する（確認ダイアログを参照）。ただし、ラップを消す（L1）は確認しない。消したラップは打ち直せるから | 不参加と答える（R2）、招待を取り消す（R4）、エントリーから外す（R6）、退会にする・紐付けを外す（M1）、却下（M3）、ラップを消す（L1） |

| 状態 | いつ |
| --- | --- |
| 通常 | |
| 押せない | 今はできない操作。例えば、空きの区があるときの確定（R7）。理由をそばに書く |
| 実行中 | この操作のリクエストが動いている。画面のほかの操作も押せなくする |

**LINE ログインのボタン**（A1、A5）は、これらの種類に入らない。色、形、文言は LINE のガイドラインに従う（`docs/design/ui-references.md`）。

コード：各機能の `ui/` にある素の `<button>`。

### 入力

| 入力 | 何のためか | 使う画面 |
| --- | --- | --- |
| 文字 | 1 行：大会名、エントリー名、名簿のメンバーの名前 | R3、R6、M1 |
| 複数行の文字 | 大会の自由な文章 | R3 |
| 数 | km と分：区の距離、1 周の距離、制限時間 | R3 |
| 日付 | 開催日 | R3 |
| 日付と時刻 | 締め切り。ブラウザの時間帯で | R3 |
| `MM:SS` の時間 | 1 km あたりのラップタイム | L1 |
| `H:MM:SS` の時間 | 距離制の目標タイム | R3 |
| 選択 | 区の走者。エントリーのメンバーから | R7 |
| 選択肢（ラジオ） | 大会の種類、出欠の答え | R3、R2 |
| 読むだけの文字 | 作った招待リンク。選んでコピーする | M1 |

どの入力にも、見えるラベルを付ける。ラベルの下に説明を付けるものもある。例えば「Enter the time per km as M:SS, for example 3:25.」（L1）、「Names cannot be changed later.」（M1）。

| 状態 | いつ |
| --- | --- |
| 通常 | |
| 不正 | 値が決まりを破っている。送る前に見つけたものか、backend が返したもの。メッセージは入力の下に出し、スクリーンリーダーのために入力と結びつける（`aria-invalid`、`aria-describedby`） |
| 押せない | フォームを送っている間 |

繰り返す入力：同じ入力を並べ、足すボタンと消すボタンを付ける。区の距離（R3）、区（R7、時間制）、新しいラップタイム（L1）。

コード：`features/races/ui/RaceFormView.tsx` の `Field` と、ほかの View の素の入力。

### 確認ダイアログ

取り消せない影響のある操作の前に聞く。タイトル（質問）、1 つ以上のメッセージ（何が起きるか）、確定のボタン、キャンセルがある。

| 場所 | タイトル（コードのとおり） | 確定のボタン |
| --- | --- | --- |
| R2 | 「Answer not attending?」／「Change your answer?」 | 危険／主 |
| R4 | 「Cancel {name}'s invitation?」 | 危険 |
| R6 | 「Remove {name} from the entry?」 | 危険 |
| R7 | 「Save and delete lap times?」。ラップが消える走者の名前を出す | 危険 |
| M1 | 「Mark {name} as left?」、「Unlink {name}'s user?」 | 危険 |
| M3 | 「Reject {name}?」 | 危険 |

| 状態 | いつ |
| --- | --- |
| 開いている | フォーカスはダイアログの中に移り、後ろの画面は使えない |
| 実行中 | リクエストが動いている。どちらのボタンも押せない |

コード：`shared/ui/Confirmation.tsx`（`role="alertdialog"`）。

### 一覧と表

| 一覧 | 出すもの | 使う画面 |
| --- | --- | --- |
| 大会の一覧 | 名前、開催日、走順が確定したか。メンバーは 2 つの欄：自分の大会、招待 | R1 |
| 走順 | 区の番号、距離、走者（空なら「未定」）、状態、予想 | R2、R7 |
| 招待 | 退会していない名簿のメンバーと、招待しているか | R4 |
| 出欠 | 招待したメンバーと、その答え | R5 |
| エントリーのメンバー | エントリーのメンバー | R6 |
| ラップタイム | 走者ごと（R2）か大会ごと（M2）のラップ。自分のラップは直す・消すも（L1） | R2、M2、L1 |
| 名簿 | 名前、退会したか、紐付けの状態、招待リンクの期限、操作 | M1 |
| 承認待ち | LINE の表示名、名簿のメンバー、ログインした日時、承認と却下 | M3 |

スマホで表にするか、カードを並べるかは、ワイヤーフレーム（#148）で決める。

| 状態 | いつ |
| --- | --- |
| 空 | まだ出すものがない。そのことを文で書く。例えば「No lap times yet.」 |
| 操作つき | 行ごとにボタンがある（R4、R6、L1、M1、M3）。ボタンは、スクリーンリーダーのために行の名前を含める。例えば「Unlink Taro」 |

### 状態のラベル

ものの状態を示す短いラベル。いつも文字で出す。

| もの | ラベル |
| --- | --- |
| 走順 | 下書き、確定 |
| 区 | 未定 |
| 出欠 | 参加、不参加、未定、未回答 |
| ユーザーの紐付け | ユーザーなし、承認待ち、承認済み |
| 名簿のメンバー | 退会 |
| 参加の手続き | 承認待ち（A2）、却下（A3） |

### 読み込み中、空、エラー

| 状態 | 出し方 | 例 |
| --- | --- | --- |
| 読み込み中 | データが届くまで、画面のタイトルのところに「Loading…」 | すべての画面 |
| 空 | 一覧の代わりに文 | 「No roster members yet.」、「Nobody is waiting for approval.」 |
| 見つからない | 画面全体のメッセージ | 「Race not found」、「Roster member not found」 |
| うまくいかなかった | 何をすればよいかを添えた画面全体のメッセージ | 「Something went wrong. Reload the page.」 |
| 入力のエラー | 各入力の下（入力を参照） | R3、L1、M1 |
| ほかの人が先に保存した | 入力を残したまま、フォームの上にメッセージと「Load the latest (discard my changes)」 | R3、R7（`409 version_conflict`、ADR-0013） |

### 通知

操作の後の短いメッセージ。例えば「Race saved.」、「The roster changed in the meantime. The latest is shown.」

| 種類 | スクリーンリーダーが読むとき |
| --- | --- |
| 成功 | 手が空いたとき（`aria-live="polite"`） |
| エラー | すぐに（`role="alert"`） |

コード：`shared/ui/NotificationList.tsx`。`shared/notification/` から渡す。

### 画面の枠

すべての画面が共有する配置：「tasuki」のヘッダーと、その下の画面。R2 は、ボタンでその大会の画面（R3〜R7、L1）へ進む。

コード：`app/Layout.tsx`。

## ワイヤーフレーム（#148）

まだ始めていない。

</details>
