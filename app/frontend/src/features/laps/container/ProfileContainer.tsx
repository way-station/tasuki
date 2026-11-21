import { useProfile } from "../hooks/useProfile.ts";
import { LapsMessage } from "../ui/LapsMessage.tsx";
import { ProfileView } from "../ui/ProfileView.tsx";

// P1 (docs/design/screens.md).
export function ProfileContainer() {
  const profile = useProfile();
  if (profile.error !== null) {
    return (
      <LapsMessage title="うまくいきませんでした" message="ページを読み込み直してください。" />
    );
  }
  if (profile.data === undefined) {
    return <LapsMessage title="読み込み中…" />;
  }
  return <ProfileView profile={profile.data} />;
}
