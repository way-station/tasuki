import type { RunningOrderView } from "../types/runningOrder.ts";
import { notDecided, statusLabels } from "./labels.ts";
import { PredictionText } from "./PredictionText.tsx";

type RunningOrderTableProps = {
  order: RunningOrderView;
};

// A saved running order on R2: the legs with their runners, the status, and
// the prediction.
export function RunningOrderTable({ order }: RunningOrderTableProps) {
  return (
    <section aria-label={`Running order of ${order.entryName}`}>
      <h2>Running order: {order.entryName}</h2>
      <p>{statusLabels[order.status]}</p>
      <table>
        <thead>
          <tr>
            <th scope="col">Leg</th>
            <th scope="col">Distance</th>
            <th scope="col">Runner</th>
          </tr>
        </thead>
        <tbody>
          {order.legs.map((leg) => (
            <tr key={leg.legNumber}>
              <td>{leg.legNumber}</td>
              <td>{leg.distanceKm === undefined ? "" : `${leg.distanceKm} km`}</td>
              <td>{leg.runner?.name ?? notDecided}</td>
            </tr>
          ))}
        </tbody>
      </table>
      <PredictionText prediction={order.prediction} />
    </section>
  );
}
