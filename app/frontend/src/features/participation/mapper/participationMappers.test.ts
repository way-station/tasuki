import { describe, expect, it } from "vitest";
import type { InvitationListResponse } from "../types/participation.ts";
import { toAttendanceRows } from "./toAttendanceRows.ts";
import { toEntryEditor } from "./toEntryEditor.ts";
import { toInvitationRows } from "./toInvitationRows.ts";

const invitations: InvitationListResponse = {
  members: [
    { roster_member_id: "m1", name: "Taro", invited: true, attendance: "attending" },
    { roster_member_id: "m2", name: "Jiro", invited: true, attendance: null },
    { roster_member_id: "m3", name: "Hanako", invited: false, attendance: null },
    { roster_member_id: "m4", name: "Saburo", invited: true, attendance: "not_attending" },
    { roster_member_id: "m5", name: "Shiro", invited: true, attendance: "attending" },
  ],
};

describe("toInvitationRows", () => {
  it("lists every member with whether they are invited", () => {
    expect(toInvitationRows(invitations).map((row) => [row.name, row.invited])).toEqual([
      ["Taro", true],
      ["Jiro", true],
      ["Hanako", false],
      ["Saburo", true],
      ["Shiro", true],
    ]);
  });
});

describe("toAttendanceRows", () => {
  it("lists invited members with their answer, including not answered", () => {
    expect(toAttendanceRows(invitations)).toEqual([
      { rosterMemberId: "m1", name: "Taro", answer: "attending" },
      { rosterMemberId: "m2", name: "Jiro", answer: "not_answered" },
      { rosterMemberId: "m4", name: "Saburo", answer: "not_attending" },
      { rosterMemberId: "m5", name: "Shiro", answer: "attending" },
    ]);
  });
});

describe("toEntryEditor", () => {
  it("offers every attending member when there is no entry yet", () => {
    expect(toEntryEditor({ entries: [] }, invitations)).toEqual({
      entry: null,
      candidates: [
        { rosterMemberId: "m1", name: "Taro" },
        { rosterMemberId: "m5", name: "Shiro" },
      ],
    });
  });

  it("shows the entry and offers attending members not in it", () => {
    const editor = toEntryEditor(
      {
        entries: [
          { id: "e1", name: "Team A", members: [{ roster_member_id: "m1", name: "Taro" }] },
        ],
      },
      invitations,
    );

    expect(editor.entry).toEqual({
      id: "e1",
      name: "Team A",
      members: [{ rosterMemberId: "m1", name: "Taro" }],
    });
    expect(editor.candidates).toEqual([{ rosterMemberId: "m5", name: "Shiro" }]);
  });
});
