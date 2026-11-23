import { useLocation, useNavigate } from "@tanstack/react-router";
import type { ReactNode } from "react";
import { useMe } from "../features/auth/hooks/useMe.ts";
import type { CurrentUser } from "../features/auth/types/auth.ts";
import { useRace } from "../features/races/hooks/useRace.ts";
import { usePendingLinks } from "../features/roster/hooks/usePendingLinks.ts";
import { type Tab, TabBar } from "../shared/ui/TabBar.tsx";
import { TopBar } from "../shared/ui/TopBar.tsx";

type TeamFrameProps = {
  children: ReactNode;
};

// The page frame of the team's screens (docs/design/components.md, Page
// frame): a top bar on the screens that are not in the tab bar, and the tab
// bar at the bottom. The joining and access screens (A1–A5) have neither.
export function TeamFrame({ children }: TeamFrameProps) {
  const me = useMe();
  const { pathname } = useLocation();
  const navigate = useNavigate();
  const go = (href: string) => void navigate({ to: href });
  const user = me.data;
  const showTabs = user !== undefined && (user.isOrganizer || user.entryScreen === "home");
  const pending = usePendingLinks(showTabs && user.isOrganizer);

  const tabs = showTabs ? tabsFor(user, pathname, pending.data?.length ?? 0) : [];
  return (
    <>
      {showTabs && <BackBar pathname={pathname} onBack={go} />}
      <div className={showTabs ? "pb-[calc(49px+env(safe-area-inset-bottom))]" : undefined}>
        {children}
      </div>
      {showTabs && <TabBar tabs={tabs} onSelect={go} />}
    </>
  );
}

// The tabs per person (docs/design/wireframes/index.html): organizers get 名簿
// and 承認, and anyone linked to a roster member gets 自分.
function tabsFor(user: CurrentUser, pathname: string, waiting: number): Tab[] {
  const tabs: Tab[] = [
    { label: "大会", href: "/", current: pathname === "/" || pathname.startsWith("/races") },
  ];
  if (user.isOrganizer) {
    tabs.push(
      { label: "名簿", href: "/roster", current: pathname.startsWith("/roster") },
      { label: "承認", href: "/approvals", current: pathname === "/approvals", badge: waiting },
    );
  }
  if (user.rosterMember !== null) {
    tabs.push({ label: "自分", href: "/me", current: pathname === "/me" });
  }
  return tabs;
}

type BackBarProps = {
  pathname: string;
  onBack: (href: string) => void;
};

// Where each screen outside the tab bar goes back to.
function BackBar({ pathname, onBack }: BackBarProps) {
  const raceChild = /^\/races\/([^/]+)\/[^/]+\/?$/.exec(pathname);
  if (raceChild !== null && raceChild[1] !== "new") {
    return <RaceBackBar raceId={raceChild[1]} onBack={onBack} />;
  }
  if (/^\/races\/[^/]+\/?$/.test(pathname)) {
    return <TopBar backLabel="大会" backHref="/" onBack={onBack} />;
  }
  if (/^\/roster\/[^/]+\/?$/.test(pathname)) {
    return <TopBar backLabel="名簿" backHref="/roster" onBack={onBack} />;
  }
  return null;
}

type RaceBackBarProps = {
  raceId: string;
  onBack: (href: string) => void;
};

// R3 (edit) to R7 and L1 go back to the race, named by the race's name.
function RaceBackBar({ raceId, onBack }: RaceBackBarProps) {
  const race = useRace(raceId);
  return (
    <TopBar backLabel={race.data?.name ?? "大会"} backHref={`/races/${raceId}`} onBack={onBack} />
  );
}
