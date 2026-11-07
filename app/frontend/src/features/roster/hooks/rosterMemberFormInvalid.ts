import { emptyRosterMemberForm } from "../mapper/rosterMemberForm.ts";
import type { RosterMemberForm, RosterMemberFormErrors } from "../types/roster.ts";

// The roster member being added breaks a rule, found before anything is sent.
export class RosterMemberFormInvalidError extends Error {
  readonly errors: RosterMemberFormErrors;

  constructor(errors: RosterMemberFormErrors) {
    super("the roster member form is invalid");
    this.name = "RosterMemberFormInvalidError";
    this.errors = errors;
  }
}

// The per-field errors of a failed add, if the form itself was invalid.
export function rosterMemberFormErrors(error: unknown): RosterMemberFormErrors | undefined {
  return error instanceof RosterMemberFormInvalidError ? error.errors : undefined;
}

// The form M1 starts from. Containers may not import the Mapper, so this
// reaches them here.
export function newRosterMemberForm(): RosterMemberForm {
  return emptyRosterMemberForm();
}
