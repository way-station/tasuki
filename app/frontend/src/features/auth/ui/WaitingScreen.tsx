import { label, waitingLabel } from "../../../shared/ui/classes.ts";
import { CenteredScreen } from "./CenteredScreen.tsx";

// A2. Shows nothing of the team's data.
export function WaitingScreen() {
  return (
    <CenteredScreen>
      <p>
        <span className={`${label} ${waitingLabel}`}>承認待ち</span>
      </p>
      <h1 className="text-title2 font-bold">幹事の承認を待っています</h1>
      <p className="text-body">承認されると、この画面を開き直したときに大会の一覧が出ます。</p>
    </CenteredScreen>
  );
}
