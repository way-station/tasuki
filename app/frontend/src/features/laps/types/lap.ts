import type { components } from "../../../api/schema.gen.ts";

// What the backend returns and takes (docs/design/api.md).
export type MyLapsResponse = components["schemas"]["MyLaps"];
export type LapsSubmitBody = components["schemas"]["LapsSubmit"];
export type LapEditBody = components["schemas"]["LapEdit"];
export type RaceLapsResponse = components["schemas"]["RaceLaps"];
export type ProfileResponse = components["schemas"]["Profile"];

// A lap time per km.
export type LapView = {
  id: string;
  // MM:SS
  time: string;
};

// The caller's average over all their laps, written back to the roster member
// (docs/design/database.md).
export type AverageView = {
  // MM:SS. `3:40` while no lap is submitted.
  averageTime: string;
  lapCount: number;
};

// L1: the caller's laps for one race, in the order they were submitted, and
// their average after them.
export type MyLapsView = AverageView & {
  laps: LapView[];
};

// The laps of one runner, as organizers see them on R2. View only.
export type RunnerLapsView = {
  rosterMemberId: string;
  name: string;
  laps: LapView[];
};

// P1. Members enter nothing here.
export type ProfileView = AverageView & {
  name: string;
};

// The lap times the caller is adding on L1, as typed (MM:SS). They are sent
// together.
export type LapsForm = {
  times: string[];
};

// An error per lap time, by its position in the form.
export type LapsFormErrors = Partial<Record<`times.${number}` | "times", string>>;
