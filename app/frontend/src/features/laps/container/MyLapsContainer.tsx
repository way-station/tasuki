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
import { useRunnerLegs } from "../hooks/useRunnerLegs.ts";
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
  // L1 is open only to runners of a confirmed running order, who may see it.
  const legsOf = useRunnerLegs(raceId, myLaps.data !== undefined);
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
          title="ラップタイムはまだ入れられません"
          message="確定した走順でこの大会を走ると、入れられるようになります。"
        />
      );
    }
    return (
      <LapsMessage title="うまくいきませんでした" message="ページを読み込み直してください。" />
    );
  }
  if (myLaps.data === undefined || draft.value === undefined) {
    return <LapsMessage title="読み込み中…" />;
  }
  const form = draft.value;

  // The lap is gone or the caller no longer runs: the latest laps are shown.
  const notifyGone = () =>
    notify("error", "ラップタイムがほかの画面で変えられていたので、最新を出しています。");

  const submit = async () => {
    setErrors({});
    try {
      await submitLaps.mutateAsync(form);
      setRound((draft.version ?? round) + 1);
      draft.clear();
      notify("success", "ラップタイムを足しました。");
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
          notify("error", "ラップタイムを足せませんでした。入力を確かめてください。");
          return;
        default:
          notify("error", "ラップタイムを足せませんでした。もう一度お試しください。");
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
      notify("success", "ラップタイムを保存しました。");
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
          setEditing({ ...editing, error: "タイムを確かめてください。" });
          return;
        default:
          notify("error", "ラップタイムを保存できませんでした。もう一度お試しください。");
      }
    }
  };

  const remove = async (lapId: string) => {
    try {
      await deleteLap.mutateAsync(lapId);
      notify("success", "ラップタイムを消しました。");
    } catch (caught) {
      if (errorCode(caught) === "not_found") {
        notifyGone();
        return;
      }
      notify("error", "ラップタイムを消せませんでした。もう一度お試しください。");
    }
  };

  return (
    <MyLapsEditorView
      myLaps={myLaps.data}
      legs={me.data?.rosterMember == null ? [] : legsOf(me.data.rosterMember.id)}
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
