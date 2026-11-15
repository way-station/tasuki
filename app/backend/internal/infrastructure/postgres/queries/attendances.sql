-- name: GetAttendance :one
SELECT race_id, roster_member_id, status
FROM attendances
WHERE race_id = $1 AND roster_member_id = $2;

-- name: SaveAttendance :exec
INSERT INTO attendances (race_id, roster_member_id, status)
VALUES ($1, $2, $3)
ON CONFLICT (race_id, roster_member_id) DO UPDATE
SET status     = EXCLUDED.status,
    updated_at = now();
