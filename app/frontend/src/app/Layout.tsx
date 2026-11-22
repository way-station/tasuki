import { Outlet } from "@tanstack/react-router";

// The layout every screen shares. The team's screens add the page frame
// (TeamFrame); the joining and access screens (A1–A5) have none.
export function Layout() {
  return (
    <main>
      <Outlet />
    </main>
  );
}
