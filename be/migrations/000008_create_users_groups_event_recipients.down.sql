DROP TABLE IF EXISTS event_recipients;

ALTER TABLE events
DROP CONSTRAINT IF EXISTS events_audience_type_check;

ALTER TABLE events
DROP COLUMN IF EXISTS audience_type;

DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS users;

