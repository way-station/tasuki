CREATE TABLE invite_links (
  id               uuid PRIMARY KEY,
  roster_member_id uuid NOT NULL REFERENCES roster_members (id),
  token_hash       bytea NOT NULL UNIQUE,              -- SHA-256 of the token; the token itself is not stored
  expires_at       timestamptz NOT NULL,
  used_at          timestamptz,
  revoked_at       timestamptz,                        -- set when a newer link replaces it
  created_at       timestamptz NOT NULL DEFAULT now()
);
-- At most one link per roster member that is neither used nor replaced.
CREATE UNIQUE INDEX invite_links_one_open ON invite_links (roster_member_id)
  WHERE used_at IS NULL AND revoked_at IS NULL;
