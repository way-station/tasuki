-- name: TeamHasRace :one
SELECT EXISTS (SELECT 1 FROM races WHERE id = @race_id AND team_id = @team_id);

-- name: ListInvitationViews :many
-- Every roster member of the team who has not left, with their invitation to the race
-- and their answer.
SELECT roster_members.id,
       roster_members.name,
       (race_invitations.race_id IS NOT NULL)::boolean AS invited,
       attendances.status AS attendance
FROM roster_members
LEFT JOIN race_invitations
  ON race_invitations.race_id = @race_id AND race_invitations.roster_member_id = roster_members.id
LEFT JOIN attendances
  ON attendances.race_id = race_invitations.race_id AND attendances.roster_member_id = race_invitations.roster_member_id
WHERE roster_members.team_id = @team_id AND roster_members.left_at IS NULL
ORDER BY roster_members.created_at, roster_members.id;
