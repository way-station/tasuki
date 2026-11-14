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

-- name: GetInviteLinkWithTeamByTokenHash :one
-- The link whose token has the hash, with the name of the team of its roster member.
SELECT l.id, l.roster_member_id, l.token_hash, l.expires_at, l.used_at, l.revoked_at, t.name AS team_name
FROM invite_links l
JOIN roster_members m ON m.id = l.roster_member_id
JOIN teams t ON t.id = m.team_id
WHERE l.token_hash = $1;

-- name: GetInviteLinkWithTeamByID :one
-- The link, with the name of the team of its roster member. LINE Login's callback uses the
-- link up, so the row stays locked until the transaction ends: two callbacks cannot both
-- use it.
SELECT l.id, l.roster_member_id, l.token_hash, l.expires_at, l.used_at, l.revoked_at, t.name AS team_name
FROM invite_links l
JOIN roster_members m ON m.id = l.roster_member_id
JOIN teams t ON t.id = m.team_id
WHERE l.id = $1
FOR UPDATE OF l;
