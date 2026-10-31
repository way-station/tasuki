-- name: GetUserByID :one
SELECT id, line_user_id, display_name
FROM users
WHERE id = $1;

-- name: GetUserByLineUserID :one
SELECT id, line_user_id, display_name
FROM users
WHERE line_user_id = $1;

-- name: SaveUser :exec
INSERT INTO users (id, line_user_id, display_name)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE
SET display_name = EXCLUDED.display_name,
    updated_at   = now();
