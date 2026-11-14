-- The race screens (R1, R2, R3). roster_member_id is whose answers to read; with NULL,
-- nobody is invited. Visibility is decided by the use case (ADR-0011).

-- name: ListRaceSummaries :many
SELECT races.id,
       races.name,
       races.description,
       (race_invitations.roster_member_id IS NOT NULL)::boolean AS invited,
       attendances.status AS attendance,
       EXISTS (
         SELECT 1
         FROM entries
         JOIN running_orders ON running_orders.entry_id = entries.id
         WHERE entries.race_id = races.id AND running_orders.status = 'confirmed'
       ) AS running_order_confirmed
FROM races
LEFT JOIN race_invitations
  ON race_invitations.race_id = races.id
 AND race_invitations.roster_member_id = sqlc.narg(roster_member_id)::uuid
LEFT JOIN attendances
  ON attendances.race_id = race_invitations.race_id
 AND attendances.roster_member_id = race_invitations.roster_member_id
WHERE races.team_id = sqlc.arg(team_id)
ORDER BY races.created_at DESC, races.id DESC;

-- name: GetRaceView :one
SELECT races.id,
       races.name,
       races.description,
       races.version,
       (race_invitations.roster_member_id IS NOT NULL)::boolean AS invited,
       attendances.status AS attendance,
       EXISTS (
         SELECT 1
         FROM entries
         JOIN running_orders ON running_orders.entry_id = entries.id
         WHERE entries.race_id = races.id AND running_orders.status = 'confirmed'
       ) AS running_order_confirmed,
       EXISTS (
         SELECT 1
         FROM entries
         JOIN running_orders ON running_orders.entry_id = entries.id
         JOIN running_order_legs ON running_order_legs.running_order_id = running_orders.id
         WHERE entries.race_id = races.id
           AND running_orders.status = 'confirmed'
           AND running_order_legs.roster_member_id = sqlc.narg(roster_member_id)::uuid
       ) AS runs_in_confirmed_order
FROM races
LEFT JOIN race_invitations
  ON race_invitations.race_id = races.id
 AND race_invitations.roster_member_id = sqlc.narg(roster_member_id)::uuid
LEFT JOIN attendances
  ON attendances.race_id = race_invitations.race_id
 AND attendances.roster_member_id = race_invitations.roster_member_id
WHERE races.team_id = sqlc.arg(team_id) AND races.id = sqlc.arg(race_id);
