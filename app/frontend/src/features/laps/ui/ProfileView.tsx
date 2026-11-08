import type { ProfileView as Profile } from "../types/lap.ts";
import { AverageText } from "./AverageText.tsx";

type ProfileViewProps = {
  profile: Profile;
};

// P1 (docs/design/screens.md). Members enter nothing here.
export function ProfileView({ profile }: ProfileViewProps) {
  return (
    <section>
      <h1>{profile.name}</h1>
      <AverageText average={profile} />
    </section>
  );
}
