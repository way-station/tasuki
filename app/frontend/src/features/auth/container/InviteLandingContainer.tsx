import { Navigate, useParams } from "@tanstack/react-router";
import { errorCode } from "../../../shared/errors/apiErrors.ts";
import { useInviteLink } from "../hooks/useInviteLink.ts";
import { AuthMessage } from "../ui/AuthMessage.tsx";
import { InviteLanding } from "../ui/InviteLanding.tsx";
import { useLineLogin } from "./useLineLogin.ts";

// A1 (docs/design/screens.md).
export function InviteLandingContainer() {
  const { token } = useParams({ from: "/invite/$token" });
  const inviteLink = useInviteLink(token);
  const login = useLineLogin({ inviteToken: token });

  if (inviteLink.isPending) {
    return <AuthMessage title="読み込み中…" />;
  }
  if (inviteLink.isError) {
    if (errorCode(inviteLink.error) === "invite_unusable") {
      return <Navigate to="/unavailable" replace />;
    }
    return (
      <AuthMessage title="うまくいきませんでした" message="ページを読み込み直してください。" />
    );
  }
  return <InviteLanding teamName={inviteLink.data.teamName} {...login} />;
}
