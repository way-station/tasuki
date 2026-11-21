import { useParams } from "@tanstack/react-router";
import { useState } from "react";
import { errorCode } from "../../../shared/errors/apiErrors.ts";
import { useNotifications } from "../../../shared/notification/useNotifications.ts";
import { useAddEntryMember, useCreateEntry, useRemoveEntryMember } from "../hooks/mutations.ts";
import { useEntryEditor } from "../hooks/useEntryEditor.ts";
import type { MemberRef } from "../types/participation.ts";
import { Confirmation } from "../../../shared/ui/Confirmation.tsx";
import { EntryEditorView } from "../ui/EntryEditorView.tsx";
import { ParticipationMessage } from "../ui/ParticipationMessage.tsx";

// R6 (docs/design/screens.md).
export function EntryContainer() {
  const { raceId } = useParams({ from: "/team/races/$raceId" });
  const editor = useEntryEditor(raceId);
  const createEntry = useCreateEntry(raceId);
  const addMember = useAddEntryMember(raceId);
  const removeMember = useRemoveEntryMember(raceId);
  const { notify } = useNotifications();
  const [newName, setNewName] = useState("");
  const [nameError, setNameError] = useState<string>();
  const [removing, setRemoving] = useState<MemberRef | null>(null);

  if (editor.error !== null) {
    return (
      <ParticipationMessage
        title="うまくいきませんでした"
        message="ページを読み込み直してください。"
      />
    );
  }
  if (editor.data === undefined) {
    return <ParticipationMessage title="読み込み中…" />;
  }
  const entry = editor.data.entry;

  const create = async () => {
    const name = newName.trim();
    if (name === "") {
      setNameError("名前を入れてください。");
      return;
    }
    setNameError(undefined);
    try {
      await createEntry.mutateAsync(name);
      notify("success", "エントリーを作りました。");
    } catch (caught) {
      notify(
        "error",
        errorCode(caught) === "state_conflict"
          ? "この大会にはもうエントリーがあります。"
          : "エントリーを作れませんでした。もう一度お試しください。",
      );
    }
  };

  const add = async (member: MemberRef) => {
    if (entry === null) {
      return;
    }
    try {
      await addMember.mutateAsync({ entryId: entry.id, rosterMemberId: member.rosterMemberId });
    } catch (caught) {
      notify(
        "error",
        errorCode(caught) === "validation_failed"
          ? `${member.name} さんはもう「参加」ではありません。`
          : `${member.name} さんを入れられませんでした。もう一度お試しください。`,
      );
    }
  };

  const remove = async (member: MemberRef) => {
    if (entry === null) {
      return;
    }
    try {
      await removeMember.mutateAsync({ entryId: entry.id, rosterMemberId: member.rosterMemberId });
    } catch {
      notify("error", `${member.name} さんを外せませんでした。もう一度お試しください。`);
    } finally {
      setRemoving(null);
    }
  };

  const busy = createEntry.isPending || addMember.isPending || removeMember.isPending;
  return (
    <>
      <EntryEditorView
        editor={editor.data}
        newName={newName}
        nameError={nameError}
        busy={busy}
        onNewNameChange={setNewName}
        onCreate={() => void create()}
        onAdd={(member) => void add(member)}
        onRemove={setRemoving}
      />
      {removing !== null && (
        <Confirmation
          title={`${removing.name} さんをエントリーから外しますか？`}
          messages={[
            "走順からも外れます。走順が確定していれば、この大会のラップタイムは消え、平均タイムは計算し直されます。",
          ]}
          confirmLabel="外す"
          busyLabel="外し中…"
          tone="danger"
          busy={busy}
          onConfirm={() => void remove(removing)}
          onCancel={() => setRemoving(null)}
        />
      )}
    </>
  );
}
