import { Route, Routes } from "react-router";
import { Layout } from "./Layout.tsx";
import { NotFound } from "./NotFound.tsx";
import { Placeholder } from "./Placeholder.tsx";

// One route per screen in docs/design/screens.md. Each placeholder is replaced
// by its feature's container. Routes only pick the screen: no loaders or
// actions (ADR-0013).
export function AppRoutes() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route
          path="invite/:token"
          element={<Placeholder screen="A1" title="Invite link landing" />}
        />
        <Route path="auth/complete" element={<Placeholder title="Completing login" />} />
        <Route path="waiting" element={<Placeholder screen="A2" title="Waiting for approval" />} />
        <Route path="rejected" element={<Placeholder screen="A3" title="Rejected" />} />
        <Route path="unavailable" element={<Placeholder screen="A4" title="Cannot use tasuki" />} />

        <Route index element={<Placeholder screen="R1" title="Race list" />} />
        <Route path="races/new" element={<Placeholder screen="R3" title="Register race" />} />
        <Route path="races/:raceId">
          <Route index element={<Placeholder screen="R2" title="Race detail" />} />
          <Route path="edit" element={<Placeholder screen="R3" title="Edit race" />} />
          <Route
            path="invitations"
            element={<Placeholder screen="R4" title="Race invitations" />}
          />
          <Route path="attendance" element={<Placeholder screen="R5" title="Attendance list" />} />
          <Route path="entry" element={<Placeholder screen="R6" title="Entry" />} />
          <Route
            path="running-order"
            element={<Placeholder screen="R7" title="Running order editor" />}
          />
          <Route path="laps" element={<Placeholder screen="L1" title="My lap times" />} />
        </Route>

        <Route path="me" element={<Placeholder screen="P1" title="My profile" />} />

        <Route path="roster" element={<Placeholder screen="M1" title="Roster" />} />
        <Route
          path="roster/:rosterMemberId"
          element={<Placeholder screen="M2" title="Member profile" />}
        />
        <Route path="approvals" element={<Placeholder screen="M3" title="Approvals" />} />

        <Route path="*" element={<NotFound />} />
      </Route>
    </Routes>
  );
}
