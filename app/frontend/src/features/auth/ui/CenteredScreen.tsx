import type { ReactNode } from "react";

type CenteredScreenProps = {
  children: ReactNode;
};

// The joining and access screens (A1–A5): a white card in the middle of the
// screen with the app's name on top, 24 px between its parts, centered text.
export function CenteredScreen({ children }: CenteredScreenProps) {
  return (
    <div className="flex min-h-[70dvh] items-center justify-center p-4">
      <section className="flex w-full max-w-md flex-col gap-6 rounded-xl bg-white px-6 py-10 text-center shadow-lg">
        <p className="text-footnote text-slate-600">tasuki</p>
        {children}
      </section>
    </div>
  );
}
