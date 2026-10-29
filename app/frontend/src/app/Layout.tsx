import { Outlet } from "react-router";

// The layout every screen shares. Its look comes with the styling decision (#33).
export function Layout() {
  return (
    <>
      <header>tasuki</header>
      <main>
        <Outlet />
      </main>
    </>
  );
}
