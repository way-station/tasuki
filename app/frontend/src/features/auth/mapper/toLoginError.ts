import type { LoginError } from "../types/auth.ts";

// Converts the `error` the LINE callback redirects with (docs/design/api.md).
// An unknown code is treated as a failed login.
export function toLoginError(code: string): LoginError {
  switch (code) {
    case "invite_unusable":
    case "left":
      return "unavailable";
    default:
      return "failed";
  }
}
