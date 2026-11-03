-- A one-time code the callback hands to the frontend, exchanged for a session by
-- POST /session. It lasts 1 minute. The exchange deletes the row, so a code works once.
CREATE TABLE login_codes (
  id         uuid PRIMARY KEY,
  user_id    uuid NOT NULL REFERENCES users (id),
  code_hash  bytea NOT NULL UNIQUE,                    -- SHA-256 of the code; the code itself is not stored
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
