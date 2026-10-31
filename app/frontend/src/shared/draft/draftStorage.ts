// Drafts of edits not yet saved, kept in localStorage per user and per
// edited resource, so one user never sees another's draft (ADR-0013).
// Storage can be blocked (private mode): editing then still works, and only
// the draft is not kept.
const prefix = "tasuki.draft.";

// What the user typed, and the `version` of the data it was copied from. Saving
// it sends that version, so a draft older than the saved data gets
// `409 version_conflict`.
export type Draft<T> = {
  value: T;
  version: number;
};

export function draftKey(userId: string, resource: string): string {
  return `${userPrefix(userId)}${resource}`;
}

function userPrefix(userId: string): string {
  return `${prefix}${userId}.`;
}

export function readDraft<T>(key: string): Draft<T> | undefined {
  try {
    const stored = localStorage.getItem(key);
    if (stored === null) {
      return undefined;
    }
    const parsed: unknown = JSON.parse(stored);
    if (isDraft(parsed)) {
      return parsed as Draft<T>;
    }
  } catch {
    // Blocked storage or a broken entry: start from the saved data.
  }
  return undefined;
}

export function writeDraft<T>(key: string, draft: Draft<T>): void {
  try {
    localStorage.setItem(key, JSON.stringify(draft));
  } catch {
    // Blocked or full storage: the edit stays on screen, unsaved.
  }
}

export function removeDraft(key: string): void {
  try {
    localStorage.removeItem(key);
  } catch {
    // Blocked storage holds no draft.
  }
}

// Removes every draft of the user. Called when they log out.
export function removeUserDrafts(userId: string): void {
  try {
    const keys: string[] = [];
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key?.startsWith(userPrefix(userId))) {
        keys.push(key);
      }
    }
    for (const key of keys) {
      localStorage.removeItem(key);
    }
  } catch {
    // Same as above.
  }
}

function isDraft(value: unknown): value is Draft<unknown> {
  return (
    typeof value === "object" &&
    value !== null &&
    "value" in value &&
    "version" in value &&
    typeof value.version === "number"
  );
}
