import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";
import type { ApiClient } from "../../../api/client.ts";
import { createFakeApiClient, problem } from "../../../api/fakeApiClient.ts";
import { TestProviders } from "../../../test/TestProviders.tsx";
import { createTestQueryClient } from "../../../test/testQueryClient.ts";
import { useMemberProfile } from "./useMemberProfile.ts";
import { usePendingLinks } from "./usePendingLinks.ts";
import { useRosterMembers } from "./useRosterMembers.ts";

function renderWith<T>(hook: () => T, client: ApiClient) {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <TestProviders client={client} queryClient={createTestQueryClient()}>
      {children}
    </TestProviders>
  );
  return renderHook(hook, { wrapper });
}

describe("useRosterMembers", () => {
  it("gets the roster for M1", async () => {
    const { result } = renderWith(
      () => useRosterMembers(),
      createFakeApiClient({
        listRosterMembers: () =>
          Promise.resolve({
            roster_members: [
              {
                id: "m1",
                name: "Taro",
                left: false,
                link_status: "approved",
                invite_link_expires_at: null,
              },
            ],
          }),
      }),
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.map((member) => [member.name, member.canUnlink])).toEqual([
      ["Taro", true],
    ]);
  });
});

describe("useMemberProfile", () => {
  it("gets a member's profile for M2", async () => {
    const { result } = renderWith(
      () => useMemberProfile("m1"),
      createFakeApiClient({
        getRosterMember: () =>
          Promise.resolve({
            id: "m1",
            name: "Taro",
            left: false,
            average_time: "3:40",
            lap_count: 0,
            laps_by_race: [],
          }),
      }),
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.averageTime).toBe("3:40");
  });

  it("reports not_found for an unknown member", async () => {
    const { result } = renderWith(
      () => useMemberProfile("unknown"),
      createFakeApiClient({ getRosterMember: () => Promise.reject(problem(404, "not_found")) }),
    );

    await waitFor(() => expect(result.current.error).toMatchObject({ code: "not_found" }));
  });
});

describe("usePendingLinks", () => {
  it("gets the waiting links for M3", async () => {
    const { result } = renderWith(
      () => usePendingLinks(),
      createFakeApiClient({
        listPendingMemberLinks: () =>
          Promise.resolve({
            links: [
              {
                id: "link-1",
                line_display_name: "taro_line",
                roster_member: { id: "m1", name: "Taro" },
                created_at: "2026-11-07T09:00:00Z",
              },
            ],
          }),
      }),
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.map((link) => link.lineDisplayName)).toEqual(["taro_line"]);
  });
});
