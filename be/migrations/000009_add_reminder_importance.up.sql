ALTER TABLE reminder_rules
ADD COLUMN IF NOT EXISTS importance text NOT NULL DEFAULT 'normal';

ALTER TABLE reminder_rules
DROP CONSTRAINT IF EXISTS reminder_rules_importance_check;

ALTER TABLE reminder_rules
ADD CONSTRAINT reminder_rules_importance_check CHECK (importance IN ('normal', 'urgent'));

ALTER TABLE reminder_jobs
ADD COLUMN IF NOT EXISTS importance text NOT NULL DEFAULT 'normal';

ALTER TABLE reminder_jobs
DROP CONSTRAINT IF EXISTS reminder_jobs_importance_check;

ALTER TABLE reminder_jobs
ADD CONSTRAINT reminder_jobs_importance_check CHECK (importance IN ('normal', 'urgent'));
