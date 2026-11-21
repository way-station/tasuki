import { blockButton, content, primaryButton } from "../../../shared/ui/classes.ts";

type NoEntryViewProps = {
  onOpenEntry: () => void;
};

// R7 before the race has an entry: runners are picked from the entry.
export function NoEntryView({ onOpenEntry }: NoEntryViewProps) {
  return (
    <section className={content}>
      <h1 className="text-title1 font-bold">走順</h1>
      <p className="text-body">
        先にエントリーを作ってください。走者はエントリーのメンバーから選びます。
      </p>
      <button type="button" onClick={onOpenEntry} className={`${primaryButton} ${blockButton}`}>
        エントリーへ
      </button>
    </section>
  );
}
