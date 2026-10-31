CREATE TABLE member_links (
  id               uuid PRIMARY KEY,
  account_id       uuid NOT NULL REFERENCES accounts (id),
  roster_member_id uuid NOT NULL REFERENCES roster_members (id),
  invite_link_id   uuid NOT NULL UNIQUE REFERENCES invite_links (id),  -- an invite link is used once
  approved_at      timestamptz,                        -- not NULL = approved
  rejected_at      timestamptz,                        -- not NULL = rejected
  unlinked_at      timestamptz,                        -- not NULL = unlinked (approved_at stays)
  created_at       timestamptz NOT NULL DEFAULT now(), -- pending = all three NULL
  CHECK (approved_at IS NULL OR rejected_at IS NULL),  -- not both approved and rejected
  CHECK (unlinked_at IS NULL OR approved_at IS NOT NULL)  -- only approved links are unlinked
);
-- One approved account per roster member, and one roster member per account (ADR-0008).
CREATE UNIQUE INDEX member_links_one_approved_per_member  ON member_links (roster_member_id) WHERE approved_at IS NOT NULL AND unlinked_at IS NULL;
CREATE UNIQUE INDEX member_links_one_approved_per_account ON member_links (account_id)       WHERE approved_at IS NOT NULL AND unlinked_at IS NULL;
-- No new invite link while a link is pending; an account waits for one roster member at a time.
CREATE UNIQUE INDEX member_links_one_pending_per_member   ON member_links (roster_member_id) WHERE approved_at IS NULL AND rejected_at IS NULL;
CREATE UNIQUE INDEX member_links_one_pending_per_account  ON member_links (account_id)       WHERE approved_at IS NULL AND rejected_at IS NULL;
