import type { LapView } from "../types/lap.ts";

// The mean of the laps as MM:SS, rounded half up to the second like the
// backend's average (docs/design/database.md). `null` without laps.
export function averageTime(laps: LapView[]): string | null {
  if (laps.length === 0) {
    return null;
  }
  const total = laps.reduce((sum, lap) => sum + toSeconds(lap.time), 0);
  const seconds = Math.floor(total / laps.length + 0.5);
  return `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, "0")}`;
}

function toSeconds(time: string): number {
  const [minutes, seconds] = time.split(":").map(Number);
  return minutes * 60 + seconds;
}
