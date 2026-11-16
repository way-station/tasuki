import { useId, type FormEvent, type ReactNode } from "react";
import {
  blockButton,
  content,
  dangerButton,
  fieldError,
  input,
  numbers,
  primaryButton,
  regularButton,
  secondaryButton,
  smallButton,
  warningNote,
} from "../../../shared/ui/classes.ts";
import type { RaceForm, RaceFormErrors, RaceFormField, RaceType } from "../types/race.ts";
import { raceTypeLabels } from "./labels.ts";

type RaceFormViewProps = {
  title: string;
  form: RaceForm;
  errors: RaceFormErrors;
  saving: boolean;
  // Someone else saved the race first (409 version_conflict).
  conflict: boolean;
  onChange: (form: RaceForm) => void;
  onSave: () => void;
  // Drops the draft for the latest saved race, after a conflict.
  onLoadLatest: () => void;
};

const fieldLabel = "text-subheadline text-slate-600";
const unit = "shrink-0 text-slate-600";

// R3 (docs/design/screens.md, wireframes/r3.html). Shows the fields of the
// chosen type only; the other type's input is kept in the form.
export function RaceFormView({
  title,
  form,
  errors,
  saving,
  conflict,
  onChange,
  onSave,
  onLoadLatest,
}: RaceFormViewProps) {
  const set = <Key extends keyof RaceForm>(key: Key, value: RaceForm[Key]) =>
    onChange({ ...form, [key]: value });
  const setLeg = (index: number, value: string) =>
    set(
      "legDistancesKm",
      form.legDistancesKm.map((leg, legIndex) => (legIndex === index ? value : leg)),
    );
  const submit = (event: FormEvent) => {
    event.preventDefault();
    onSave();
  };

  return (
    <section className={content}>
      <h1 className="text-title1 font-bold">{title}</h1>
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
      <form onSubmit={submit} noValidate className="flex flex-col gap-4">
        <Field label="大会名" error={errors.name}>
          {(props) => (
            <input
              {...props}
              className={input}
              value={form.name}
              onChange={(event) => set("name", event.target.value)}
            />
          )}
        </Field>
        <fieldset className="flex min-w-0 flex-col gap-1">
          <legend className={`${fieldLabel} mb-1`}>種別</legend>
          <div className="grid auto-cols-fr grid-flow-col gap-2">
            {(["distance", "time"] as const satisfies RaceType[]).map((type) => (
              <label
                key={type}
                className="flex min-h-11 cursor-pointer items-center justify-center rounded-lg border border-slate-500 bg-white p-2 text-headline font-semibold has-checked:border-2 has-checked:border-blue-700 has-checked:bg-blue-100 has-checked:text-blue-700 has-focus-visible:outline-2 has-focus-visible:outline-offset-2 has-focus-visible:outline-blue-700"
              >
                <input
                  type="radio"
                  name="type"
                  value={type}
                  checked={form.type === type}
                  onChange={() => set("type", type)}
                  className="sr-only"
                />
                {raceTypeLabels[type]}
              </label>
            ))}
          </div>
        </fieldset>
        <Field label="開催日" error={errors.raceDate}>
          {(props) => (
            <input
              {...props}
              className={input}
              type="date"
              value={form.raceDate}
              onChange={(event) => set("raceDate", event.target.value)}
            />
          )}
        </Field>
        <Field label="締め切り" error={errors.deadline}>
          {(props) => (
            <input
              {...props}
              className={input}
              type="datetime-local"
              value={form.deadline}
              onChange={(event) => set("deadline", event.target.value)}
            />
          )}
        </Field>
        {form.type === "distance" ? (
          <>
            <Field label="目標タイム（時:分:秒）" error={errors.targetTotalTime}>
              {(props) => (
                <input
                  {...props}
                  className={`${input} ${numbers}`}
                  inputMode="numeric"
                  value={form.targetTotalTime}
                  onChange={(event) => set("targetTotalTime", event.target.value)}
                />
              )}
            </Field>
            <fieldset className="flex min-w-0 flex-col gap-2">
              <legend className={`${fieldLabel} mb-1`}>区の距離</legend>
              {errors.legDistancesKm !== undefined && (
                <p role="alert" className={fieldError}>
                  {errors.legDistancesKm}
                </p>
              )}
              <ol className="flex flex-col gap-2">
                {form.legDistancesKm.map((leg, index) => (
                  <li key={index}>
                    <LegField
                      leg={index + 1}
                      value={leg}
                      error={errors[legField(index)]}
                      onChange={(value) => setLeg(index, value)}
                      onRemove={() =>
                        set(
                          "legDistancesKm",
                          form.legDistancesKm.filter((_, legIndex) => legIndex !== index),
                        )
                      }
                    />
                  </li>
                ))}
              </ol>
              <button
                type="button"
                onClick={() => set("legDistancesKm", [...form.legDistancesKm, ""])}
                className={`${secondaryButton} ${regularButton}`}
              >
                区を足す
              </button>
            </fieldset>
          </>
        ) : (
          <>
            <Field label="目標周回数" error={errors.targetLaps}>
              {(props) => (
                <span className="flex items-center gap-2">
                  <input
                    {...props}
                    className={`${input} ${numbers}`}
                    inputMode="numeric"
                    value={form.targetLaps}
                    onChange={(event) => set("targetLaps", event.target.value)}
                  />
                  <span className={unit}>周</span>
                </span>
              )}
            </Field>
            <Field label="制限時間" error={errors.timeLimitMinutes}>
              {(props) => (
                <span className="flex items-center gap-2">
                  <input
                    {...props}
                    className={`${input} ${numbers}`}
                    inputMode="numeric"
                    value={form.timeLimitMinutes}
                    onChange={(event) => set("timeLimitMinutes", event.target.value)}
                  />
                  <span className={unit}>分</span>
                </span>
              )}
            </Field>
            <Field label="1周の距離" error={errors.lapDistanceKm}>
              {(props) => (
                <span className="flex items-center gap-2">
                  <input
                    {...props}
                    className={`${input} ${numbers}`}
                    inputMode="decimal"
                    value={form.lapDistanceKm}
                    onChange={(event) => set("lapDistanceKm", event.target.value)}
                  />
                  <span className={unit}>km</span>
                </span>
              )}
            </Field>
          </>
        )}
        <Field label="メモ">
          {(props) => (
            <textarea
              {...props}
              className={`${input} min-h-22 resize-y py-2.5`}
              value={form.text}
              onChange={(event) => set("text", event.target.value)}
            />
          )}
        </Field>
        <button type="submit" disabled={saving} className={`${primaryButton} ${blockButton}`}>
          {saving ? "保存中…" : "保存"}
        </button>
      </form>
    </section>
  );
}

function legField(index: number): RaceFormField {
  return `legDistancesKm.${index}`;
}

type LegFieldProps = {
  leg: number;
  value: string;
  error?: string;
  onChange: (value: string) => void;
  onRemove: () => void;
};

// One leg distance: 1区, the input, km, and 消す on one row.
function LegField({ leg, value, error, onChange, onRemove }: LegFieldProps) {
  const errorId = `${useId()}-error`;
  return (
    <div className="flex flex-col gap-1">
      <div className="flex items-center gap-2">
        <span aria-hidden="true" className="w-10 shrink-0 text-slate-600">
          {leg}区
        </span>
        <input
          aria-label={`${leg}区の距離`}
          className={`${input} ${numbers}`}
          inputMode="decimal"
          value={value}
          aria-invalid={error !== undefined}
          aria-describedby={error === undefined ? undefined : errorId}
          onChange={(event) => onChange(event.target.value)}
        />
        <span className={unit}>km</span>
        <button
          type="button"
          onClick={onRemove}
          aria-label={`${leg}区を消す`}
          className={`${dangerButton} ${smallButton} shrink-0`}
        >
          消す
        </button>
      </div>
      {error !== undefined && (
        <p id={errorId} className={fieldError}>
          {error}
        </p>
      )}
    </div>
  );
}

type FieldProps = {
  label: string;
  error?: string;
  children: (props: {
    id: string;
    "aria-invalid": boolean;
    "aria-describedby": string | undefined;
  }) => ReactNode;
};

// A labelled input with its error, linked for screen readers.
function Field({ label, error, children }: FieldProps) {
  const id = useId();
  const errorId = `${id}-error`;
  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={id} className={fieldLabel}>
        {label}
      </label>
      {children({
        id,
        "aria-invalid": error !== undefined,
        "aria-describedby": error === undefined ? undefined : errorId,
      })}
      {error !== undefined && (
        <p id={errorId} className={fieldError}>
          {error}
        </p>
      )}
    </div>
  );
}
