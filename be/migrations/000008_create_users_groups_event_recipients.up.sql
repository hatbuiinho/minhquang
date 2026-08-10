CREATE TABLE IF NOT EXISTS users (
  id text PRIMARY KEY,
  name text NOT NULL,
  email text NOT NULL DEFAULT '',
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email
ON users (email)
WHERE email <> '';

CREATE INDEX IF NOT EXISTS idx_users_active_name
ON users (active, name, id);

CREATE TABLE IF NOT EXISTS groups (
  id text PRIMARY KEY,
  name text NOT NULL,
  description text NOT NULL DEFAULT '',
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_groups_active_name
ON groups (active, name, id);

CREATE TABLE IF NOT EXISTS group_members (
  group_id text NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL,
  PRIMARY KEY (group_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_group_members_user_id
ON group_members (user_id, group_id);

ALTER TABLE events
ADD COLUMN IF NOT EXISTS audience_type text NOT NULL DEFAULT 'self';

ALTER TABLE events
DROP CONSTRAINT IF EXISTS events_audience_type_check;

ALTER TABLE events
ADD CONSTRAINT events_audience_type_check
CHECK (audience_type IN ('self', 'selected_users', 'selected_groups', 'all_users'));

CREATE TABLE IF NOT EXISTS event_recipients (
  event_id text NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  user_id text NOT NULL,
  source_type text NOT NULL,
  source_id text,
  created_at timestamptz NOT NULL,
  PRIMARY KEY (event_id, user_id),
  CONSTRAINT event_recipients_source_type_check CHECK (source_type IN ('self', 'user', 'group', 'all_users'))
);

CREATE INDEX IF NOT EXISTS idx_event_recipients_user_event
ON event_recipients (user_id, event_id);

INSERT INTO users (id, name, email, active, created_at, updated_at)
SELECT DISTINCT user_id, user_id, '', true, NOW(), NOW()
FROM events
ON CONFLICT (id) DO NOTHING;

INSERT INTO users (id, name, email, active, created_at, updated_at)
SELECT DISTINCT user_id, user_id, '', true, NOW(), NOW()
FROM user_devices
ON CONFLICT (id) DO NOTHING;

INSERT INTO event_recipients (event_id, user_id, source_type, source_id, created_at)
SELECT id, user_id, 'self', NULL, created_at
FROM events
ON CONFLICT (event_id, user_id) DO NOTHING;

