import type { InviteLinkCreatedResponse, InviteLinkView } from "../types/roster.ts";

// M1: the link just created, to copy.
export function toInviteLinkView(created: InviteLinkCreatedResponse): InviteLinkView {
  return { url: created.url, expiresAt: created.expires_at };
}
