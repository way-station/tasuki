import { Navigate, useRouter, useSearch } from "@tanstack/react-router";
import { useEffect, useRef } from "react";
import { useCompleteLogin } from "../hooks/useCompleteLogin.ts";
import { AuthMessage } from "../ui/AuthMessage.tsx";
import { LoginScreen } from "../ui/LoginScreen.tsx";
import { takeReturnPath } from "./returnPath.ts";
import { useLineLogin } from "./useLineLogin.ts";

// `/auth/complete`, where the LINE callback redirects (docs/design/api.md).
export function AuthCompleteContainer() {
  const search = useSearch({ from: "/auth/complete" });
  const state = useCompleteLogin(search);
  const router = useRouter();
  const retry = useLineLogin({});
  // The return path is read once, and Strict Mode runs effects twice.
  const returned = useRef(false);

  useEffect(() => {
    if (state === "done" && !returned.current) {
      returned.current = true;
      router.history.replace(takeReturnPath());
    }
  }, [state, router]);

  switch (state) {
    case "pending":
    case "done":
      return <AuthMessage title="Logging in…" />;
    case "unavailable":
      return <Navigate to="/unavailable" replace />;
    case "failed":
      return <LoginScreen {...retry} failed />;
  }
}
