import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { type Draft, draftKey, readDraft, writeDraft } from "./draftStorage.ts";
import { useDraft } from "./useDraft.ts";

afterEach(() => {
  localStorage.clear();
  vi.restoreAllMocks();
});

type Props = { userId: string | undefined; source: Draft<string> | undefined };

function renderDraft(props: Props) {
  return renderHook(({ userId, source }: Props) => useDraft(userId, "race-1", source), {
    initialProps: props,
  });
}

const saved = { value: "saved", version: 1 };

describe("useDraft", () => {
  it("copies the saved data when editing starts", () => {
    const { result, rerender } = renderDraft({ userId: "user-1", source: undefined });
    expect(result.current.value).toBeUndefined();

    rerender({ userId: "user-1", source: saved });

    expect(result.current.value).toBe("saved");
    expect(result.current.version).toBe(1);
  });

  it("waits for the user", () => {
    const { result } = renderDraft({ userId: undefined, source: saved });

    expect(result.current.value).toBeUndefined();
  });

  it("saves every edit with the version it was copied from", () => {
    const { result } = renderDraft({ userId: "user-1", source: saved });

    act(() => result.current.edit("typed"));

    expect(result.current.value).toBe("typed");
    expect(readDraft(draftKey("user-1", "race-1"))).toEqual({ value: "typed", version: 1 });
  });

  it("is not overwritten by data fetched again", () => {
    const { result, rerender } = renderDraft({ userId: "user-1", source: saved });
    act(() => result.current.edit("typed"));

    rerender({ userId: "user-1", source: { value: "someone else's", version: 2 } });

    expect(result.current.value).toBe("typed");
    expect(result.current.version).toBe(1);
  });

  it("is restored after a reload, with the version it was copied from", () => {
    const first = renderDraft({ userId: "user-1", source: saved });
    act(() => first.result.current.edit("typed"));
    first.unmount();

    const { result } = renderDraft({
      userId: "user-1",
      source: { value: "someone else's", version: 2 },
    });

    expect(result.current.value).toBe("typed");
    expect(result.current.version).toBe(1);
  });

  it("never shows another user's draft", () => {
    writeDraft(draftKey("user-2", "race-1"), { value: "theirs", version: 1 });

    const { result, rerender } = renderDraft({ userId: "user-1", source: saved });
    expect(result.current.value).toBe("saved");

    rerender({ userId: "user-2", source: saved });
    expect(result.current.value).toBe("theirs");
  });

  it("is deleted by clear, and copied again from the next version", () => {
    const { result, rerender } = renderDraft({ userId: "user-1", source: saved });
    act(() => result.current.edit("typed"));

    act(() => result.current.clear());

    expect(readDraft(draftKey("user-1", "race-1"))).toBeUndefined();
    expect(result.current.value).toBeUndefined();

    rerender({ userId: "user-1", source: { value: "typed", version: 2 } });
    expect(result.current.value).toBe("typed");
    expect(result.current.version).toBe(2);
  });

  it("keeps editing when storage is blocked", () => {
    vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new DOMException("blocked", "SecurityError");
    });
    const { result } = renderDraft({ userId: "user-1", source: saved });

    act(() => result.current.edit("typed"));

    expect(result.current.value).toBe("typed");
    expect(readDraft(draftKey("user-1", "race-1"))).toBeUndefined();
  });
});
