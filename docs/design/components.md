# UI common components

- Status: Draft
- Issue: #147
- Related: `docs/design/ui.md` (#146), `docs/design/wireframes/` (#148), `docs/design/screens.md`, ADR-0012 (frontend structure), ADR-0013 (drafts and version conflicts)

This document lists the components every MVP screen is built from: what each is for, its variants and states, how it looks, and where it is used. The values (colors, type, spacing) come from `docs/design/ui.md`; this document only says which value each part uses. The components are drawn in [`wireframes/components.html`](wireframes/components.html), and the screens that use them in the other wireframes.

The screens today render plain HTML versions of these components (ADR-0012). "In the code" names them. Their English text is replaced by the Japanese text below when each screen is styled.

## Rules for every component

- A state is never shown by color alone. It also has text, a mark (○ △ ×), or a different border width.
- Text meets 4.5:1 and the edges of controls 3:1 against their background (WCAG 2.2; the pairs are in `ui.md`).
- Anything that can be tapped is at least 44 px high (`min-h-11`).
- Focus is shown with a 2 px `blue-700` outline, 2 px away from the element (WCAG 2.2, Focus Visible).

## Buttons

Shape: at least 44 px high, 8 px corners (`rounded-lg`), 16 px side padding, Headline (17 px Semibold). A button that is the main action of a form spans the full width. Buttons inside a card are smaller: 12 px side padding and Subheadline (15 px), still 44 px high. References: Human Interface Guidelines, Buttons; Material 3, filled and outlined buttons.

| Variant | Look | For | Examples |
| --- | --- | --- | --- |
| Primary | `blue-700` background, white text | The main action of a screen or a form. At most one per form | 保存 (R3, R7), 送る (L1), 追加 (M1), 確定する (R7), 承認 (M3) |
| Secondary | White background, `blue-700` text, `slate-500` border | Other actions | やめる, 区を足す (R3, R7), 行を足す (L1), コピー (M1), 最新を読み込む (R3, R7), 直す (L1) |
| Danger | White background, `red-700` text and border | Actions that cannot be undone or that delete data. Each asks first (see Confirmation dialog), except 消す on L1: a deleted lap can be typed again | 不参加 (R2), 取り消す (R4), 外す (R6), 退会にする and 紐付けを解除 (M1), 却下 (M3), 消す (L1) |

| State | Look | When |
| --- | --- | --- |
| Default | As above | |
| Disabled | `slate-100` background, `slate-600` text (6.9:1), no border. The reason is written below in Footnote | The action is not possible now, for example 確定する while a leg is empty (R7) |
| Busy | Disabled look, and the label changes to what is happening, for example 保存中… | A request for this action is running. The other actions of the screen are disabled too |

The LINE Login button (A1, A5) is not one of these variants: its colors, logo, and wording (LINEでログイン) follow the LINE Login button design guidelines.

In the code: plain `<button>` elements in each feature's `ui/`.

## Form inputs

Shape: the label is above the input in Subheadline and secondary text (`slate-600`). The input is at least 44 px high, white, with a 1 px `slate-500` border, 8 px corners, 12 px side padding, and Body text (17 px; at 16 px or more, iOS Safari does not zoom in when the input gets focus). Inputs are 16 px apart in a form.

| Input | For | Used on |
| --- | --- | --- |
| Text | One line: race name, entry name, roster member name | R3, R6, M1 |
| Multi-line text | The race's free text (メモ), at least 88 px high | R3 |
| Number | km, minutes, laps. Opens the number keyboard (`inputmode="decimal"` or `"numeric"`), with the unit after the input | R3 |
| Date | Race date | R3 |
| Date and time | Deadline, in the browser's time zone | R3 |
| `MM:SS` time | Lap times per km. Number keyboard | L1 |
| `H:MM:SS` time | Target total time of a distance race. Number keyboard | R3 |
| Select | The runner of a leg, from the entry's members, and 選ばない to leave it empty | R7 |
| Choice | A row of options that works like radio buttons, each at least 44 px high. The selected one has a `blue-100` background and a 2 px `blue-700` border; the others a 1 px `slate-500` border | Race type (R3); attendance with ○ △ × and the meaning of each written below (R2) |
| Read-only text | The created invite link, to copy | M1 |

Some inputs have a help text below the label in Footnote: 1km あたりのタイムを 分:秒 で入れてください（例：3:25） on L1, and 名前はあとで変えられません on M1.

| State | Look | When |
| --- | --- | --- |
| Default | As above | |
| Focused | The focus outline (see the rules) | |
| Invalid | `red-700` border, and the message below the input in Footnote and `red-700`. The message is tied to the input for screen readers (`aria-invalid`, `aria-describedby`) | The value breaks a rule, found before sending or returned by the backend |
| Disabled | `slate-100` background, `slate-600` text | While the form is being sent |

Repeating inputs: rows of the same input with a Danger button (消す) on each row and a Secondary button below (区を足す, 行を足す), for leg distances (R3), legs of a time race (R7), and new lap times (L1).

In the code: `Field` in `features/races/ui/RaceFormView.tsx`, and plain inputs in the other views.

## Confirmation dialog

Asks before an action with effects the user cannot undo. Reference: Human Interface Guidelines, Alerts.

Look: the screen behind is dimmed (`slate-900` at 40 %). The dialog is a white card, 12 px corners, 24 px padding, at most 320 px wide. The title is the question in Headline, the messages say what will happen in Body, and a warning, if any, is in the warning colors (`amber-800` on `amber-100`). The buttons are stacked at full width: the confirm button first, then やめる (Secondary).

| Where | Title | Confirm button |
| --- | --- | --- |
| R2 | 不参加にしますか？ | 不参加にする (Danger) |
| R2 | 保留にしますか？ (changing from 参加) | 保留にする (Primary) |
| R4 | {名前} さんの招待を取り消しますか？ | 取り消す (Danger) |
| R6 | {名前} さんをエントリーから外しますか？ | 外す (Danger) |
| R7 | 確定した走順を変えますか？, naming the runners who lose lap times | 保存する (Danger) |
| M1 | {名前} さんを退会にしますか？ | 退会にする (Danger) |
| M1 | {名前} さんの紐付けを解除しますか？ | 解除する (Danger) |
| M3 | {名簿の名前}（{LINE の名前}）を却下しますか？ | 却下する (Danger) |

| State | When |
| --- | --- |
| Open | Focus moves into the dialog, and the screen behind cannot be used |
| Busy | The request is running. Both buttons are disabled and the confirm button shows what is happening, for example 取り消し中… |

In the code: `shared/ui/Confirmation.tsx` (`role="alertdialog"`).

## Lists and tables

Lists are columns of cards, as `ui.md` decides (white, `slate-200` border, 12 px corners, at least 72 px high, 8 px apart). Times are tables inside a card, as in Strava's splits: a Footnote header row, one row per km or leg, `slate-200` lines between rows, numbers with equal widths (`tabular-nums`).

| List | Shows | Used on |
| --- | --- | --- |
| Race list | Name, race date, 走順確定. Members see two groups: 参加する大会 and 招待 | R1 |
| Running order | Leg, distance, runner (未定 when empty), the state, and the prediction | R2, R7 |
| Invitations | Each roster member who has not left, and whether they are invited | R4 |
| Attendance | Each invited member and their answer, grouped by answer | R5 |
| Entry members | Members of the entry, and the attending members who can be added | R6 |
| Lap times | Laps per runner (R2) or per race (M2); the user's own laps with 直す and 消す (L1) | R2, M2, L1 |
| Roster | Name, left or not, link status, invite link expiry, actions | M1 |
| Waiting links | Roster member, LINE display name, when they logged in, 承認 and 却下 | M3 |

| State | When |
| --- | --- |
| Empty | Nothing to show yet: a sentence in Subheadline in place of the cards, for example まだラップタイムはありません |
| With actions | Cards with their own buttons (R4, R6, L1, M1, M3). Each button names its card for screen readers, for example 佐藤 健 さんの紐付けを解除 |

## Status labels

Short labels for the state of a thing, always as text, in Footnote, fully round, 8 px side padding. The color pairs are those of `ui.md`.

| Thing | Label | Colors |
| --- | --- | --- |
| Running order | 走順確定 | Confirmed (`green-800` on `green-100`) |
| Running order | 下書き | Neutral (`slate-700` on `slate-100`) |
| Leg | 未定 | Neutral |
| Attendance | ○ 参加 | Confirmed |
| Attendance | △ 保留, × 不参加 | Neutral |
| Attendance | 未回答 | Waiting (`amber-800` on `amber-100`) |
| User link | ユーザーなし | Neutral |
| User link | 承認待ち | Waiting |
| User link | 承認済み | Confirmed |
| Roster member | 退会 | Neutral |
| Joining | 承認待ち (A2) | Waiting |

## Loading, empty, and error states

| State | Look | Text |
| --- | --- | --- |
| Loading | The screen's title area shows the text in Subheadline until its data arrives. The top bar and the tab bar stay | 読み込み中… (ログイン中… after LINE Login) |
| Empty | See Lists and tables | まだ大会はありません, 承認待ちはありません, まだ名簿にメンバーはいません |
| Not found | A screen-level message: the title in Title 2 and a sentence in Body, centered | 大会が見つかりません, 名簿のメンバーが見つかりません |
| Something went wrong | A screen-level message, as above, with what to do | うまくいきませんでした。ページを読み込み直してください |
| Lap times not open | A screen-level message, as above | ラップタイムはまだ入れられません。確定した走順でこの大会を走ると、入れられるようになります (L1, for anyone who is not a runner in a confirmed running order) |
| Field errors | See Form inputs | |
| Someone else saved first | A warning above the form (`amber-800` on `amber-100`) that keeps the user's input, with a Secondary button (ADR-0013, `409 version_conflict`) | ほかの人が先に保存しました。 / 最新を読み込む（自分の変更を捨てる） (R3, R7) |

## Notifications

Short messages after an action, for example 大会を保存しました or 名簿がほかの人に変えられていたので、最新を出しています. Reference: Material 3, Snackbar.

Look: a white card above the tab bar, 16 px from the sides and from the tab bar, 8 px corners, a `slate-200` border, Body text. Errors have a `red-700` border and text. Several stack 8 px apart. Each disappears after 4 seconds (`notificationDuration` in the code).

| Variant | Read by screen readers |
| --- | --- |
| Success | When they are free (`aria-live="polite"`) |
| Error | At once (`role="alert"`) |

In the code: `shared/ui/NotificationList.tsx`, fed by `shared/notification/`.

## Page frame

The layout every team screen shares, from top to bottom. References: Human Interface Guidelines, Navigation bars and Tab bars; Material 3, Navigation bar.

- **Top of the screen.** On the screens of the tab bar (R1, M1, M3, P1): a white header with the title in Large Title (34 px), and the screen's main action on the right if it has one (大会を登録 on R1). On the other screens: a white bar at least 44 px high with ‹ and the name of the screen it goes back to, in `blue-700` Headline, and the title in Title 1 at the top of the content.
- **Content.** 16 px from each side, 24 px between groups. A group has a heading in Footnote and secondary text.
- **Tab bar.** White, a `slate-200` line above, at least 49 px high. Each tab has an icon and its name in Footnote (13 px: `ui.md` uses nothing smaller, so the tab names are larger than the 10 pt of iOS). The current tab is `blue-700` and Semibold, the others `slate-600`. 承認 shows the number of waiting links in a `red-700` badge, hidden at 0. The tabs per person are in `wireframes/index.html`.

The joining and access screens (A1–A5) have no top bar and no tab bar: their content is centered.

In the code: `app/Layout.tsx`.

## Not decided here

- The icons of the tabs. The wireframes draw a box in their place.
- Dark mode (`ui.md`).

## Sources

- Apple Human Interface Guidelines: [Buttons](https://developer.apple.com/design/human-interface-guidelines/buttons), [Alerts](https://developer.apple.com/design/human-interface-guidelines/alerts), [Navigation bars](https://developer.apple.com/design/human-interface-guidelines/navigation-bars), [Tab bars](https://developer.apple.com/design/human-interface-guidelines/tab-bars)
- Material Design 3: [Buttons](https://m3.material.io/components/all-buttons), [Snackbar](https://m3.material.io/components/snackbar/overview), [Navigation bar](https://m3.material.io/components/navigation-bar/overview)
- WCAG 2.2: [Focus Visible](https://www.w3.org/WAI/WCAG22/Understanding/focus-visible.html), [Non-text Contrast](https://www.w3.org/WAI/WCAG22/Understanding/non-text-contrast.html)
- LINE Developers: [LINE Login button design guidelines](https://developers.line.biz/en/docs/line-login/login-button/)

---

<details>
<summary>日本語</summary>

# UI の共通の部品

- Status: Draft
- Issue: #147
- Related: `docs/design/ui.md`（#146）、`docs/design/wireframes/`（#148）、`docs/design/screens.md`、ADR-0012（frontend の構成）、ADR-0013（下書きと version の衝突）

この文書は、MVP のすべての画面を組み立てる部品を一覧にする。部品ごとに、何のためか、種類と状態、見た目、使う場所を書く。値（色、文字、余白）は `docs/design/ui.md` から取り、この文書はどの部分にどの値を使うかだけを書く。部品は [`wireframes/components.html`](wireframes/components.html) に、部品を使った画面はほかのワイヤーフレームに描いた。

今の画面は、これらの部品の素の HTML の版を出している（ADR-0012）。「コード」にその名前を書く。英語の文言は、各画面に見た目を付けるときに、下の日本語の文言に置き換える。

## すべての部品の決まり

- 状態を色だけで示さない。文字、印（○ △ ×）、枠の太さの違いも付ける
- 文字は 4.5:1、操作できる部品の縁は 3:1 を、背景に対して満たす（WCAG 2.2。組み合わせは `ui.md`）
- 押せるものは、高さ 44 px 以上（`min-h-11`）
- フォーカスは、要素から 2 px 離した 2 px の `blue-700` の線で示す（WCAG 2.2 の Focus Visible）

## ボタン

形：高さ 44 px 以上、角丸 8 px（`rounded-lg`）、左右の内側の余白 16 px、Headline（17 px Semibold）。入力画面の主な操作のボタンは、横いっぱいに広げる。カードの中のボタンは小さくする（左右 12 px、Subheadline 15 px、高さは 44 px のまま）。参考：Human Interface Guidelines の Buttons、Material 3 の filled と outlined のボタン。

| 種類 | 見た目 | 使い道 | 例 |
| --- | --- | --- | --- |
| 主 | `blue-700` の背景、白い文字 | 画面や入力画面の主な操作。1 つの入力画面に 1 つまで | 保存（R3、R7）、送る（L1）、追加（M1）、確定する（R7）、承認（M3） |
| 副 | 白い背景、`blue-700` の文字、`slate-500` の枠 | ほかの操作 | やめる、区を足す（R3、R7）、行を足す（L1）、コピー（M1）、最新を読み込む（R3、R7）、直す（L1） |
| 危険 | 白い背景、`red-700` の文字と枠 | 元に戻せない操作、データを消す操作。どれも先に確かめる（確認ダイアログを参照）。ただし L1 の「消す」は確かめない。消したラップはまた打てるから | 不参加（R2）、取り消す（R4）、外す（R6）、退会にする・紐付けを解除（M1）、却下（M3）、消す（L1） |

| 状態 | 見た目 | いつ |
| --- | --- | --- |
| 通常 | 上のとおり | |
| 押せない | `slate-100` の背景、`slate-600` の文字（6.9:1）、枠なし。理由を下に Footnote で書く | 今はできない操作。たとえば区が空いているときの「確定する」（R7） |
| 実行中 | 押せないときの見た目で、文言を今していることに変える（例：保存中…） | その操作の通信をしている間。画面のほかの操作も押せなくする |

LINE ログインのボタン（A1、A5）は、この種類のどれでもない。色、ロゴ、文言（LINEでログイン）は LINE ログインボタンのデザインガイドラインに従う。

コード：各機能の `ui/` の素の `<button>`。

## 入力

形：ラベルは入力の上に、Subheadline・補足の文字（`slate-600`）で置く。入力は高さ 44 px 以上、白、1 px の `slate-500` の枠、角丸 8 px、左右の内側の余白 12 px、Body の文字（17 px。16 px 以上なら、iOS の Safari は入力に触れたときに拡大しない）。入力画面の入力どうしは 16 px あける。

| 入力 | 使い道 | 使う画面 |
| --- | --- | --- |
| 文字 | 1 行：大会名、エントリーの名前、名簿のメンバーの名前 | R3、R6、M1 |
| 複数行の文字 | 大会の自由な文章（メモ）。高さ 88 px 以上 | R3 |
| 数 | km、分、周。数字のキーボードを開き（`inputmode="decimal"` か `"numeric"`）、入力の後ろに単位を書く | R3 |
| 日付 | 開催日 | R3 |
| 日付と時刻 | 締め切り。ブラウザのタイムゾーン | R3 |
| `MM:SS` の時間 | 1km ごとのラップタイム。数字のキーボード | L1 |
| `H:MM:SS` の時間 | 距離制の大会の目標の合計タイム。数字のキーボード | R3 |
| 選択 | 区の走者。エントリーのメンバーと、区を空ける「選ばない」から選ぶ | R7 |
| 選択肢 | ラジオボタンのように働く選択肢の並び。それぞれ高さ 44 px 以上。選んでいるものは `blue-100` の背景と 2 px の `blue-700` の枠、ほかは 1 px の `slate-500` の枠 | 種別（R3）。出欠は ○ △ × と、その意味を下に書く（R2） |
| 読むだけの文字 | 作った招待リンク。コピーするため | M1 |

ラベルの下に Footnote の説明がある入力もある：L1 の「1km あたりのタイムを 分:秒 で入れてください（例：3:25）」、M1 の「名前はあとで変えられません」。

| 状態 | 見た目 | いつ |
| --- | --- | --- |
| 通常 | 上のとおり | |
| フォーカス | フォーカスの線（決まりを参照） | |
| 不正 | `red-700` の枠と、入力の下の Footnote・`red-700` の文。文はスクリーンリーダー向けに入力と結び付ける（`aria-invalid`、`aria-describedby`） | 値が決まりに合わない。送る前に分かったときと、backend から返ってきたとき |
| 使えない | `slate-100` の背景、`slate-600` の文字 | 入力画面を送っている間 |

繰り返す入力：同じ入力の行に、行ごとの危険のボタン（消す）と、下の副のボタン（区を足す、行を足す）を付ける。区の距離（R3）、時間制の区（R7）、足すラップタイム（L1）に使う。

コード：`features/races/ui/RaceFormView.tsx` の `Field` と、ほかの画面の素の入力。

## 確認ダイアログ

元に戻せない操作の前に確かめる。参考：Human Interface Guidelines の Alerts。

見た目：後ろの画面を暗くする（`slate-900` の 40 %）。ダイアログは白いカードで、角丸 12 px、内側の余白 24 px、幅は 320 px まで。題は問いを Headline で、文は何が起きるかを Body で書き、注意があれば注意の色（`amber-100` の上の `amber-800`）で書く。ボタンは横いっぱいに縦に並べ、確かめるボタンを先に、その下に「やめる」（副）を置く。

| 場所 | 題 | 確かめるボタン |
| --- | --- | --- |
| R2 | 不参加にしますか？ | 不参加にする（危険） |
| R2 | 保留にしますか？（「参加」から変えるとき） | 保留にする（主） |
| R4 | {名前} さんの招待を取り消しますか？ | 取り消す（危険） |
| R6 | {名前} さんをエントリーから外しますか？ | 外す（危険） |
| R7 | 確定した走順を変えますか？（ラップタイムが消える走者の名前を出す） | 保存する（危険） |
| M1 | {名前} さんを退会にしますか？ | 退会にする（危険） |
| M1 | {名前} さんの紐付けを解除しますか？ | 解除する（危険） |
| M3 | {名簿の名前}（{LINE の名前}）を却下しますか？ | 却下する（危険） |

| 状態 | いつ |
| --- | --- |
| 開いている | フォーカスがダイアログに移り、後ろの画面は使えない |
| 実行中 | 通信をしている間。どちらのボタンも押せず、確かめるボタンは今していることを出す（例：取り消し中…） |

コード：`shared/ui/Confirmation.tsx`（`role="alertdialog"`）。

## 一覧と表

一覧は、`ui.md` のとおりカードの列にする（白、`slate-200` の枠、角丸 12 px、高さ 72 px 以上、8 px ずつ離す）。時間は、Strava のスプリットのようにカードの中の表にする：Footnote の見出しの行、1km か 1 区ごとに 1 行、行の間に `slate-200` の線、数字は同じ幅（`tabular-nums`）。

| 一覧 | 出すもの | 使う画面 |
| --- | --- | --- |
| 大会の一覧 | 大会名、開催日、走順確定。メンバーは「参加する大会」と「招待」の 2 つに分ける | R1 |
| 走順 | 区、距離、走者（空きは「未定」）、状態、予想 | R2、R7 |
| 招待 | 退会していない名簿のメンバーと、誘っているか | R4 |
| 出欠 | 誘ったメンバーと回答。回答ごとにまとめる | R5 |
| エントリーのメンバー | エントリーのメンバーと、入れられる「参加」のメンバー | R6 |
| ラップタイム | 走者ごと（R2）か大会ごと（M2）のラップ。自分のラップは「直す」「消す」付き（L1） | R2、M2、L1 |
| 名簿 | 名前、退会したかどうか、紐付けの状態、招待リンクの期限、操作 | M1 |
| 承認待ち | 名簿のメンバー、LINE の表示名、ログインした日時、承認と却下 | M3 |

| 状態 | いつ |
| --- | --- |
| 空 | まだ出すものがない。カードの代わりに Subheadline の文を出す（例：まだラップタイムはありません） |
| 操作付き | ボタンを持つカード（R4、R6、L1、M1、M3）。ボタンはスクリーンリーダー向けにカードの名前を含める（例：佐藤 健 さんの紐付けを解除） |

## 状態のラベル

ものの状態を示す短いラベル。いつも文字で、Footnote、丸い形、左右の内側の余白 8 px。色の組み合わせは `ui.md` のもの。

| もの | ラベル | 色 |
| --- | --- | --- |
| 走順 | 走順確定 | 確定（`green-100` の上の `green-800`） |
| 走順 | 下書き | 中立（`slate-100` の上の `slate-700`） |
| 区 | 未定 | 中立 |
| 出欠 | ○ 参加 | 確定 |
| 出欠 | △ 保留、× 不参加 | 中立 |
| 出欠 | 未回答 | 待ち（`amber-100` の上の `amber-800`） |
| ユーザーとの紐付け | ユーザーなし | 中立 |
| ユーザーとの紐付け | 承認待ち | 待ち |
| ユーザーとの紐付け | 承認済み | 確定 |
| 名簿のメンバー | 退会 | 中立 |
| 参加 | 承認待ち（A2） | 待ち |

## 読み込み中・空・エラー

| 状態 | 見た目 | 文言 |
| --- | --- | --- |
| 読み込み中 | データが届くまで、画面の題の場所に Subheadline で出す。上のバーと下のタブバーはそのまま | 読み込み中…（LINE ログインの後は ログイン中…） |
| 空 | 一覧と表を参照 | まだ大会はありません、承認待ちはありません、まだ名簿にメンバーはいません |
| 見つからない | 画面全体の文：題を Title 2、文を Body で、中央に | 大会が見つかりません、名簿のメンバーが見つかりません |
| うまくいかなかった | 画面全体の文（上と同じ）と、どうすればよいか | うまくいきませんでした。ページを読み込み直してください |
| ラップをまだ入れられない | 画面全体の文（上と同じ） | ラップタイムはまだ入れられません。確定した走順でこの大会を走ると、入れられるようになります（確定した走順の走者でない人が開いた L1） |
| 入力の誤り | 入力を参照 | |
| ほかの人が先に保存した | 入力画面の上の注意（`amber-100` の上の `amber-800`）。入力した値は残し、副のボタンを付ける（ADR-0013、`409 version_conflict`） | ほかの人が先に保存しました。／最新を読み込む（自分の変更を捨てる）（R3、R7） |

## 通知

操作の後の短い知らせ。例：「大会を保存しました」「名簿がほかの人に変えられていたので、最新を出しています」。参考：Material 3 の Snackbar。

見た目：タブバーの上に出す白いカード。左右とタブバーから 16 px 離し、角丸 8 px、`slate-200` の枠、Body の文字。エラーは `red-700` の枠と文字。いくつか出るときは 8 px ずつ離して重ねる。どれも 4 秒で消える（コードの `notificationDuration`）。

| 種類 | スクリーンリーダー |
| --- | --- |
| 成功 | 手が空いたときに読む（`aria-live="polite"`） |
| エラー | すぐ読む（`role="alert"`） |

コード：`shared/ui/NotificationList.tsx`。`shared/notification/` から受け取る。

## 画面の枠

チームのすべての画面が共有する並び（上から下へ）。参考：Human Interface Guidelines の Navigation bars と Tab bars、Material 3 の Navigation bar。

- **画面の上。** タブバーの画面（R1、M1、M3、P1）：白いヘッダーに Large Title（34 px）の題。画面の主な操作があれば右に置く（R1 の「大会を登録」）。ほかの画面：高さ 44 px 以上の白いバーに、‹ と戻る先の画面の名前を `blue-700` の Headline で置き、題は内容のいちばん上に Title 1 で書く
- **内容。** 左右から 16 px、グループの間は 24 px。グループには Footnote・補足の文字の見出しを付ける
- **タブバー。** 白、上に `slate-200` の線、高さ 49 px 以上。タブごとにアイコンと名前（Footnote）。`ui.md` は 13 px より小さい文字を使わないので、タブの名前は iOS の 10 pt より大きい。今のタブは `blue-700` の Semibold、ほかは `slate-600`。「承認」には承認待ちの数を `red-700` の丸で出し、0 のときは出さない。人ごとのタブは `wireframes/index.html` にある

参加とアクセスの画面（A1〜A5）には、上のバーもタブバーもない。内容は中央に置く。

コード：`app/Layout.tsx`。

## ここでは決めないこと

- タブのアイコン。ワイヤーフレームでは、その場所に四角を描いた
- ダークモード（`ui.md`）

## 出典

英語の節を参照。

</details>
