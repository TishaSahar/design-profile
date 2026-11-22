-- Outbox of background tasks, e.g. emailing the admin about a new project request.
-- A task is written in the same transaction as the change it reports, so it is
-- never lost and never sent for a change that was rolled back. The backend picks
-- up due tasks and deletes each one once done:
--   new   — waiting for its first run;
--   error — failed for a temporary reason (server down, 4xx SMTP reply); retried at run_at;
--   fatal — cannot ever succeed (bad address, broken body); kept for a human to look at.
-- Migrations run on every startup, so everything here must be idempotent.
CREATE TABLE IF NOT EXISTS tasks (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type       VARCHAR(64) NOT NULL,
    status     VARCHAR(16) NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'error', 'fatal')),
    body       JSONB NOT NULL,
    attempts   INTEGER NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    run_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Due tasks are looked up by run_at among the ones still to run.
CREATE INDEX IF NOT EXISTS idx_tasks_due ON tasks (run_at) WHERE status IN ('new', 'error');
