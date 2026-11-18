import { useId, type FormEvent } from "react";
import type { EntryEditor, MemberRef } from "../types/participation.ts";
import {
  blockButton,
  card,
  cards,
  content,
  dangerButton,
  fieldError,
  group,
  groupHeading,
  input,
  primaryButton,
  smallButton,
} from "../../../shared/ui/classes.ts";

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

const title = "text-title1 font-bold";

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
      <section className={content}>
        <h1 className={title}>エントリー</h1>
        <form onSubmit={submit} noValidate className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <label htmlFor={nameId} className="text-subheadline text-slate-600">
              エントリーの名前
            </label>
            <input
              id={nameId}
              value={newName}
              placeholder="例：A チーム"
              aria-invalid={nameError !== undefined}
              aria-describedby={nameError === undefined ? undefined : `${nameId}-error`}
              onChange={(event) => onNewNameChange(event.target.value)}
              className={input}
            />
            {nameError !== undefined && (
              <p id={`${nameId}-error`} className={fieldError}>
                {nameError}
              </p>
            )}
          </div>
          <button type="submit" disabled={busy} className={`${primaryButton} ${blockButton}`}>
            エントリーを作る
          </button>
        </form>
      </section>
    );
  }

  const { members } = editor.entry;
  return (
    <section className={content}>
      <div className={group}>
        <h1 className={title}>エントリー</h1>
        <p className="text-subheadline text-slate-600">{editor.entry.name}</p>
      </div>
      <section aria-label="エントリーのメンバー" className={group}>
        <h2 className={groupHeading}>エントリーのメンバー（{members.length}）</h2>
        {members.length === 0 ? (
          <p className="text-subheadline text-slate-600">まだメンバーはいません</p>
        ) : (
          <ul className={cards}>
            {members.map((member) => (
              <li key={member.rosterMemberId} className={card}>
                <span className="min-w-0 grow text-headline font-semibold">{member.name}</span>
                <button
                  type="button"
                  disabled={busy}
                  onClick={() => onRemove(member)}
                  aria-label={`${member.name} さんを外す`}
                  className={`${dangerButton} ${smallButton} shrink-0`}
                >
                  外す
                </button>
              </li>
            ))}
          </ul>
        )}
      </section>
      <section aria-label="入れられるメンバー" className={group}>
        <h2 className={groupHeading}>入れられるメンバー（出欠が「参加」）</h2>
        {editor.candidates.length === 0 ? (
          <p className="text-subheadline text-slate-600">ほかに「参加」の人はいません</p>
        ) : (
          <ul className={cards}>
            {editor.candidates.map((member) => (
              <li key={member.rosterMemberId} className={card}>
                <span className="min-w-0 grow text-headline font-semibold">{member.name}</span>
                <button
                  type="button"
                  disabled={busy}
                  onClick={() => onAdd(member)}
                  aria-label={`${member.name} さんを入れる`}
                  className={`${primaryButton} ${smallButton} shrink-0`}
                >
                  入れる
                </button>
              </li>
            ))}
          </ul>
        )}
      </section>
    </section>
  );
}
