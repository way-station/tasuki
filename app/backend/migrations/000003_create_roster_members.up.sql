CREATE TABLE roster_members (
  id              uuid PRIMARY KEY,
  team_id         uuid NOT NULL REFERENCES teams (id),
  name            text NOT NULL CHECK (name <> ''),
  left_at         timestamptz,                         -- NULL = not left
  average_seconds integer NOT NULL DEFAULT 220 CHECK (average_seconds > 0),  -- 3:40
  lap_count       integer NOT NULL DEFAULT 0 CHECK (lap_count >= 0),
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX roster_members_team_id ON roster_members (team_id);
