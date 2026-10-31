import { createContext } from "react";

export type NotificationKind = "success" | "error";

// A message shown after saving, for example "Saved", or that someone else
// saved first (`409 version_conflict`).
export type Notification = {
  id: number;
  kind: NotificationKind;
  message: string;
};

export type Notifier = {
  // Shows a notification, and returns its id. It disappears by itself.
  notify: (kind: NotificationKind, message: string) => number;
  dismiss: (id: number) => void;
};

// Notifications are UI state shared across screens, the one thing ADR-0013
// keeps in Context. Nothing from the backend goes in.
export const NotificationContext = createContext<Notifier | null>(null);
