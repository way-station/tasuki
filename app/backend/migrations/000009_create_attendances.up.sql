CREATE TABLE attendances (
  race_id          uuid NOT NULL,
  roster_member_id uuid NOT NULL,
  status           text NOT NULL CHECK (status IN ('attending', 'not_attending', 'undecided')),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (race_id, roster_member_id),
  FOREIGN KEY (race_id, roster_member_id) REFERENCES race_invitations ON DELETE CASCADE
);
