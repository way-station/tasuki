import { content } from "../../../shared/ui/classes.ts";
import type { ProfileView as Profile } from "../types/lap.ts";
import { AverageText } from "./AverageText.tsx";

type ProfileViewProps = {
  profile: Profile;
};

// P1 (docs/design/screens.md, wireframes/p1.html). Members enter nothing here.
export function ProfileView({ profile }: ProfileViewProps) {
  return (
    <section>
      <header className="border-b border-slate-200 bg-white p-4">
        <h1 className="text-large-title font-bold">{profile.name}</h1>
      </header>
      <div className={content}>
        <AverageText average={profile} />
      </div>
    </section>
  );
}
