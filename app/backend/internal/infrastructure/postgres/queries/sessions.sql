-- name: GetSessionByTokenHash :one
SELECT id, user_id, token_hash, expires_at, revoked_at
FROM sessions
WHERE token_hash = $1;

-- name: SaveSession :exec
-- Only the revocation time changes after a session is created.
INSERT INTO sessions (id, user_id, token_hash, expires_at, revoked_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE
SET revoked_at = EXCLUDED.revoked_at;

-- name: GetActiveSessionUserID :one
SELECT user_id
FROM sessions
WHERE token_hash = $1
  AND revoked_at IS NULL
  AND expires_at > sqlc.arg(now);
