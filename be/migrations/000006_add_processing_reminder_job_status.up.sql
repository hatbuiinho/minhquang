ALTER TABLE reminder_jobs
DROP CONSTRAINT IF EXISTS reminder_jobs_status_check;

ALTER TABLE reminder_jobs
ADD CONSTRAINT reminder_jobs_status_check
CHECK (status IN ('pending', 'processing', 'sent', 'cancelled', 'failed'));

