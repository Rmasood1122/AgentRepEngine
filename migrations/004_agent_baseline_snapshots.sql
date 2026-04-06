-- Migration 004: agent_baseline_snapshots
-- Purpose: Weekly snapshots of agent_baselines std_dev values.
--          Required by CheckVarianceGrowthRate() in internal/scoring/policy.go.
--          Without this table, variance growth rate detection is silent —
--          the function runs but LEFT JOIN always returns NULL for prev_std_dev,
--          causing all growth rate checks to be skipped (prevVariance < 0.001).
-- Source:  TW-6 gap identified April 5, 2026
-- Rule:    Additive only. Zero impact on existing tables.

CREATE TABLE IF NOT EXISTS agent_baseline_snapshots (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        TEXT        NOT NULL,
    agent_did     TEXT        NOT NULL,
    feature_name  TEXT        NOT NULL,
    std_dev       FLOAT8      NOT NULL CHECK (std_dev >= 0),
    sample_count  INTEGER     NOT NULL CHECK (sample_count >= 0),
    snapshot_date TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Generated column: pre-computed UTC date for idempotency checks
ALTER TABLE agent_baseline_snapshots
    ADD COLUMN IF NOT EXISTS snapshot_day DATE
    GENERATED ALWAYS AS (DATE(timezone('UTC', snapshot_date))) STORED;

-- Index for the exact JOIN pattern used in CheckVarianceGrowthRate
CREATE INDEX IF NOT EXISTS idx_baseline_snapshots_lookup
    ON agent_baseline_snapshots (org_id, agent_did, feature_name, snapshot_date DESC);

-- Unique constraint: one snapshot per agent per feature per day
CREATE UNIQUE INDEX IF NOT EXISTS idx_baseline_snapshots_daily
    ON agent_baseline_snapshots (org_id, agent_did, feature_name, snapshot_day);

COMMENT ON TABLE agent_baseline_snapshots IS
    'Weekly snapshots of agent baseline std_dev values. '
    'Used by CheckVarianceGrowthRate() to detect slow-walk baseline poisoning. '
    'A snapshot is taken once per day by the retention job (internal/store/retention.go). '
    'The variance growth check compares current std_dev against the snapshot '
    'from VarianceWindowDays (7) days ago. '
    'If variance has grown 2x or more in 7 days, the agent is flagged HIGH_RISK '
    'before the attack score has time to recover to a normal range.';