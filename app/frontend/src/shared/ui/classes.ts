// Tailwind classes of the common components (docs/design/components.md),
// shared by every feature's ui/ so each screen looks the same.

const buttonBase =
  "inline-flex min-h-11 items-center justify-center rounded-lg font-semibold disabled:border-0 disabled:bg-slate-100 disabled:text-slate-600";

// Variants (docs/design/components.md, Buttons). Add one size below.
// The main action of a screen or a form. At most one per form.
export const primaryButton = `${buttonBase} bg-blue-700 text-white`;
export const secondaryButton = `${buttonBase} border border-slate-500 bg-white text-blue-700`;
export const dangerButton = `${buttonBase} border border-red-700 bg-white text-red-700`;

// Sizes. The main action of a form spans the full width. Buttons inside a
// card are smaller: 12 px side padding and Subheadline, still 44 px high.
export const blockButton = "flex w-full px-4";
export const regularButton = "px-4";
export const smallButton = "px-3 text-subheadline";

export const input =
  "min-h-11 w-full rounded-lg border border-slate-500 bg-white px-3 text-body text-slate-900 aria-invalid:border-red-700 disabled:bg-slate-100 disabled:text-slate-600";
export const fieldError = "text-footnote text-red-700";
export const fieldHelp = "text-footnote text-slate-600";

// A screen's content: 16 px from each side, 24 px between groups.
export const content = "flex flex-col gap-6 p-4";
export const group = "flex flex-col gap-2";
export const groupHeading = "text-footnote text-slate-600";
export const cards = "flex flex-col gap-2";
export const card =
  "flex min-h-18 items-center gap-3 rounded-xl border border-slate-200 bg-white px-4 py-3";
export const cardStack = "flex-col items-stretch";
export const panel = "flex flex-col gap-3 rounded-xl border border-slate-200 bg-white p-4";
// Numbers with equal widths, so times line up.
export const numbers = "tabular-nums";

// State labels: always text, Footnote, fully round (docs/design/components.md).
export const label = "shrink-0 whitespace-nowrap rounded-full px-2 py-0.5 text-footnote";
export const confirmedLabel = "bg-green-100 text-green-800";
export const waitingLabel = "bg-amber-100 text-amber-800";
export const neutralLabel = "bg-slate-100 text-slate-700";

// Messages above a form or in a panel.
export const warningNote = "rounded-lg bg-amber-100 px-4 py-3 text-subheadline text-amber-800";
