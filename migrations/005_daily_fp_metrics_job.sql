-- Migration 005: Daily FP metrics aggregation job
-- Populates daily_fp_metrics from fp_candidates (confirmed_fp = true)
-- Run via pg_cron or external scheduler (cron job calling psql)
-- Safe to run multiple times — INSERT ... ON CONFLICT DO UPDATE

-- Step 1: Add missing columns to daily_fp_metrics for full audit trail
ALTER TABLE daily_fp_metrics
    ADD COLUMN IF NOT EXISTS tp_count        integer      NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reviewed_count  integer      NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS precision_pct   numeric(5,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS updated_at      timestamptz  NOT NULL DEFAULT now();

-- Step 2: Create the aggregation function
CREATE OR REPLACE FUNCTION compute_daily_fp_metrics(target_date date DEFAULT CURRENT_DATE - 1)
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
    v_fp_count      integer;
    v_tp_count      integer;
    v_reviewed      integer;
    v_total_blocked integer;
    v_fp_rate       numeric(5,2);
    v_precision     numeric(5,2);
BEGIN
    -- Count confirmed FPs for target date
    SELECT COUNT(*)
    INTO v_fp_count
    FROM fp_candidates
    WHERE confirmed_fp = true
      AND flagged_at::date = target_date;

    -- Count confirmed TPs (reviewed but NOT false positive)
    SELECT COUNT(*)
    INTO v_tp_count
    FROM fp_candidates
    WHERE confirmed_fp = false
      AND reviewed_at IS NOT NULL
      AND flagged_at::date = target_date;

    -- Total reviewed decisions
    SELECT COUNT(*)
    INTO v_reviewed
    FROM fp_candidates
    WHERE reviewed_at IS NOT NULL
      AND flagged_at::date = target_date;

    -- Total blocked decisions (all BLOCKED + RESTRICTED bands)
    SELECT COUNT(*)
    INTO v_total_blocked
    FROM fp_candidates
    WHERE band IN ('BLOCKED', 'RESTRICTED')
      AND flagged_at::date = target_date;

    -- FP rate: confirmed FPs / total blocked (avoid division by zero)
    IF v_total_blocked > 0 THEN
        v_fp_rate := ROUND((v_fp_count::numeric / v_total_blocked::numeric) * 100, 2);
    ELSE
        v_fp_rate := 0.00;
    END IF;

    -- Precision: TPs / (TPs + FPs) among reviewed
    IF (v_tp_count + v_fp_count) > 0 THEN
        v_precision := ROUND((v_tp_count::numeric / (v_tp_count + v_fp_count)::numeric) * 100, 2);
    ELSE
        v_precision := 0.00;
    END IF;

    -- Upsert — safe to run multiple times per day
    INSERT INTO daily_fp_metrics
        (date, fp_count, total_blocked, fp_rate, tp_count, reviewed_count, precision_pct, updated_at)
    VALUES
        (target_date, v_fp_count, v_total_blocked, v_fp_rate, v_tp_count, v_reviewed, v_precision, now())
    ON CONFLICT (date) DO UPDATE SET
        fp_count       = EXCLUDED.fp_count,
        total_blocked  = EXCLUDED.total_blocked,
        fp_rate        = EXCLUDED.fp_rate,
        tp_count       = EXCLUDED.tp_count,
        reviewed_count = EXCLUDED.reviewed_count,
        precision_pct  = EXCLUDED.precision_pct,
        updated_at     = now();

    RAISE NOTICE 'daily_fp_metrics updated for %: fp=% tp=% blocked=% fp_rate=% precision=%',
        target_date, v_fp_count, v_tp_count, v_total_blocked, v_fp_rate, v_precision;
END;
$$;

-- Step 3: Create view for last 30 days (CISO dashboard query)
CREATE OR REPLACE VIEW fp_metrics_30d AS
SELECT
    date,
    fp_count,
    tp_count,
    total_blocked,
    reviewed_count,
    fp_rate,
    precision_pct,
    updated_at
FROM daily_fp_metrics
WHERE date >= CURRENT_DATE - 30
ORDER BY date DESC;

-- Step 4: Self-verifying SQL proof query (C13 claim)
-- "Run this yourself. The result is your FP rate."
CREATE OR REPLACE VIEW fp_rate_current AS
SELECT
    CURRENT_DATE - 1                          AS for_date,
    COALESCE(fp_count, 0)                     AS confirmed_fps,
    COALESCE(total_blocked, 0)                AS total_enforcement_decisions,
    COALESCE(fp_rate, 0.00)                   AS fp_rate_pct,
    COALESCE(precision_pct, 0.00)             AS precision_pct,
    CASE
        WHEN fp_rate <= 1.00 THEN 'WITHIN_SLA'
        WHEN fp_rate <= 2.00 THEN 'APPROACHING_THRESHOLD'
        ELSE                      'EXCEEDS_SLA — auto-rollback should have fired'
    END                                       AS sla_status
FROM daily_fp_metrics
WHERE date = CURRENT_DATE - 1;

COMMENT ON VIEW fp_rate_current IS
'Self-verifying FP rate proof. Run this query yourself — no ARE involvement required.
Result is derived from your confirmed fp_candidates reviewed by your security team.
SLA: fp_rate_pct <= 2.00. Auto-rollback fires at 2.00. Target: <= 1.00.';