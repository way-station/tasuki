-- A LINE Login in progress, from POST /auth/line/authorize to the callback (ADR-0007).
-- It lasts 10 minutes. The callback deletes the row, so a state works once.
CREATE TABLE login_attempts (
  id             uuid PRIMARY KEY,
  state_hash     bytea NOT NULL UNIQUE,                -- SHA-256 of the state; the state itself is not stored
  nonce          text NOT NULL,                        -- compared with the nonce in the ID token
  code_verifier  text NOT NULL,                        -- PKCE verifier; sent to LINE in the code exchange, so kept as is
  invite_link_id uuid REFERENCES invite_links (id),    -- when the login started from an invite link
  expires_at     timestamptz NOT NULL,
  created_at     timestamptz NOT NULL DEFAULT now()
);
