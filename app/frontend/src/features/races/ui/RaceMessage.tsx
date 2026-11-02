type RaceMessageProps = {
  title: string;
  message?: string;
};

// A race screen that only tells the user something: loading and errors.
export function RaceMessage({ title, message }: RaceMessageProps) {
  return (
    <section>
      <h1>{title}</h1>
      {message !== undefined && <p>{message}</p>}
    </section>
  );
}
