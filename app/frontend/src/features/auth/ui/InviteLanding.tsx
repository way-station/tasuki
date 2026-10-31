import { LineLoginButton } from "./LineLoginButton.tsx";

type InviteLandingProps = {
  teamName: string;
  onLogin: () => void;
  loggingIn: boolean;
  failed: boolean;
};

// A1. Shows the team name, never the roster member's name (docs/design/screens.md).
export function InviteLanding({ teamName, onLogin, loggingIn, failed }: InviteLandingProps) {
  return (
    <section>
      <h1>{teamName}</h1>
      {failed && <p role="alert">Login failed. Please try again.</p>}
      <LineLoginButton onLogin={onLogin} loggingIn={loggingIn} />
    </section>
  );
}
