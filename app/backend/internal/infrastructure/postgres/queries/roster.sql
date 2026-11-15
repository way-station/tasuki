-- name: ListRosterMembersOfTeam :many
SELECT id, name, left_at
FROM roster_members
WHERE team_id = $1
ORDER BY name, id;

-- name: ListMemberLinksOfTeam :many
SELECT member_links.id, member_links.user_id, member_links.roster_member_id, member_links.invite_link_id,
       member_links.created_at, member_links.approved_at, member_links.rejected_at, member_links.unlinked_at
FROM member_links
JOIN roster_members ON roster_members.id = member_links.roster_member_id
WHERE roster_members.team_id = $1
ORDER BY member_links.created_at, member_links.id;

-- name: ListOpenInviteLinksOfTeam :many
-- The links of the team's members that are neither used nor replaced.
SELECT invite_links.id, invite_links.roster_member_id, invite_links.expires_at
FROM invite_links
JOIN roster_members ON roster_members.id = invite_links.roster_member_id
WHERE roster_members.team_id = $1
  AND invite_links.used_at IS NULL
  AND invite_links.revoked_at IS NULL;

-- name: GetRosterMemberOfTeam :one
SELECT id, name, left_at, average_seconds, lap_count
FROM roster_members
WHERE id = $1
  AND team_id = $2;

-- name: ListPendingLinksOfTeam :many
-- The links that are neither approved nor rejected, oldest first.
SELECT member_links.id, users.display_name, roster_members.id AS roster_member_id,
       roster_members.name AS roster_member_name, member_links.created_at
FROM member_links
JOIN users ON users.id = member_links.user_id
JOIN roster_members ON roster_members.id = member_links.roster_member_id
WHERE roster_members.team_id = $1
  AND member_links.approved_at IS NULL
  AND member_links.rejected_at IS NULL
ORDER BY member_links.created_at, member_links.id;

-- name: ListInvitedRaces :many
SELECT races.id, (races.description ->> 'race_date')::text AS race_date
FROM race_invitations
JOIN races ON races.id = race_invitations.race_id
WHERE race_invitations.roster_member_id = $1;

-- name: ListRacesOfTeamByID :many
-- The names and dates of the team's races among the ids, newest race date first.
SELECT id, name, (description ->> 'race_date')::text AS race_date
FROM races
WHERE team_id = sqlc.arg(team_id)
  AND id = ANY(sqlc.arg(ids)::uuid[])
ORDER BY description ->> 'race_date' DESC, id;
