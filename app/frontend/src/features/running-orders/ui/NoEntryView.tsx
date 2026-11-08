type NoEntryViewProps = {
  onOpenEntry: () => void;
};

// R7 before the race has an entry: runners are picked from the entry.
export function NoEntryView({ onOpenEntry }: NoEntryViewProps) {
  return (
    <section>
      <h1>Running order</h1>
      <p>Create the entry first. Runners are picked from its members.</p>
      <button type="button" onClick={onOpenEntry}>
        Go to the entry
      </button>
    </section>
  );
}
