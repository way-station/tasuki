import { useApiClient } from "../../../api/useApiClient.ts";
import type { AttendanceAnswer } from "../types/participation.ts";
import { useRaceMutation } from "./useRaceMutation.ts";

type EntryMember = {
  entryId: string;
  rosterMemberId: string;
};

// R4. Fails with validation_failed when the member has left.
export function useInvite(raceId: string) {
  const api = useApiClient();
  return useRaceMutation((rosterMemberId: string) => api.invite(raceId, rosterMemberId));
}

// R4. Also removes the member from attendance, the entry, and the running
// order (docs/design/api.md).
export function useCancelInvitation(raceId: string) {
  const api = useApiClient();
  return useRaceMutation((rosterMemberId: string) => api.cancelInvitation(raceId, rosterMemberId));
}

// R2. Fails with not_found when the caller is not invited or already answered
// not attending.
export function useAnswerAttendance(raceId: string) {
  const api = useApiClient();
  return useRaceMutation((answer: AttendanceAnswer) => api.answerAttendance(raceId, answer));
}

// R6. Fails with state_conflict when the race already has an entry.
export function useCreateEntry(raceId: string) {
  const api = useApiClient();
  return useRaceMutation((name: string) => api.createEntry(raceId, name));
}

// R6. Fails with validation_failed when the member's answer is not attending.
export function useAddEntryMember(raceId: string) {
  const api = useApiClient();
  return useRaceMutation(({ entryId, rosterMemberId }: EntryMember) =>
    api.addEntryMember(raceId, entryId, rosterMemberId),
  );
}

// R6. Also removes the member from the running order.
export function useRemoveEntryMember(raceId: string) {
  const api = useApiClient();
  return useRaceMutation(({ entryId, rosterMemberId }: EntryMember) =>
    api.removeEntryMember(raceId, entryId, rosterMemberId),
  );
}
