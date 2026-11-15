-- The remaining work of a use case that writes to both PostgreSQL and MongoDB (ADR-0016).
-- The steps run right after the commit, and the row is deleted together with the final
-- step. Until then the retry worker picks the row up when next_attempt_at has come, up to
-- 10 attempts in all; a row that keeps failing stays for someone to look at.
CREATE TABLE outbox_entries (
  id              uuid PRIMARY KEY,
  steps           jsonb NOT NULL,                      -- [ { kind, payload } ] in the order they run
  attempts        integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
  next_attempt_at timestamptz NOT NULL,
  last_error      text,
  created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX outbox_entries_next_attempt_at ON outbox_entries (next_attempt_at);
