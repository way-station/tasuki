CREATE TABLE running_orders (
  id           uuid PRIMARY KEY,
  entry_id     uuid NOT NULL UNIQUE REFERENCES entries (id),   -- one running order per entry
  status       text NOT NULL CHECK (status IN ('draft', 'confirmed')),
  version      integer NOT NULL DEFAULT 0,
  confirmed_at timestamptz,
  created_at   timestamptz NOT NULL DEFAULT now(),
  updated_at   timestamptz NOT NULL DEFAULT now(),
  UNIQUE (id, entry_id)                                -- target of the composite key in running_order_legs
);
