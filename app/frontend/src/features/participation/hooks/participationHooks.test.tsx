import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import type { ApiClient, InvitationList } from "../../../api/client.ts";
import { createFakeApiClient, problem } from "../../../api/fakeApiClient.ts";
import { queryKeys } from "../../../api/queryKeys.ts";
import { TestProviders } from "../../../test/TestProviders.tsx";
import { createTestQueryClient } from "../../../test/testQueryClient.ts";
import {
  useAddEntryMember,
  useAnswerAttendance,
  useCancelInvitation,
  useCreateEntry,
  useInvite,
  useRemoveEntryMember,
} from "./mutations.ts";
import { useAttendanceRows } from "./useAttendanceRows.ts";
import { useEntryEditor } from "./useEntryEditor.ts";
import { useInvitationRows } from "./useInvitationRows.ts";

function renderWith<T>(hook: () => T, client: ApiClient, queryClient = createTestQueryClient()) {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <TestProviders client={client} queryClient={queryClient}>
      {children}
    </TestProviders>
  );
  return { ...renderHook(hook, { wrapper }), queryClient };
}

// A cache holding the race list and the race's invitations, as on R4.
function cachedRace() {
  const queryClient = createTestQueryClient();
  queryClient.setQueryData(queryKeys.races, { races: [] });
  queryClient.setQueryData(queryKeys.invitations("race-1"), invitations);
  return queryClient;
}

function invalidated(queryClient: ReturnType<typeof createTestQueryClient>) {
  return [queryKeys.races, queryKeys.invitations("race-1")].map(
    (key) => queryClient.getQueryState(key)?.isInvalidated ?? false,
  );
}

const invitations: InvitationList = {
  members: [
    { roster_member_id: "m1", name: "Taro", invited: true, attendance: "attending" },
    { roster_member_id: "m2", name: "Jiro", invited: false, attendance: null },
  ],
};

describe("queries", () => {
  it("useInvitationRows lists who is invited", async () => {
    const { result } = renderWith(
      () => useInvitationRows("race-1"),
      createFakeApiClient({ listInvitations: () => Promise.resolve(invitations) }),
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.map((row) => row.invited)).toEqual([true, false]);
  });

  it("useAttendanceRows lists invited members' answers", async () => {
    const { result } = renderWith(
      () => useAttendanceRows("race-1"),
      createFakeApiClient({ listInvitations: () => Promise.resolve(invitations) }),
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([
      { rosterMemberId: "m1", name: "Taro", answer: "attending" },
    ]);
  });

  it("useEntryEditor combines the entries and the invitations", async () => {
    const { result } = renderWith(
      () => useEntryEditor("race-1"),
      createFakeApiClient({
        listEntries: () => Promise.resolve({ entries: [] }),
        listInvitations: () => Promise.resolve(invitations),
      }),
    );

    await waitFor(() => expect(result.current.data).toBeDefined());
    expect(result.current.data).toEqual({
      entry: null,
      candidates: [{ rosterMemberId: "m1", name: "Taro" }],
    });
  });

  it("useEntryEditor fails when either list fails", async () => {
    const { result } = renderWith(
      () => useEntryEditor("race-1"),
      createFakeApiClient({
        listEntries: () => Promise.reject(problem(404, "not_found")),
        listInvitations: () => Promise.resolve(invitations),
      }),
    );

    await waitFor(() => expect(result.current.error).toMatchObject({ code: "not_found" }));
    expect(result.current.data).toBeUndefined();
  });
});

describe("mutations", () => {
  it("useInvite invites the member and fetches the race's data again", async () => {
    const invite = vi.fn(() => Promise.resolve());
    const { result, queryClient } = renderWith(
      () => useInvite("race-1"),
      createFakeApiClient({ invite }),
      cachedRace(),
    );

    await result.current.mutateAsync("m2");

    expect(invite).toHaveBeenCalledWith("race-1", "m2");
    expect(invalidated(queryClient)).toEqual([true, true]);
  });

  it("useInvite fails with validation_failed for a member who left, and still fetches again", async () => {
    const { result, queryClient } = renderWith(
      () => useInvite("race-1"),
      createFakeApiClient({ invite: () => Promise.reject(problem(422, "validation_failed")) }),
      cachedRace(),
    );

    await expect(result.current.mutateAsync("m2")).rejects.toMatchObject({
      code: "validation_failed",
    });
    await waitFor(() => expect(invalidated(queryClient)).toEqual([true, true]));
  });

  it("useCancelInvitation cancels the invitation", async () => {
    const cancelInvitation = vi.fn(() => Promise.resolve());
    const { result } = renderWith(
      () => useCancelInvitation("race-1"),
      createFakeApiClient({ cancelInvitation }),
    );

    await result.current.mutateAsync("m1");

    expect(cancelInvitation).toHaveBeenCalledWith("race-1", "m1");
  });

  it("useAnswerAttendance sends the answer and fetches races again", async () => {
    const answerAttendance = vi.fn(() => Promise.resolve());
    const { result, queryClient } = renderWith(
      () => useAnswerAttendance("race-1"),
      createFakeApiClient({ answerAttendance }),
      cachedRace(),
    );

    await result.current.mutateAsync("undecided");

    expect(answerAttendance).toHaveBeenCalledWith("race-1", "undecided");
    expect(invalidated(queryClient)[0]).toBe(true);
  });

  it("useAnswerAttendance fails with not_found after answering not attending", async () => {
    const { result } = renderWith(
      () => useAnswerAttendance("race-1"),
      createFakeApiClient({ answerAttendance: () => Promise.reject(problem(404, "not_found")) }),
    );

    await expect(result.current.mutateAsync("attending")).rejects.toMatchObject({
      code: "not_found",
    });
  });

  it("useCreateEntry creates the entry, and fails with state_conflict for a second one", async () => {
    const createEntry = vi
      .fn<ApiClient["createEntry"]>()
      .mockResolvedValueOnce({ id: "e1", name: "Team A", members: [] })
      .mockRejectedValueOnce(problem(409, "state_conflict"));
    const { result } = renderWith(
      () => useCreateEntry("race-1"),
      createFakeApiClient({ createEntry }),
    );

    await result.current.mutateAsync("Team A");
    await expect(result.current.mutateAsync("Team B")).rejects.toMatchObject({
      code: "state_conflict",
    });
    expect(createEntry).toHaveBeenCalledWith("race-1", "Team A");
  });

  it("useAddEntryMember adds the member, and fails with validation_failed when not attending", async () => {
    const addEntryMember = vi
      .fn<ApiClient["addEntryMember"]>()
      .mockResolvedValueOnce()
      .mockRejectedValueOnce(problem(422, "validation_failed"));
    const { result } = renderWith(
      () => useAddEntryMember("race-1"),
      createFakeApiClient({ addEntryMember }),
    );

    await result.current.mutateAsync({ entryId: "e1", rosterMemberId: "m1" });
    await expect(
      result.current.mutateAsync({ entryId: "e1", rosterMemberId: "m2" }),
    ).rejects.toMatchObject({ code: "validation_failed" });
    expect(addEntryMember).toHaveBeenCalledWith("race-1", "e1", "m1");
  });

  it("useRemoveEntryMember removes the member", async () => {
    const removeEntryMember = vi.fn(() => Promise.resolve());
    const { result } = renderWith(
      () => useRemoveEntryMember("race-1"),
      createFakeApiClient({ removeEntryMember }),
    );

    await result.current.mutateAsync({ entryId: "e1", rosterMemberId: "m1" });

    expect(removeEntryMember).toHaveBeenCalledWith("race-1", "e1", "m1");
  });
});
