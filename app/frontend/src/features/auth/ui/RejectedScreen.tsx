import { CenteredScreen } from "./CenteredScreen.tsx";

// A3. The reason is told outside the app; shows nothing of the team's data.
export function RejectedScreen() {
  return (
    <CenteredScreen>
      <h1 className="text-title2 font-bold">却下されました</h1>
      <p className="text-body">くわしくは幹事に聞いてください。</p>
    </CenteredScreen>
  );
}
