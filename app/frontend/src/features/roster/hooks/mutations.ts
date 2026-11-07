import { type QueryKey, useMutation, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "../../../api/queryKeys.ts";
import { useApiClient } from "../../../api/useApiClient.ts";
import { toRosterMemberBody } from "../mapper/rosterMemberForm.ts";
import { toInviteLinkView } from "../mapper/toInviteLinkView.ts";
import type { RosterMemberForm } from "../types/roster.ts";
import { RosterMemberFormInvalidError } from "./rosterMemberFormInvalid.ts";

type SetLeft = {
  rosterMemberId: string;
  left: boolean;
};

// The queries a change affects are fetched again afterwards, after a failure
// too. Nothing is updated before the backend answers (ADR-0013).
function useRosterMutation<Variables, Result>(
  mutationFn: (variables: Variables) => Promise<Result>,
  affected: QueryKey[],
) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn,
    onSettled: () =>
      Promise.all(affected.map((queryKey) => queryClient.invalidateQueries({ queryKey }))),
  });
}

// M1. A new member can be invited to races, so the races' lists are fetched
// again too. An invalid form fails with RosterMemberFormInvalidError before
// anything is sent.
export function useAddRosterMember() {
  const api = useApiClient();
  return useRosterMutation(
    async (form: RosterMemberForm) => {
      const result = toRosterMemberBody(form);
      if (!result.ok) {
        throw new RosterMemberFormInvalidError(result.errors);
      }
      await api.createRosterMember(result.body.name);
    },
    [queryKeys.rosterMembers, queryKeys.races],
  );
}

// M1. Leaving cancels invitations to races dated today or later, with what
// cancelling changes (docs/design/api.md), so every race query is fetched
// again too.
export function useSetLeft() {
  const api = useApiClient();
  return useRosterMutation(
    async ({ rosterMemberId, left }: SetLeft) => {
      await api.setRosterMemberLeft(rosterMemberId, left);
    },
    [queryKeys.rosterMembers, queryKeys.races],
  );
}

// M1. Fails with state_conflict when the member has no approved user.
export function useUnlink() {
  const api = useApiClient();
  return useRosterMutation(
    (rosterMemberId: string) => api.unlinkRosterMember(rosterMemberId),
    [queryKeys.rosterMembers],
  );
}

// M1. Resolves to the link to copy. Fails with state_conflict when the member
// has left or has a user, approved or waiting.
export function useCreateInviteLink() {
  const api = useApiClient();
  return useRosterMutation(
    async (rosterMemberId: string) => toInviteLinkView(await api.createInviteLink(rosterMemberId)),
    [queryKeys.rosterMembers],
  );
}

// M3. Both fail with state_conflict when the link is no longer waiting. The
// roster shows the new link status.
export function useApproveLink() {
  const api = useApiClient();
  return useRosterMutation(
    (linkId: string) => api.approveMemberLink(linkId),
    [queryKeys.memberLinks, queryKeys.rosterMembers],
  );
}

export function useRejectLink() {
  const api = useApiClient();
  return useRosterMutation(
    (linkId: string) => api.rejectMemberLink(linkId),
    [queryKeys.memberLinks, queryKeys.rosterMembers],
  );
}
