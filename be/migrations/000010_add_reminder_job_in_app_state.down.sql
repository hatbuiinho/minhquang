DROP INDEX IF EXISTS idx_reminder_jobs_user_dismissed_scheduled;

ALTER TABLE reminder_jobs
DROP COLUMN IF EXISTS dismissed_at,
DROP COLUMN IF EXISTS read_at;
