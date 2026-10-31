import { ApiError, type ProblemCode } from "./ApiError.ts";
import type { ApiClient } from "./client.ts";

// A fake ApiClient for tests. A test stubs only the methods it needs; the
// others reject, so an unexpected call fails the test.
//
// A fake must keep the real client's contract (ADR-0012): stubs return what
// the real method returns, and report errors with problem(), which throws the
// same ApiError the real client throws.
export function createFakeApiClient(stubs: Partial<ApiClient> = {}): ApiClient {
  return {
    getHealth: notStubbed("getHealth"),
    getInviteLink: notStubbed("getInviteLink"),
    authorizeLine: notStubbed("authorizeLine"),
    createSession: notStubbed("createSession"),
    logout: notStubbed("logout"),
    getMe: notStubbed("getMe"),
    ...stubs,
  };
}

// The ApiError the real client throws for a problem+json response.
export function problem(status: number, code: ProblemCode): ApiError {
  return new ApiError(status, { type: "about:blank", title: code, status, code });
}

function notStubbed(name: keyof ApiClient) {
  return () => Promise.reject(new Error(`${name} is not stubbed`));
}
