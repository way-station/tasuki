-- Server-side sessions (ADR-0014). The browser holds only the session token, in a cookie.
CREATE TABLE sessions (
  id         uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users (id),
  token_hash bytea NOT NULL UNIQUE,                    -- SHA-256 of the token; the token itself is not stored
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz,                              -- set on logout, and when the roster member leaves
  created_at timestamptz NOT NULL DEFAULT now()
);
-- Revoking every session of a user.
CREATE INDEX sessions_user_id ON sessions (user_id);
