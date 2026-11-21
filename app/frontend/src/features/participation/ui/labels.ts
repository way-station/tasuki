import type { AttendanceAnswer } from "../types/participation.ts";
import { confirmedLabel, neutralLabel, waitingLabel } from "../../../shared/ui/classes.ts";

// Each answer has a mark as well as a word, so it is not told by color alone
// (docs/design/components.md, Status labels).
export const answerMarks: Record<AttendanceAnswer, string> = {
  attending: "○",
  undecided: "△",
  not_attending: "×",
};

export const answerWords: Record<AttendanceAnswer, string> = {
  attending: "参加",
  undecided: "保留",
  not_attending: "不参加",
};

export const answerLabels: Record<AttendanceAnswer | "not_answered", string> = {
  attending: "○ 参加",
  undecided: "△ 保留",
  not_attending: "× 不参加",
  not_answered: "未回答",
};

// The colors of each label. Not answered waits for the member.
export const answerColors: Record<AttendanceAnswer | "not_answered", string> = {
  attending: confirmedLabel,
  undecided: neutralLabel,
  not_attending: neutralLabel,
  not_answered: waitingLabel,
};
