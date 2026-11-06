import type { ProfileResponse, ProfileView } from "../types/lap.ts";

// P1.
export function toProfileView(profile: ProfileResponse): ProfileView {
  return {
    name: profile.name,
    averageTime: profile.average_time,
    lapCount: profile.lap_count,
  };
}
