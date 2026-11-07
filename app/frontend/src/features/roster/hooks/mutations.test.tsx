import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import type { ApiClient, RosterMember } from "../../../api/client.ts";
import { createFakeApiClient, problem } from "../../../api/fakeApiClient.ts";
import { queryKeys } from "../../../api/queryKeys.ts";
import { TestProviders } from "../../../test/TestProviders.tsx";
import { createTestQueryClient } from "../../../test/testQueryClient.ts";
import { useAddRosterMember, useCreateInviteLink, useSetLeft, useUnlink } from "./mutations.ts";
import { rosterMemberFormErrors } from "./rosterMemberFormInvalid.ts";

// A cache holding what a roster change may affect.
const cachedKeys = {
  rosterMembers: queryKeys.rosterMembers,
  rosterMember: queryKeys.rosterMember("m1"),
  memberLinks: queryKeys.memberLinks,
  race: queryKeys.race("race-1"),
};

function invalidated(queryClient: ReturnType<typeof createTestQueryClient>) {
  return Object.fromEntries(
    Object.entries(cachedKeys).map(([name, key]) => [
      name,
      queryClient.getQueryState(key)?.isInvalidated ?? false,
    ]),
  );
}

function renderWith<T>(hook: () => T, client: ApiClient) {
  const queryClient = createTestQueryClient();
  for (const key of Object.values(cachedKeys)) {
    queryClient.setQueryData(key, {});
  }
  const wrapper = ({ children }: { children: ReactNode }) => (
    <TestProviders client={client} queryClient={queryClient}>
      {children}
    </TestProviders>
  );
  return { ...renderHook(hook, { wrapper }), queryClient };
}

const rosterMember: RosterMember = {
  id: "m1",
  name: "Taro",
  left: false,
  link_status: "none",
  invite_link_expires_at: null,
};

describe("useAddRosterMember", () => {
  it("sends the trimmed name and fetches the roster and races again", async () => {
    const createRosterMember = vi.fn(() => Promise.resolve(rosterMember));
    const { result, queryClient } = renderWith(
      () => useAddRosterMember(),
      createFakeApiClient({ createRosterMember }),
    );

    await result.current.mutateAsync({ name: " Taro " });

    expect(createRosterMember).toHaveBeenCalledWith("Taro");
    expect(invalidated(queryClient)).toEqual({
      rosterMembers: true,
      rosterMember: true,
      memberLinks: false,
      race: true,
    });
  });

  it("sends nothing without a name", async () => {
    const createRosterMember = vi.fn(() => Promise.resolve(rosterMember));
    const { result } = renderWith(
      () => useAddRosterMember(),
      createFakeApiClient({ createRosterMember }),
    );

    result.current.mutate({ name: " " });

    await waitFor(() =>
      expect(rosterMemberFormErrors(result.current.error)).toEqual({ name: "Enter a name." }),
    );
    expect(createRosterMember).not.toHaveBeenCalled();
  });
});

describe("useSetLeft", () => {
  it("marks a member as left and fetches the roster and races again", async () => {
    const setRosterMemberLeft = vi.fn(() => Promise.resolve({ ...rosterMember, left: true }));
    const { result, queryClient } = renderWith(
      () => useSetLeft(),
      createFakeApiClient({ setRosterMemberLeft }),
    );

    await result.current.mutateAsync({ rosterMemberId: "m1", left: true });

    expect(setRosterMemberLeft).toHaveBeenCalledWith("m1", true);
    expect(invalidated(queryClient)).toMatchObject({ rosterMembers: true, race: true });
  });
});

describe("useUnlink", () => {
  it("reports state_conflict and still fetches the roster again", async () => {
    const { result, queryClient } = renderWith(
      () => useUnlink(),
      createFakeApiClient({
        unlinkRosterMember: () => Promise.reject(problem(409, "state_conflict")),
      }),
    );

    result.current.mutate("m1");

    await waitFor(() => expect(result.current.error).toMatchObject({ code: "state_conflict" }));
    expect(invalidated(queryClient)).toEqual({
      rosterMembers: true,
      rosterMember: true,
      memberLinks: false,
      race: false,
    });
  });
});

describe("useCreateInviteLink", () => {
  it("resolves to the link to copy", async () => {
    const { result } = renderWith(
      () => useCreateInviteLink(),
      createFakeApiClient({
        createInviteLink: () =>
          Promise.resolve({
            url: "https://tasuki.example/invite/abc",
            expires_at: "2026-11-14T09:00:00Z",
          }),
      }),
    );

    await expect(result.current.mutateAsync("m1")).resolves.toEqual({
      url: "https://tasuki.example/invite/abc",
      expiresAt: "2026-11-14T09:00:00Z",
    });
  });

  it("reports state_conflict for a member who cannot get a link", async () => {
    const { result } = renderWith(
      () => useCreateInviteLink(),
      createFakeApiClient({
        createInviteLink: () => Promise.reject(problem(409, "state_conflict")),
      }),
    );

    result.current.mutate("m1");

    await waitFor(() => expect(result.current.error).toMatchObject({ code: "state_conflict" }));
  });
});
