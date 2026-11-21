import { useId, type FormEvent } from "react";
import type {
  RosterMemberForm,
  RosterMemberFormErrors,
  RosterMemberView,
} from "../types/roster.ts";
import {
  card,
  cards,
  cardStack,
  content,
  dangerButton,
  fieldError,
  fieldHelp,
  group,
  groupHeading,
  input,
  label,
  neutralLabel,
  primaryButton,
  regularButton,
  secondaryButton,
  smallButton,
} from "../../../shared/ui/classes.ts";
import { formatDateTime, linkStatusColors, linkStatusLabels } from "./labels.ts";

type RosterViewProps = {
  members: RosterMemberView[];
  form: RosterMemberForm;
  errors: RosterMemberFormErrors;
  busy: boolean;
  onChange: (form: RosterMemberForm) => void;
  onAdd: () => void;
  onCreateInviteLink: (member: RosterMemberView) => void;
  onMarkLeft: (member: RosterMemberView) => void;
  onUndoLeft: (member: RosterMemberView) => void;
  onUnlink: (member: RosterMemberView) => void;
  onOpenProfile: (member: RosterMemberView) => void;
};

// M1 (docs/design/screens.md, wireframes/m1.html). Names cannot be changed.
// The name opens M2; the actions of each member are under it.
export function RosterView({
  members,
  form,
  errors,
  busy,
  onChange,
  onAdd,
  onCreateInviteLink,
  onMarkLeft,
  onUndoLeft,
  onUnlink,
  onOpenProfile,
}: RosterViewProps) {
  const id = useId();
  const submit = (event: FormEvent) => {
    event.preventDefault();
    onAdd();
  };

  return (
    <section>
      <header className="border-b border-slate-200 bg-white p-4">
        <h1 className="text-large-title font-bold">名簿</h1>
      </header>
      <div className={content}>
        <form onSubmit={submit} noValidate className="flex flex-col gap-1">
          <label htmlFor={id} className="text-subheadline text-slate-600">
            メンバーを追加（名前だけ）
          </label>
          <div className="flex items-center gap-2">
            <input
              id={id}
              className={input}
              placeholder="例：山本 蓮"
              value={form.name}
              aria-invalid={errors.name !== undefined}
              aria-describedby={errors.name === undefined ? `${id}-help` : `${id}-error ${id}-help`}
              onChange={(event) => onChange({ name: event.target.value })}
            />
            <button
              type="submit"
              disabled={busy}
              className={`${primaryButton} ${regularButton} shrink-0`}
            >
              追加
            </button>
          </div>
          {errors.name !== undefined && (
            <p id={`${id}-error`} className={fieldError}>
              {errors.name}
            </p>
          )}
          <p id={`${id}-help`} className={fieldHelp}>
            名前はあとで変えられません
          </p>
        </form>
        <section aria-label="メンバー" className={group}>
          <h2 className={groupHeading}>メンバー</h2>
          {members.length === 0 ? (
            <p className="text-subheadline text-slate-600">まだ名簿にメンバーはいません</p>
          ) : (
            <ul className={cards}>
              {members.map((member) => (
                <li key={member.id} className={`${card} ${cardStack}`}>
                  <button
                    type="button"
                    onClick={() => onOpenProfile(member)}
                    className="flex min-h-11 items-center gap-3 text-left"
                  >
                    <span className="flex min-w-0 grow flex-col gap-0.5">
                      <span className="text-headline font-semibold">{member.name}</span>
                      {member.inviteLinkExpiresAt !== undefined && (
                        <span className="text-subheadline text-slate-600">
                          招待リンク：
                          <time dateTime={member.inviteLinkExpiresAt}>
                            {formatDateTime(member.inviteLinkExpiresAt)}
                          </time>{" "}
                          まで
                        </span>
                      )}
                    </span>
                    {member.left ? (
                      <span className={`${label} ${neutralLabel}`}>退会</span>
                    ) : (
                      <span className={`${label} ${linkStatusColors[member.linkStatus]}`}>
                        {linkStatusLabels[member.linkStatus]}
                      </span>
                    )}
                    <span aria-hidden="true" className="text-slate-600">
                      ›
                    </span>
                  </button>
                  <div className="flex flex-wrap gap-2">
                    {member.canCreateInviteLink && (
                      <button
                        type="button"
                        disabled={busy}
                        aria-label={`${member.name} さんの${
                          member.inviteLinkExpiresAt === undefined
                            ? "招待リンクを作る"
                            : "招待リンクを作り直す"
                        }`}
                        onClick={() => onCreateInviteLink(member)}
                        className={`${
                          member.inviteLinkExpiresAt === undefined && member.linkStatus === "none"
                            ? primaryButton
                            : secondaryButton
                        } ${smallButton}`}
                      >
                        {member.inviteLinkExpiresAt === undefined
                          ? "招待リンクを作る"
                          : "リンクを作り直す"}
                      </button>
                    )}
                    {member.canUnlink && (
                      <button
                        type="button"
                        disabled={busy}
                        aria-label={`${member.name} さんの紐付けを解除`}
                        onClick={() => onUnlink(member)}
                        className={`${secondaryButton} ${smallButton}`}
                      >
                        紐付けを解除
                      </button>
                    )}
                    {member.left ? (
                      <button
                        type="button"
                        disabled={busy}
                        aria-label={`${member.name} さんの退会を取り消す`}
                        onClick={() => onUndoLeft(member)}
                        className={`${secondaryButton} ${smallButton}`}
                      >
                        退会を取り消す
                      </button>
                    ) : (
                      <button
                        type="button"
                        disabled={busy}
                        aria-label={`${member.name} さんを退会にする`}
                        onClick={() => onMarkLeft(member)}
                        className={`${dangerButton} ${smallButton}`}
                      >
                        退会にする
                      </button>
                    )}
                  </div>
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>
    </section>
  );
}
