CREATE TABLE races (
  id          uuid PRIMARY KEY,
  team_id     uuid NOT NULL REFERENCES teams (id),
  name        text NOT NULL CHECK (name <> ''),
  description jsonb NOT NULL,                          -- validated by the RaceDescription value object
  version     integer NOT NULL DEFAULT 0,
  created_at  timestamptz NOT NULL DEFAULT now(),      -- R1 sorts by this, newest first
  updated_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX races_team_id_created_at ON races (team_id, created_at DESC);
-- Marking a member as left looks for races dated today or later.
CREATE INDEX races_team_id_race_date ON races (team_id, ((description ->> 'race_date')));
