import type { ReactNode } from "react";

type CenteredScreenProps = {
  children: ReactNode;
};

// The joining and access screens (A1–A5): one column in the middle, 24 px
// between its parts, centered text (docs/design/wireframes/access.html).
export function CenteredScreen({ children }: CenteredScreenProps) {
  return (
    <section className="flex min-h-[70dvh] flex-col justify-center gap-6 p-4 text-center">
      {children}
    </section>
  );
}
