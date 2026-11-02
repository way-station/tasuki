import type { RaceFormErrors } from "../types/race.ts";

// The race form breaks a rule, found before anything is sent.
export class RaceFormInvalidError extends Error {
  readonly errors: RaceFormErrors;

  constructor(errors: RaceFormErrors) {
    super("the race form is invalid");
    this.name = "RaceFormInvalidError";
    this.errors = errors;
  }
}

// The per-field errors of a failed save, if the form itself was invalid.
export function raceFormErrors(error: unknown): RaceFormErrors | undefined {
  return error instanceof RaceFormInvalidError ? error.errors : undefined;
}
