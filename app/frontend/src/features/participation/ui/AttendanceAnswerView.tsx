import type { AttendanceAnswer, CurrentAnswer } from "../types/participation.ts";
import { answerMarks, answerWords } from "./labels.ts";
import { group, groupHeading } from "../../../shared/ui/classes.ts";

const answers: AttendanceAnswer[] = ["attending", "undecided", "not_attending"];

// What each answer means, under the choices (R2 in docs/design/wireframes).
const answerHelp: Record<AttendanceAnswer, string> = {
  attending: "走ります。エントリーと走順の候補になります",
  undecided: "まだ決められません。あとで変えられます",
  not_attending: "走りません。この大会は表示されなくなり、変えられません",
};

type AttendanceAnswerViewProps = {
  current: CurrentAnswer;
  saving: boolean;
  onAnswer: (answer: AttendanceAnswer) => void;
};

// The attendance answer on R2: three segments, the current answer checked.
// Nothing is checked before the member answers.
export function AttendanceAnswerView({ current, saving, onAnswer }: AttendanceAnswerViewProps) {
  return (
    <fieldset className={`${group} min-w-0`}>
      <legend className={`${groupHeading} mb-2`}>あなたの出欠</legend>
      <div className="grid grid-cols-3 gap-2">
        {answers.map((answer) => (
          <label
            key={answer}
            className="flex min-h-11 cursor-pointer flex-col items-center justify-center rounded-lg border border-slate-500 bg-white p-2 has-checked:border-2 has-checked:border-blue-700 has-checked:bg-blue-100 has-checked:text-blue-700 has-focus-visible:outline-2 has-focus-visible:outline-offset-2 has-focus-visible:outline-blue-700 has-disabled:cursor-default has-disabled:opacity-60"
          >
            <input
              type="radio"
              name="attendance"
              value={answer}
              checked={current === answer}
              disabled={saving}
              onChange={() => onAnswer(answer)}
              className="sr-only"
            />
            <span aria-hidden="true" className="text-title2">
              {answerMarks[answer]}
            </span>
            <span className="text-headline font-semibold">{answerWords[answer]}</span>
          </label>
        ))}
      </div>
      <dl className="grid grid-cols-[24px_minmax(0,1fr)] gap-x-2 gap-y-1 text-subheadline text-slate-600">
        {answers.map((answer) => (
          <div key={answer} className="contents">
            <dt aria-hidden="true">{answerMarks[answer]}</dt>
            <dd>{answerHelp[answer]}</dd>
          </div>
        ))}
      </dl>
    </fieldset>
  );
}
