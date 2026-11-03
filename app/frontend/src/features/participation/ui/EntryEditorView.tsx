import { useId, type FormEvent } from "react";
import type { EntryEditor, MemberRef } from "../types/participation.ts";

type EntryEditorViewProps = {
  editor: EntryEditor;
  // The name typed for a new entry, and its error.
  newName: string;
  nameError?: string;
  busy: boolean;
  onNewNameChange: (name: string) => void;
  onCreate: () => void;
  onAdd: (member: MemberRef) => void;
  onRemove: (member: MemberRef) => void;
};

// R6 (docs/design/screens.md): create the entry, then add attending members
// and remove members.
export function EntryEditorView({
  editor,
  newName,
  nameError,
  busy,
  onNewNameChange,
  onCreate,
  onAdd,
  onRemove,
}: EntryEditorViewProps) {
  const nameId = useId();
  const submit = (event: FormEvent) => {
    event.preventDefault();
    onCreate();
  };

  if (editor.entry === null) {
    return (
      <section>
        <h1>Entry</h1>
        <form onSubmit={submit} noValidate>
          <label htmlFor={nameId}>Entry name</label>
          <input
            id={nameId}
            value={newName}
            aria-invalid={nameError !== undefined}
            aria-describedby={nameError === undefined ? undefined : `${nameId}-error`}
            onChange={(event) => onNewNameChange(event.target.value)}
          />
          {nameError !== undefined && <p id={`${nameId}-error`}>{nameError}</p>}
          <button type="submit" disabled={busy}>
            Create entry
          </button>
        </form>
      </section>
    );
  }

  return (
    <section>
      <h1>Entry</h1>
      <h2>{editor.entry.name}</h2>
      <section aria-label="Members">
        <h3>Members</h3>
        {editor.entry.members.length === 0 ? (
          <p>No members yet.</p>
        ) : (
          <ul>
            {editor.entry.members.map((member) => (
              <li key={member.rosterMemberId}>
                {member.name}{" "}
                <button type="button" disabled={busy} onClick={() => onRemove(member)}>
                  Remove {member.name}
                </button>
              </li>
            ))}
          </ul>
        )}
      </section>
      <section aria-label="Attending members">
        <h3>Attending members</h3>
        {editor.candidates.length === 0 ? (
          <p>Nobody else is attending.</p>
        ) : (
          <ul>
            {editor.candidates.map((member) => (
              <li key={member.rosterMemberId}>
                {member.name}{" "}
                <button type="button" disabled={busy} onClick={() => onAdd(member)}>
                  Add {member.name}
                </button>
              </li>
            ))}
          </ul>
        )}
      </section>
    </section>
  );
}
