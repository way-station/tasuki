-- name: InsertLoginAttempt :exec
INSERT INTO login_attempts (id, state_hash, nonce, code_verifier, invite_link_id, expires_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: TakeLoginAttempt :one
-- Deleting and returning in one statement lets only one caller take a state.
DELETE FROM login_attempts
WHERE state_hash = $1
RETURNING id, state_hash, nonce, code_verifier, invite_link_id, expires_at;
