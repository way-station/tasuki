import { useParams } from "@tanstack/react-router";
import { useState } from "react";
import { useDraft } from "../../../shared/draft/useDraft.ts";
import { errorCode } from "../../../shared/errors/apiErrors.ts";
import { useNotifications } from "../../../shared/notification/useNotifications.ts";
import { useMe } from "../../auth/hooks/useMe.ts";
import { lapsDraftSource } from "../hooks/lapsDraft.ts";
import { lapsFormErrors, lapTimeError } from "../hooks/lapsFormInvalid.ts";
import { useDeleteLap, useEditLap, useSubmitLaps } from "../hooks/mutations.ts";
import { useMyLaps } from "../hooks/useMyLaps.ts";
import type { LapsForm, LapsFormErrors } from "../types/lap.ts";
import { LapsMessage } from "../ui/LapsMessage.tsx";
import { type LapEdit, MyLapsEditorView } from "../ui/MyLapsEditorView.tsx";

// L1 (docs/design/screens.md). Only the lap times being added are kept as a
// draft until they are submitted (ADR-0013). Laps have no version: only the
// caller changes their own laps, so there is nothing to conflict with. Editing
// one lap is done in place and is not kept.
export function MyLapsContainer() {
  const { raceId } = useParams({ from: "/team/races/$raceId" });
  const me = useMe();
  const myLaps = useMyLaps(raceId);
  // useDraft copies the source again only for a version it has not cleared,
  // so each submit starts the next empty form with a new round.
  const [round, setRound] = useState(0);
  const draft = useDraft<LapsForm>(me.data?.userId, `my-laps-${raceId}`, lapsDraftSource(round));
  const submitLaps = useSubmitLaps(raceId);
  const editLap = useEditLap(raceId);
  const deleteLap = useDeleteLap(raceId);
  const { notify } = useNotifications();
  const [errors, setErrors] = useState<LapsFormErrors>({});
  const [editing, setEditing] = useState<LapEdit | null>(null);

  if (myLaps.error !== null) {
    // Callers who are not a runner in a confirmed running order get not_found
    // (docs/design/api.md).
    if (errorCode(myLaps.error) === "not_found") {
      return (
        <LapsMessage
          title="Lap times are not open"
          message="You can add lap times once you run in a confirmed running order of this race."
        />
      );
    }
    return <LapsMessage title="Something went wrong" message="Reload the page." />;
  }
  if (myLaps.data === undefined || draft.value === undefined) {
    return <LapsMessage title="Loading…" />;
  }
  const form = draft.value;

  // The lap is gone or the caller no longer runs: the latest laps are shown.
  const notifyGone = () =>
    notify("error", "Your lap times changed in the meantime. The latest are shown.");

  const submit = async () => {
    setErrors({});
    try {
      await submitLaps.mutateAsync(form);
      setRound((draft.version ?? round) + 1);
      draft.clear();
      notify("success", "Lap times added.");
    } catch (caught) {
      const fieldErrors = lapsFormErrors(caught);
      if (fieldErrors !== undefined) {
        setErrors(fieldErrors);
        return;
      }
      switch (errorCode(caught)) {
        case "not_found":
          notifyGone();
          return;
        case "validation_failed":
          notify("error", "The lap times could not be added. Check the inputs.");
          return;
        default:
          notify("error", "The lap times could not be added. Try again.");
      }
    }
  };

  const startEdit = (lapId: string) => {
    const lap = myLaps.data.laps.find((candidate) => candidate.id === lapId);
    if (lap !== undefined) {
      setEditing({ lapId, time: lap.time });
    }
  };

  const saveEdit = async () => {
    if (editing === null) {
      return;
    }
    try {
      await editLap.mutateAsync({ lapId: editing.lapId, time: editing.time });
      setEditing(null);
      notify("success", "Lap time saved.");
    } catch (caught) {
      const timeError = lapTimeError(caught);
      if (timeError !== undefined) {
        setEditing({ ...editing, error: timeError });
        return;
      }
      switch (errorCode(caught)) {
        case "not_found":
          setEditing(null);
          notifyGone();
          return;
        case "validation_failed":
          setEditing({ ...editing, error: "Check the time." });
          return;
        default:
          notify("error", "The lap time could not be saved. Try again.");
      }
    }
  };

  const remove = async (lapId: string) => {
    try {
      await deleteLap.mutateAsync(lapId);
      notify("success", "Lap time deleted.");
    } catch (caught) {
      if (errorCode(caught) === "not_found") {
        notifyGone();
        return;
      }
      notify("error", "The lap time could not be deleted. Try again.");
    }
  };

  return (
    <MyLapsEditorView
      myLaps={myLaps.data}
      form={form}
      errors={errors}
      editing={editing}
      busy={submitLaps.isPending || editLap.isPending || deleteLap.isPending}
      onChange={draft.edit}
      onSubmit={() => void submit()}
      onStartEdit={startEdit}
      onEditTime={(time) => {
        if (editing !== null) {
          setEditing({ ...editing, time });
        }
      }}
      onSaveEdit={() => void saveEdit()}
      onCancelEdit={() => setEditing(null)}
      onDelete={(lapId) => void remove(lapId)}
    />
  );
}
