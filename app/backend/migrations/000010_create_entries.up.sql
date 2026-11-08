CREATE TABLE entries (
  id         uuid PRIMARY KEY,
  race_id    uuid NOT NULL REFERENCES races (id),
  name       text NOT NULL CHECK (name <> ''),
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (id, race_id)                                 -- target of the composite key in entry_members
);
CREATE INDEX entries_race_id ON entries (race_id);
