type RunningOrderMessageProps = {
  title: string;
  message?: string;
};

// A screen with nothing but a heading and a line, while loading or on errors.
export function RunningOrderMessage({ title, message }: RunningOrderMessageProps) {
  return (
    <section>
      <h1>{title}</h1>
      {message !== undefined && <p>{message}</p>}
    </section>
  );
}
