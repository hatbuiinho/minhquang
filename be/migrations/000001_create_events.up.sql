CREATE TABLE IF NOT EXISTS events (
  id text PRIMARY KEY,
  user_id text NOT NULL,
  title text NOT NULL,
  description text NOT NULL DEFAULT '',
  starts_at timestamptz NOT NULL,
  timezone text NOT NULL DEFAULT 'UTC',
  status text NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT events_status_check CHECK (status IN ('active', 'archived', 'cancelled'))
);

CREATE INDEX IF NOT EXISTS idx_events_user_starts_at
ON events (user_id, starts_at, id);

CREATE INDEX IF NOT EXISTS idx_events_user_created_at
ON events (user_id, created_at DESC, id);
