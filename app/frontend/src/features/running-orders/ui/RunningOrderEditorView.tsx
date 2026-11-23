import type {
  LegRule,
  PredictionView,
  Runner,
  RunningOrderForm,
  RunningOrderStatus,
  TimeGoal,
} from "../types/runningOrder.ts";
import {
  blockButton,
  card,
  cards,
  cardStack,
  content,
  dangerButton,
  group,
  input,
  label,
  neutralLabel,
  numbers,
  panel,
  primaryButton,
  regularButton,
  secondaryButton,
  smallButton,
  warningNote,
} from "../../../shared/ui/classes.ts";
import { noRunner, notDecided, statusColors, statusLabels } from "./labels.ts";
import { PredictionText } from "./PredictionText.tsx";

type RunningOrderEditorViewProps = {
  entryName: string;
  legRule: LegRule;
  // Time races only, shown under the prediction.
  timeGoal: TimeGoal | null;
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

// R7 (docs/design/screens.md, wireframes/r7.html). Runners are picked from the entry; time races
// add and remove legs.
export function RunningOrderEditorView({
  entryName,
  legRule,
  timeGoal,
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
    <section className={content}>
      <div className={group}>
        <h1 className="text-title1 font-bold">走順</h1>
        <p className="flex items-center gap-2">
          <span className="text-subheadline text-slate-600">{entryName}</span>
          {status !== null && (
            <span className={`${label} ${statusColors[status]}`}>{statusLabels[status]}</span>
          )}
        </p>
      </div>
      {conflict && (
        <div role="alert" className={`${warningNote} flex flex-col gap-3`}>
          <p>ほかの人が先に保存しました。</p>
          <button
            type="button"
            onClick={onLoadLatest}
            className={`${secondaryButton} ${regularButton}`}
          >
            最新を読み込む（自分の変更を捨てる）
          </button>
        </div>
      )}
      <ol aria-label="区と走者" className={cards}>
        {form.runnerIds.map((runnerId, index) => {
          const legNumber = index + 1;
          const distanceKm = legRule.type === "distance" ? legRule.distancesKm[index] : undefined;
          return (
            <li key={index} className={`${card} ${cardStack}`}>
              <div className="flex items-center gap-3">
                <span className="flex min-w-0 grow flex-col gap-0.5">
                  <span className="text-headline font-semibold">{legNumber}区</span>
                  {distanceKm !== undefined && (
                    <span className={`text-subheadline text-slate-600 ${numbers}`}>
                      {distanceKm} km
                    </span>
                  )}
                </span>
                {runnerId === null && (
                  <span className={`${label} ${neutralLabel}`}>{notDecided}</span>
                )}
                {legRule.type === "time" && form.runnerIds.length > 1 && (
                  <button
                    type="button"
                    aria-label={`${legNumber}区を消す`}
                    onClick={() => removeLeg(index)}
                    className={`${dangerButton} ${smallButton}`}
                  >
                    区を消す
                  </button>
                )}
              </div>
              <select
                aria-label={`${legNumber}区の走者`}
                className={input}
                value={runnerId ?? ""}
                onChange={(event) =>
                  setRunner(index, event.target.value === "" ? null : event.target.value)
                }
              >
                <option value="">{noRunner}</option>
                {runners.map((runner) => (
                  <option key={runner.rosterMemberId} value={runner.rosterMemberId}>
                    {runner.name}
                  </option>
                ))}
              </select>
            </li>
          );
        })}
      </ol>
      {legRule.type === "time" && (
        <button
          type="button"
          onClick={() => onChange({ runnerIds: [...form.runnerIds, null] })}
          className={`${secondaryButton} ${blockButton}`}
        >
          区を足す
        </button>
      )}
      <section aria-label="予想" className={panel}>
        <PredictionText
          prediction={prediction}
          note={unsaved && status !== null ? "（保存した走順から）" : ""}
          changesNote={
            unsaved && status !== null
              ? "保存していない変更があります。保存すると予想が更新されます。"
              : undefined
          }
        />
        {timeGoal !== null && (
          <p className={`text-subheadline text-slate-600 ${numbers}`}>{timeGoalText(timeGoal)}</p>
        )}
      </section>
      <div className={group}>
        <button
          type="button"
          onClick={onSave}
          disabled={busy}
          className={`${primaryButton} ${blockButton}`}
        >
          保存
        </button>
        {status !== "confirmed" && (
          <>
            <button
              type="button"
              onClick={onConfirm}
              disabled={busy || !canConfirm}
              className={`${primaryButton} ${blockButton}`}
            >
              確定する
            </button>
            {unsaved ? (
              <p className="text-footnote text-slate-600">確定するには、変更を保存してください。</p>
            ) : (
              !everyRunner && (
                <p className="text-footnote text-slate-600">
                  確定するには、すべての区に走者を入れてください。
                </p>
              )
            )}
          </>
        )}
      </div>
    </section>
  );
}

// 目標 75 周・制限時間 360 分・1周 2 km (wireframes/r7.html). Without a target,
// only the time limit and the lap distance.
function timeGoalText(goal: TimeGoal): string {
  const parts = [`制限時間 ${goal.timeLimitMinutes} 分`, `1周 ${goal.lapDistanceKm} km`];
  if (goal.targetLaps !== undefined) {
    parts.unshift(`目標 ${goal.targetLaps} 周`);
  }
  return parts.join("・");
}
