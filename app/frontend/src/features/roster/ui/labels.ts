import type { LinkStatus } from "../types/roster.ts";

export const linkStatusLabels: Record<LinkStatus, string> = {
  none: "No user",
  pending: "Waiting for approval",
  approved: "Approved",
};

export function formatDateTime(rfc3339: string): string {
  return new Date(rfc3339).toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" });
}
