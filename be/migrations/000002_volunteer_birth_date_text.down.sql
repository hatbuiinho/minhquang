ALTER TABLE volunteers
ADD COLUMN IF NOT EXISTS birth_year integer;

UPDATE volunteers
SET birth_year = birth_date::integer
WHERE birth_date ~ '^[0-9]{4}$';

ALTER TABLE volunteers
DROP COLUMN IF EXISTS birth_date;
