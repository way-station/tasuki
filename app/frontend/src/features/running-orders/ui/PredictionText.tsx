import { numbers } from "../../../shared/ui/classes.ts";
import type { PredictionView } from "../types/runningOrder.ts";

type PredictionTextProps = {
  prediction: PredictionView | null;
  // Shown after the name, for example （保存した走順から） on R7.
  note?: string;
};

// The prediction as a large number, which leaves empty legs out
// (docs/design/screens.md, wireframes/r2.html and r7.html).
export function PredictionText({ prediction, note = "" }: PredictionTextProps) {
  if (prediction === null) {
    return (
      <p className="text-subheadline text-slate-600">
        まだ予想はありません。区に走者を入れて保存すると出ます。
      </p>
    );
  }
  return (
    <>
      <p className="flex items-center gap-3">
        <span className="grow text-subheadline text-slate-600">
          {prediction.type === "distance" ? "予想タイム" : "予想周回数"}
          {note}
        </span>
        <span className={`text-title2 font-bold ${numbers}`}>
          {prediction.type === "distance" ? prediction.totalTime : `${prediction.laps} 周`}
        </span>
      </p>
      {prediction.hasEmptyLegs && (
        <p className="rounded-lg bg-slate-100 px-4 py-3 text-subheadline text-slate-700">
          空きの区を除いた予想です。
        </p>
      )}
    </>
  );
}
