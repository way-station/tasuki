import { afterEach, describe, expect, it, vi } from "vitest";
import { draftKey, readDraft, removeUserDrafts, removeDraft, writeDraft } from "./draftStorage.ts";

afterEach(() => {
  localStorage.clear();
  vi.restoreAllMocks();
});

describe("draftStorage", () => {
  it("reads back a draft with its version", () => {
    writeDraft(draftKey("user-1", "race-1"), { value: { name: "Ekiden" }, version: 3 });

    expect(readDraft(draftKey("user-1", "race-1"))).toEqual({
      value: { name: "Ekiden" },
      version: 3,
    });
  });

  it("keeps each user's and each resource's drafts apart", () => {
    writeDraft(draftKey("user-1", "race-1"), { value: "mine", version: 1 });

    expect(readDraft(draftKey("user-2", "race-1"))).toBeUndefined();
    expect(readDraft(draftKey("user-1", "race-2"))).toBeUndefined();
  });

  it("removes one draft", () => {
    writeDraft(draftKey("user-1", "race-1"), { value: "mine", version: 1 });

    removeDraft(draftKey("user-1", "race-1"));

    expect(readDraft(draftKey("user-1", "race-1"))).toBeUndefined();
  });

  it("removes every draft of a user, and only that user's", () => {
    writeDraft(draftKey("user-1", "race-1"), { value: "mine", version: 1 });
    writeDraft(draftKey("user-1", "running-order-1"), { value: "mine", version: 1 });
    writeDraft(draftKey("user-2", "race-1"), { value: "theirs", version: 1 });
    localStorage.setItem("tasuki.returnPath", "/races/1");

    removeUserDrafts("user-1");

    expect(readDraft(draftKey("user-1", "race-1"))).toBeUndefined();
    expect(readDraft(draftKey("user-1", "running-order-1"))).toBeUndefined();
    expect(readDraft(draftKey("user-2", "race-1"))).toEqual({ value: "theirs", version: 1 });
    expect(localStorage.getItem("tasuki.returnPath")).toBe("/races/1");
  });

  it.each(["not json", '{"value":"no version"}', '{"value":"x","version":"1"}'])(
    "ignores a broken entry %s",
    (stored) => {
      localStorage.setItem(draftKey("user-1", "race-1"), stored);

      expect(readDraft(draftKey("user-1", "race-1"))).toBeUndefined();
    },
  );

  it("does not throw when storage is blocked", () => {
    for (const method of ["getItem", "setItem", "removeItem", "key"] as const) {
      vi.spyOn(Storage.prototype, method).mockImplementation(() => {
        throw new DOMException("blocked", "SecurityError");
      });
    }

    expect(() => {
      writeDraft(draftKey("user-1", "race-1"), { value: "mine", version: 1 });
      removeDraft(draftKey("user-1", "race-1"));
      removeUserDrafts("user-1");
    }).not.toThrow();
    expect(readDraft(draftKey("user-1", "race-1"))).toBeUndefined();
  });
});
