-- migrations/004_session_index.sql
-- PURPOSE: Support /enforcement/session-status endpoint (operator-only)
-- SAFETY:  Additive index only. No schema changes. No data modified.
--          session_id column may or may not exist yet — index is conditional.
-- DATE:    2026-04-16

-- Add session_id column if not present
-- Nullable: existing rows have no session context — that is correct
ALTER TABLE enforcement_decisions
  ADD COLUMN IF NOT EXISTS session_id TEXT;

-- Index for session-status query performance
-- Partial index: only rows where session_id IS NOT NULL
CREATE INDEX IF NOT EXISTS idx_enforcement_session_id
  ON enforcement_decisions(session_id)
  WHERE session_id IS NOT NULL;

-- Index for time-bounded session queries
CREATE INDEX IF NOT EXISTS idx_enforcement_created_at
  ON enforcement_decisions(created_at DESC)
  WHERE session_id IS NOT NULL;
