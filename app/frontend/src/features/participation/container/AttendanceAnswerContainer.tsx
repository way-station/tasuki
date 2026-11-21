import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { errorCode } from "../../../shared/errors/apiErrors.ts";
import { useNotifications } from "../../../shared/notification/useNotifications.ts";
import { useAnswerAttendance } from "../hooks/mutations.ts";
import type { AttendanceAnswer, CurrentAnswer } from "../types/participation.ts";
import { AttendanceAnswerView } from "../ui/AttendanceAnswerView.tsx";
import { Confirmation } from "../../../shared/ui/Confirmation.tsx";

// What changing the answer can do (docs/design/screens.md, R2).
const finalWarning = "この大会は表示されなくなり、回答は変えられません。";
// Members cannot see the entry, so this is said whenever they leave
// "attending".
const entryWarning =
  "エントリーに入っていれば、エントリーと走順から外れます。走順が確定していれば、この大会のラップタイムは消え、平均タイムは計算し直されます。";

type AttendanceAnswerContainerProps = {
  raceId: string;
  current: CurrentAnswer;
};

// The attendance answer on R2, for invited members.
export function AttendanceAnswerContainer({ raceId, current }: AttendanceAnswerContainerProps) {
  const answerAttendance = useAnswerAttendance(raceId);
  const { notify } = useNotifications();
  const navigate = useNavigate();
  // An answer waiting for the member to confirm it.
  const [confirming, setConfirming] = useState<AttendanceAnswer | null>(null);

  const send = async (answer: AttendanceAnswer) => {
    try {
      await answerAttendance.mutateAsync(answer);
      setConfirming(null);
      notify("success", "回答を保存しました。");
      if (answer === "not_attending") {
        // The race is no longer shown to the member.
        void navigate({ to: "/" });
      }
    } catch (caught) {
      setConfirming(null);
      notify(
        "error",
        errorCode(caught) === "not_found"
          ? "この大会にはもう回答できません。"
          : "回答を保存できませんでした。もう一度お試しください。",
      );
    }
  };

  const choose = (answer: AttendanceAnswer) => {
    if (answer === current) {
      return;
    }
    if (answer === "not_attending" || current === "attending") {
      setConfirming(answer);
      return;
    }
    void send(answer);
  };

  return (
    <>
      <AttendanceAnswerView
        current={current}
        saving={answerAttendance.isPending}
        onAnswer={choose}
      />
      {confirming !== null && (
        <Confirmation
          title={confirming === "not_attending" ? "不参加にしますか？" : "保留にしますか？"}
          messages={[
            ...(confirming === "not_attending" ? [finalWarning] : []),
            ...(current === "attending" ? [entryWarning] : []),
          ]}
          confirmLabel={confirming === "not_attending" ? "不参加にする" : "保留にする"}
          busyLabel="保存中…"
          tone={confirming === "not_attending" ? "danger" : "primary"}
          busy={answerAttendance.isPending}
          onConfirm={() => void send(confirming)}
          onCancel={() => setConfirming(null)}
        />
      )}
    </>
  );
}
