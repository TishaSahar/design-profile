-- Work direction of a portfolio project: 'residential' or 'commercial'.
-- Existing projects become residential. The service layer validates values.
-- Migrations run on every startup, so everything here must be idempotent.
ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS category VARCHAR(20) NOT NULL DEFAULT 'residential';
