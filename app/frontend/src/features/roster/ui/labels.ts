import type { LinkStatus } from "../types/roster.ts";
import { confirmedLabel, neutralLabel, waitingLabel } from "../../../shared/ui/classes.ts";

export const linkStatusLabels: Record<LinkStatus, string> = {
  none: "ユーザーなし",
  pending: "承認待ち",
  approved: "承認済み",
};

// The colors of each label (docs/design/components.md, Status labels).
export const linkStatusColors: Record<LinkStatus, string> = {
  none: neutralLabel,
  pending: waitingLabel,
  approved: confirmedLabel,
};

const weekdays = ["日", "月", "火", "水", "木", "金", "土"];

// RFC 3339 as 11月22日（日）18:00, in the browser's time zone. Invite links
// and logins are recent, so the year is left out (wireframes/m1.html, m3.html).
export function formatDateTime(rfc3339: string): string {
  const time = new Date(rfc3339);
  const minutes = String(time.getMinutes()).padStart(2, "0");
  return `${time.getMonth() + 1}月${time.getDate()}日（${weekdays[time.getDay()]}）${time.getHours()}:${minutes}`;
}
