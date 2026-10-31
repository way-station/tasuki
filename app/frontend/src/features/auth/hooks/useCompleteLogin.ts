import { useEffect, useRef } from "react";
import { toLoginError } from "../mapper/toLoginError.ts";
import type { LoginState } from "../types/auth.ts";
import { useCreateSession } from "./useCreateSession.ts";

type CompleteLoginSearch = {
  code?: string;
  error?: string;
};

// Finishes LINE Login from what the callback redirected with: an `error`, or
// a one-time `code` exchanged for a session (docs/design/api.md).
export function useCompleteLogin({ code, error }: CompleteLoginSearch): LoginState {
  const { mutate, isSuccess, isError } = useCreateSession();
  // The code can be exchanged only once, and Strict Mode runs effects twice.
  const exchanged = useRef(false);

  useEffect(() => {
    if (exchanged.current || error !== undefined || code === undefined) {
      return;
    }
    exchanged.current = true;
    mutate(code);
  }, [code, error, mutate]);

  if (error !== undefined) {
    return toLoginError(error);
  }
  if (code === undefined) {
    return "failed";
  }
  if (isSuccess) {
    return "done";
  }
  if (isError) {
    return "failed";
  }
  return "pending";
}
