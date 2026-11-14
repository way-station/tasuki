-- name: GetRosterMember :one
-- Locks the row in a unit of work, so that a member who is saved later in the same
-- transaction is not changed by another one in between (for example the average time
-- recalculated by a lap step while the member is marked as left).
SELECT id, team_id, name, left_at, average_seconds, lap_count
FROM roster_members
WHERE id = $1
FOR UPDATE;

-- name: SaveRosterMember :exec
-- Inserts a new member, or updates whether they have left and their average time and lap
-- count. Names cannot be changed.
INSERT INTO roster_members (id, team_id, name, left_at, average_seconds, lap_count)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE
SET left_at         = EXCLUDED.left_at,
    average_seconds = EXCLUDED.average_seconds,
    lap_count       = EXCLUDED.lap_count,
    updated_at      = now();
