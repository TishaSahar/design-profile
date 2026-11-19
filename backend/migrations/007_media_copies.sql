-- Web copies of project photos (JPEG, 800 and 1600 px wide), served to visitors
-- instead of the uploaded originals. copies_done is set once copies were made —
-- or could not be (GIF, SVG): the backend makes the missing ones on startup.
-- Migrations run on every startup, so everything here must be idempotent.
ALTER TABLE project_media
    ADD COLUMN IF NOT EXISTS small_data  BYTEA,
    ADD COLUMN IF NOT EXISTS large_data  BYTEA,
    ADD COLUMN IF NOT EXISTS copies_done BOOLEAN NOT NULL DEFAULT FALSE;
