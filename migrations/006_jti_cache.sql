-- A5 Hardening Sprint: jti_cache table for Redis restart recovery.
-- Redis = fast lookup. Postgres = durable recovery source.
-- On Redis restart, rebuild jti cache from this table.
CREATE TABLE IF NOT EXISTS jti_cache (
    jti         TEXT PRIMARY KEY,
    agent_did   TEXT NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for cleanup of expired entries
CREATE INDEX IF NOT EXISTS idx_jti_cache_expires ON jti_cache (expires_at);

-- Automatic cleanup of expired JTIs (run periodically or via pg_cron)
-- DELETE FROM jti_cache WHERE expires_at < NOW();
