import {
  group,
  groupHeading,
  label,
  neutralLabel,
  numbers,
  panel,
} from "../../../shared/ui/classes.ts";
import type { RunningOrderView } from "../types/runningOrder.ts";
import { notDecided, statusColors, statusLabels } from "./labels.ts";
import { PredictionText } from "./PredictionText.tsx";

type RunningOrderTableProps = {
  order: RunningOrderView;
};

// A saved running order on R2: the legs with their runners, the status, and
// the prediction (wireframes/r2.html).
export function RunningOrderTable({ order }: RunningOrderTableProps) {
  const hasDistance = order.legs.some((leg) => leg.distanceKm !== undefined);
  return (
    <section aria-label={`${order.entryName} の走順`} className={group}>
      <div className="flex items-center gap-2">
        <h2 className={groupHeading}>走順と予想（{order.entryName}）</h2>
        <span className={`${label} ${statusColors[order.status]}`}>
          {statusLabels[order.status]}
        </span>
      </div>
      <div className={panel}>
        <table className={`w-full border-collapse ${numbers}`}>
          <thead>
            <tr>
              <th scope="col" className={headerCell}>
                区
              </th>
              <th scope="col" className={headerCell}>
                走者
              </th>
              {hasDistance && (
                <th scope="col" className={`${headerCell} text-right`}>
                  距離
                </th>
              )}
            </tr>
          </thead>
          <tbody>
            {order.legs.map((leg) => (
              <tr key={leg.legNumber}>
                <td className={bodyCell}>{leg.legNumber}区</td>
                <td className={bodyCell}>
                  {leg.runner === null ? (
                    <span className={`${label} ${neutralLabel}`}>{notDecided}</span>
                  ) : (
                    leg.runner.name
                  )}
                </td>
                {hasDistance && (
                  <td className={`${bodyCell} text-right`}>
                    {leg.distanceKm === undefined ? "" : `${leg.distanceKm} km`}
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
        <PredictionText prediction={order.prediction} />
      </div>
    </section>
  );
}

const headerCell = "pb-1 text-left text-footnote font-normal text-slate-600";
const bodyCell = "border-t border-slate-200 py-2";
