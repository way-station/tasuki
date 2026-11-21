import { CenteredScreen } from "./CenteredScreen.tsx";
import { LineLoginButton } from "./LineLoginButton.tsx";
import { LoginFailedNote } from "./LoginFailedNote.tsx";

type InviteLandingProps = {
  teamName: string;
  onLogin: () => void;
  loggingIn: boolean;
  failed: boolean;
};

// A1. Shows the team name, never the roster member's name (docs/design/screens.md).
export function InviteLanding({ teamName, onLogin, loggingIn, failed }: InviteLandingProps) {
  return (
    <CenteredScreen>
      <div className="flex flex-col gap-2">
        <p className="text-subheadline text-slate-600">チームに招待されています</p>
        <h1 className="text-title1 font-bold">{teamName}</h1>
      </div>
      {failed && <LoginFailedNote />}
      <LineLoginButton onLogin={onLogin} loggingIn={loggingIn} />
      <p className="text-footnote text-slate-600">ログインすると、幹事が承認するまで待ちます。</p>
    </CenteredScreen>
  );
}
