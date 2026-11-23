import { numbers } from "../../../shared/ui/classes.ts";
import type { PredictionView } from "../types/runningOrder.ts";

type PredictionTextProps = {
  prediction: PredictionView | null;
  // Shown after the name, for example （保存した走順から） on R7.
  note?: string;
  // Shown in the note box before the empty legs, for example that there are
  // unsaved changes on R7.
  changesNote?: string;
};

// The prediction as a large number, which leaves empty legs out and names them
// (docs/design/screens.md, wireframes/r2.html and r7.html).
export function PredictionText({ prediction, note = "", changesNote }: PredictionTextProps) {
  if (prediction === null) {
    return (
      <>
        <p className="text-subheadline text-slate-600">
          まだ予想はありません。区に走者を入れて保存すると出ます。
        </p>
        <NoteBox notes={changesNote === undefined ? [] : [changesNote]} />
      </>
    );
  }
  const emptyLegs = prediction.emptyLegNumbers.map((legNumber) => `${legNumber}区`).join("・");
  const notes = changesNote === undefined ? [] : [changesNote];
  if (emptyLegs !== "") {
    // After the changes note the sentence goes on from it (wireframes/r7.html).
    notes.push(
      changesNote === undefined
        ? `空きの区（${emptyLegs}）を除いた予想です。`
        : `空きの区（${emptyLegs}）は除いています。`,
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
      <NoteBox notes={notes} />
    </>
  );
}

// The notes under the prediction in one box, or nothing.
function NoteBox({ notes }: { notes: string[] }) {
  if (notes.length === 0) {
    return null;
  }
  return (
    <p className="rounded-lg bg-slate-100 px-4 py-3 text-subheadline text-slate-700">
      {notes.join("")}
    </p>
  );
}
