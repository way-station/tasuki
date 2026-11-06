import type {
  LegRule,
  PredictionView,
  Runner,
  RunningOrderForm,
  RunningOrderStatus,
} from "../types/runningOrder.ts";
import { notDecided, statusLabels } from "./labels.ts";
import { PredictionText } from "./PredictionText.tsx";

type RunningOrderEditorViewProps = {
  entryName: string;
  legRule: LegRule;
  runners: Runner[];
  form: RunningOrderForm;
  // `null` before the first save.
  status: RunningOrderStatus | null;
  prediction: PredictionView | null;
  unsaved: boolean;
  everyRunner: boolean;
  canConfirm: boolean;
  busy: boolean;
  // Someone else saved the running order first (409 version_conflict).
  conflict: boolean;
  onChange: (form: RunningOrderForm) => void;
  onSave: () => void;
  onConfirm: () => void;
  // Drops the draft for the latest saved running order, after a conflict.
  onLoadLatest: () => void;
};

// R7 (docs/design/screens.md). Runners are picked from the entry; time races
// add and remove legs.
export function RunningOrderEditorView({
  entryName,
  legRule,
  runners,
  form,
  status,
  prediction,
  unsaved,
  everyRunner,
  canConfirm,
  busy,
  conflict,
  onChange,
  onSave,
  onConfirm,
  onLoadLatest,
}: RunningOrderEditorViewProps) {
  const setRunner = (index: number, runnerId: string | null) =>
    onChange({
      runnerIds: form.runnerIds.map((current, legIndex) =>
        legIndex === index ? runnerId : current,
      ),
    });
  const removeLeg = (index: number) =>
    onChange({ runnerIds: form.runnerIds.filter((_, legIndex) => legIndex !== index) });

  return (
    <section>
      <h1>Running order</h1>
      <p>
        {entryName} · {status === null ? "Not saved yet" : statusLabels[status]}
      </p>
      {conflict && (
        <div role="alert">
          <p>Someone else saved this running order first. Your changes are kept here.</p>
          <button type="button" onClick={onLoadLatest}>
            Load the latest (discard my changes)
          </button>
        </div>
      )}
      <ol>
        {form.runnerIds.map((runnerId, index) => {
          const legNumber = index + 1;
          const distanceKm = legRule.type === "distance" ? legRule.distancesKm[index] : undefined;
          return (
            <li key={index}>
              <label>
                Leg {legNumber}
                {distanceKm !== undefined && ` (${distanceKm} km)`}
                <select
                  value={runnerId ?? ""}
                  onChange={(event) =>
                    setRunner(index, event.target.value === "" ? null : event.target.value)
                  }
                >
                  <option value="">{notDecided}</option>
                  {runners.map((runner) => (
                    <option key={runner.rosterMemberId} value={runner.rosterMemberId}>
                      {runner.name}
                    </option>
                  ))}
                </select>
              </label>
              {legRule.type === "time" && form.runnerIds.length > 1 && (
                <button type="button" onClick={() => removeLeg(index)}>
                  Remove leg {legNumber}
                </button>
              )}
            </li>
          );
        })}
      </ol>
      {legRule.type === "time" && (
        <button type="button" onClick={() => onChange({ runnerIds: [...form.runnerIds, null] })}>
          Add leg
        </button>
      )}
      <PredictionText prediction={prediction} />
      {unsaved && status !== null && (
        <p>This is the prediction of the saved order. Saving updates it.</p>
      )}
      <button type="button" onClick={onSave} disabled={busy}>
        Save
      </button>
      {status !== "confirmed" && (
        <>
          <button type="button" onClick={onConfirm} disabled={busy || !canConfirm}>
            Confirm
          </button>
          {unsaved ? (
            <p>Save your changes before confirming.</p>
          ) : (
            !everyRunner && <p>Every leg needs a runner before confirming.</p>
          )}
        </>
      )}
    </section>
  );
}
