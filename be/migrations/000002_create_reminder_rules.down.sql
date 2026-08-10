DROP TABLE IF EXISTS reminder_rules;

ALTER TABLE events
DROP COLUMN IF EXISTS reminder_generation;
