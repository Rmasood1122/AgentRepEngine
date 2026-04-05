-- Migration 002: fp_candidates
-- Purpose: Log every RESTRICTED/BLOCKED decision during observe mode
--          for human review. Foundation of production FP measurement.
-- Source:  FP Legitimacy Panel output, April 5, 2026
-- Rule:    Additive only. Zero impact on existing tables.
--          enforcement_decisions is INSERT-only and untouched.

CREATE TABLE IF NOT EXISTS fp_candidates (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_did       TEXT        NOT NULL,
    org_id          TEXT        NOT NULL,
    decision_id     BIGINT,
    score           INTEGER     NOT NULL CHECK (score >= 0 AND score <= 1000),
    band            TEXT        NOT NULL CHECK (band IN ('BLOCKED','RESTRICTED','MONITORED','TRUSTED')),
    reason_object   JSONB       NOT NULL,
    flagged_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_by     TEXT,
    confirmed_fp    BOOLEAN,
    reviewed_at     TIMESTAMPTZ,
    notes           TEXT
);

-- Index for fast lookup during pilot review workflow
CREATE INDEX IF NOT EXISTS idx_fp_candidates_unreviewed
    ON fp_candidates (org_id, flagged_at DESC)
    WHERE confirmed_fp IS NULL;

CREATE INDEX IF NOT EXISTS idx_fp_candidates_agent
    ON fp_candidates (agent_did, flagged_at DESC);

-- Daily FP rate view — used by auto-rollback calibration and CISO reporting
CREATE OR REPLACE VIEW daily_fp_metrics AS
SELECT
    DATE(flagged_at)            AS day,
    org_id,
    COUNT(*)                    AS total_candidates,
    COUNT(*) FILTER (WHERE confirmed_fp = true)  AS confirmed_fps,
    COUNT(*) FILTER (WHERE confirmed_fp = false) AS confirmed_tps,
    COUNT(*) FILTER (WHERE confirmed_fp IS NULL) AS pending_review,
    ROUND(
        COUNT(*) FILTER (WHERE confirmed_fp = true)::DECIMAL /
        NULLIF(COUNT(*) FILTER (WHERE confirmed_fp IS NOT NULL), 0) * 100,
    4) AS confirmed_fp_rate_pct
FROM fp_candidates
GROUP BY DATE(flagged_at), org_id;

COMMENT ON TABLE fp_candidates IS
    'Candidate false positives logged during observe mode. '
    'Every RESTRICTED or BLOCKED decision is inserted here for human review. '
    'confirmed_fp=true means the block was wrong (FP). '
    'confirmed_fp=false means the block was correct (TP). '
    'confirmed_fp=NULL means pending human review. '
    'This table drives the production FP rate claim (Tier 3 and Tier 4).';

COMMENT ON VIEW daily_fp_metrics IS
    'Daily FP rate per org. Used by CISO weekly report and auto-rollback '
    'calibration. confirmed_fp_rate_pct is the operative production FP metric.';
