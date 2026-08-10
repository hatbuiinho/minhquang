ALTER TABLE reminder_jobs
ADD COLUMN IF NOT EXISTS read_at timestamptz,
ADD COLUMN IF NOT EXISTS dismissed_at timestamptz;

CREATE INDEX IF NOT EXISTS idx_reminder_jobs_user_dismissed_scheduled
ON reminder_jobs (user_id, dismissed_at, scheduled_at DESC, id DESC);
