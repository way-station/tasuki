import type { RaceCreateBody, RaceDetail, RaceForm, RaceFormErrors } from "../types/race.ts";

// The form for registering a race: one empty leg to start from.
export function emptyRaceForm(): RaceForm {
  return {
    name: "",
    type: "distance",
    raceDate: "",
    deadline: "",
    text: "",
    legDistancesKm: [""],
    targetTotalTime: "",
    timeLimitMinutes: "",
    lapDistanceKm: "",
    targetLaps: "",
  };
}

// The form for editing a race, filled with what is saved. The other type's
// fields start empty, as when registering.
export function toRaceForm(race: RaceDetail): RaceForm {
  const form: RaceForm = {
    ...emptyRaceForm(),
    name: race.name,
    type: race.detail.type,
    raceDate: race.raceDate,
    deadline: race.deadline === undefined ? "" : toDateTimeInput(race.deadline),
    text: race.text,
  };
  if (race.detail.type === "distance") {
    return {
      ...form,
      legDistancesKm: race.detail.legDistancesKm.map(String),
      targetTotalTime: race.detail.targetTotalTime ?? "",
    };
  }
  return {
    ...form,
    timeLimitMinutes: String(race.detail.timeLimitMinutes),
    lapDistanceKm: String(race.detail.lapDistanceKm),
    targetLaps: race.detail.targetLaps === undefined ? "" : String(race.detail.targetLaps),
  };
}

export type RaceBodyResult =
  | { ok: true; body: RaceCreateBody }
  | { ok: false; errors: RaceFormErrors };

// Checks the form and converts it into the request body of `POST /races`
// (`PUT /races/{race_id}` adds the version). Only the chosen type's fields are
// sent. The backend checks the same rules again (`422 validation_failed`).
export function toRaceBody(form: RaceForm): RaceBodyResult {
  const errors: RaceFormErrors = {};
  const name = form.name.trim();
  if (name === "") {
    errors.name = "Enter a name.";
  }
  if (form.raceDate === "") {
    errors.raceDate = "Enter the race date.";
  }
  const deadline = toRfc3339(form.deadline);
  if (deadline === undefined) {
    errors.deadline = "Enter the deadline.";
  }

  const description =
    form.type === "distance" ? distanceFields(form, errors) : timeFields(form, errors);
  if (Object.keys(errors).length > 0 || deadline === undefined) {
    return { ok: false, errors };
  }

  const text = form.text.trim();
  return {
    ok: true,
    body: {
      name,
      description: {
        type: form.type,
        race_date: form.raceDate,
        deadline,
        ...(text === "" ? {} : { text }),
        ...description,
      },
    },
  };
}

type TypeFields = Omit<RaceCreateBody["description"], "type" | "race_date" | "deadline" | "text">;

function distanceFields(form: RaceForm, errors: RaceFormErrors): TypeFields {
  if (form.legDistancesKm.length === 0) {
    errors.legDistancesKm = "Add at least one leg.";
  }
  const legDistancesKm = form.legDistancesKm.map((input, index) => {
    const km = positiveNumber(input);
    if (km === undefined) {
      errors[`legDistancesKm.${index}`] = "Enter a distance above 0 km.";
    }
    return km ?? 0;
  });

  const totalTime = form.targetTotalTime.trim();
  if (totalTime !== "" && !totalTimePattern.test(totalTime)) {
    errors.targetTotalTime = "Enter the time as H:MM:SS.";
  }
  return {
    leg_distances_km: legDistancesKm,
    ...(totalTime === "" ? {} : { target: { total_time: totalTime } }),
  };
}

function timeFields(form: RaceForm, errors: RaceFormErrors): TypeFields {
  const timeLimitMinutes = positiveInteger(form.timeLimitMinutes);
  if (timeLimitMinutes === undefined) {
    errors.timeLimitMinutes = "Enter the time limit in whole minutes.";
  }
  const lapDistanceKm = positiveNumber(form.lapDistanceKm);
  if (lapDistanceKm === undefined) {
    errors.lapDistanceKm = "Enter a distance above 0 km.";
  }

  const lapsInput = form.targetLaps.trim();
  const laps = positiveInteger(lapsInput);
  if (lapsInput !== "" && laps === undefined) {
    errors.targetLaps = "Enter the laps as a whole number.";
  }
  return {
    time_limit_minutes: timeLimitMinutes ?? 0,
    lap_distance_km: lapDistanceKm ?? 0,
    ...(laps === undefined ? {} : { target: { laps } }),
  };
}

// The pattern of HMmSs in the OpenAPI document.
const totalTimePattern = /^[0-9]+:[0-5][0-9]:[0-5][0-9]$/;

function positiveNumber(input: string): number | undefined {
  const value = Number(input.trim());
  return input.trim() !== "" && Number.isFinite(value) && value > 0 ? value : undefined;
}

function positiveInteger(input: string): number | undefined {
  const value = positiveNumber(input);
  return value !== undefined && Number.isInteger(value) ? value : undefined;
}

// RFC 3339 → the value of <input type="datetime-local">, in the browser's time
// zone.
function toDateTimeInput(rfc3339: string): string {
  const date = new Date(rfc3339);
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

// The value of <input type="datetime-local"> → RFC 3339 with the browser's
// offset at that time, as the API takes timestamps (docs/design/api.md).
function toRfc3339(input: string): string | undefined {
  if (!/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/.test(input)) {
    return undefined;
  }
  // A date-time without an offset is read in the browser's time zone.
  const date = new Date(input);
  if (Number.isNaN(date.getTime())) {
    return undefined;
  }
  const offsetMinutes = -date.getTimezoneOffset();
  const sign = offsetMinutes >= 0 ? "+" : "-";
  const absolute = Math.abs(offsetMinutes);
  return `${input}:00${sign}${pad(Math.floor(absolute / 60))}:${pad(absolute % 60)}`;
}

function pad(value: number): string {
  return String(value).padStart(2, "0");
}
