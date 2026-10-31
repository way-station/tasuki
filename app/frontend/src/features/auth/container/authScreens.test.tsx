import { createMemoryHistory, RouterProvider } from "@tanstack/react-router";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { ApiClient, Me } from "../../../api/client.ts";
import { createFakeApiClient, problem } from "../../../api/fakeApiClient.ts";
import { createAppRouter } from "../../../app/router.tsx";
import { TestProviders } from "../../../test/TestProviders.tsx";
import { leaveApp } from "./leaveApp.ts";

vi.mock("./leaveApp.ts", () => ({ leaveApp: vi.fn() }));

afterEach(() => {
  vi.mocked(leaveApp).mockClear();
  localStorage.clear();
});

function renderAt(path: string, stubs: Partial<ApiClient>) {
  const router = createAppRouter(createMemoryHistory({ initialEntries: [path] }));
  const client = createFakeApiClient({
    getHealth: () => Promise.resolve({ status: "ok", mongo: "ok", postgres: "ok" }),
    ...stubs,
  });
  render(
    <TestProviders client={client}>
      <RouterProvider router={router} />
    </TestProviders>,
  );
  return router;
}

function me(overrides: Partial<Me>): Partial<ApiClient> {
  return {
    getMe: () =>
      Promise.resolve({
        user_id: "user-1",
        display_name: "Taro",
        is_organizer: false,
        status: "approved",
        roster_member: { id: "member-1", name: "Taro Yamada" },
        ...overrides,
      }),
  };
}

const unlinked = { status: "none", roster_member: null } as const;

describe("route guard", () => {
  describe("without a session", () => {
    it("sends the user to the login screen, remembering where they were", async () => {
      const router = renderAt("/races/1/running-order", {
        getMe: () => Promise.reject(problem(401, "unauthenticated")),
      });

      expect(await screen.findByRole("heading", { name: "Log in" })).toBeInTheDocument();
      expect(router.state.location.pathname).toBe("/login");
      expect(router.state.location.search).toEqual({ redirect: "/races/1/running-order" });
    });
  });

  describe("member", () => {
    it.each([
      ["pending", "Waiting for approval"],
      ["rejected", "You were rejected"],
      ["none", "tasuki cannot be used"],
    ] as const)("with status %s sees %s instead of the team's screens", async (status, heading) => {
      renderAt("/", me({ status, roster_member: null }));

      expect(await screen.findByRole("heading", { name: heading })).toBeInTheDocument();
    });

    it.each(["/", "/races/1", "/races/1/laps", "/me"])("approved opens %s", async (path) => {
      const router = renderAt(path, me({}));

      await screen.findByRole("heading", { level: 1 });
      expect(router.state.location.pathname).toBe(path);
    });

    it.each(["/races/new", "/races/1/edit", "/races/1/running-order", "/roster", "/approvals"])(
      "approved goes home from the organizer screen %s",
      async (path) => {
        const router = renderAt(path, me({}));

        expect(await screen.findByRole("heading", { name: "R1 Race list" })).toBeInTheDocument();
        expect(router.state.location.pathname).toBe("/");
      },
    );
  });

  describe("organizer", () => {
    it.each(["pending", "rejected", "none"] as const)(
      "with own status %s still opens organizer screens",
      async (status) => {
        renderAt("/roster", me({ is_organizer: true, status, roster_member: null }));

        expect(await screen.findByRole("heading", { name: "M1 Roster" })).toBeInTheDocument();
      },
    );

    it.each(["/races/1/laps", "/me"])(
      "without a link goes home from the member screen %s",
      async (path) => {
        const router = renderAt(path, me({ is_organizer: true, ...unlinked }));

        expect(await screen.findByRole("heading", { name: "R1 Race list" })).toBeInTheDocument();
        expect(router.state.location.pathname).toBe("/");
      },
    );

    it("with a link opens member screens", async () => {
      renderAt("/me", me({ is_organizer: true }));

      expect(await screen.findByRole("heading", { name: "P1 My profile" })).toBeInTheDocument();
    });
  });

  it("tells the user when the backend fails", async () => {
    renderAt("/", { getMe: () => Promise.reject(problem(403, "forbidden")) });

    expect(
      await screen.findByRole("heading", { name: "Something went wrong" }),
    ).toBeInTheDocument();
  });
});

describe("login screen", () => {
  it("starts LINE Login and returns to the screen the user was on", async () => {
    const authorizeLine = vi.fn(() => Promise.resolve("https://line.example/auth"));
    renderAt("/login?redirect=%2Fraces%2F1", { authorizeLine });

    fireEvent.click(await screen.findByRole("button", { name: "Log in with LINE" }));

    await waitFor(() => expect(leaveApp).toHaveBeenCalledWith("https://line.example/auth"));
    expect(authorizeLine).toHaveBeenCalledWith(undefined);
    expect(localStorage.getItem("tasuki.returnPath")).toBe("/races/1");
  });

  it("shows an error when LINE Login cannot start", async () => {
    renderAt("/login", { authorizeLine: () => Promise.reject(problem(400, "invalid_request")) });

    fireEvent.click(await screen.findByRole("button", { name: "Log in with LINE" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Login failed");
    expect(leaveApp).not.toHaveBeenCalled();
  });
});

describe("invite link landing (A1)", () => {
  it("shows the team name and starts LINE Login with the invite token", async () => {
    const authorizeLine = vi.fn(() => Promise.resolve("https://line.example/auth"));
    renderAt("/invite/abc", {
      getInviteLink: () => Promise.resolve({ team_name: "Team" }),
      authorizeLine,
    });

    expect(await screen.findByRole("heading", { name: "Team" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Log in with LINE" }));

    await waitFor(() => expect(leaveApp).toHaveBeenCalledWith("https://line.example/auth"));
    expect(authorizeLine).toHaveBeenCalledWith("abc");
  });

  it("goes to A4 for an unusable invite link", async () => {
    const router = renderAt("/invite/abc", {
      getInviteLink: () => Promise.reject(problem(404, "invite_unusable")),
    });

    expect(
      await screen.findByRole("heading", { name: "tasuki cannot be used" }),
    ).toBeInTheDocument();
    expect(router.state.location.pathname).toBe("/unavailable");
  });
});

describe("login completion (/auth/complete)", () => {
  it("exchanges the code once and returns to the saved screen", async () => {
    localStorage.setItem("tasuki.returnPath", "/races/1");
    const createSession = vi.fn(() => Promise.resolve());
    const router = renderAt("/auth/complete?code=one-time", { createSession, ...me({}) });

    expect(await screen.findByRole("heading", { name: "R2 Race detail" })).toBeInTheDocument();
    expect(router.state.location.pathname).toBe("/races/1");
    expect(createSession).toHaveBeenCalledTimes(1);
    expect(createSession).toHaveBeenCalledWith("one-time");
  });

  it("goes home when no screen was saved, where the guard picks the screen", async () => {
    renderAt("/auth/complete?code=one-time", {
      createSession: () => Promise.resolve(),
      ...me({ status: "pending", roster_member: null }),
    });

    expect(
      await screen.findByRole("heading", { name: "Waiting for approval" }),
    ).toBeInTheDocument();
  });

  it.each(["invite_unusable", "left"])("goes to A4 for the error %s", async (error) => {
    const router = renderAt(`/auth/complete?error=${error}`, {});

    expect(
      await screen.findByRole("heading", { name: "tasuki cannot be used" }),
    ).toBeInTheDocument();
    expect(router.state.location.pathname).toBe("/unavailable");
  });

  it("shows an error for a failed or cancelled login", async () => {
    renderAt("/auth/complete?error=login_failed", {});

    expect(await screen.findByRole("alert")).toHaveTextContent("Login failed");
    expect(screen.getByRole("button", { name: "Log in with LINE" })).toBeInTheDocument();
  });

  it("shows an error for an unknown, expired, or used code", async () => {
    renderAt("/auth/complete?code=used", {
      createSession: () => Promise.reject(problem(400, "invalid_request")),
    });

    expect(await screen.findByRole("alert")).toHaveTextContent("Login failed");
  });
});
