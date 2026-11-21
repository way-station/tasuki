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
    return (
      <RosterMessage title="うまくいきませんでした" message="ページを読み込み直してください。" />
    );
  }
  if (members.data === undefined) {
    return <RosterMessage title="読み込み中…" />;
  }
  const busy =
    addMember.isPending || createInviteLink.isPending || setLeft.isPending || unlink.isPending;

  // The roster changed in the meantime; it is fetched again after every change.
  const notifyChanged = () =>
    notify("error", "名簿がほかの人に変えられていたので、最新を出しています。");

  const add = async () => {
    setErrors({});
    try {
      await addMember.mutateAsync(form);
      setForm(newRosterMemberForm());
      notify("success", "名簿にメンバーを足しました。");
    } catch (caught) {
      const fieldErrors = rosterMemberFormErrors(caught);
      if (fieldErrors !== undefined) {
        setErrors(fieldErrors);
        return;
      }
      notify("error", "名簿にメンバーを足せませんでした。もう一度お試しください。");
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
          notify("error", "招待リンクを作れませんでした。もう一度お試しください。");
      }
    }
  };

  const copy = async (link: InviteLinkView) => {
    try {
      await navigator.clipboard.writeText(link.url);
      notify("success", "招待リンクをコピーしました。");
    } catch {
      notify("error", "リンクをコピーできませんでした。選んで手でコピーしてください。");
    }
  };

  const changeLeft = async (member: RosterMemberView, left: boolean) => {
    try {
      await setLeft.mutateAsync({ rosterMemberId: member.id, left });
      setAsking(null);
      notify(
        "success",
        left ? `${member.name} さんを退会にしました。` : `${member.name} さんを戻しました。`,
      );
    } catch (caught) {
      setAsking(null);
      if (errorCode(caught) === "not_found") {
        notifyChanged();
        return;
      }
      notify("error", "名簿のメンバーを変えられませんでした。もう一度お試しください。");
    }
  };

  const removeLink = async (member: RosterMemberView) => {
    try {
      await unlink.mutateAsync(member.id);
      setAsking(null);
      notify("success", `${member.name} さんの紐付けを解除しました。`);
    } catch (caught) {
      setAsking(null);
      switch (errorCode(caught)) {
        case "state_conflict":
        case "not_found":
          notifyChanged();
          return;
        default:
          notify("error", "紐付けを解除できませんでした。もう一度お試しください。");
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
          title={`${asking.member.name} さんを退会にしますか？`}
          messages={[
            "開催日が今日以降の大会への招待は取り消され、出欠、エントリー、走順からも外れます。走順が確定していれば、その大会のラップタイムは消えます。過去の大会は変わりません。",
            "ログインもできなくなります。退会を取り消しても、招待は戻りません。",
          ]}
          confirmLabel="退会にする"
          busyLabel="退会にしています…"
          tone="danger"
          busy={setLeft.isPending}
          onConfirm={() => void changeLeft(asking.member, true)}
          onCancel={() => setAsking(null)}
        />
      )}
      {asking?.action === "unlink" && (
        <Confirmation
          title={`${asking.member.name} さんの紐付けを解除しますか？`}
          messages={[
            "いまの LINE のアカウントでは tasuki を使えなくなります。つなぎ直すには、新しい招待リンクと承認が要ります。",
          ]}
          confirmLabel="解除する"
          busyLabel="解除中…"
          tone="danger"
          busy={unlink.isPending}
          onConfirm={() => void removeLink(asking.member)}
          onCancel={() => setAsking(null)}
        />
      )}
    </>
  );
}
