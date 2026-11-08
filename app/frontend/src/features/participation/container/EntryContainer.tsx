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
    return <ParticipationMessage title="Something went wrong" message="Reload the page." />;
  }
  if (editor.data === undefined) {
    return <ParticipationMessage title="Loading…" />;
  }
  const entry = editor.data.entry;

  const create = async () => {
    const name = newName.trim();
    if (name === "") {
      setNameError("Enter a name.");
      return;
    }
    setNameError(undefined);
    try {
      await createEntry.mutateAsync(name);
      notify("success", "The entry was created.");
    } catch (caught) {
      notify(
        "error",
        errorCode(caught) === "state_conflict"
          ? "This race already has an entry."
          : "The entry could not be created. Try again.",
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
          ? `${member.name} is no longer attending.`
          : `${member.name} could not be added. Try again.`,
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
      notify("error", `${member.name} could not be removed. Try again.`);
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
          title={`Remove ${removing.name} from the entry?`}
          messages={[
            `${removing.name} is also removed from the running order. If the running order is confirmed, their lap times for this race are deleted and their average time is recalculated.`,
          ]}
          confirmLabel="Remove"
          busy={busy}
          onConfirm={() => void remove(removing)}
          onCancel={() => setRemoving(null)}
        />
      )}
    </>
  );
}
