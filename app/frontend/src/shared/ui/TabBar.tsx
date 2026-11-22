import type { MouseEvent } from "react";

export type Tab = {
  label: string;
  href: string;
  current: boolean;
  // A number shown in a red badge, hidden at 0 (承認: the links waiting).
  badge?: number;
};

type TabBarProps = {
  tabs: Tab[];
  onSelect: (href: string) => void;
};

// The tab bar at the bottom of the team's screens (docs/design/components.md,
// Page frame): white, a slate-200 line above, at least 49 px high. The tabs
// have no icons yet: they are not decided.
export function TabBar({ tabs, onSelect }: TabBarProps) {
  const select = (event: MouseEvent, href: string) => {
    event.preventDefault();
    onSelect(href);
  };
  return (
    <nav
      aria-label="メイン"
      data-tabbar=""
      className="fixed inset-x-0 bottom-0 z-[5] border-t border-slate-200 bg-white pb-[env(safe-area-inset-bottom)]"
    >
      <ul className="flex">
        {tabs.map((tab) => (
          <li key={tab.href} className="flex-1">
            <a
              href={tab.href}
              onClick={(event) => select(event, tab.href)}
              aria-current={tab.current ? "page" : undefined}
              aria-label={
                tab.badge !== undefined && tab.badge > 0
                  ? `${tab.label} ${tab.badge} 件`
                  : undefined
              }
              className={`flex min-h-[49px] items-center justify-center gap-1 text-footnote ${
                tab.current ? "font-semibold text-blue-700" : "text-slate-600"
              }`}
            >
              {tab.label}
              {tab.badge !== undefined && tab.badge > 0 && (
                <span className="min-w-5 rounded-full bg-red-700 px-1.5 text-center text-footnote font-semibold text-white">
                  {tab.badge}
                </span>
              )}
            </a>
          </li>
        ))}
      </ul>
    </nav>
  );
}
