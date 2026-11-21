import { CenteredScreen } from "./CenteredScreen.tsx";

// A4. The same text for a member who has left and for an unusable invite link.
export function UnavailableScreen() {
  return (
    <CenteredScreen>
      <h1 className="text-title2 font-bold">tasuki を利用できません</h1>
      <p className="text-body">幹事に連絡してください。</p>
    </CenteredScreen>
  );
}
