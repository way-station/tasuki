-- name: GetTeam :one
SELECT id, name, created_at
FROM teams
WHERE id = $1;

-- name: CreateTeamIfMissing :exec
INSERT INTO teams (id, name)
VALUES ($1, $2)
ON CONFLICT (id) DO NOTHING;
