import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { errorCode } from "../../../shared/errors/apiErrors.ts";
import { useNotifications } from "../../../shared/notification/useNotifications.ts";
import { Confirmation } from "../../../shared/ui/Confirmation.tsx";
import {
  useAddRosterMember,
  useCreateInviteLink,
  useSetLeft,
  useUnlink,
} from "../hooks/mutations.ts";
import { newRosterMemberForm, rosterMemberFormErrors } from "../hooks/rosterMemberFormInvalid.ts";
import { useRosterMembers } from "../hooks/useRosterMembers.ts";
import type {
  InviteLinkView,
  RosterMemberForm,
  RosterMemberFormErrors,
  RosterMemberView,
} from "../types/roster.ts";
import { InviteLinkPanel } from "../ui/InviteLinkPanel.tsx";
import { RosterMessage } from "../ui/RosterMessage.tsx";
import { RosterView } from "../ui/RosterView.tsx";

type Asking = { action: "left" | "unlink"; member: RosterMemberView };

// M1 (docs/design/screens.md). Marking as left and unlinking ask first.
export function RosterContainer() {
  const members = useRosterMembers();
  const addMember = useAddRosterMember();
  const createInviteLink = useCreateInviteLink();
  const setLeft = useSetLeft();
  const unlink = useUnlink();
  const { notify } = useNotifications();
  const navigate = useNavigate();
  const [form, setForm] = useState<RosterMemberForm>(newRosterMemberForm);
  const [errors, setErrors] = useState<RosterMemberFormErrors>({});
  const [created, setCreated] = useState<{ name: string; link: InviteLinkView } | null>(null);
  const [asking, setAsking] = useState<Asking | null>(null);

  if (members.error !== null) {
    return <RosterMessage title="Something went wrong" message="Reload the page." />;
  }
  if (members.data === undefined) {
    return <RosterMessage title="Loading…" />;
  }
  const busy =
    addMember.isPending || createInviteLink.isPending || setLeft.isPending || unlink.isPending;

  // The roster changed in the meantime; it is fetched again after every change.
  const notifyChanged = () =>
    notify("error", "The roster changed in the meantime. The latest is shown.");

  const add = async () => {
    setErrors({});
    try {
      await addMember.mutateAsync(form);
      setForm(newRosterMemberForm());
      notify("success", "Roster member added.");
    } catch (caught) {
      const fieldErrors = rosterMemberFormErrors(caught);
      if (fieldErrors !== undefined) {
        setErrors(fieldErrors);
        return;
      }
      notify("error", "The roster member could not be added. Try again.");
    }
  };

  const invite = async (member: RosterMemberView) => {
    try {
      const link = await createInviteLink.mutateAsync(member.id);
      setCreated({ name: member.name, link });
    } catch (caught) {
      switch (errorCode(caught)) {
        case "state_conflict":
        case "not_found":
          notifyChanged();
          return;
        default:
          notify("error", "The invite link could not be created. Try again.");
      }
    }
  };

  const copy = async (link: InviteLinkView) => {
    try {
      await navigator.clipboard.writeText(link.url);
      notify("success", "Invite link copied.");
    } catch {
      notify("error", "The link could not be copied. Select it and copy it by hand.");
    }
  };

  const changeLeft = async (member: RosterMemberView, left: boolean) => {
    try {
      await setLeft.mutateAsync({ rosterMemberId: member.id, left });
      setAsking(null);
      notify("success", left ? `${member.name} is marked as left.` : `${member.name} is back.`);
    } catch (caught) {
      setAsking(null);
      if (errorCode(caught) === "not_found") {
        notifyChanged();
        return;
      }
      notify("error", "The roster member could not be changed. Try again.");
    }
  };

  const removeLink = async (member: RosterMemberView) => {
    try {
      await unlink.mutateAsync(member.id);
      setAsking(null);
      notify("success", `${member.name} is unlinked.`);
    } catch (caught) {
      setAsking(null);
      switch (errorCode(caught)) {
        case "state_conflict":
        case "not_found":
          notifyChanged();
          return;
        default:
          notify("error", "The user could not be unlinked. Try again.");
      }
    }
  };

  return (
    <>
      <RosterView
        members={members.data}
        form={form}
        errors={errors}
        busy={busy}
        onChange={setForm}
        onAdd={() => void add()}
        onCreateInviteLink={(member) => void invite(member)}
        onMarkLeft={(member) => setAsking({ action: "left", member })}
        onUndoLeft={(member) => void changeLeft(member, false)}
        onUnlink={(member) => setAsking({ action: "unlink", member })}
        onOpenProfile={(member) =>
          void navigate({
            to: "/roster/$rosterMemberId",
            params: { rosterMemberId: member.id },
          })
        }
      />
      {created !== null && (
        <InviteLinkPanel
          name={created.name}
          link={created.link}
          onCopy={() => void copy(created.link)}
          onClose={() => setCreated(null)}
        />
      )}
      {asking?.action === "left" && (
        <Confirmation
          title={`Mark ${asking.member.name} as left?`}
          messages={[
            `${asking.member.name}'s invitations to races dated today or later are cancelled, with the same effects as cancelling them one by one. Past races are not changed.`,
            `${asking.member.name} can no longer log in. Undoing this does not bring the invitations back.`,
          ]}
          confirmLabel="Mark as left"
          busy={setLeft.isPending}
          onConfirm={() => void changeLeft(asking.member, true)}
          onCancel={() => setAsking(null)}
        />
      )}
      {asking?.action === "unlink" && (
        <Confirmation
          title={`Unlink ${asking.member.name}'s user?`}
          messages={[
            `The user can no longer act as ${asking.member.name}. To link again, ${asking.member.name} needs a new invite link and approval.`,
          ]}
          confirmLabel="Unlink"
          busy={unlink.isPending}
          onConfirm={() => void removeLink(asking.member)}
          onCancel={() => setAsking(null)}
        />
      )}
    </>
  );
}
