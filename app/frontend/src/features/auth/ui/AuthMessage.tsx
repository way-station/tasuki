type AuthMessageProps = {
  title: string;
  message?: string;
};

// A screen that only tells the user something: loading, errors, and A2–A4.
export function AuthMessage({ title, message }: AuthMessageProps) {
  return (
    <section>
      <h1>{title}</h1>
      {message !== undefined && <p>{message}</p>}
    </section>
  );
}
