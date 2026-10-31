import { useState } from "react";
import { type Draft, draftKey, readDraft, removeDraft, writeDraft } from "./draftStorage.ts";

type Held<T> = {
  key: string | undefined;
  draft: Draft<T> | undefined;
  // The version of the draft that was cleared. The data is copied again only
  // once the query has a different version.
  clearedVersion: number | undefined;
};

// The draft of an editing screen, used by its Container (ADR-0013).
//
// `source` is the saved data from the query, and `undefined` while it loads.
// The draft is restored from localStorage, or else copied from `source` once;
// data fetched again in the background does not overwrite it. `edit` saves the
// draft on every edit. Call `clear` after a successful save, and to load the
// latest data after `409 version_conflict`: the draft is then copied again
// when the query brings a new version.
export function useDraft<T>(
  userId: string | undefined,
  resource: string,
  source: Draft<T> | undefined,
) {
  const key = userId === undefined ? undefined : draftKey(userId, resource);
  const [held, setHeld] = useState<Held<T>>({
    key: undefined,
    draft: undefined,
    clearedVersion: undefined,
  });

  let current = held;
  if (current.key !== key) {
    // Another user or resource: start over from its own draft.
    current = {
      key,
      draft: key === undefined ? undefined : readDraft<T>(key),
      clearedVersion: undefined,
    };
  }
  if (
    current.key !== undefined &&
    current.draft === undefined &&
    source !== undefined &&
    source.version !== current.clearedVersion
  ) {
    current = { ...current, draft: source };
  }
  if (current !== held) {
    setHeld(current);
  }

  const edit = (value: T) => {
    if (current.key === undefined || current.draft === undefined) {
      return;
    }
    const draft = { value, version: current.draft.version };
    writeDraft(current.key, draft);
    setHeld({ ...current, draft });
  };

  const clear = () => {
    if (current.key === undefined) {
      return;
    }
    removeDraft(current.key);
    setHeld({ key: current.key, draft: undefined, clearedVersion: current.draft?.version });
  };

  return {
    value: current.draft?.value,
    version: current.draft?.version,
    edit,
    clear,
  };
}
