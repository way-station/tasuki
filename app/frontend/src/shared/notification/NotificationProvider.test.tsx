import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
  useNavigate,
} from "@tanstack/react-router";
import { act, fireEvent, render, renderHook, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { NotificationProvider, notificationDuration } from "./NotificationProvider.tsx";
import { useNotifications } from "./useNotifications.ts";

afterEach(() => {
  vi.useRealTimers();
});

function renderNotifications() {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <NotificationProvider>{children}</NotificationProvider>
  );
  const hook = renderHook(useNotifications, { wrapper });
  return hook.result;
}

describe("NotificationProvider", () => {
  it("shows a notification", () => {
    const notifier = renderNotifications();

    act(() => {
      notifier.current.notify("success", "Saved");
    });

    expect(screen.getByText("Saved")).toBeInTheDocument();
  });

  it("announces errors at once", () => {
    const notifier = renderNotifications();

    act(() => {
      notifier.current.notify("error", "Someone else saved first");
    });

    expect(screen.getByRole("alert")).toHaveTextContent("Someone else saved first");
  });

  it("dismisses a notification", () => {
    const notifier = renderNotifications();
    let id = 0;
    act(() => {
      id = notifier.current.notify("success", "Saved");
      notifier.current.notify("success", "Also saved");
    });

    act(() => {
      notifier.current.dismiss(id);
    });

    expect(screen.queryByText("Saved")).not.toBeInTheDocument();
    expect(screen.getByText("Also saved")).toBeInTheDocument();
  });

  it.each(["success", "error"] as const)("lets a %s notification disappear by itself", (kind) => {
    vi.useFakeTimers();
    const notifier = renderNotifications();
    act(() => {
      notifier.current.notify(kind, "Saved");
    });

    act(() => {
      vi.advanceTimersByTime(notificationDuration - 1);
    });
    expect(screen.getByText("Saved")).toBeInTheDocument();

    act(() => {
      vi.advanceTimersByTime(1);
    });
    expect(screen.queryByText("Saved")).not.toBeInTheDocument();
  });

  // Any registered path works here; only leaving the screen matters.
  it("keeps a notification visible on the next screen", async () => {
    function EditScreen() {
      const { notify } = useNotifications();
      const navigate = useNavigate();
      const save = () => {
        notify("success", "Saved");
        void navigate({ to: "/waiting" });
      };
      return (
        <button type="button" onClick={save}>
          Save
        </button>
      );
    }
    const rootRoute = createRootRoute();
    const router = createRouter({
      routeTree: rootRoute.addChildren([
        createRoute({ getParentRoute: () => rootRoute, path: "/", component: EditScreen }),
        createRoute({
          getParentRoute: () => rootRoute,
          path: "waiting",
          component: () => <h1>Next screen</h1>,
        }),
      ]),
      history: createMemoryHistory({ initialEntries: ["/"] }),
    });
    render(
      <NotificationProvider>
        <RouterProvider router={router} />
      </NotificationProvider>,
    );

    fireEvent.click(await screen.findByRole("button", { name: "Save" }));

    expect(await screen.findByRole("heading", { name: "Next screen" })).toBeInTheDocument();
    expect(screen.getByText("Saved")).toBeInTheDocument();
  });

  it("must be inside the provider", () => {
    vi.spyOn(console, "error").mockImplementation(() => {});

    expect(() => renderHook(useNotifications)).toThrow("inside NotificationProvider");
  });
});
