import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import type { ApiClient, Me } from "../../../api/client.ts";
import { createFakeApiClient, problem } from "../../../api/fakeApiClient.ts";
import { TestProviders } from "../../../test/TestProviders.tsx";
import { createTestQueryClient } from "../../../test/testQueryClient.ts";
import { useCreateSession } from "./useCreateSession.ts";
import { useInviteLink } from "./useInviteLink.ts";
import { useLogout } from "./useLogout.ts";
import { useMe } from "./useMe.ts";
import { useStartLineLogin } from "./useStartLineLogin.ts";

// A cache that already holds data, standing in for what a user could see.
function seededQueryClient() {
  const queryClient = createTestQueryClient();
  queryClient.setQueryData(["races"], "visible to the previous user");
  return queryClient;
}

function renderWith<T>(hook: () => T, client: ApiClient, queryClient = createTestQueryClient()) {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <TestProviders client={client} queryClient={queryClient}>
      {children}
    </TestProviders>
  );
  return { ...renderHook(hook, { wrapper }), queryClient };
}

const approvedMe: Me = {
  user_id: "user-1",
  display_name: "Taro",
  is_organizer: false,
  status: "approved",
  roster_member: { id: "member-1", name: "Taro Yamada" },
};

describe("useMe", () => {
  it("returns the signed-in user", async () => {
    const { result } = renderWith(
      useMe,
      createFakeApiClient({ getMe: () => Promise.resolve(approvedMe) }),
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual({
      userId: "user-1",
      displayName: "Taro",
      isOrganizer: false,
      rosterMember: { id: "member-1", name: "Taro Yamada" },
      entryScreen: "home",
    });
  });

  it("fails with unauthenticated without a session, and the cache is cleared", async () => {
    const { result, queryClient } = renderWith(
      useMe,
      createFakeApiClient({ getMe: () => Promise.reject(problem(401, "unauthenticated")) }),
      seededQueryClient(),
    );

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toMatchObject({ code: "unauthenticated" });
    expect(queryClient.getQueryData(["races"])).toBeUndefined();
  });
});

describe("useInviteLink", () => {
  it("returns the team name", async () => {
    const getInviteLink = vi.fn(() => Promise.resolve({ team_name: "Team" }));
    const { result } = renderWith(
      () => useInviteLink("abc"),
      createFakeApiClient({ getInviteLink }),
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual({ teamName: "Team" });
    expect(getInviteLink).toHaveBeenCalledWith("abc");
  });

  it("fails with invite_unusable once, without retrying", async () => {
    const getInviteLink = vi.fn(() => Promise.reject(problem(404, "invite_unusable")));
    const { result } = renderWith(
      () => useInviteLink("abc"),
      createFakeApiClient({ getInviteLink }),
    );

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toMatchObject({ code: "invite_unusable" });
    expect(getInviteLink).toHaveBeenCalledTimes(1);
  });
});

describe("useStartLineLogin", () => {
  it("resolves to the LINE authorization URL, passing the invite token", async () => {
    const authorizeLine = vi.fn(() => Promise.resolve("https://line.example/auth"));
    const { result } = renderWith(useStartLineLogin, createFakeApiClient({ authorizeLine }));

    await expect(result.current.mutateAsync("abc")).resolves.toBe("https://line.example/auth");
    expect(authorizeLine).toHaveBeenCalledWith("abc");
  });

  it("starts without an invite token", async () => {
    const authorizeLine = vi.fn(() => Promise.resolve("https://line.example/auth"));
    const { result } = renderWith(useStartLineLogin, createFakeApiClient({ authorizeLine }));

    await result.current.mutateAsync(undefined);

    expect(authorizeLine).toHaveBeenCalledWith(undefined);
  });

  it("fails with invalid_request", async () => {
    const { result } = renderWith(
      useStartLineLogin,
      createFakeApiClient({
        authorizeLine: () => Promise.reject(problem(400, "invalid_request")),
      }),
    );

    await expect(result.current.mutateAsync("abc")).rejects.toMatchObject({
      code: "invalid_request",
    });
  });
});

describe("useCreateSession", () => {
  it("creates a session and clears the cache", async () => {
    const createSession = vi.fn(() => Promise.resolve());
    const { result, queryClient } = renderWith(
      useCreateSession,
      createFakeApiClient({ createSession }),
      seededQueryClient(),
    );

    await result.current.mutateAsync("one-time");

    expect(createSession).toHaveBeenCalledWith("one-time");
    expect(queryClient.getQueryData(["races"])).toBeUndefined();
  });

  it("fails with invalid_request for an unknown, expired, or used code, and keeps the cache", async () => {
    const { result, queryClient } = renderWith(
      useCreateSession,
      createFakeApiClient({
        createSession: () => Promise.reject(problem(400, "invalid_request")),
      }),
      seededQueryClient(),
    );

    await expect(result.current.mutateAsync("used")).rejects.toMatchObject({
      code: "invalid_request",
    });
    expect(queryClient.getQueryData(["races"])).toBe("visible to the previous user");
  });
});

describe("useLogout", () => {
  it("logs out and clears the cache", async () => {
    const logout = vi.fn(() => Promise.resolve());
    const { result, queryClient } = renderWith(
      useLogout,
      createFakeApiClient({ logout }),
      seededQueryClient(),
    );

    await result.current.mutateAsync();

    expect(logout).toHaveBeenCalledTimes(1);
    expect(queryClient.getQueryData(["races"])).toBeUndefined();
  });

  it("clears the cache when the session had already ended", async () => {
    const { result, queryClient } = renderWith(
      useLogout,
      createFakeApiClient({ logout: () => Promise.reject(problem(401, "unauthenticated")) }),
      seededQueryClient(),
    );

    await expect(result.current.mutateAsync()).rejects.toMatchObject({ code: "unauthenticated" });
    expect(queryClient.getQueryData(["races"])).toBeUndefined();
  });
});
