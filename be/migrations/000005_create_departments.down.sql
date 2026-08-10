ALTER TABLE volunteers ADD COLUMN IF NOT EXISTS department varchar(60) NOT NULL DEFAULT '';

UPDATE volunteers AS v
SET department = d.name
FROM departments AS d
WHERE v.department_id = d.id;

ALTER TABLE volunteers DROP CONSTRAINT IF EXISTS volunteers_department_id_fkey;
DROP INDEX IF EXISTS idx_volunteers_department_id;
ALTER TABLE volunteers DROP COLUMN IF EXISTS department_id;
DROP TABLE IF EXISTS departments;
