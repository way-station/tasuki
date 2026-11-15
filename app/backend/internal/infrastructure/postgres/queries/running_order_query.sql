-- name: GetRunningOrderRace :one
-- The race of the team, with the invitation and the answer of roster_member_id.
SELECT races.description,
       (race_invitations.race_id IS NOT NULL)::boolean AS invited,
       attendances.status AS attendance
FROM races
LEFT JOIN race_invitations
  ON race_invitations.race_id = races.id AND race_invitations.roster_member_id = sqlc.narg(roster_member_id)::uuid
LEFT JOIN attendances
  ON attendances.race_id = race_invitations.race_id AND attendances.roster_member_id = race_invitations.roster_member_id
WHERE races.id = @race_id AND races.team_id = @team_id;

-- name: ListRunningOrderViews :many
-- The saved running orders of the race's entries, oldest entry first.
SELECT running_orders.id, running_orders.entry_id, entries.name AS entry_name, running_orders.status, running_orders.version
FROM running_orders
JOIN entries ON entries.id = running_orders.entry_id
WHERE entries.race_id = $1
ORDER BY entries.created_at, entries.id;

-- name: ListRunningOrderLegViews :many
-- The legs of the race's running orders, with the runners' names and average times.
SELECT running_order_legs.running_order_id,
       running_order_legs.leg_number,
       roster_members.id AS runner_id,
       roster_members.name AS runner_name,
       roster_members.average_seconds AS runner_average_seconds
FROM running_order_legs
JOIN entries ON entries.id = running_order_legs.entry_id
LEFT JOIN roster_members ON roster_members.id = running_order_legs.roster_member_id
WHERE entries.race_id = $1
ORDER BY running_order_legs.running_order_id, running_order_legs.leg_number;
