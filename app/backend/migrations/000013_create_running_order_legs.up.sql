-- Refers to entry_members, so removing a member from the entry empties their legs.
-- ON DELETE SET NULL (column) needs PostgreSQL 15 or later.
CREATE TABLE running_order_legs (
  running_order_id uuid NOT NULL,
  entry_id         uuid NOT NULL,
  leg_number       integer NOT NULL CHECK (leg_number > 0),
  roster_member_id uuid,                               -- NULL = empty leg
  PRIMARY KEY (running_order_id, leg_number),
  FOREIGN KEY (running_order_id, entry_id) REFERENCES running_orders (id, entry_id) ON DELETE CASCADE,
  FOREIGN KEY (entry_id, roster_member_id) REFERENCES entry_members (entry_id, roster_member_id)
    ON DELETE SET NULL (roster_member_id)
);
CREATE INDEX running_order_legs_runner ON running_order_legs (entry_id, roster_member_id);
