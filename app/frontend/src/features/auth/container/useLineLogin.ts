import { useStartLineLogin } from "../hooks/useStartLineLogin.ts";
import { leaveApp } from "./leaveApp.ts";
import { saveReturnPath } from "./returnPath.ts";

type LineLoginOptions = {
  inviteToken?: string;
  // Where to come back to after logging in.
  returnPath?: string;
};

// Starts LINE Login and leaves the app for it.
export function useLineLogin({ inviteToken, returnPath }: LineLoginOptions) {
  const start = useStartLineLogin();
  return {
    onLogin: () => {
      saveReturnPath(returnPath);
      start.mutate(inviteToken, { onSuccess: (url) => leaveApp(url) });
    },
    // Still true after success, while the browser leaves.
    loggingIn: start.isPending || start.isSuccess,
    failed: start.isError,
  };
}
