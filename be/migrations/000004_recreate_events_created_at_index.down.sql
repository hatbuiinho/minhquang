DROP INDEX IF EXISTS idx_events_user_created_at;

CREATE INDEX IF NOT EXISTS idx_events_user_created_at
ON events (user_id, created_at DESC, id);

