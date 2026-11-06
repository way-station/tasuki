import type { LapEditBody, LapsForm, LapsFormErrors, LapsSubmitBody } from "../types/lap.ts";

// The form for adding laps on L1: one empty lap to start from.
export function emptyLapsForm(): LapsForm {
  return { times: [""] };
}

export type LapsBodyResult =
  | { ok: true; body: LapsSubmitBody }
  | { ok: false; errors: LapsFormErrors };

// Checks the form and converts it into the request body of
// `POST /races/{race_id}/my-laps`. The backend checks the same rules again
// (`422 validation_failed`).
export function toLapsBody(form: LapsForm): LapsBodyResult {
  const errors: LapsFormErrors = {};
  if (form.times.length === 0) {
    errors.times = "Add at least one lap time.";
  }
  const times = form.times.map((input, index) => {
    const time = input.trim();
    if (!lapTimePattern.test(time)) {
      errors[`times.${index}`] = lapTimeError;
    }
    return time;
  });
  return Object.keys(errors).length > 0 ? { ok: false, errors } : { ok: true, body: { times } };
}

export type LapEditBodyResult = { ok: true; body: LapEditBody } | { ok: false; error: string };

// Checks an edited lap time and converts it into the request body of
// `PATCH /races/{race_id}/my-laps/{lap_id}`.
export function toLapEditBody(input: string): LapEditBodyResult {
  const time = input.trim();
  return lapTimePattern.test(time)
    ? { ok: true, body: { time } }
    : { ok: false, error: lapTimeError };
}

// The pattern of MmSs in the OpenAPI document.
const lapTimePattern = /^[0-9]{1,2}:[0-5][0-9]$/;
const lapTimeError = "Enter the time per km as M:SS, for example 3:25.";
