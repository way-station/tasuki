import { useId, type FormEvent } from "react";
import type { LapsForm, LapsFormErrors, MyLapsView } from "../types/lap.ts";
import { AverageText } from "./AverageText.tsx";

// The one lap being edited, in place.
export type LapEdit = {
  lapId: string;
  time: string;
  error?: string;
};

type MyLapsEditorViewProps = {
  myLaps: MyLapsView;
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

// L1 (docs/design/screens.md): the caller's laps for the race, adding several
// at once, and editing or deleting one.
export function MyLapsEditorView({
  myLaps,
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

  return (
    <section>
      <h1>My lap times</h1>
      <AverageText average={myLaps} />
      {myLaps.laps.length === 0 ? (
        <p>No lap times yet.</p>
      ) : (
        <ol aria-label="Submitted lap times">
          {myLaps.laps.map((lap, index) =>
            editing?.lapId === lap.id ? (
              <li key={lap.id}>
                <form onSubmit={saveEdit} noValidate>
                  <label htmlFor={`${id}-edit`}>Lap {index + 1}</label>
                  <input
                    id={`${id}-edit`}
                    value={editing.time}
                    aria-invalid={editing.error !== undefined}
                    aria-describedby={editing.error === undefined ? undefined : `${id}-edit-error`}
                    onChange={(event) => onEditTime(event.target.value)}
                  />
                  {editing.error !== undefined && <p id={`${id}-edit-error`}>{editing.error}</p>}
                  <button type="submit" disabled={busy}>
                    Save lap {index + 1}
                  </button>
                  <button type="button" onClick={onCancelEdit}>
                    Cancel
                  </button>
                </form>
              </li>
            ) : (
              <li key={lap.id}>
                <span>
                  Lap {index + 1}: {lap.time}
                </span>
                <button type="button" disabled={busy} onClick={() => onStartEdit(lap.id)}>
                  Edit lap {index + 1}
                </button>
                <button type="button" disabled={busy} onClick={() => onDelete(lap.id)}>
                  Delete lap {index + 1}
                </button>
              </li>
            ),
          )}
        </ol>
      )}
      <form onSubmit={submit} noValidate>
        <h2>Add lap times</h2>
        <p>Enter the time per km as M:SS, for example 3:25.</p>
        {form.times.map((time, index) => {
          const error = errors[`times.${index}`];
          const inputId = `${id}-new-${index}`;
          return (
            <div key={index}>
              <label htmlFor={inputId}>New lap {index + 1}</label>
              <input
                id={inputId}
                value={time}
                aria-invalid={error !== undefined}
                aria-describedby={error === undefined ? undefined : `${inputId}-error`}
                onChange={(event) => setTime(index, event.target.value)}
              />
              {form.times.length > 1 && (
                <button
                  type="button"
                  onClick={() =>
                    onChange({ times: form.times.filter((_, timeIndex) => timeIndex !== index) })
                  }
                >
                  Remove new lap {index + 1}
                </button>
              )}
              {error !== undefined && <p id={`${inputId}-error`}>{error}</p>}
            </div>
          );
        })}
        {errors.times !== undefined && <p>{errors.times}</p>}
        <button type="button" onClick={() => onChange({ times: [...form.times, ""] })}>
          Add another lap
        </button>
        <button type="submit" disabled={busy}>
          Submit
        </button>
      </form>
    </section>
  );
}
