type RosterMessageProps = {
  title: string;
  message?: string;
};

// A screen that only tells the user something: loading and errors.
export function RosterMessage({ title, message }: RosterMessageProps) {
  return (
    <section>
      <h1>{title}</h1>
      {message !== undefined && <p>{message}</p>}
    </section>
  );
}
