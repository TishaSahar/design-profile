-- Position of a project within its category on the public site (0 = first).
-- Added once: existing projects keep the order they had (newest first).
-- Migrations run on every startup, so everything here must be idempotent.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'projects' AND column_name = 'sort_order'
    ) THEN
        ALTER TABLE projects ADD COLUMN sort_order INT NOT NULL DEFAULT 0;
        UPDATE projects p SET sort_order = ranked.position
        FROM (
            SELECT id, ROW_NUMBER() OVER (PARTITION BY category ORDER BY created_at DESC) - 1 AS position
            FROM projects
        ) ranked
        WHERE p.id = ranked.id;
    END IF;
END $$;
