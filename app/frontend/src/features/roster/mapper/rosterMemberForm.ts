import type {
  RosterMemberCreateBody,
  RosterMemberForm,
  RosterMemberFormErrors,
} from "../types/roster.ts";

export function emptyRosterMemberForm(): RosterMemberForm {
  return { name: "" };
}

export type RosterMemberBodyResult =
  | { ok: true; body: RosterMemberCreateBody }
  | { ok: false; errors: RosterMemberFormErrors };

// Checks the form and converts it into the request body of
// `POST /roster-members`. The backend checks the same rule again.
export function toRosterMemberBody(form: RosterMemberForm): RosterMemberBodyResult {
  const name = form.name.trim();
  return name === ""
    ? { ok: false, errors: { name: "Enter a name." } }
    : { ok: true, body: { name } };
}
