CREATE TABLE IF NOT EXISTS reminder_jobs (
  id text PRIMARY KEY,
  user_id text NOT NULL,
  event_id text NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  reminder_rule_id text NOT NULL,
  event_title text NOT NULL,
  event_starts_at timestamptz NOT NULL,
  offset_minutes integer NOT NULL,
  channel text NOT NULL DEFAULT 'push',
  status text NOT NULL DEFAULT 'pending',
  scheduled_at timestamptz NOT NULL,
  sent_at timestamptz,
  cancelled_at timestamptz,
  reminder_generation integer NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT reminder_jobs_offset_check CHECK (offset_minutes >= 0),
  CONSTRAINT reminder_jobs_channel_check CHECK (channel IN ('push')),
  CONSTRAINT reminder_jobs_status_check CHECK (status IN ('pending', 'sent', 'cancelled', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_reminder_jobs_user_status_scheduled
ON reminder_jobs (user_id, status, scheduled_at, id);

CREATE INDEX IF NOT EXISTS idx_reminder_jobs_event_id
ON reminder_jobs (event_id);

