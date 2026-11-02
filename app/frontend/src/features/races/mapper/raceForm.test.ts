import { describe, expect, it } from "vitest";
import type { RaceDetail, RaceForm } from "../types/race.ts";
import { emptyRaceForm, toRaceBody, toRaceForm } from "./raceForm.ts";

function form(overrides: Partial<RaceForm>): RaceForm {
  return {
    ...emptyRaceForm(),
    name: "Hakone",
    raceDate: "2026-11-23",
    deadline: "2026-11-01T18:00",
    legDistancesKm: ["5", "3.2"],
    ...overrides,
  };
}

function bodyOf(input: RaceForm) {
  const result = toRaceBody(input);
  if (!result.ok) {
    throw new Error(`invalid form: ${JSON.stringify(result.errors)}`);
  }
  return result.body;
}

function errorsOf(input: RaceForm) {
  const result = toRaceBody(input);
  if (result.ok) {
    throw new Error("the form passed");
  }
  return result.errors;
}

describe("toRaceBody", () => {
  it("converts a distance race, sending only its own fields", () => {
    const body = bodyOf(
      form({
        name: "  Hakone  ",
        text: "Meet at 7:00",
        targetTotalTime: "2:45:00",
        timeLimitMinutes: "180",
        lapDistanceKm: "2",
      }),
    );

    expect(body).toEqual({
      name: "Hakone",
      description: {
        type: "distance",
        race_date: "2026-11-23",
        deadline: body.description.deadline,
        text: "Meet at 7:00",
        leg_distances_km: [5, 3.2],
        target: { total_time: "2:45:00" },
      },
    });
  });

  it("converts a time race, sending only its own fields", () => {
    const body = bodyOf(
      form({ type: "time", timeLimitMinutes: "180", lapDistanceKm: "2", targetLaps: "30" }),
    );

    expect(body.description).toEqual({
      type: "time",
      race_date: "2026-11-23",
      deadline: body.description.deadline,
      time_limit_minutes: 180,
      lap_distance_km: 2,
      target: { laps: 30 },
    });
  });

  it("leaves out an empty text and target", () => {
    const body = bodyOf(form({ text: "  " }));

    expect(body.description).not.toHaveProperty("text");
    expect(body.description).not.toHaveProperty("target");
  });

  it("sends the deadline as RFC 3339 with the browser's offset", () => {
    const deadline = bodyOf(form({ deadline: "2026-11-01T18:00" })).description.deadline ?? "";

    expect(deadline).toMatch(/^2026-11-01T18:00:00[+-]\d{2}:\d{2}$/);
    expect(new Date(deadline).getTime()).toBe(new Date("2026-11-01T18:00").getTime());
  });

  it("requires the name, the race date, and the deadline", () => {
    expect(errorsOf(form({ name: " ", raceDate: "", deadline: "" }))).toEqual({
      name: "Enter a name.",
      raceDate: "Enter the race date.",
      deadline: "Enter the deadline.",
    });
  });

  it("requires a distance race to have legs above 0 km", () => {
    expect(errorsOf(form({ legDistancesKm: [] }))).toEqual({
      legDistancesKm: "Add at least one leg.",
    });
    expect(errorsOf(form({ legDistancesKm: ["5", "", "0", "-1", "abc"] }))).toEqual({
      "legDistancesKm.1": "Enter a distance above 0 km.",
      "legDistancesKm.2": "Enter a distance above 0 km.",
      "legDistancesKm.3": "Enter a distance above 0 km.",
      "legDistancesKm.4": "Enter a distance above 0 km.",
    });
  });

  it("requires a target time as H:MM:SS", () => {
    expect(errorsOf(form({ targetTotalTime: "2:5:00" }))).toEqual({
      targetTotalTime: "Enter the time as H:MM:SS.",
    });
  });

  it("requires a time race to have a whole time limit and a lap distance", () => {
    expect(
      errorsOf(
        form({ type: "time", timeLimitMinutes: "90.5", lapDistanceKm: "", targetLaps: "2.5" }),
      ),
    ).toEqual({
      timeLimitMinutes: "Enter the time limit in whole minutes.",
      lapDistanceKm: "Enter a distance above 0 km.",
      targetLaps: "Enter the laps as a whole number.",
    });
  });

  it("does not check the fields of the other type", () => {
    expect(toRaceBody(form({ type: "distance", timeLimitMinutes: "abc" })).ok).toBe(true);
    expect(
      toRaceBody(
        form({
          type: "time",
          legDistancesKm: ["abc"],
          timeLimitMinutes: "180",
          lapDistanceKm: "2",
        }),
      ).ok,
    ).toBe(true);
  });
});

describe("toRaceForm", () => {
  const distanceRace: RaceDetail = {
    id: "race-1",
    name: "Hakone",
    version: 3,
    raceDate: "2026-11-23",
    deadline: "2026-11-01T18:00:00+09:00",
    text: "Meet at 7:00",
    detail: { type: "distance", legDistancesKm: [5, 3.2], targetTotalTime: "2:45:00" },
    canSubmitLaps: false,
  };

  it("fills the form with a distance race", () => {
    expect(toRaceForm(distanceRace)).toMatchObject({
      name: "Hakone",
      type: "distance",
      raceDate: "2026-11-23",
      text: "Meet at 7:00",
      legDistancesKm: ["5", "3.2"],
      targetTotalTime: "2:45:00",
      timeLimitMinutes: "",
      lapDistanceKm: "",
      targetLaps: "",
    });
  });

  it("fills the form with a time race", () => {
    const race: RaceDetail = {
      ...distanceRace,
      detail: { type: "time", timeLimitMinutes: 180, lapDistanceKm: 2, targetLaps: 30 },
    };

    expect(toRaceForm(race)).toMatchObject({
      type: "time",
      legDistancesKm: [""],
      timeLimitMinutes: "180",
      lapDistanceKm: "2",
      targetLaps: "30",
    });
  });

  it("keeps the deadline's moment through the form and back", () => {
    const deadline = bodyOf(toRaceForm(distanceRace)).description.deadline ?? "";

    expect(new Date(deadline).getTime()).toBe(new Date("2026-11-01T18:00:00+09:00").getTime());
  });
});
