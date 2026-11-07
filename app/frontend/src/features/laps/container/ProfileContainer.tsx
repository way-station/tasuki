import { useProfile } from "../hooks/useProfile.ts";
import { LapsMessage } from "../ui/LapsMessage.tsx";
import { ProfileView } from "../ui/ProfileView.tsx";

// P1 (docs/design/screens.md).
export function ProfileContainer() {
  const profile = useProfile();
  if (profile.error !== null) {
    return <LapsMessage title="Something went wrong" message="Reload the page." />;
  }
  if (profile.data === undefined) {
    return <LapsMessage title="Loading…" />;
  }
  return <ProfileView profile={profile.data} />;
}
