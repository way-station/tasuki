import type { RunningOrderStatus } from "../types/runningOrder.ts";

export const statusLabels: Record<RunningOrderStatus, string> = {
  draft: "Draft",
  confirmed: "Confirmed",
};

// An empty leg (docs/design/screens.md).
export const notDecided = "Not decided";
