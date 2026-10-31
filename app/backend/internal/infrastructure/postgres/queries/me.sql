-- name: GetMeAccount :one
SELECT line_user_id, display_name
FROM accounts
WHERE id = $1;

-- name: ListMeLinks :many
SELECT member_links.id,
       member_links.roster_member_id,
       roster_members.name AS roster_member_name,
       member_links.invite_link_id,
       member_links.created_at,
       member_links.approved_at,
       member_links.rejected_at,
       member_links.unlinked_at
FROM member_links
JOIN roster_members ON roster_members.id = member_links.roster_member_id
WHERE member_links.account_id = $1
ORDER BY member_links.created_at;
