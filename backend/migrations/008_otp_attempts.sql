-- Number of times a login code was checked: a code is spent after a few wrong
-- guesses, and the sum over the last hour limits guessing across codes.
-- Migrations run on every startup, so everything here must be idempotent.
ALTER TABLE otp_tokens
    ADD COLUMN IF NOT EXISTS attempts INTEGER NOT NULL DEFAULT 0;
