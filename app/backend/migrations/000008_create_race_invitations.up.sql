-- Cancelling an invitation deletes its row; attendance and entry membership follow
-- by their foreign keys.
CREATE TABLE race_invitations (
  race_id          uuid NOT NULL REFERENCES races (id),
  roster_member_id uuid NOT NULL REFERENCES roster_members (id),
  created_at       timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (race_id, roster_member_id)
);
CREATE INDEX race_invitations_roster_member_id ON race_invitations (roster_member_id);
