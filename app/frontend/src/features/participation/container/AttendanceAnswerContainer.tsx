import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { errorCode } from "../../../shared/errors/apiErrors.ts";
import { useNotifications } from "../../../shared/notification/useNotifications.ts";
import { useAnswerAttendance } from "../hooks/mutations.ts";
import type { AttendanceAnswer, CurrentAnswer } from "../types/participation.ts";
import { AttendanceAnswerView } from "../ui/AttendanceAnswerView.tsx";
import { Confirmation } from "../../../shared/ui/Confirmation.tsx";

// What changing the answer can do (docs/design/screens.md, R2).
const finalWarning =
  "Not attending is final: the race disappears from your screens and you cannot change your answer.";
// Members cannot see the entry, so this is said whenever they leave
// "attending".
const entryWarning =
  "If you are in the entry, you are removed from it and from the running order. If the running order is confirmed, your lap times for this race are deleted and your average time is recalculated.";

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
      notify("success", "Your answer was saved.");
      if (answer === "not_attending") {
        // The race is no longer shown to the member.
        void navigate({ to: "/" });
      }
    } catch (caught) {
      setConfirming(null);
      notify(
        "error",
        errorCode(caught) === "not_found"
          ? "You can no longer answer for this race."
          : "Your answer could not be saved. Try again.",
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
          title={confirming === "not_attending" ? "Answer not attending?" : "Change your answer?"}
          messages={[
            ...(confirming === "not_attending" ? [finalWarning] : []),
            ...(current === "attending" ? [entryWarning] : []),
          ]}
          confirmLabel={confirming === "not_attending" ? "Answer not attending" : "Change answer"}
          busy={answerAttendance.isPending}
          onConfirm={() => void send(confirming)}
          onCancel={() => setConfirming(null)}
        />
      )}
    </>
  );
}
