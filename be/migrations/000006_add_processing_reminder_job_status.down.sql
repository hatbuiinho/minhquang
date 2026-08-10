UPDATE reminder_jobs
SET status = 'pending'
WHERE status = 'processing';

ALTER TABLE reminder_jobs
DROP CONSTRAINT IF EXISTS reminder_jobs_status_check;

ALTER TABLE reminder_jobs
ADD CONSTRAINT reminder_jobs_status_check
CHECK (status IN ('pending', 'sent', 'cancelled', 'failed'));

