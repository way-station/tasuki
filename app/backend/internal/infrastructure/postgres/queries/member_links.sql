-- name: GetMemberLink :one
SELECT id, user_id, roster_member_id, invite_link_id, approved_at, rejected_at, unlinked_at, created_at
FROM member_links
WHERE id = $1;

-- name: ListMemberLinksByUser :many
SELECT id, user_id, roster_member_id, invite_link_id, approved_at, rejected_at, unlinked_at, created_at
FROM member_links
WHERE user_id = $1
ORDER BY created_at, id;

-- name: ListMemberLinksByMember :many
SELECT id, user_id, roster_member_id, invite_link_id, approved_at, rejected_at, unlinked_at, created_at
FROM member_links
WHERE roster_member_id = $1
ORDER BY created_at, id;

-- name: SaveMemberLink :exec
-- Inserts a new link, or records that it was approved, rejected, or unlinked.
INSERT INTO member_links (id, user_id, roster_member_id, invite_link_id, created_at, approved_at, rejected_at, unlinked_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (id) DO UPDATE
SET approved_at = EXCLUDED.approved_at,
    rejected_at = EXCLUDED.rejected_at,
    unlinked_at = EXCLUDED.unlinked_at;
