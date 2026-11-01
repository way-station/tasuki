-- name: InsertLoginCode :exec
INSERT INTO login_codes (id, user_id, code_hash, expires_at)
VALUES ($1, $2, $3, $4);

-- name: TakeLoginCode :one
-- Deleting and returning in one statement lets only one caller take a code.
DELETE FROM login_codes
WHERE code_hash = $1
RETURNING id, user_id, code_hash, expires_at;
