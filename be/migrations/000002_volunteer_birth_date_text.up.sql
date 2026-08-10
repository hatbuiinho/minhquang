ALTER TABLE volunteers
ADD COLUMN IF NOT EXISTS birth_date text NOT NULL DEFAULT '';

UPDATE volunteers
SET birth_date = birth_year::text
WHERE birth_date = '' AND birth_year IS NOT NULL;

ALTER TABLE volunteers
DROP COLUMN IF EXISTS birth_year;
