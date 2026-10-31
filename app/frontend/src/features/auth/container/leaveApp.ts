// Leaves the app for another site, such as LINE Login. Its own module so
// tests can replace it: jsdom cannot navigate.
export function leaveApp(url: string): void {
  window.location.assign(url);
}
