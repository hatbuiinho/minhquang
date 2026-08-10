CREATE TABLE IF NOT EXISTS user_devices (
  id text PRIMARY KEY,
  user_id text NOT NULL,
  platform text NOT NULL,
  push_token text NOT NULL,
  enabled boolean NOT NULL DEFAULT true,
  last_seen_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT user_devices_platform_check CHECK (platform IN ('android', 'ios', 'web'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_devices_user_push_token
ON user_devices (user_id, push_token);

CREATE INDEX IF NOT EXISTS idx_user_devices_user_enabled
ON user_devices (user_id, enabled, updated_at DESC);

