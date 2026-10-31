import { Navigate, Outlet, useLocation } from "@tanstack/react-router";
import { isUnauthenticated } from "../hooks/apiErrors.ts";
import { useMe } from "../hooks/useMe.ts";
import { AuthMessage } from "../ui/AuthMessage.tsx";
import { type Access, accessRedirect } from "./accessRedirect.ts";

type TeamGuardProps = {
  access: Access;
};

// The route guard: shows the screens below it only to users who may open
// them, and sends everyone else to the right screen (#113).
export function TeamGuard({ access }: TeamGuardProps) {
  const me = useMe();
  const location = useLocation();

  if (me.data === undefined) {
    if (me.isPending) {
      return <AuthMessage title="Loading…" />;
    }
    if (!isUnauthenticated(me.error)) {
      return <AuthMessage title="Something went wrong" message="Reload the page." />;
    }
  }

  const redirect = accessRedirect(me.data ?? "unauthenticated", access);
  if (redirect === null) {
    return <Outlet />;
  }
  // While the router moves to the next screen, this guard renders once more
  // with that screen's URL; redirecting again would loop.
  if (location.pathname === redirect) {
    return null;
  }
  if (redirect === "/login") {
    return <Navigate to="/login" search={{ redirect: location.href }} replace />;
  }
  return <Navigate to={redirect} replace />;
}
