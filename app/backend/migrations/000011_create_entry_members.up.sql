-- Refers to race_invitations, so cancelling an invitation removes the member from
-- the entry.
CREATE TABLE entry_members (
  entry_id         uuid NOT NULL,
  race_id          uuid NOT NULL,
  roster_member_id uuid NOT NULL,
  created_at       timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (entry_id, roster_member_id),
  FOREIGN KEY (entry_id, race_id) REFERENCES entries (id, race_id),
  FOREIGN KEY (race_id, roster_member_id) REFERENCES race_invitations ON DELETE CASCADE
);
