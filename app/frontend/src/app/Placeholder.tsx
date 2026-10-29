type PlaceholderProps = {
  // The screen ID in docs/design/screens.md, if the route has one.
  screen?: string;
  title: string;
};

// Stands in for a screen until its feature's container is built.
export function Placeholder({ screen, title }: PlaceholderProps) {
  return <h1>{screen ? `${screen} ${title}` : title}</h1>;
}
