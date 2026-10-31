// Where to return after LINE Login, which leaves the app. Saved in
// localStorage just before leaving and read once at `/auth/complete`, so the
// user comes back to the screen they were on (ADR-0013).
const storageKey = "tasuki.returnPath";

export function saveReturnPath(path: string | undefined): void {
  try {
    if (path !== undefined && isAppPath(path)) {
      localStorage.setItem(storageKey, path);
    } else {
      localStorage.removeItem(storageKey);
    }
  } catch {
    // Storage can be blocked; the user then lands on the home screen.
  }
}

export function takeReturnPath(): string {
  try {
    const path = localStorage.getItem(storageKey);
    localStorage.removeItem(storageKey);
    if (path !== null && isAppPath(path)) {
      return path;
    }
  } catch {
    // Same as above.
  }
  return "/";
}

// Only a path inside the app, and never a login screen, which would loop.
function isAppPath(path: string): boolean {
  return (
    path.startsWith("/") &&
    !path.startsWith("//") &&
    !path.startsWith("/login") &&
    !path.startsWith("/auth/")
  );
}
