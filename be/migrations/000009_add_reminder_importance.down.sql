ALTER TABLE reminder_jobs
DROP CONSTRAINT IF EXISTS reminder_jobs_importance_check;

ALTER TABLE reminder_jobs
DROP COLUMN IF EXISTS importance;

ALTER TABLE reminder_rules
DROP CONSTRAINT IF EXISTS reminder_rules_importance_check;

ALTER TABLE reminder_rules
DROP COLUMN IF EXISTS importance;
