CREATE TABLE users (
  id           uuid PRIMARY KEY,
  line_user_id text NOT NULL UNIQUE,
  display_name text NOT NULL,
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now()
);
