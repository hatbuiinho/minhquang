CREATE INDEX IF NOT EXISTS idx_reminder_jobs_status_scheduled
ON reminder_jobs (status, scheduled_at, id);

