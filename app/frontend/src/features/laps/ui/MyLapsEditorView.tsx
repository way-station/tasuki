import { useId, type FormEvent } from "react";
import type { LapsForm, LapsFormErrors, MyLapsView, RunnerLeg } from "../types/lap.ts";
import { legsText } from "./legsText.ts";
import {
  blockButton,
  card,
  cards,
  cardStack,
  content,
  dangerButton,
  fieldError,
  fieldHelp,
  group,
  groupHeading,
  input,
  numbers,
  panel,
  primaryButton,
  secondaryButton,
  smallButton,
} from "../../../shared/ui/classes.ts";

// The one lap being edited, in place.
export type LapEdit = {
  lapId: string;
  time: string;
  error?: string;
};

type MyLapsEditorViewProps = {
  myLaps: MyLapsView;
  // The caller's legs in the running order, shown under the title.
  legs: RunnerLeg[];
  form: LapsForm;
  errors: LapsFormErrors;
  editing: LapEdit | null;
  busy: boolean;
  onChange: (form: LapsForm) => void;
  onSubmit: () => void;
  onStartEdit: (lapId: string) => void;
  onEditTime: (time: string) => void;
  onSaveEdit: () => void;
  onCancelEdit: () => void;
  onDelete: (lapId: string) => void;
};

// L1 (docs/design/screens.md, wireframes/l1.html): the caller's laps for the
// race, adding several at once, and editing or deleting one. The list is left
// out until there is a lap.
export function MyLapsEditorView({
  myLaps,
  legs,
  form,
  errors,
  editing,
  busy,
  onChange,
  onSubmit,
  onStartEdit,
  onEditTime,
  onSaveEdit,
  onCancelEdit,
  onDelete,
}: MyLapsEditorViewProps) {
  const id = useId();
  const setTime = (index: number, value: string) =>
    onChange({ times: form.times.map((time, timeIndex) => (timeIndex === index ? value : time)) });
  const submit = (event: FormEvent) => {
    event.preventDefault();
    onSubmit();
  };
  const saveEdit = (event: FormEvent) => {
    event.preventDefault();
    onSaveEdit();
  };

  // Laps are per km, so the n-th lap is the n-th km (wireframes/l1.html).
  const firstNewKm = myLaps.laps.length + 1;

  return (
    <section className={content}>
      <div className={group}>
        <h1 className="text-title1 font-bold">自分のラップタイム</h1>
        {legs.length > 0 && (
          <p className={`text-subheadline text-slate-600 ${numbers}`}>{legsText(legs, true)}</p>
        )}
      </div>
      {myLaps.laps.length > 0 && (
        <section className={group}>
          <h2 className={groupHeading}>送ったラップタイム（1km あたり）</h2>
          <ol aria-label="送ったラップタイム" className={cards}>
            {myLaps.laps.map((lap, index) =>
              editing?.lapId === lap.id ? (
                <li key={lap.id} className={`${card} ${cardStack}`}>
                  <form onSubmit={saveEdit} noValidate className="flex flex-col gap-1">
                    <label htmlFor={`${id}-edit`} className="text-subheadline text-slate-600">
                      {index + 1} km
                    </label>
                    <div className="flex items-center gap-2">
                      <input
                        id={`${id}-edit`}
                        className={`${input} ${numbers}`}
                        inputMode="numeric"
                        value={editing.time}
                        aria-invalid={editing.error !== undefined}
                        aria-describedby={
                          editing.error === undefined ? undefined : `${id}-edit-error`
                        }
                        onChange={(event) => onEditTime(event.target.value)}
                      />
                      <button
                        type="submit"
                        disabled={busy}
                        aria-label={`${index + 1} km のラップタイムを保存`}
                        className={`${primaryButton} ${smallButton} shrink-0`}
                      >
                        保存
                      </button>
                      <button
                        type="button"
                        onClick={onCancelEdit}
                        className={`${secondaryButton} ${smallButton} shrink-0`}
                      >
                        やめる
                      </button>
                    </div>
                    {editing.error !== undefined && (
                      <p id={`${id}-edit-error`} className={fieldError}>
                        {editing.error}
                      </p>
                    )}
                  </form>
                </li>
              ) : (
                <li key={lap.id} className={card}>
                  <span className="flex min-w-0 grow flex-col gap-0.5">
                    <span className="text-subheadline text-slate-600">{index + 1} km</span>
                    <span className={`text-headline font-semibold ${numbers}`}>{lap.time}</span>
                  </span>
                  <span className="flex flex-wrap gap-2">
                    <button
                      type="button"
                      disabled={busy}
                      aria-label={`${index + 1} km のラップタイムを直す`}
                      className={`${secondaryButton} ${smallButton}`}
                      onClick={() => onStartEdit(lap.id)}
                    >
                      直す
                    </button>
                    <button
                      type="button"
                      disabled={busy}
                      aria-label={`${index + 1} km のラップタイムを消す`}
                      className={`${dangerButton} ${smallButton}`}
                      onClick={() => onDelete(lap.id)}
                    >
                      消す
                    </button>
                  </span>
                </li>
              ),
            )}
          </ol>
        </section>
      )}
      <form onSubmit={submit} noValidate className={group}>
        <h2 className={groupHeading}>ラップタイムを足す</h2>
        <p id={`${id}-help`} className={fieldHelp}>
          1km あたりのタイムを 分:秒 で入れてください（例：3:25）
        </p>
        <div className={panel}>
          {form.times.map((time, index) => {
            const error = errors[`times.${index}`];
            const inputId = `${id}-new-${index}`;
            const km = firstNewKm + index;
            return (
              <div key={index} className="flex flex-col gap-1">
                <div className="flex items-center gap-2">
                  <label htmlFor={inputId} className="w-12 shrink-0 text-slate-600">
                    {km} km
                  </label>
                  <input
                    id={inputId}
                    className={`${input} ${numbers}`}
                    inputMode="numeric"
                    placeholder="分:秒"
                    value={time}
                    aria-invalid={error !== undefined}
                    aria-describedby={
                      error === undefined ? `${id}-help` : `${inputId}-error ${id}-help`
                    }
                    onChange={(event) => setTime(index, event.target.value)}
                  />
                  {form.times.length > 1 && (
                    <button
                      type="button"
                      aria-label={`${km} km の行を消す`}
                      className={`${dangerButton} ${smallButton} shrink-0`}
                      onClick={() =>
                        onChange({
                          times: form.times.filter((_, timeIndex) => timeIndex !== index),
                        })
                      }
                    >
                      消す
                    </button>
                  )}
                </div>
                {error !== undefined && (
                  <p id={`${inputId}-error`} className={fieldError}>
                    {error}
                  </p>
                )}
              </div>
            );
          })}
          {errors.times !== undefined && <p className={fieldError}>{errors.times}</p>}
          <button
            type="button"
            className={`${secondaryButton} ${blockButton}`}
            onClick={() => onChange({ times: [...form.times, ""] })}
          >
            行を足す
          </button>
        </div>
        <button type="submit" disabled={busy} className={`${primaryButton} ${blockButton}`}>
          送る
        </button>
        <p className="text-footnote text-slate-600">
          送るまでは、この端末に下書きとして残ります。送るたびに平均タイムを計算し直します。
        </p>
      </form>
    </section>
  );
}
