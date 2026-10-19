# MVP API design

- Status: Draft
- Issue: #70
- Related: ADR-0006 (access policy), ADR-0007 (LINE Login), ADR-0009 (dev-only login), ADR-0010 (spec-first OpenAPI), `docs/design/screens.md`, `docs/design/database.md`

This document lists the endpoints the MVP screens need. The OpenAPI document (`app/backend/api/openapi.yaml`) is written from it in #71.

## Conventions

| Topic | Rule |
| --- | --- |
| Base path | `/api/v1` |
| Authentication | The session token is carried in an HttpOnly cookie `tasuki_session` (`Secure`, `SameSite=Lax`, `Path=/api`). Scripts on the page cannot read it, so the frontend never handles the token. The frontend and the API are served from the same site. When the token expires, the user logs in with LINE again; there is no refresh token |
| CSRF | Requests that change data (`POST`, `PUT`, `PATCH`, `DELETE`) must carry an `Origin` header equal to the frontend's origin; otherwise `403 forbidden`. With `SameSite=Lax`, this rejects requests sent from other sites |
| Errors | RFC 9457 `application/problem+json` with an extra `code` field the frontend uses to choose what to show |
| Hidden resources | A race (and anything under it) that the caller may not see returns `404`, the same as a race that does not exist |
| Times | Lap times and average times: `"MM:SS"` (e.g. `"3:28"`). Total times: `"H:MM:SS"` (e.g. `"2:45:30"`) |
| Dates | Race date: `YYYY-MM-DD`. Deadline and other timestamps: RFC 3339 with offset (e.g. `2026-11-01T18:00:00+09:00`) |
| Distances | km as numbers (e.g. `30.5`) |
| Concurrent edits | Races and running orders carry a `version`. Updates send the version they read; if it changed, `409` with `code: version_conflict` |
| Lists | No pagination in the MVP |
| IDs | Opaque strings (the type is decided in #68) |
| JSON field names | `snake_case` |

### Who can call

| Caller | Meaning |
| --- | --- |
| Public | No token |
| Account | Any valid token, including accounts waiting for approval or rejected |
| Member | Account whose link to a roster member is approved and who has not left |
| Organizer | Account listed as an organizer in configuration. Does not need a link |

An organizer who is linked is also a member. "Member" endpoints below are open to such organizers too.

### Error codes

| Status | `code` | When |
| --- | --- | --- |
| 400 | `invalid_request` | Malformed body or parameters |
| 401 | `unauthenticated` | No session cookie, or its token is invalid or expired |
| 403 | `forbidden` | The caller's role does not allow the endpoint (e.g. a member calling an organizer endpoint, a waiting account calling a member endpoint), or a request that changes data has a missing or foreign `Origin` |
| 404 | `not_found` | Does not exist, or the caller may not see it |
| 404 | `invite_unusable` | Invite link expired, used, or replaced |
| 409 | `version_conflict` | Someone else changed it since it was read |
| 409 | `state_conflict` | The action is not allowed in the current state (details per endpoint) |
| 422 | `validation_failed` | The body breaks a business rule (details per endpoint), with `errors` listing the fields |

## Authentication and joining

| Method and path | Caller | Screen | Purpose |
| --- | --- | --- | --- |
| `GET /invite-links/{token}` | Public | A1 | Check an invite link before login |
| `POST /auth/line/authorize` | Public | A1, login | Start LINE Login |
| `GET /auth/line/callback` | Public (LINE redirects here) | — | Finish LINE Login, redirect to the frontend |
| `POST /auth/token` | Public | — | Exchange the one-time code for a session |
| `POST /dev/login` | Public, development only | — | Dev-only login (ADR-0009) |
| `DELETE /session` | Account | — | Log out |
| `GET /me` | Account | All | Who am I, and which screens to show |

### `GET /invite-links/{token}`

- `200`: `{ team_name }`. The roster member's name is not returned.
- `404 invite_unusable`: expired, used, or replaced → A4.

### `POST /auth/line/authorize`

- Body: `{ invite_token? }`. Included when the user came from an invite link.
- `200`: `{ authorization_url }`. The frontend navigates to it. The backend keeps the state, the PKCE verifier, and the invite token for the callback.

### `GET /auth/line/callback`

- LINE redirects here with `code` and `state`. The backend verifies the ID token (ADR-0007).
- With an invite token: creates a pending link between the account and the roster member and makes the invite link unusable. If the link became unusable in the meantime, redirects with an error.
- Redirects to the frontend `/auth/complete?code=<one-time code>`, or `/auth/complete?error=<code>`:
  - `invite_unusable`: the invite link cannot be used → A4
  - `left`: the account's roster member has left → A4 (no token is issued)
  - `login_failed`: LINE Login failed or was cancelled
- The one-time code is short-lived and can be exchanged once.

### `POST /auth/token`

- Body: `{ code }`.
- `200`: `{ expires_at }`, and sets the session cookie (`Set-Cookie`). The token is not in the body.
- `400 invalid_request`: unknown, expired, or used code.

### `POST /dev/login`

- Exists only when `DEV_LOGIN_ENABLED=true` (ADR-0009). Otherwise the route is not registered (`404`).
- Body: `{ line_user_id, display_name }`. Response is the same as `POST /auth/token`.

### `DELETE /session`

- `204`. The token can no longer be used, and the session cookie is cleared.

### `GET /me`

```
{
  "display_name": "LINE display name",
  "is_organizer": true,
  "status": "approved" | "pending" | "rejected" | "none",
  "roster_member": { "id": "...", "name": "..." } | null
}
```

| `status` | Meaning | Screen |
| --- | --- | --- |
| `approved` | Linked and approved | R1 |
| `pending` | Waiting for approval | A2 (organizers can still use organizer screens) |
| `rejected` | The latest link was rejected | A3 (same as above for organizers) |
| `none` | No link | A4 for non-organizers; organizers use organizer screens |

- `roster_member` is set only when `status` is `approved`.
- Left members never get a token (see the callback), so there is no `left` status.

## Races

| Method and path | Caller | Screen |
| --- | --- | --- |
| `GET /races` | Member, Organizer | R1 |
| `POST /races` | Organizer | R3 |
| `GET /races/{race_id}` | Member (invited, not "not attending"), Organizer | R2, R3 |
| `PUT /races/{race_id}` | Organizer | R3 |

### Race description object

```
{
  "type": "distance" | "time",
  "race_date": "2026-11-23",
  "deadline": "2026-11-01T18:00:00+09:00",   // organizers only
  "text": "free text",
  "leg_distances_km": [5.0, 3.2, ...],       // distance only
  "time_limit_minutes": 180,                 // time only
  "lap_distance_km": 2.0,                    // time only
  "target": { "total_time": "2:45:00" } | { "laps": 30 }   // distance → total_time, time → laps
}
```

### `GET /races`

- Organizers: every race of the team. Members: races they are invited to, except those answered "not attending".
- Newest registered first.

```
{ "races": [
  {
    "id": "...",
    "name": "...",
    "race_date": "2026-11-23",
    "my_attendance": "attending" | "undecided" | null,   // null = invited, not answered; field omitted if not invited (organizers)
    "running_order_confirmed": true                     // organizers, and members who answered attending
  }
] }
```

- The frontend puts races with `my_attendance = attending` in "My races" and the rest in "Invitations".

### `POST /races` and `PUT /races/{race_id}`

- Body: `{ name, description }` (`PUT` also sends `version`). The description must match its type: distance needs at least one leg distance; time needs the time limit and the lap distance.
- `POST` → `201` race. `PUT` → `200` race, `409 version_conflict`, `422 validation_failed`.
- Changing the type or the number of legs adjusts existing running orders (see Running order: legs). A confirmed running order stays confirmed.

### `GET /races/{race_id}`

```
{
  "id": "...",
  "name": "...",
  "version": 3,                     // organizers only
  "description": { ... },           // deadline only for organizers
  "my_attendance": "attending" | "undecided" | null,   // omitted if not invited
  "running_order_confirmed": true,  // organizers, and members who answered attending
  "can_submit_laps": true           // true when the caller is a runner in a confirmed running order
}
```

## Invitations and attendance

| Method and path | Caller | Screen |
| --- | --- | --- |
| `GET /races/{race_id}/invitations` | Organizer | R4, R5 |
| `PUT /races/{race_id}/invitations/{roster_member_id}` | Organizer | R4 |
| `DELETE /races/{race_id}/invitations/{roster_member_id}` | Organizer | R4 |
| `PUT /races/{race_id}/attendance` | Member (invited) | R2 |

### `GET /races/{race_id}/invitations`

- Every roster member who has not left, with whether they are invited and their answer.

```
{ "members": [
  { "roster_member_id": "...", "name": "...", "invited": true,
    "attendance": "attending" | "not_attending" | "undecided" | null }
] }
```

### `PUT .../invitations/{roster_member_id}`

- Invites the member. `204`. Inviting an invited member does nothing.
- `422 validation_failed`: the member has left.

### `DELETE .../invitations/{roster_member_id}`

- Cancels the invitation. `204`.
- Side effects: removes the attendance answer and the entry membership; removes the member from every leg of the running order. If the running order is confirmed, the member's laps for this race are deleted and their average time is recalculated from all remaining laps.

### `PUT /races/{race_id}/attendance`

- Body: `{ status: "attending" | "not_attending" | "undecided" }`. `200`: `{ status }`.
- `404`: not invited, or already answered "not attending".
- Changing from "attending" to anything else has the side effects of removing the member from the entry (below).

## Entries

| Method and path | Caller | Screen |
| --- | --- | --- |
| `GET /races/{race_id}/entries` | Organizer | R6 |
| `POST /races/{race_id}/entries` | Organizer | R6 |
| `PUT /races/{race_id}/entries/{entry_id}/members/{roster_member_id}` | Organizer | R6 |
| `DELETE /races/{race_id}/entries/{entry_id}/members/{roster_member_id}` | Organizer | R6 |

- `GET`: `{ "entries": [ { "id", "name", "members": [ { "roster_member_id", "name" } ] } ] }`.
- `POST`: body `{ name }` → `201` entry. `409 state_conflict` when the race already has an entry (MVP limit: one).
- Entries cannot be renamed or deleted.
- `PUT .../members/...`: `204`. `422 validation_failed` when the member's attendance is not "attending".
- `DELETE .../members/...`: `204`. Removes the member from every leg of the running order. If the running order is confirmed, their laps for this race are deleted and their average time is recalculated.

## Running order

| Method and path | Caller | Screen |
| --- | --- | --- |
| `GET /races/{race_id}/running-orders` | Member (attending), Organizer | R2, R7 |
| `PUT /races/{race_id}/entries/{entry_id}/running-order` | Organizer | R7 |
| `POST /races/{race_id}/entries/{entry_id}/running-order/confirm` | Organizer | R7 |

### `GET /races/{race_id}/running-orders`

- One running order per entry. Organizers see drafts and confirmed orders; attending members see confirmed orders only (an empty list otherwise).

```
{ "running_orders": [
  {
    "entry_id": "...",
    "entry_name": "...",
    "status": "draft" | "confirmed",
    "version": 5,                   // organizers only
    "legs": [
      { "leg_number": 1, "distance_km": 5.0,   // distance races only
        "runner": { "roster_member_id": "...", "name": "..." } | null }
    ],
    "prediction": {
      "total_time": "2:43:10" | null,   // distance races
      "laps": 25 | null,                // time races
      "has_empty_legs": false
    } | null                            // null when no leg has a runner
  }
] }
```

- The prediction is calculated by the backend (`docs/design/database.md`). Empty legs are left out. Members never receive other members' average times.

### `PUT .../running-order`

- Body: `{ version?, legs: [ { leg_number, roster_member_id | null } ] }`. `version` is omitted when creating the first draft.
- `200`: the running order as above.
- Rules (`422 validation_failed`):
  - Distance races: exactly one leg per leg distance. Time races: one or more legs, numbered from 1 without gaps.
  - Runners must be members of the entry.
- A confirmed running order stays confirmed and the change is visible right away.
- Every save is kept as one history entry (`running_order_history`).
- If a runner's number of legs goes down in a confirmed running order, all their laps for this race are deleted and their average time is recalculated. Moving a runner to another leg without reducing their number of legs keeps the laps.

### `POST .../running-order/confirm`

- Body: `{ version }`. `200`: the running order.
- `422 validation_failed`: some legs have no runner.
- `409 state_conflict`: already confirmed.

### Running order: legs when the race changes

When `PUT /races/{race_id}` changes the type or the number of legs:

- Distance race with fewer legs: the last legs are removed. With more legs: empty legs are added at the end.
- Time → distance: the order is cut or padded to the number of leg distances. Distance → time: the legs are kept.
- Removed legs follow the lap rule above. A confirmed running order stays confirmed, with empty legs shown as not decided.

## Lap times

| Method and path | Caller | Screen |
| --- | --- | --- |
| `GET /races/{race_id}/my-laps` | Member (runner in the confirmed order) | L1 |
| `POST /races/{race_id}/my-laps` | Same | L1 |
| `PATCH /races/{race_id}/my-laps/{lap_id}` | Same | L1 |
| `DELETE /races/{race_id}/my-laps/{lap_id}` | Same | L1 |
| `GET /races/{race_id}/laps` | Organizer | R2 |

- `GET my-laps`: `{ "laps": [ { "id", "time": "3:25" } ], "average_time": "3:28", "lap_count": 8 }`. Laps are in the order they were submitted.
- `POST my-laps`: body `{ times: ["3:20", "3:25"] }` → `201` with the same shape. The average uses the submission formula.
- `PATCH`: body `{ time }` → `200` same shape. `DELETE` → `200` same shape. The average is recalculated from all of the member's laps.
- `404`: not a runner in a confirmed running order of this race (also for organizers who are not runners).
- `GET /races/{race_id}/laps` (organizers): `{ "runners": [ { "roster_member_id", "name", "laps": [ { "id", "time" } ] } ] }`. View only.

## Profile

| Method and path | Caller | Screen |
| --- | --- | --- |
| `GET /me/profile` | Member | P1 |

- `200`: `{ "name", "average_time": "3:28", "lap_count": 8 }`.

## Roster and approvals

| Method and path | Caller | Screen |
| --- | --- | --- |
| `GET /roster-members` | Organizer | M1 |
| `POST /roster-members` | Organizer | M1 |
| `GET /roster-members/{roster_member_id}` | Organizer | M2 |
| `PATCH /roster-members/{roster_member_id}` | Organizer | M1 |
| `DELETE /roster-members/{roster_member_id}/link` | Organizer | M1 |
| `POST /roster-members/{roster_member_id}/invite-links` | Organizer | M1 |
| `GET /member-links` | Organizer | M3 |
| `POST /member-links/{link_id}/approve` | Organizer | M3 |
| `POST /member-links/{link_id}/reject` | Organizer | M3 |

### `GET /roster-members`

```
{ "roster_members": [
  { "id", "name", "left": false,
    "link_status": "none" | "pending" | "approved",
    "invite_link_expires_at": "..." | null }
] }
```

- A rejected link shows as `none`.

### `POST /roster-members`

- Body: `{ name }` → `201` roster member.

### `GET /roster-members/{roster_member_id}`

- `200`: `{ "id", "name", "left", "average_time", "lap_count", "laps_by_race": [ { "race_id", "race_name", "laps": [ { "id", "time" } ] } ] }`.

### `PATCH /roster-members/{roster_member_id}`

- Body: `{ left: true | false }`. Names cannot be changed. `200` roster member.
- `left: true`: cancels invitations to races dated today or later, with the side effects of cancelling an invitation. The member can no longer log in, and existing tokens stop working.
- `left: false`: the member can log in again. Cancelled invitations do not come back.

### `DELETE /roster-members/{roster_member_id}/link`

- Removes the approved link. `204`. The account's tokens stop working as a member.
- `409 state_conflict`: no approved link.

### `POST /roster-members/{roster_member_id}/invite-links`

- `201`: `{ "url": "https://.../invite/<token>", "expires_at": "..." }`. The previous link becomes unusable.
- `409 state_conflict`: the member has left, has an approved link, or has a pending link.

### `GET /member-links`

- Pending links: `{ "links": [ { "id", "line_display_name", "roster_member": { "id", "name" }, "created_at" } ] }`.

### `POST /member-links/{link_id}/approve` and `/reject`

- `204`. `409 state_conflict`: not pending.

## Screens and endpoints

| Screen | Endpoints |
| --- | --- |
| A1 | `GET /invite-links/{token}`, `POST /auth/line/authorize` |
| A2, A3, A4 | `GET /me` |
| R1 | `GET /races` |
| R2 | `GET /races/{id}`, `PUT .../attendance`, `GET .../running-orders`, `GET .../laps` (organizers) |
| R3 | `POST /races`, `GET /races/{id}`, `PUT /races/{id}` |
| R4, R5 | `GET/PUT/DELETE .../invitations` |
| R6 | `.../entries` |
| R7 | `GET .../running-orders`, `PUT .../running-order`, `POST .../confirm` |
| L1 | `.../my-laps` |
| P1 | `GET /me/profile` |
| M1 | `/roster-members`, `.../link`, `.../invite-links` |
| M2 | `GET /roster-members/{id}` |
| M3 | `/member-links` |

---

<details>
<summary>日本語</summary>

# MVP の API の設計

- Status: Draft
- Issue: #70
- 関連：ADR-0006（アクセスの方針）、ADR-0007（LINE ログイン）、ADR-0009（ローカル開発専用の仮ログイン）、ADR-0010（仕様を先に書く OpenAPI）、`docs/design/screens.md`、`docs/design/database.md`

MVP の画面が必要とするエンドポイントを一覧にする。OpenAPI の文書（`app/backend/api/openapi.yaml`）は、これをもとに #71 で書く。

## 決まりごと

| 項目 | ルール |
| --- | --- |
| パスの頭 | `/api/v1` |
| 認証 | セッションのトークンは HttpOnly Cookie `tasuki_session`（`Secure`、`SameSite=Lax`、`Path=/api`）で運ぶ。ページ上のスクリプトから読めないので、frontend はトークンを扱わない。frontend と API は同じサイトから出す。期限が切れたら LINE でログインし直す。リフレッシュトークンはない |
| CSRF | データを変えるリクエスト（`POST`、`PUT`、`PATCH`、`DELETE`）は、frontend のオリジンと同じ `Origin` ヘッダーを持つ必要がある。なければ `403 forbidden`。`SameSite=Lax` と合わせて、ほかのサイトから送られたリクエストを断る |
| エラー | RFC 9457 の `application/problem+json`。画面が出し分けるための `code` を足す |
| 見られないもの | 見てはいけない大会（とその下のもの）は `404`。存在しない大会と同じ |
| タイム | ラップと平均タイム：`"MM:SS"`（例：`"3:28"`）。合計タイム：`"H:MM:SS"`（例：`"2:45:30"`） |
| 日付 | 開催日：`YYYY-MM-DD`。締め切りなどの日時：オフセット付きの RFC 3339（例：`2026-11-01T18:00:00+09:00`） |
| 距離 | km の数値（例：`30.5`） |
| 同時編集 | 大会と走順は `version` を持つ。更新では読んだときの版を送り、変わっていたら `409`（`code: version_conflict`） |
| 一覧 | MVP ではページ分けしない |
| ID | 中身を意味を持たない文字列として扱う（型は #68 で決める） |
| JSON の項目名 | `snake_case` |

### 呼べる人

| 呼ぶ人 | 意味 |
| --- | --- |
| 公開 | トークンなし |
| アカウント | 有効なトークンがあれば誰でも。承認待ちや却下のアカウントも含む |
| メンバー | 名簿のメンバーとの紐付けが承認され、退会していないアカウント |
| 幹事 | 設定で幹事として並べられたアカウント。紐付けは不要 |

紐付いた幹事はメンバーでもある。下の「メンバー」のエンドポイントは、そういう幹事も呼べる。

### エラーのコード

| ステータス | `code` | いつ |
| --- | --- | --- |
| 400 | `invalid_request` | 本文やパラメータの形が正しくない |
| 401 | `unauthenticated` | セッションの Cookie がない、またはそのトークンが正しくない、期限切れ |
| 403 | `forbidden` | 役割がそのエンドポイントを許さない（例：メンバーが幹事のエンドポイントを呼ぶ、承認待ちがメンバーのエンドポイントを呼ぶ）、またはデータを変えるリクエストの `Origin` がない、ほかのサイトのもの |
| 404 | `not_found` | 存在しない、または見てはいけない |
| 404 | `invite_unusable` | 招待リンクが期限切れ、使用済み、置き換え済み |
| 409 | `version_conflict` | 読んだ後にほかの人が変えた |
| 409 | `state_conflict` | 今の状態ではその操作ができない（内容はエンドポイントごと） |
| 422 | `validation_failed` | 本文が業務のルールに反する（内容はエンドポイントごと）。`errors` に項目を並べる |

## 認証と参加

| メソッドとパス | 呼べる人 | 画面 | 目的 |
| --- | --- | --- | --- |
| `GET /invite-links/{token}` | 公開 | A1 | ログイン前に招待リンクを確かめる |
| `POST /auth/line/authorize` | 公開 | A1、ログイン | LINE ログインを始める |
| `GET /auth/line/callback` | 公開（LINE がここに戻す） | — | LINE ログインを終え、frontend に戻す |
| `POST /auth/token` | 公開 | — | 一度きりのコードをセッションに交換する |
| `POST /dev/login` | 公開、開発のときだけ | — | ローカル開発専用の仮ログイン（ADR-0009） |
| `DELETE /session` | アカウント | — | ログアウト |
| `GET /me` | アカウント | すべて | 自分は誰か、どの画面を出すか |

### `GET /invite-links/{token}`

- `200`：`{ team_name }`。名簿のメンバーの名前は返さない
- `404 invite_unusable`：期限切れ、使用済み、置き換え済み → A4

### `POST /auth/line/authorize`

- 本文：`{ invite_token? }`。招待リンクから来たときに付ける
- `200`：`{ authorization_url }`。frontend はそこへ移動する。backend は state、PKCE の verifier、招待トークンをコールバックのために持っておく

### `GET /auth/line/callback`

- LINE が `code` と `state` を付けてここに戻す。backend は ID トークンを検証する（ADR-0007）
- 招待トークンがあれば、アカウントと名簿のメンバーの紐付けを承認待ちで作り、招待リンクを使えなくする。その間にリンクが使えなくなっていたら、エラーで戻す
- frontend の `/auth/complete?code=<一度きりのコード>` か、`/auth/complete?error=<コード>` に戻す
  - `invite_unusable`：招待リンクが使えない → A4
  - `left`：アカウントの名簿のメンバーが退会している → A4（トークンは出さない）
  - `login_failed`：LINE ログインが失敗した、または取り消された
- 一度きりのコードは短い期限で、1回だけ交換できる

### `POST /auth/token`

- 本文：`{ code }`
- `200`：`{ expires_at }`。セッションの Cookie を設定する（`Set-Cookie`）。本文にトークンは入れない
- `400 invalid_request`：知らない、期限切れ、使用済みのコード

### `POST /dev/login`

- `DEV_LOGIN_ENABLED=true` のときだけある（ADR-0009）。それ以外はルートを登録しない（`404`）
- 本文：`{ line_user_id, display_name }`。レスポンスは `POST /auth/token` と同じ

### `DELETE /session`

- `204`。そのトークンは使えなくなり、セッションの Cookie を消す

### `GET /me`

形は英語の節を参照。

| `status` | 意味 | 画面 |
| --- | --- | --- |
| `approved` | 紐付けが承認済み | R1 |
| `pending` | 承認待ち | A2（幹事は幹事の画面を使える） |
| `rejected` | 最後の紐付けが却下された | A3（幹事は同上） |
| `none` | 紐付けがない | 幹事でなければ A4。幹事は幹事の画面を使う |

- `roster_member` は `status` が `approved` のときだけ入る
- 退会したメンバーにはトークンを出さない（コールバックを参照）ので、`left` の状態はない

## 大会

| メソッドとパス | 呼べる人 | 画面 |
| --- | --- | --- |
| `GET /races` | メンバー、幹事 | R1 |
| `POST /races` | 幹事 | R3 |
| `GET /races/{race_id}` | メンバー（誘われていて、不参加ではない）、幹事 | R2、R3 |
| `PUT /races/{race_id}` | 幹事 | R3 |

### 大会の説明欄

形は英語の節を参照。`type`（距離制／時間制）、開催日、締め切り（幹事だけ）、自由な文章、区ごとの距離（距離制）、制限時間と1周の距離（時間制）、目標（距離制は `total_time`、時間制は `laps`）。

### `GET /races`

- 幹事：チームのすべての大会。メンバー：誘われた大会のうち、「不参加」と答えたものを除いたもの
- 新しく登録した順
- `my_attendance`：`attending`／`undecided`／`null`（誘われていて未回答）。誘われていなければ項目を出さない（幹事）
- `running_order_confirmed`：幹事と、参加と答えたメンバーにだけ出す
- frontend は `my_attendance = attending` の大会を「参加する大会」に、それ以外を「招待」に出す

### `POST /races` と `PUT /races/{race_id}`

- 本文：`{ name, description }`（`PUT` は `version` も送る）。説明欄は種別に合っていること。距離制は区の距離が1つ以上、時間制は制限時間と1周の距離が必要
- `POST` → `201`。`PUT` → `200`、`409 version_conflict`、`422 validation_failed`
- 種別や区の数を変えると、今ある走順を合わせる（「走順：大会が変わったときの区」を参照）。確定した走順は確定のまま

### `GET /races/{race_id}`

- `version`（幹事だけ）、説明欄（締め切りは幹事だけ）、`my_attendance`（誘われていなければ出さない）、`running_order_confirmed`（幹事と参加のメンバー）、`can_submit_laps`（確定した走順の走者なら true）

## 招待と出欠

| メソッドとパス | 呼べる人 | 画面 |
| --- | --- | --- |
| `GET /races/{race_id}/invitations` | 幹事 | R4、R5 |
| `PUT /races/{race_id}/invitations/{roster_member_id}` | 幹事 | R4 |
| `DELETE /races/{race_id}/invitations/{roster_member_id}` | 幹事 | R4 |
| `PUT /races/{race_id}/attendance` | メンバー（誘われている） | R2 |

- `GET invitations`：退会していない名簿のメンバー全員と、誘われているか、その回答
- `PUT invitations/...`：誘う。`204`。誘い済みなら何もしない。`422 validation_failed`：退会している
- `DELETE invitations/...`：招待を取り消す。`204`。出欠の回答とエントリーから外し、走順のすべての区から外す。走順が確定していれば、この大会のラップを消し、残りの全ラップから平均タイムを計算し直す
- `PUT attendance`：本文 `{ status }`。`200`。`404`：誘われていない、またはもう「不参加」と答えた。「参加」から変えると、エントリーから外したときと同じことが起きる

## エントリー

| メソッドとパス | 呼べる人 | 画面 |
| --- | --- | --- |
| `GET /races/{race_id}/entries` | 幹事 | R6 |
| `POST /races/{race_id}/entries` | 幹事 | R6 |
| `PUT /races/{race_id}/entries/{entry_id}/members/{roster_member_id}` | 幹事 | R6 |
| `DELETE /races/{race_id}/entries/{entry_id}/members/{roster_member_id}` | 幹事 | R6 |

- `POST`：本文 `{ name }` → `201`。大会にもうエントリーがあれば `409 state_conflict`（MVP では1つまで）
- エントリーは名前の変更も削除もできない
- `PUT .../members/...`：`204`。出欠が「参加」でなければ `422 validation_failed`
- `DELETE .../members/...`：`204`。走順のすべての区から外す。走順が確定していれば、この大会のラップを消し、平均タイムを計算し直す

## 走順

| メソッドとパス | 呼べる人 | 画面 |
| --- | --- | --- |
| `GET /races/{race_id}/running-orders` | メンバー（参加）、幹事 | R2、R7 |
| `PUT /races/{race_id}/entries/{entry_id}/running-order` | 幹事 | R7 |
| `POST /races/{race_id}/entries/{entry_id}/running-order/confirm` | 幹事 | R7 |

- `GET`：エントリーごとに1つ。幹事は下書きと確定を、参加のメンバーは確定だけを見る（なければ空の一覧）。区ごとの走者と、予想（距離制は `total_time`、時間制は `laps`、空きの区があるか）を返す。誰も入っていなければ予想は `null`
- 予想は backend で計算する。空きの区は除く。メンバーには、ほかのメンバーの平均タイムを返さない
- `PUT`：本文 `{ version?, legs: [ { leg_number, roster_member_id | null } ] }`。最初の下書きを作るときは `version` を付けない。`200`
  - ルール（`422 validation_failed`）：距離制は区の距離と同じ数の区。時間制は1区以上で、1から欠けずに番号を振る。走者はエントリーのメンバー
  - 確定した走順は確定のまま、すぐに見える
  - 保存のたびに履歴を1件残す（`running_order_history`）
  - 確定した走順で、ある走者の区の数が減ったら、その人のこの大会のラップはすべて消え、平均タイムは計算し直される。区の数を減らさずに別の区に移すだけなら、ラップは残る
- `POST confirm`：本文 `{ version }`。`200`。空きの区があれば `422 validation_failed`。確定済みなら `409 state_conflict`

### 走順：大会が変わったときの区

`PUT /races/{race_id}` で種別や区の数が変わったとき：

- 距離制で区が減った：後ろの区を消す。増えた：後ろに空きの区を足す
- 時間制 → 距離制：区の距離の数に合わせて切るか足す。距離制 → 時間制：区はそのまま
- 消えた区には上のラップのルールが当てはまる。確定した走順は確定のままで、空きの区は未定と出す

## ラップタイム

| メソッドとパス | 呼べる人 | 画面 |
| --- | --- | --- |
| `GET /races/{race_id}/my-laps` | メンバー（確定した走順の走者） | L1 |
| `POST /races/{race_id}/my-laps` | 同上 | L1 |
| `PATCH /races/{race_id}/my-laps/{lap_id}` | 同上 | L1 |
| `DELETE /races/{race_id}/my-laps/{lap_id}` | 同上 | L1 |
| `GET /races/{race_id}/laps` | 幹事 | R2 |

- `GET my-laps`：自分のラップ（提出した順）、平均タイム、ラップ数
- `POST my-laps`：本文 `{ times: ["3:20", "3:25"] }` → `201`。平均は提出の式で計算する
- `PATCH`：本文 `{ time }` → `200`。`DELETE` → `200`。平均はその人の全ラップから計算し直す
- `404`：この大会の確定した走順の走者ではない（走者でない幹事も同じ）
- `GET laps`（幹事）：走者ごとのラップ。見るだけ

## プロフィール

| メソッドとパス | 呼べる人 | 画面 |
| --- | --- | --- |
| `GET /me/profile` | メンバー | P1 |

- `200`：名前、平均タイム、ラップ数

## 名簿と承認

| メソッドとパス | 呼べる人 | 画面 |
| --- | --- | --- |
| `GET /roster-members` | 幹事 | M1 |
| `POST /roster-members` | 幹事 | M1 |
| `GET /roster-members/{roster_member_id}` | 幹事 | M2 |
| `PATCH /roster-members/{roster_member_id}` | 幹事 | M1 |
| `DELETE /roster-members/{roster_member_id}/link` | 幹事 | M1 |
| `POST /roster-members/{roster_member_id}/invite-links` | 幹事 | M1 |
| `GET /member-links` | 幹事 | M3 |
| `POST /member-links/{link_id}/approve` | 幹事 | M3 |
| `POST /member-links/{link_id}/reject` | 幹事 | M3 |

- `GET /roster-members`：名前、退会したか、紐付けの状態（`none`／`pending`／`approved`。却下は `none`）、有効な招待リンクの期限
- `POST /roster-members`：本文 `{ name }` → `201`
- `GET /roster-members/{id}`：名前、退会したか、平均タイム、ラップ数、大会ごとのラップ
- `PATCH /roster-members/{id}`：本文 `{ left: true | false }`。名前は変えられない。`left: true` で、開催日が今日以降の大会への招待を取り消し（招待の取り消しと同じことが起きる）、ログインできなくなり、今のトークンも使えなくなる。`left: false` でまたログインできる。取り消された招待は戻らない
- `DELETE .../link`：承認済みの紐付けを外す。`204`。承認済みの紐付けがなければ `409 state_conflict`
- `POST .../invite-links`：`201`：`{ url, expires_at }`。前のリンクは使えなくなる。退会している、承認済みの紐付けがある、承認待ちの紐付けがあるときは `409 state_conflict`
- `GET /member-links`：承認待ちの紐付け（LINE の表示名、名簿のメンバー、日時）
- `POST .../approve`、`.../reject`：`204`。承認待ちでなければ `409 state_conflict`

## 画面とエンドポイント

英語の節の表を参照。

</details>
