-- name: GetRunningOrderByEntry :one
SELECT id, entry_id, status, version
FROM running_orders
WHERE entry_id = $1;

-- name: ListRunningOrderLegs :many
SELECT leg_number, roster_member_id
FROM running_order_legs
WHERE running_order_id = $1
ORDER BY leg_number;

-- name: SaveRunningOrder :execrows
-- Inserts a new running order, or updates it only if the stored version is loaded_version.
-- No row is affected on a version conflict. confirmed_at keeps the time it was first
-- confirmed.
INSERT INTO running_orders (id, entry_id, status, version, confirmed_at)
VALUES (@id, @entry_id, @status, @version, CASE WHEN @status = 'confirmed' THEN now() END)
ON CONFLICT (id) DO UPDATE
SET status       = EXCLUDED.status,
    version      = EXCLUDED.version,
    confirmed_at = CASE WHEN EXCLUDED.status = 'confirmed' THEN COALESCE(running_orders.confirmed_at, now()) END,
    updated_at   = now()
WHERE running_orders.version = @loaded_version;

-- name: DeleteRunningOrderLegs :exec
DELETE FROM running_order_legs
WHERE running_order_id = $1;

-- name: InsertRunningOrderLeg :exec
INSERT INTO running_order_legs (running_order_id, entry_id, leg_number, roster_member_id)
VALUES ($1, $2, $3, $4);
