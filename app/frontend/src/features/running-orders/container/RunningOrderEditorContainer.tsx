import { useNavigate, useParams } from "@tanstack/react-router";
import { useState } from "react";
import { useDraft } from "../../../shared/draft/useDraft.ts";
import { errorCode } from "../../../shared/errors/apiErrors.ts";
import { useNotifications } from "../../../shared/notification/useNotifications.ts";
import { Confirmation } from "../../../shared/ui/Confirmation.tsx";
import { useMe } from "../../auth/hooks/useMe.ts";
import { useConfirmRunningOrder, useSaveRunningOrder } from "../hooks/mutations.ts";
import { checkRunningOrder } from "../hooks/runningOrderChecks.ts";
import { useRunningOrderEditor } from "../hooks/useRunningOrderEditor.ts";
import type { Runner } from "../types/runningOrder.ts";
import { NoEntryView } from "../ui/NoEntryView.tsx";
import { RunningOrderEditorView } from "../ui/RunningOrderEditorView.tsx";
import { RunningOrderMessage } from "../ui/RunningOrderMessage.tsx";

// R7 (docs/design/screens.md). Every edit is kept as a draft until it is
// saved (ADR-0013).
export function RunningOrderEditorContainer() {
  const { raceId } = useParams({ from: "/team/races/$raceId" });
  const me = useMe();
  const editor = useRunningOrderEditor(raceId);
  const draft = useDraft(me.data?.userId, `running-order-${raceId}`, editor.data?.source);
  const saveOrder = useSaveRunningOrder(raceId);
  const confirmOrder = useConfirmRunningOrder(raceId);
  const { notify } = useNotifications();
  const navigate = useNavigate();
  const [conflict, setConflict] = useState(false);
  const [losing, setLosing] = useState<Runner[] | null>(null);

  if (editor.error !== null) {
    if (errorCode(editor.error) === "not_found") {
      return <RunningOrderMessage title="大会が見つかりません" />;
    }
    return (
      <RunningOrderMessage
        title="うまくいきませんでした"
        message="ページを読み込み直してください。"
      />
    );
  }
  if (editor.data === undefined) {
    return <RunningOrderMessage title="読み込み中…" />;
  }
  const data = editor.data.editor;
  if (data.entry === null) {
    return (
      <NoEntryView
        onOpenEntry={() => void navigate({ to: "/races/$raceId/entry", params: { raceId } })}
      />
    );
  }
  if (draft.value === undefined) {
    return <RunningOrderMessage title="読み込み中…" />;
  }
  const entry = data.entry;
  const form = draft.value;
  const checks = checkRunningOrder(data.saved, form);

  const onVersionConflict = () => {
    // The draft stays; the organizer chooses to load the latest (ADR-0013).
    setConflict(true);
    notify("error", "ほかの人が先にこの走順を保存しました。");
  };

  const save = async () => {
    setLosing(null);
    try {
      await saveOrder.mutateAsync({ entryId: entry.id, form, version: draft.version ?? 0 });
      draft.clear();
      notify("success", "走順を保存しました。");
    } catch (caught) {
      switch (errorCode(caught)) {
        case "version_conflict":
          onVersionConflict();
          return;
        case "validation_failed":
          notify("error", "走順が大会かエントリーと合わなくなりました。最新を読み込んでください。");
          setConflict(true);
          return;
        default:
          notify("error", "走順を保存できませんでした。もう一度お試しください。");
      }
    }
  };

  // Saving a confirmed order can delete lap times: name who loses them first.
  const requestSave = () => {
    if (checks.losingLaps.length > 0) {
      setLosing(checks.losingLaps);
      return;
    }
    void save();
  };

  const confirm = async () => {
    const version = data.saved?.version;
    if (version === undefined) {
      return;
    }
    try {
      await confirmOrder.mutateAsync({ entryId: entry.id, version });
      notify("success", "走順を確定しました。参加するメンバーに見えるようになりました。");
    } catch (caught) {
      switch (errorCode(caught)) {
        case "version_conflict":
          onVersionConflict();
          return;
        case "state_conflict":
          notify("error", "この走順はもう確定しています。");
          return;
        case "validation_failed":
          notify("error", "確定するには、すべての区に走者を入れてください。");
          return;
        default:
          notify("error", "走順を確定できませんでした。もう一度お試しください。");
      }
    }
  };

  const loadLatest = () => {
    draft.clear();
    setConflict(false);
  };

  return (
    <>
      <RunningOrderEditorView
        entryName={entry.name}
        legRule={data.legRule}
        runners={data.runners}
        form={form}
        status={data.saved?.status ?? null}
        prediction={data.saved?.prediction ?? null}
        unsaved={checks.unsaved}
        everyRunner={checks.everyRunner}
        canConfirm={checks.canConfirm}
        busy={saveOrder.isPending || confirmOrder.isPending}
        conflict={conflict}
        onChange={draft.edit}
        onSave={requestSave}
        onConfirm={() => void confirm()}
        onLoadLatest={loadLatest}
      />
      {losing !== null && (
        <Confirmation
          title="確定した走順を変えますか？"
          messages={["保存すると、すぐメンバーに見えます。走順は確定のままです。"]}
          warning={`走る区が減るため、この大会のラップタイムが消える人：${losing.map((runner) => runner.name).join("、")}。平均タイムは計算し直されます。`}
          confirmLabel="保存する"
          busyLabel="保存中…"
          tone="danger"
          busy={saveOrder.isPending}
          onConfirm={() => void save()}
          onCancel={() => setLosing(null)}
        />
      )}
    </>
  );
}
