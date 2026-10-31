type LineLoginButtonProps = {
  onLogin: () => void;
  loggingIn: boolean;
};

export function LineLoginButton({ onLogin, loggingIn }: LineLoginButtonProps) {
  return (
    <button type="button" onClick={onLogin} disabled={loggingIn}>
      Log in with LINE
    </button>
  );
}
