import type { MouseEvent } from "react";

type TopBarProps = {
  // The screen it goes back to: its name and its path.
  backLabel: string;
  backHref: string;
  onBack: (href: string) => void;
};

// The bar at the top of the screens that are not in the tab bar: ‹ and the
// screen it goes back to, in blue-700 Headline (docs/design/components.md,
// Page frame).
export function TopBar({ backLabel, backHref, onBack }: TopBarProps) {
  const back = (event: MouseEvent) => {
    event.preventDefault();
    onBack(backHref);
  };
  return (
    <nav
      aria-label="戻る"
      className="flex min-h-11 items-center border-b border-slate-200 bg-white px-2"
    >
      <a
        href={backHref}
        onClick={back}
        className="flex min-h-11 min-w-0 items-center gap-1 px-2 text-headline font-semibold text-blue-700"
      >
        <span aria-hidden="true">‹</span>
        <span className="truncate">{backLabel}</span>
      </a>
    </nav>
  );
}
