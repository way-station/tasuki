-- name: GetTeam :one
SELECT id, name, created_at
FROM teams
WHERE id = $1;
