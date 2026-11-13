-- name: GetInviteLinkByTokenHash :one
SELECT id, roster_member_id, token_hash, expires_at, used_at, revoked_at
FROM invite_links
WHERE token_hash = $1;

-- name: GetOpenInviteLinkOfMember :one
-- The member's link that is neither used nor replaced; there is at most one.
SELECT id, roster_member_id, token_hash, expires_at, used_at, revoked_at
FROM invite_links
WHERE roster_member_id = $1
  AND used_at IS NULL
  AND revoked_at IS NULL;

-- name: SaveInviteLink :exec
-- Inserts a new link, or records that it was used or replaced.
INSERT INTO invite_links (id, roster_member_id, token_hash, expires_at, used_at, revoked_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE
SET used_at    = EXCLUDED.used_at,
    revoked_at = EXCLUDED.revoked_at;
