import {
  createRootRoute,
  createRoute,
  createRouter,
  type RouterHistory,
} from "@tanstack/react-router";
import { Layout } from "./Layout.tsx";
import { NotFound } from "./NotFound.tsx";
import { Placeholder } from "./Placeholder.tsx";

// One route per screen in docs/design/screens.md, all inside the common
// layout. Each placeholder is replaced by its feature's container. Routes only
// pick the screen: no loaders (ADR-0013).
const rootRoute = createRootRoute({ component: Layout });

const inviteRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "invite/$token",
  component: () => <Placeholder screen="A1" title="Invite link landing" />,
});
const authCompleteRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "auth/complete",
  component: () => <Placeholder title="Completing login" />,
});
const waitingRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "waiting",
  component: () => <Placeholder screen="A2" title="Waiting for approval" />,
});
const rejectedRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "rejected",
  component: () => <Placeholder screen="A3" title="Rejected" />,
});
const unavailableRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "unavailable",
  component: () => <Placeholder screen="A4" title="Cannot use tasuki" />,
});

const homeRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  component: () => <Placeholder screen="R1" title="Race list" />,
});
const raceNewRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "races/new",
  component: () => <Placeholder screen="R3" title="Register race" />,
});
const raceRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "races/$raceId",
});
const raceDetailRoute = createRoute({
  getParentRoute: () => raceRoute,
  path: "/",
  component: () => <Placeholder screen="R2" title="Race detail" />,
});
const raceEditRoute = createRoute({
  getParentRoute: () => raceRoute,
  path: "edit",
  component: () => <Placeholder screen="R3" title="Edit race" />,
});
const raceInvitationsRoute = createRoute({
  getParentRoute: () => raceRoute,
  path: "invitations",
  component: () => <Placeholder screen="R4" title="Race invitations" />,
});
const raceAttendanceRoute = createRoute({
  getParentRoute: () => raceRoute,
  path: "attendance",
  component: () => <Placeholder screen="R5" title="Attendance list" />,
});
const raceEntryRoute = createRoute({
  getParentRoute: () => raceRoute,
  path: "entry",
  component: () => <Placeholder screen="R6" title="Entry" />,
});
const raceRunningOrderRoute = createRoute({
  getParentRoute: () => raceRoute,
  path: "running-order",
  component: () => <Placeholder screen="R7" title="Running order editor" />,
});
const raceLapsRoute = createRoute({
  getParentRoute: () => raceRoute,
  path: "laps",
  component: () => <Placeholder screen="L1" title="My lap times" />,
});

const profileRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "me",
  component: () => <Placeholder screen="P1" title="My profile" />,
});

const rosterRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "roster",
  component: () => <Placeholder screen="M1" title="Roster" />,
});
const rosterMemberRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "roster/$rosterMemberId",
  component: () => <Placeholder screen="M2" title="Member profile" />,
});
const approvalsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "approvals",
  component: () => <Placeholder screen="M3" title="Approvals" />,
});

const routeTree = rootRoute.addChildren([
  inviteRoute,
  authCompleteRoute,
  waitingRoute,
  rejectedRoute,
  unavailableRoute,
  homeRoute,
  raceNewRoute,
  raceRoute.addChildren([
    raceDetailRoute,
    raceEditRoute,
    raceInvitationsRoute,
    raceAttendanceRoute,
    raceEntryRoute,
    raceRunningOrderRoute,
    raceLapsRoute,
  ]),
  profileRoute,
  rosterRoute,
  rosterMemberRoute,
  approvalsRoute,
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
