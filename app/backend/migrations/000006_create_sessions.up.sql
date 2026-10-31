-- Server-side sessions (ADR-0014). The browser holds only the session token, in a cookie.
CREATE TABLE sessions (
  id         uuid PRIMARY KEY,
  account_id uuid NOT NULL REFERENCES accounts (id),
  token_hash bytea NOT NULL UNIQUE,                    -- SHA-256 of the token; the token itself is not stored
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz,                              -- set on logout, and when the roster member leaves
  created_at timestamptz NOT NULL DEFAULT now()
);
-- Revoking every session of an account.
CREATE INDEX sessions_account_id ON sessions (account_id);
