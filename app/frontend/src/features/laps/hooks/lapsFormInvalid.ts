import type { LapsFormErrors } from "../types/lap.ts";

// The lap times being added break a rule, found before anything is sent.
export class LapsFormInvalidError extends Error {
  readonly errors: LapsFormErrors;

  constructor(errors: LapsFormErrors) {
    super("the lap times are invalid");
    this.name = "LapsFormInvalidError";
    this.errors = errors;
  }
}

// An edited lap time is not M:SS, found before anything is sent.
export class LapTimeInvalidError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "LapTimeInvalidError";
  }
}

// The per-time errors of a failed submit, if the form itself was invalid.
export function lapsFormErrors(error: unknown): LapsFormErrors | undefined {
  return error instanceof LapsFormInvalidError ? error.errors : undefined;
}

// The error of a failed edit, if the time itself was invalid.
export function lapTimeError(error: unknown): string | undefined {
  return error instanceof LapTimeInvalidError ? error.message : undefined;
}
