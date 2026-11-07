import { useParams } from "@tanstack/react-router";
import { errorCode } from "../../../shared/errors/apiErrors.ts";
import { useMemberProfile } from "../hooks/useMemberProfile.ts";
import { MemberProfileDetails } from "../ui/MemberProfileDetails.tsx";
import { RosterMessage } from "../ui/RosterMessage.tsx";

// M2 (docs/design/screens.md).
export function MemberProfileContainer() {
  const { rosterMemberId } = useParams({ from: "/team/organizer/roster/$rosterMemberId" });
  const profile = useMemberProfile(rosterMemberId);
  if (profile.error !== null) {
    if (errorCode(profile.error) === "not_found") {
      return <RosterMessage title="Roster member not found" />;
    }
    return <RosterMessage title="Something went wrong" message="Reload the page." />;
  }
  if (profile.data === undefined) {
    return <RosterMessage title="Loading…" />;
  }
  return <MemberProfileDetails profile={profile.data} />;
}
