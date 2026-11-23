import { useRunningOrders } from "../../running-orders/hooks/useRunningOrders.ts";
import { runnerLegs } from "../mapper/runnerLegs.ts";
import type { RunnerLeg } from "../types/lap.ts";

// The legs each runner runs in the race's running order (L1, R2), read from
// the running order the caller may already see. Pass `enabled` only for
// callers who may ask: organizers, and runners of a confirmed order.
export function useRunnerLegs(
  raceId: string,
  enabled: boolean,
): (rosterMemberId: string) => RunnerLeg[] {
  const orders = useRunningOrders(raceId, enabled);
  // The MVP allows one entry per race, so a race has at most one order.
  const legs = orders.data?.[0]?.legs ?? [];
  return (rosterMemberId) => runnerLegs(legs, rosterMemberId);
}
