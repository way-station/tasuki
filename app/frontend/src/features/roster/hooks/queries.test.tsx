import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";
import type { ApiClient } from "../../../api/client.ts";
import { createFakeApiClient } from "../../../api/fakeApiClient.ts";
import { TestProviders } from "../../../test/TestProviders.tsx";
import { createTestQueryClient } from "../../../test/testQueryClient.ts";
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
