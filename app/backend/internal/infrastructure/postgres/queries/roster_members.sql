-- name: GetRosterMember :one
SELECT id, team_id, name, left_at
FROM roster_members
WHERE id = $1;

-- name: SaveRosterMember :exec
-- Inserts a new member, or updates whether they have left. Names cannot be changed, and
-- the average time and the lap count are left to the lap records.
INSERT INTO roster_members (id, team_id, name, left_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE
SET left_at    = EXCLUDED.left_at,
    updated_at = now();
