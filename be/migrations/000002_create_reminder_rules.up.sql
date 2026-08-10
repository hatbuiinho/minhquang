ALTER TABLE events
ADD COLUMN IF NOT EXISTS reminder_generation integer NOT NULL DEFAULT 1;

CREATE TABLE IF NOT EXISTS reminder_rules (
  id text PRIMARY KEY,
  event_id text NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  offset_minutes integer NOT NULL,
  enabled boolean NOT NULL DEFAULT true,
  channel text NOT NULL DEFAULT 'push',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT reminder_rules_offset_check CHECK (offset_minutes >= 0),
  CONSTRAINT reminder_rules_channel_check CHECK (channel IN ('push'))
);

CREATE INDEX IF NOT EXISTS idx_reminder_rules_event_id
ON reminder_rules (event_id);
