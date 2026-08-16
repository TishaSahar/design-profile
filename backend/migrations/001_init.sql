-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Migration tracking table
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- OTP tokens for admin authentication
CREATE TABLE IF NOT EXISTS otp_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       VARCHAR(255) NOT NULL,
    code        VARCHAR(6) NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    used        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_otp_tokens_email ON otp_tokens(email);

-- Designer's projects
CREATE TABLE IF NOT EXISTS projects (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title          VARCHAR(255) NOT NULL,
    description    VARCHAR(500) NOT NULL DEFAULT '',
    cover_media_id UUID,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Project media files (stored as binary)
CREATE TABLE IF NOT EXISTS project_media (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id   UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    data         BYTEA NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    filename     VARCHAR(255) NOT NULL,
    sort_order   INTEGER NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_project_media_project_id ON project_media(project_id);

-- Add FK for cover photo after both tables exist (idempotent via DO block)
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_projects_cover_media'
    ) THEN
        ALTER TABLE projects
            ADD CONSTRAINT fk_projects_cover_media
            FOREIGN KEY (cover_media_id) REFERENCES project_media(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Client project requests
CREATE TABLE IF NOT EXISTS project_requests (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name  VARCHAR(100) NOT NULL,
    last_name   VARCHAR(100) NOT NULL,
    contact     VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    consented   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Attachments for project requests (up to 10 photos or 1 PDF)
CREATE TABLE IF NOT EXISTS request_attachments (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id   UUID NOT NULL REFERENCES project_requests(id) ON DELETE CASCADE,
    data         BYTEA NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    filename     VARCHAR(255) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_request_attachments_request_id ON request_attachments(request_id);

-- Designer contact information (single row)
CREATE TABLE IF NOT EXISTS contacts (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram  VARCHAR(255) NOT NULL DEFAULT '',
    instagram VARCHAR(255) NOT NULL DEFAULT '',
    email     VARCHAR(255) NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed initial contacts record if not exists
INSERT INTO contacts (telegram, instagram, email)
SELECT '', '', ''
WHERE NOT EXISTS (SELECT 1 FROM contacts);
