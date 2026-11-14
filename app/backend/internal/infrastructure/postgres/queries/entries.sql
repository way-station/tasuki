-- name: GetEntry :one
SELECT id, race_id, name
FROM entries
WHERE id = $1;

-- name: ListEntriesByRace :many
SELECT id, race_id, name
FROM entries
WHERE race_id = $1
ORDER BY created_at, id;

-- name: ListEntryMembers :many
-- The members of the entries, in the order they were added.
SELECT entry_id, roster_member_id
FROM entry_members
WHERE entry_id = ANY(@entry_ids::uuid[])
ORDER BY created_at, roster_member_id;

-- name: SaveEntry :exec
-- Entries cannot be renamed, so saving a stored entry leaves its row as it is.
INSERT INTO entries (id, race_id, name)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO NOTHING;

-- name: AddEntryMember :exec
-- Adding a member who is in the entry does nothing.
INSERT INTO entry_members (entry_id, race_id, roster_member_id)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: RemoveEntryMembersExcept :exec
-- Removes the members of the entry who are not in kept. Their legs of the running order
-- are emptied by the foreign key.
DELETE FROM entry_members
WHERE entry_id = @entry_id AND NOT (roster_member_id = ANY(@kept::uuid[]));
