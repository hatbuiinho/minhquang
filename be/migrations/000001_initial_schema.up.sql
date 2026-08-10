CREATE TABLE IF NOT EXISTS users (
  id text PRIMARY KEY,
  username text NOT NULL UNIQUE,
  display_name text NOT NULL,
  password_hash text NOT NULL,
  role text NOT NULL DEFAULT 'admin' CHECK (role IN ('admin')),
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_active_display_name ON users (active, display_name, id);

CREATE TABLE IF NOT EXISTS user_sessions (
  token_hash text PRIMARY KEY,
  user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions (expires_at);

CREATE TABLE IF NOT EXISTS user_devices (
  id text PRIMARY KEY,
  user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  platform text NOT NULL CHECK (platform IN ('android', 'ios', 'web')),
  push_token text NOT NULL,
  enabled boolean NOT NULL DEFAULT true,
  last_seen_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  UNIQUE (user_id, push_token)
);

CREATE INDEX IF NOT EXISTS idx_user_devices_user_enabled ON user_devices (user_id, enabled, updated_at DESC);

CREATE TABLE IF NOT EXISTS volunteers (
  id text PRIMARY KEY,
  full_name text NOT NULL,
  dharma_name text NOT NULL DEFAULT '',
  birth_year integer CHECK (birth_year IS NULL OR birth_year BETWEEN 1900 AND 2100),
  cultivation_place text NOT NULL DEFAULT '',
  phone text NOT NULL DEFAULT '',
  avatar_url text NOT NULL DEFAULT '',
  arrival_date date NOT NULL,
  departure_date date,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT volunteers_departure_after_arrival CHECK (
    departure_date IS NULL OR departure_date >= arrival_date
  )
);

CREATE INDEX IF NOT EXISTS idx_volunteers_arrival_date ON volunteers (arrival_date DESC, id);
CREATE INDEX IF NOT EXISTS idx_volunteers_active_name ON volunteers (full_name, id) WHERE departure_date IS NULL;
