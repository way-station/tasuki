import {
  createRootRoute,
  createRoute,
  createRouter,
  type RouterHistory,
} from "@tanstack/react-router";
import { AuthCompleteContainer } from "../features/auth/container/AuthCompleteContainer.tsx";
import { InviteLandingContainer } from "../features/auth/container/InviteLandingContainer.tsx";
import { LoginContainer } from "../features/auth/container/LoginContainer.tsx";
import { TeamGuard } from "../features/auth/container/TeamGuard.tsx";
import { AuthMessage } from "../features/auth/ui/AuthMessage.tsx";
import { HealthContainer } from "../features/health/container/HealthContainer.tsx";
import { Layout } from "./Layout.tsx";
import { NotFound } from "./NotFound.tsx";
import { Placeholder } from "./Placeholder.tsx";

// One route per screen in docs/design/screens.md, all inside the common
// layout. Each placeholder is replaced by its feature's container. Routes only
// pick the screen: no loaders (ADR-0013).
const rootRoute = createRootRoute({ component: Layout });

// ---- Joining and access: open without a session --------------------------

const inviteRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "invite/$token",
  component: InviteLandingContainer,
});
const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "login",
  // The screen to return to after logging in.
  validateSearch: (search: Record<string, unknown>): { redirect?: string } => ({
    redirect: typeof search.redirect === "string" ? search.redirect : undefined,
  }),
  component: LoginContainer,
});
const authCompleteRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "auth/complete",
  // What the LINE callback redirects with (docs/design/api.md).
  validateSearch: (search: Record<string, unknown>): { code?: string; error?: string } => ({
    code: typeof search.code === "string" ? search.code : undefined,
    error: typeof search.error === "string" ? search.error : undefined,
  }),
  component: AuthCompleteContainer,
});
const waitingRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "waiting",
  component: () => (
    <AuthMessage title="Waiting for approval" message="Waiting for an organizer to approve you." />
  ),
});
const rejectedRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "rejected",
  component: () => <AuthMessage title="You were rejected" />,
});
const unavailableRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "unavailable",
  component: () => (
    <AuthMessage title="tasuki cannot be used" message="Please contact an organizer." />
  ),
});

// ---- The team's screens, behind the route guard (#113) --------------------

const teamRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: "team",
  component: () => <TeamGuard access="team" />,
});
const organizerRoute = createRoute({
  getParentRoute: () => teamRoute,
  id: "organizer",
  component: () => <TeamGuard access="organizer" />,
});
const linkedRoute = createRoute({
  getParentRoute: () => teamRoute,
  id: "linked",
  component: () => <TeamGuard access="linked" />,
});

// Until the race list (#119) replaces it, the home page also shows the
// backend health, to check frontend → backend → databases (#36).
const homeRoute = createRoute({
  getParentRoute: () => teamRoute,
  path: "/",
  component: () => (
    <>
      <Placeholder screen="R1" title="Race list" />
      <HealthContainer />
    </>
  ),
});
const raceNewRoute = createRoute({
  getParentRoute: () => organizerRoute,
  path: "races/new",
  component: () => <Placeholder screen="R3" title="Register race" />,
});
const raceRoute = createRoute({
  getParentRoute: () => teamRoute,
  path: "races/$raceId",
});
const raceDetailRoute = createRoute({
  getParentRoute: () => raceRoute,
  path: "/",
  component: () => <Placeholder screen="R2" title="Race detail" />,
});
const raceOrganizerRoute = createRoute({
  getParentRoute: () => raceRoute,
  id: "race-organizer",
  component: () => <TeamGuard access="organizer" />,
});
const raceEditRoute = createRoute({
  getParentRoute: () => raceOrganizerRoute,
  path: "edit",
  component: () => <Placeholder screen="R3" title="Edit race" />,
});
const raceInvitationsRoute = createRoute({
  getParentRoute: () => raceOrganizerRoute,
  path: "invitations",
  component: () => <Placeholder screen="R4" title="Race invitations" />,
});
const raceAttendanceRoute = createRoute({
  getParentRoute: () => raceOrganizerRoute,
  path: "attendance",
  component: () => <Placeholder screen="R5" title="Attendance list" />,
});
const raceEntryRoute = createRoute({
  getParentRoute: () => raceOrganizerRoute,
  path: "entry",
  component: () => <Placeholder screen="R6" title="Entry" />,
});
const raceRunningOrderRoute = createRoute({
  getParentRoute: () => raceOrganizerRoute,
  path: "running-order",
  component: () => <Placeholder screen="R7" title="Running order editor" />,
});
const raceLinkedRoute = createRoute({
  getParentRoute: () => raceRoute,
  id: "race-linked",
  component: () => <TeamGuard access="linked" />,
});
const raceLapsRoute = createRoute({
  getParentRoute: () => raceLinkedRoute,
  path: "laps",
  component: () => <Placeholder screen="L1" title="My lap times" />,
});

const profileRoute = createRoute({
  getParentRoute: () => linkedRoute,
  path: "me",
  component: () => <Placeholder screen="P1" title="My profile" />,
});

const rosterRoute = createRoute({
  getParentRoute: () => organizerRoute,
  path: "roster",
  component: () => <Placeholder screen="M1" title="Roster" />,
});
const rosterMemberRoute = createRoute({
  getParentRoute: () => organizerRoute,
  path: "roster/$rosterMemberId",
  component: () => <Placeholder screen="M2" title="Member profile" />,
});
const approvalsRoute = createRoute({
  getParentRoute: () => organizerRoute,
  path: "approvals",
  component: () => <Placeholder screen="M3" title="Approvals" />,
});

const routeTree = rootRoute.addChildren([
  inviteRoute,
  loginRoute,
  authCompleteRoute,
  waitingRoute,
  rejectedRoute,
  unavailableRoute,
  teamRoute.addChildren([
    homeRoute,
    raceRoute.addChildren([
      raceDetailRoute,
      raceOrganizerRoute.addChildren([
        raceEditRoute,
        raceInvitationsRoute,
        raceAttendanceRoute,
        raceEntryRoute,
        raceRunningOrderRoute,
      ]),
      raceLinkedRoute.addChildren([raceLapsRoute]),
    ]),
    organizerRoute.addChildren([raceNewRoute, rosterRoute, rosterMemberRoute, approvalsRoute]),
    linkedRoute.addChildren([profileRoute]),
  ]),
]);

// Tests pass a memory history; the app uses the browser's.
export function createAppRouter(history?: RouterHistory) {
  return createRouter({ routeTree, history, defaultNotFoundComponent: NotFound });
}

// Registers the routes so links and params are type-checked everywhere.
declare module "@tanstack/react-router" {
  interface Register {
    router: ReturnType<typeof createAppRouter>;
  }
}
