type NotificationListProps = {
  notifications: { id: number; kind: "success" | "error"; message: string }[];
};

// Shows the notifications it receives. The list is always there, so screen
// readers announce what is added: errors at once, the rest when they are free.
// White cards 16 px from the sides, and 16 px above the tab bar when there is
// one (docs/design/components.md, Notifications).
export function NotificationList({ notifications }: NotificationListProps) {
  return (
    <div
      aria-live="polite"
      className="fixed inset-x-4 bottom-4 z-[6] flex flex-col gap-2 [body:has([data-tabbar])_&]:bottom-[calc(49px+16px+env(safe-area-inset-bottom))]"
    >
      {notifications.map((notification) => (
        <p
          key={notification.id}
          role={notification.kind === "error" ? "alert" : undefined}
          className={`rounded-lg border bg-white px-4 py-3 text-body ${
            notification.kind === "error" ? "border-red-700 text-red-700" : "border-slate-200"
          }`}
        >
          {notification.message}
        </p>
      ))}
    </div>
  );
}
