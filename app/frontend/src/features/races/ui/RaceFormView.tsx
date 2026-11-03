import { useId, type FormEvent, type ReactNode } from "react";
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

// R3 (docs/design/screens.md). Shows the fields of the chosen type only; the
// other type's input is kept in the form.
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
    <section>
      <h1>{title}</h1>
      {conflict && (
        <div role="alert">
          <p>Someone else saved this race first. Your changes are kept here.</p>
          <button type="button" onClick={onLoadLatest}>
            Load the latest (discard my changes)
          </button>
        </div>
      )}
      <form onSubmit={submit} noValidate>
        <Field label="Name" error={errors.name}>
          {(props) => (
            <input
              {...props}
              value={form.name}
              onChange={(event) => set("name", event.target.value)}
            />
          )}
        </Field>
        <fieldset>
          <legend>Type</legend>
          {(["distance", "time"] as const satisfies RaceType[]).map((type) => (
            <label key={type}>
              <input
                type="radio"
                name="type"
                value={type}
                checked={form.type === type}
                onChange={() => set("type", type)}
              />
              {raceTypeLabels[type]}
            </label>
          ))}
        </fieldset>
        <Field label="Race date" error={errors.raceDate}>
          {(props) => (
            <input
              {...props}
              type="date"
              value={form.raceDate}
              onChange={(event) => set("raceDate", event.target.value)}
            />
          )}
        </Field>
        <Field label="Deadline" error={errors.deadline}>
          {(props) => (
            <input
              {...props}
              type="datetime-local"
              value={form.deadline}
              onChange={(event) => set("deadline", event.target.value)}
            />
          )}
        </Field>
        {form.type === "distance" ? (
          <>
            <fieldset>
              <legend>Leg distances (km)</legend>
              {errors.legDistancesKm !== undefined && <p role="alert">{errors.legDistancesKm}</p>}
              <ol>
                {form.legDistancesKm.map((leg, index) => (
                  <li key={index}>
                    <Field label={`Leg ${index + 1}`} error={errors[legField(index)]}>
                      {(props) => (
                        <input
                          {...props}
                          inputMode="decimal"
                          value={leg}
                          onChange={(event) => setLeg(index, event.target.value)}
                        />
                      )}
                    </Field>
                    <button
                      type="button"
                      onClick={() =>
                        set(
                          "legDistancesKm",
                          form.legDistancesKm.filter((_, legIndex) => legIndex !== index),
                        )
                      }
                    >
                      Remove leg {index + 1}
                    </button>
                  </li>
                ))}
              </ol>
              <button
                type="button"
                onClick={() => set("legDistancesKm", [...form.legDistancesKm, ""])}
              >
                Add leg
              </button>
            </fieldset>
            <Field label="Target time (H:MM:SS, optional)" error={errors.targetTotalTime}>
              {(props) => (
                <input
                  {...props}
                  value={form.targetTotalTime}
                  onChange={(event) => set("targetTotalTime", event.target.value)}
                />
              )}
            </Field>
          </>
        ) : (
          <>
            <Field label="Time limit (minutes)" error={errors.timeLimitMinutes}>
              {(props) => (
                <input
                  {...props}
                  inputMode="numeric"
                  value={form.timeLimitMinutes}
                  onChange={(event) => set("timeLimitMinutes", event.target.value)}
                />
              )}
            </Field>
            <Field label="Lap distance (km)" error={errors.lapDistanceKm}>
              {(props) => (
                <input
                  {...props}
                  inputMode="decimal"
                  value={form.lapDistanceKm}
                  onChange={(event) => set("lapDistanceKm", event.target.value)}
                />
              )}
            </Field>
            <Field label="Target laps (optional)" error={errors.targetLaps}>
              {(props) => (
                <input
                  {...props}
                  inputMode="numeric"
                  value={form.targetLaps}
                  onChange={(event) => set("targetLaps", event.target.value)}
                />
              )}
            </Field>
          </>
        )}
        <Field label="Notes (optional)">
          {(props) => (
            <textarea
              {...props}
              value={form.text}
              onChange={(event) => set("text", event.target.value)}
            />
          )}
        </Field>
        <button type="submit" disabled={saving}>
          {saving ? "Saving…" : "Save"}
        </button>
      </form>
    </section>
  );
}

function legField(index: number): RaceFormField {
  return `legDistancesKm.${index}`;
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
    <div>
      <label htmlFor={id}>{label}</label>
      {children({
        id,
        "aria-invalid": error !== undefined,
        "aria-describedby": error === undefined ? undefined : errorId,
      })}
      {error !== undefined && <p id={errorId}>{error}</p>}
    </div>
  );
}
