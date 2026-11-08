-- Separate contact fields on client project requests.
-- `contact` keeps the free-text value of requests created before this migration;
-- new requests leave it empty and fill the dedicated columns instead.
-- Migrations run on every startup, so everything here must be idempotent.
ALTER TABLE project_requests
    ADD COLUMN IF NOT EXISTS email             VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS phone             VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS telegram          VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS preferred_contact VARCHAR(20)  NOT NULL DEFAULT '';

ALTER TABLE project_requests ALTER COLUMN contact SET DEFAULT '';
