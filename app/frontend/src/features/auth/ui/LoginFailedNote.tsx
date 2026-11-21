// Shown on A1 and A5 when LINE Login failed, was cancelled, or could not start.
export function LoginFailedNote() {
  return (
    <p
      role="alert"
      className="rounded-lg border border-red-700 bg-white px-4 py-3 text-subheadline text-red-700"
    >
      ログインできませんでした。もう一度お試しください。
    </p>
  );
}
