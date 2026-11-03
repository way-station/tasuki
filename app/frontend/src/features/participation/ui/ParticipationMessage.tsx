type ParticipationMessageProps = {
  title: string;
  message?: string;
};

// A screen that only tells the user something: loading and errors.
export function ParticipationMessage({ title, message }: ParticipationMessageProps) {
  return (
    <section>
      <h1>{title}</h1>
      {message !== undefined && <p>{message}</p>}
    </section>
  );
}
