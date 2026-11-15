-- name: InsertOutboxEntry :exec
INSERT INTO outbox_entries (id, steps, next_attempt_at)
VALUES ($1, $2, $3);

-- name: FailOutboxEntry :exec
-- Counts a failed attempt of the request and sets when the retry worker tries again.
UPDATE outbox_entries
SET attempts        = attempts + 1,
    next_attempt_at = @next_attempt_at,
    last_error      = @last_error
WHERE id = @id;

-- name: ClaimDueOutboxEntries :many
-- Takes the entries whose next attempt has come and that have attempts left, counts the
-- attempt, and moves the next attempt away by the first delay times 2 to the power of the
-- attempts so far, at most the max delay. SKIP LOCKED lets more than one worker run.
UPDATE outbox_entries
SET attempts        = attempts + 1,
    next_attempt_at = sqlc.arg(now)::timestamptz
      + LEAST(sqlc.arg(first_delay_seconds)::integer * power(2, attempts), sqlc.arg(max_delay_seconds)::integer) * interval '1 second'
WHERE id IN (
  SELECT due.id
  FROM outbox_entries AS due
  WHERE due.attempts < sqlc.arg(max_attempts)::integer AND due.next_attempt_at <= sqlc.arg(now)::timestamptz
  ORDER BY due.next_attempt_at
  LIMIT sqlc.arg(batch_size)::integer
  FOR UPDATE SKIP LOCKED
)
RETURNING id, steps, attempts;

-- name: SetOutboxEntryError :exec
UPDATE outbox_entries
SET last_error = $2
WHERE id = $1;

-- name: DeleteOutboxEntry :execrows
-- Deletes the entry together with its final step. No row is affected when another attempt
-- already finished it.
DELETE FROM outbox_entries
WHERE id = $1;
