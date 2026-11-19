-- Project descriptions (Russian and English) may be up to 1000 characters.
-- Widening a VARCHAR keeps the data and is safe to run again on every startup.
ALTER TABLE projects
    ALTER COLUMN description    TYPE VARCHAR(1000),
    ALTER COLUMN description_en TYPE VARCHAR(1000);
