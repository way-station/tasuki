import { type ReactNode, useEffect, useRef, useState } from "react";
import { NotificationList } from "../ui/NotificationList.tsx";
import { type Notification, NotificationContext, type Notifier } from "./notificationContext.ts";

// How long a notification stays (#202: every notification disappears by itself).
export const notificationDuration = 4000;

type NotificationProviderProps = {
  children: ReactNode;
};

// Holds the notifications above the router, so one shown just before
// navigating is still visible on the next screen.
export function NotificationProvider({ children }: NotificationProviderProps) {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const nextId = useRef(1);
  const timers = useRef(new Map<number, ReturnType<typeof setTimeout>>());

  const [notifier] = useState<Notifier>(() => {
    const dismiss = (id: number) => {
      clearTimeout(timers.current.get(id));
      timers.current.delete(id);
      setNotifications((shown) => shown.filter((notification) => notification.id !== id));
    };
    const notify: Notifier["notify"] = (kind, message) => {
      const id = nextId.current++;
      setNotifications((shown) => [...shown, { id, kind, message }]);
      timers.current.set(
        id,
        setTimeout(() => dismiss(id), notificationDuration),
      );
      return id;
    };
    return { notify, dismiss };
  });

  useEffect(() => {
    const pending = timers.current;
    return () => {
      for (const timer of pending.values()) {
        clearTimeout(timer);
      }
    };
  }, []);

  return (
    <NotificationContext value={notifier}>
      {children}
      <NotificationList notifications={notifications} />
    </NotificationContext>
  );
}
