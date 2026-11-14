-- name: ListEntryViews :many
SELECT id, name
FROM entries
WHERE race_id = $1
ORDER BY created_at, id;

-- name: ListEntryMemberViews :many
-- The members of the race's entries with their names, in the order they were added.
SELECT entry_members.entry_id, roster_members.id, roster_members.name
FROM entry_members
JOIN roster_members ON roster_members.id = entry_members.roster_member_id
WHERE entry_members.race_id = $1
ORDER BY entry_members.created_at, roster_members.id;
