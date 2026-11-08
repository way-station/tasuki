import { useRunningOrders } from "../hooks/useRunningOrders.ts";
import { RunningOrderTable } from "../ui/RunningOrderTable.tsx";

type RunningOrderSectionContainerProps = {
  raceId: string;
  // Organizers see drafts and confirmed orders; members who answered attending
  // see confirmed orders only. Nobody else may ask (docs/design/screens.md).
  canSee: boolean;
};

// The running order and prediction on R2.
export function RunningOrderSectionContainer({
  raceId,
  canSee,
}: RunningOrderSectionContainerProps) {
  const orders = useRunningOrders(raceId, canSee);
  if (!canSee || orders.data === undefined) {
    return null;
  }
  return (
    <>
      {orders.data.map((order) => (
        <RunningOrderTable key={order.entryId} order={order} />
      ))}
    </>
  );
}
