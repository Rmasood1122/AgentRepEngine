-- Migration 003: mode_change_log
-- Purpose: Audit trail for every enforcement mode change.
--          Required for DORA Article 11 + SOC2 CC7.2 compliance.
--          Every mode change — automatic or manual — recorded here.
-- Source:  FP Legitimacy Panel output, April 5, 2026
-- Rule:    Additive only. Zero impact on existing tables.

CREATE TABLE IF NOT EXISTS mode_change_log (
    id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    changed_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    org_id                TEXT        NOT NULL,
    from_mode             TEXT        NOT NULL CHECK (from_mode IN ('observe','enforce','disabled')),
    to_mode               TEXT        NOT NULL CHECK (to_mode IN ('observe','enforce','disabled')),
    trigger_reason        TEXT        NOT NULL CHECK (trigger_reason IN (
                              'FP_THRESHOLD_EXCEEDED',
                              'MANUAL_OPERATOR',
                              'MANUAL_CISO',
                              'STARTUP_DEFAULT',
                              'PILOT_ACTIVATION',
                              'EMERGENCY_DISABLE'
                          )),
    fp_rate_at_trigger    DECIMAL(7,4),
    triggered_by          TEXT        NOT NULL,
    notes                 TEXT
);

-- Index for audit queries (most recent changes per org)
CREATE INDEX IF NOT EXISTS idx_mode_change_log_org
    ON mode_change_log (org_id, changed_at DESC);

-- Auto-rollback event view — auditor-facing
CREATE OR REPLACE VIEW auto_rollback_events AS
SELECT
    changed_at,
    org_id,
    from_mode,
    to_mode,
    fp_rate_at_trigger,
    triggered_by,
    notes
FROM mode_change_log
WHERE trigger_reason = 'FP_THRESHOLD_EXCEEDED'
ORDER BY changed_at DESC;

COMMENT ON TABLE mode_change_log IS
    'Complete audit trail of enforcement mode changes. '
    'Every change — automatic (FP_THRESHOLD_EXCEEDED) or manual — is recorded. '
    'This table is the DORA Article 11 evidence for automated business continuity. '
    'It is the SOC2 CC7.2 evidence for continuous monitoring controls. '
    'Hand this table to any auditor who asks about the auto-rollback mechanism. '
    'The absence of rows with trigger_reason=FP_THRESHOLD_EXCEEDED is the '
    'strongest possible evidence that the FP rate has never exceeded 2%.';

COMMENT ON VIEW auto_rollback_events IS
    'All auto-rollback events. If this view returns zero rows after N days '
    'of enforce-mode operation, the CISO claim is: auto-rollback has never '
    'triggered in N days. This is the Tier 4 FP claim.';
