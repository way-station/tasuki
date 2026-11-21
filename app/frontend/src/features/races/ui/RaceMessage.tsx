type RaceMessageProps = {
  title: string;
  message?: string;
};

// A race screen that only tells the user something: loading and errors. The
// title in Title 2 and the sentence in Body, centered (docs/design/components.md).
export function RaceMessage({ title, message }: RaceMessageProps) {
  return (
    <section className="flex flex-col gap-2 p-4 py-6 text-center">
      <h1 className="text-title2 font-bold">{title}</h1>
      {message !== undefined && <p className="text-body">{message}</p>}
    </section>
  );
}
