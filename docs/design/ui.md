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
