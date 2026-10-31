-- name: GetAccountByID :one
SELECT id, line_user_id, display_name
FROM accounts
WHERE id = $1;

-- name: GetAccountByLineUserID :one
SELECT id, line_user_id, display_name
FROM accounts
WHERE line_user_id = $1;

-- name: SaveAccount :exec
INSERT INTO accounts (id, line_user_id, display_name)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE
SET display_name = EXCLUDED.display_name,
    updated_at   = now();
