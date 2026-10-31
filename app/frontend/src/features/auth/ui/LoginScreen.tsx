import { LineLoginButton } from "./LineLoginButton.tsx";

type LoginScreenProps = {
  onLogin: () => void;
  loggingIn: boolean;
  // The last login failed, was cancelled, or could not start.
  failed: boolean;
};

// The login screen for returning users, also shown when a login failed.
export function LoginScreen({ onLogin, loggingIn, failed }: LoginScreenProps) {
  return (
    <section>
      <h1>Log in</h1>
      {failed && <p role="alert">Login failed. Please try again.</p>}
      <LineLoginButton onLogin={onLogin} loggingIn={loggingIn} />
    </section>
  );
}
