-- English title and description of a portfolio project, shown on the English
-- version of the site. Empty means "not translated": the site shows the Russian text.
-- Migrations run on every startup, so everything here must be idempotent.
ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS title_en       VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS description_en VARCHAR(500) NOT NULL DEFAULT '';
