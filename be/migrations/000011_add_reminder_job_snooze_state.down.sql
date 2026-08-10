DROP INDEX IF EXISTS idx_reminder_jobs_snoozed_from_id;

ALTER TABLE reminder_jobs
DROP COLUMN IF EXISTS snoozed_at,
DROP COLUMN IF EXISTS snoozed_from_id;
