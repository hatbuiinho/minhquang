CREATE TABLE IF NOT EXISTS departments (
    id text PRIMARY KEY,
    name varchar(60) NOT NULL,
    search_key varchar(60) NOT NULL UNIQUE,
    active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

INSERT INTO departments (id, name, search_key, active, created_at, updated_at)
SELECT
    'dep_' || substr(md5(search_key), 1, 24),
    MIN(department),
    search_key,
    true,
    NOW(),
    NOW()
FROM (
    SELECT trim(department) AS department, unaccent(lower(trim(department))) AS search_key
    FROM volunteers
    WHERE trim(department) <> ''
) existing_departments
GROUP BY search_key
ON CONFLICT (search_key) DO NOTHING;

ALTER TABLE volunteers ADD COLUMN IF NOT EXISTS department_id text;

UPDATE volunteers AS v
SET department_id = d.id
FROM departments AS d
WHERE d.search_key = unaccent(lower(trim(v.department)));

ALTER TABLE volunteers
    DROP CONSTRAINT IF EXISTS volunteers_department_id_fkey,
    ADD CONSTRAINT volunteers_department_id_fkey
        FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_volunteers_department_id ON volunteers(department_id);

ALTER TABLE volunteers DROP COLUMN department;
