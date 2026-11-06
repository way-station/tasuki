import type { components } from "../../../api/schema.gen.ts";
import type {
  LegRule,
  RunningOrderEditor,
  RunningOrderListResponse,
} from "../types/runningOrder.ts";
import { toRunningOrderView } from "./toRunningOrderView.ts";

type RaceResponse = components["schemas"]["Race"];
type EntryListResponse = components["schemas"]["EntryList"];

// R7 from the race, its entries, and the saved running orders. The MVP allows
// one entry per race. An entry whose running order was never saved is not in
// the running orders (docs/design/api.md).
export function toRunningOrderEditor(
  race: RaceResponse,
  entries: EntryListResponse,
  orders: RunningOrderListResponse,
): RunningOrderEditor {
  const entry = entries.entries[0];
  if (entry === undefined) {
    return { entry: null };
  }
  const saved = orders.running_orders.find((order) => order.entry_id === entry.id);
  return {
    entry: { id: entry.id, name: entry.name },
    legRule: toLegRule(race.description),
    runners: entry.members.map((member) => ({
      rosterMemberId: member.roster_member_id,
      name: member.name,
    })),
    saved: saved === undefined ? null : toRunningOrderView(saved),
  };
}

function toLegRule(description: RaceResponse["description"]): LegRule {
  return description.type === "distance"
    ? { type: "distance", distancesKm: description.leg_distances_km ?? [] }
    : { type: "time" };
}
