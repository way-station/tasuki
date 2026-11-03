import type { AttendanceAnswer, CurrentAnswer } from "../types/participation.ts";
import { answerLabels } from "./labels.ts";

const answers: AttendanceAnswer[] = ["attending", "undecided", "not_attending"];

type AttendanceAnswerViewProps = {
  current: CurrentAnswer;
  saving: boolean;
  onAnswer: (answer: AttendanceAnswer) => void;
};

// The attendance answer on R2. The current answer is pressed.
export function AttendanceAnswerView({ current, saving, onAnswer }: AttendanceAnswerViewProps) {
  return (
    <div role="group" aria-label="My attendance">
      <p>{answerLabels[current]}</p>
      {answers.map((answer) => (
        <button
          key={answer}
          type="button"
          aria-pressed={current === answer}
          disabled={saving}
          onClick={() => onAnswer(answer)}
        >
          {answerLabels[answer]}
        </button>
      ))}
    </div>
  );
}
