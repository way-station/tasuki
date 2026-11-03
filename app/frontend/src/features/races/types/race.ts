import type { components } from "../../../api/schema.gen.ts";

// What the backend returns and takes (docs/design/api.md).
export type RaceListResponse = components["schemas"]["RaceList"];
export type RaceSummaryResponse = components["schemas"]["RaceSummary"];
export type RaceResponse = components["schemas"]["Race"];
export type RaceCreateBody = components["schemas"]["RaceCreate"];
export type RaceUpdateBody = components["schemas"]["RaceUpdate"];

export type RaceType = "distance" | "time";

// The caller's answer. "not_answered" is an invitation without an answer yet.
// Races answered "not attending" never reach the frontend.
export type MyAttendance = "attending" | "undecided" | "not_answered";

// A race in the list (R1).
export type RaceListItem = {
  id: string;
  name: string;
  // YYYY-MM-DD
  raceDate: string;
  // Present when the caller is invited.
  myAttendance?: MyAttendance;
  // Present for organizers, and for members who answered attending.
  runningOrderConfirmed?: boolean;
};

// R1 is laid out by role (docs/design/screens.md).
export type RaceListView =
  | {
      role: "organizer";
      // Every race of the team.
      races: RaceListItem[];
      // Races the organizer is invited to as a roster member and has not
      // answered attending.
      invitations: RaceListItem[];
    }
  | {
      role: "member";
      // Races answered attending, past ones included.
      myRaces: RaceListItem[];
      // Races not answered yet, or answered undecided.
      invitations: RaceListItem[];
    };

// What each type of race adds to the description.
export type RaceTypeDetail =
  | {
      type: "distance";
      legDistancesKm: number[];
      // H:MM:SS
      targetTotalTime?: string;
    }
  | {
      type: "time";
      timeLimitMinutes: number;
      lapDistanceKm: number;
      targetLaps?: number;
    };

// A race (R2). Optional fields are present only when the backend returns
// them to the caller.
export type RaceDetail = {
  id: string;
  name: string;
  // Organizers only.
  version?: number;
  // YYYY-MM-DD
  raceDate: string;
  // RFC 3339. Organizers only.
  deadline?: string;
  text: string;
  detail: RaceTypeDetail;
  myAttendance?: MyAttendance;
  runningOrderConfirmed?: boolean;
  canSubmitLaps: boolean;
};

// The input of the race form (R3), as typed. Both types' fields are kept, so
// switching the type back and forth loses nothing; only the chosen type's
// fields are sent.
export type RaceForm = {
  name: string;
  type: RaceType;
  // YYYY-MM-DD, from <input type="date">
  raceDate: string;
  // YYYY-MM-DDTHH:mm in the browser's time zone, from <input type="datetime-local">
  deadline: string;
  text: string;
  legDistancesKm: string[];
  // H:MM:SS, optional
  targetTotalTime: string;
  timeLimitMinutes: string;
  lapDistanceKm: string;
  // Optional
  targetLaps: string;
};

// A form field with an error. Leg distances are reported per leg.
export type RaceFormField =
  | "name"
  | "raceDate"
  | "deadline"
  | "legDistancesKm"
  | `legDistancesKm.${number}`
  | "targetTotalTime"
  | "timeLimitMinutes"
  | "lapDistanceKm"
  | "targetLaps";

export type RaceFormErrors = Partial<Record<RaceFormField, string>>;
