import { useSearch } from "@tanstack/react-router";
import { LoginScreen } from "../ui/LoginScreen.tsx";
import { useLineLogin } from "./useLineLogin.ts";

export function LoginContainer() {
  const { redirect } = useSearch({ from: "/login" });
  const login = useLineLogin({ returnPath: redirect });
  return <LoginScreen {...login} />;
}
