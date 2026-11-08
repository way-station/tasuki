import { useId, type FormEvent } from "react";
import type {
  RosterMemberForm,
  RosterMemberFormErrors,
  RosterMemberView,
} from "../types/roster.ts";
import { formatDateTime, linkStatusLabels } from "./labels.ts";

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

// M1 (docs/design/screens.md). Names cannot be changed.
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
      <h1>Roster</h1>
      {members.length === 0 ? (
        <p>No roster members yet.</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th scope="col">Name</th>
              <th scope="col">Status</th>
              <th scope="col">User</th>
              <th scope="col">Invite link</th>
              <th scope="col">Actions</th>
            </tr>
          </thead>
          <tbody>
            {members.map((member) => (
              <tr key={member.id}>
                <td>{member.name}</td>
                <td>{member.left ? "Left" : "Active"}</td>
                <td>{linkStatusLabels[member.linkStatus]}</td>
                <td>
                  {member.inviteLinkExpiresAt === undefined ? (
                    ""
                  ) : (
                    <>
                      Valid until{" "}
                      <time dateTime={member.inviteLinkExpiresAt}>
                        {formatDateTime(member.inviteLinkExpiresAt)}
                      </time>
                    </>
                  )}
                </td>
                <td>
                  <button type="button" onClick={() => onOpenProfile(member)}>
                    Profile of {member.name}
                  </button>
                  {member.canCreateInviteLink && (
                    <button
                      type="button"
                      disabled={busy}
                      onClick={() => onCreateInviteLink(member)}
                    >
                      {member.inviteLinkExpiresAt === undefined
                        ? `Create an invite link for ${member.name}`
                        : `Create a new invite link for ${member.name}`}
                    </button>
                  )}
                  {member.canUnlink && (
                    <button type="button" disabled={busy} onClick={() => onUnlink(member)}>
                      Unlink {member.name}
                    </button>
                  )}
                  {member.left ? (
                    <button type="button" disabled={busy} onClick={() => onUndoLeft(member)}>
                      Undo left for {member.name}
                    </button>
                  ) : (
                    <button type="button" disabled={busy} onClick={() => onMarkLeft(member)}>
                      Mark {member.name} as left
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      <form onSubmit={submit} noValidate>
        <h2>Add a roster member</h2>
        <label htmlFor={id}>Name</label>
        <input
          id={id}
          value={form.name}
          aria-invalid={errors.name !== undefined}
          aria-describedby={errors.name === undefined ? undefined : `${id}-error`}
          onChange={(event) => onChange({ name: event.target.value })}
        />
        {errors.name !== undefined && <p id={`${id}-error`}>{errors.name}</p>}
        <p>Names cannot be changed later.</p>
        <button type="submit" disabled={busy}>
          Add
        </button>
      </form>
    </section>
  );
}
