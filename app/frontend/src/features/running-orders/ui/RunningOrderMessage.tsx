type RunningOrderMessageProps = {
  title: string;
  message?: string;
};

// A screen with nothing but a heading and a line, while loading or on errors.
// The title in Title 2 and the sentence in Body, centered
// (docs/design/components.md).
export function RunningOrderMessage({ title, message }: RunningOrderMessageProps) {
  return (
    <section className="flex flex-col gap-2 p-4 py-6 text-center">
      <h1 className="text-title2 font-bold">{title}</h1>
      {message !== undefined && <p className="text-body">{message}</p>}
    </section>
  );
}
