ALTER TABLE reminder_jobs
ADD COLUMN IF NOT EXISTS snoozed_from_id text,
ADD COLUMN IF NOT EXISTS snoozed_at timestamptz;

CREATE INDEX IF NOT EXISTS idx_reminder_jobs_snoozed_from_id
ON reminder_jobs (snoozed_from_id)
WHERE snoozed_from_id IS NOT NULL;
