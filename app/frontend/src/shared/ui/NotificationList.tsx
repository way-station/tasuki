type NotificationListProps = {
  notifications: { id: number; kind: "success" | "error"; message: string }[];
};

// Shows the notifications it receives. The list is always there, so screen
// readers announce what is added: errors at once, the rest when they are free.
// The look comes with #147.
export function NotificationList({ notifications }: NotificationListProps) {
  return (
    <div aria-live="polite" className="fixed inset-x-4 bottom-4 flex flex-col gap-2">
      {notifications.map((notification) => (
        <p
          key={notification.id}
          role={notification.kind === "error" ? "alert" : undefined}
          className="rounded border bg-white p-3"
        >
          {notification.message}
        </p>
      ))}
    </div>
  );
}
