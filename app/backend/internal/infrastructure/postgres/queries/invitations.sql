-- name: GetInvitation :one
SELECT race_id, roster_member_id
FROM race_invitations
WHERE race_id = $1 AND roster_member_id = $2;

-- name: ListInvitationsByMember :many
SELECT race_id, roster_member_id
FROM race_invitations
WHERE roster_member_id = $1
ORDER BY created_at, race_id;

-- name: SaveInvitation :exec
-- Inviting an invited member does nothing.
INSERT INTO race_invitations (race_id, roster_member_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteInvitation :exec
-- The attendance and the entry membership are deleted by their foreign keys.
DELETE FROM race_invitations
WHERE race_id = $1 AND roster_member_id = $2;
