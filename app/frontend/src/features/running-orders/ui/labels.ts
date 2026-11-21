import type { RunningOrderStatus } from "../types/runningOrder.ts";
import { confirmedLabel, neutralLabel } from "../../../shared/ui/classes.ts";

export const statusLabels: Record<RunningOrderStatus, string> = {
  draft: "下書き",
  confirmed: "走順確定",
};

// The colors of each label (docs/design/components.md, Status labels).
export const statusColors: Record<RunningOrderStatus, string> = {
  draft: neutralLabel,
  confirmed: confirmedLabel,
};

// An empty leg (docs/design/screens.md).
export const notDecided = "未定";
// The choice that leaves a leg empty on R7.
export const noRunner = "選ばない";
