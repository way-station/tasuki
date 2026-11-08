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

