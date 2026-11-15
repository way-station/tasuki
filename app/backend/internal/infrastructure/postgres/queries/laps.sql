-- name: IsConfirmedRunner :one
-- Whether the roster member runs a leg of a confirmed running order of the race of the team.
SELECT EXISTS (
  SELECT 1
  FROM running_order_legs
  JOIN running_orders ON running_orders.id = running_order_legs.running_order_id
  JOIN entries ON entries.id = running_order_legs.entry_id
  JOIN races ON races.id = entries.race_id
  WHERE races.id = $1
    AND races.team_id = $2
    AND running_orders.status = 'confirmed'
    AND running_order_legs.roster_member_id = $3
) AS runner;

-- name: RaceOfTeamExists :one
SELECT EXISTS (
  SELECT 1 FROM races WHERE id = $1 AND team_id = $2
) AS found;

-- name: ListConfirmedRunners :many
-- Every runner of the confirmed running orders of the race, once each, by name.
SELECT DISTINCT roster_members.id, roster_members.name
FROM running_order_legs
JOIN running_orders ON running_orders.id = running_order_legs.running_order_id
JOIN entries ON entries.id = running_order_legs.entry_id
JOIN roster_members ON roster_members.id = running_order_legs.roster_member_id
WHERE entries.race_id = $1
  AND running_orders.status = 'confirmed'
ORDER BY roster_members.name, roster_members.id;

-- name: GetProfile :one
SELECT name, average_seconds, lap_count
FROM roster_members
WHERE id = $1;
