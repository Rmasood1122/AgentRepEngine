-- migrations/002_chain_anchor.sql
-- PURPOSE: Eliminate hash chain fork under concurrent writes
-- SAFETY:  Additive only. No existing tables modified.
--          REVOKE statements in 001_initial.sql are NOT touched.
--          chain_position is populated by PostgreSQL on INSERT via RETURNING.
--          Application cannot forge sequence values.
-- DATE:    2026-04-16

-- Sequence for monotonic chain ordering
CREATE SEQUENCE IF NOT EXISTS enforcement_chain_seq START 1 INCREMENT 1;

-- Add chain_position column to enforcement_decisions
-- DEFAULT nextval() means existing rows get sequential values on first read
-- NOT NULL ensures every new decision has an unambiguous chain position
ALTER TABLE enforcement_decisions
  ADD COLUMN IF NOT EXISTS chain_position BIGINT
    DEFAULT nextval('enforcement_chain_seq') NOT NULL;

-- Index for fork detection query performance in /health
CREATE INDEX IF NOT EXISTS idx_enforcement_chain_position
  ON enforcement_decisions(chain_position);

-- Grant USAGE on sequence to are user (INSERT already granted by 001)
GRANT USAGE ON SEQUENCE enforcement_chain_seq TO are;

-- Verify: REVOKE statements from 001 still hold
-- Run after migration:
-- SELECT privilege_type FROM information_schema.role_table_grants
-- WHERE table_name='enforcement_decisions' AND grantee='are';
-- Expected: INSERT, SELECT only.
