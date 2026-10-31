import { use } from "react";
import { NotificationContext, type Notifier } from "./notificationContext.ts";

export function useNotifications(): Notifier {
  const notifier = use(NotificationContext);
  if (notifier === null) {
    throw new Error("useNotifications must be used inside NotificationProvider");
  }
  return notifier;
}
