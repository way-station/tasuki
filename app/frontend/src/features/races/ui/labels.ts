import type { MyAttendance, RaceType } from "../types/race.ts";

// Each answer has a mark as well as a word, so it is not told by color alone
// (docs/design/components.md, Status labels).
export const attendanceLabels: Record<MyAttendance, string> = {
  attending: "○ 参加",
  undecided: "△ 保留",
  not_attending: "× 不参加",
  not_answered: "未回答",
};

export const raceTypeLabels: Record<RaceType, string> = {
  distance: "距離制",
  time: "時間制",
};

const weekdays = ["日", "月", "火", "水", "木", "金", "土"];

// YYYY-MM-DD as 2026年12月20日（日）. The date has no time zone, so it is read
// as UTC and never moves to another day.
export function formatDate(date: string): string {
  const day = new Date(`${date}T00:00:00Z`);
  return `${day.getUTCFullYear()}年${day.getUTCMonth() + 1}月${day.getUTCDate()}日（${weekdays[day.getUTCDay()]}）`;
}

// RFC 3339 as 2026年12月6日（日）23:59, in the browser's time zone.
export function formatDateTime(rfc3339: string): string {
  const time = new Date(rfc3339);
  const hours = String(time.getHours()).padStart(2, "0");
  const minutes = String(time.getMinutes()).padStart(2, "0");
  return `${time.getFullYear()}年${time.getMonth() + 1}月${time.getDate()}日（${weekdays[time.getDay()]}）${hours}:${minutes}`;
}
