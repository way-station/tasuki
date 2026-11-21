import { CenteredScreen } from "./CenteredScreen.tsx";
import { LineLoginButton } from "./LineLoginButton.tsx";
import { LoginFailedNote } from "./LoginFailedNote.tsx";

type LoginScreenProps = {
  onLogin: () => void;
  loggingIn: boolean;
  // The last login failed, was cancelled, or could not start.
  failed: boolean;
};

// A5: the login screen for returning users, also shown when a login failed.
export function LoginScreen({ onLogin, loggingIn, failed }: LoginScreenProps) {
  return (
    <CenteredScreen>
      <h1 className="text-title2 font-bold">ログイン</h1>
      {failed && <LoginFailedNote />}
      <LineLoginButton onLogin={onLogin} loggingIn={loggingIn} />
      {/* Provisional: a place for other ways to log in, such as logging in
          on a PC later. It leads nowhere yet. */}
      <a
        href="#"
        onClick={(event) => event.preventDefault()}
        className="text-body text-blue-700 underline-offset-2 hover:underline"
      >
        そのほかの方法でログイン
      </a>
    </CenteredScreen>
  );
}
