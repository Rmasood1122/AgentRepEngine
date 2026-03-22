-- Agent identities
CREATE TABLE IF NOT EXISTS agent_identities (
    id                  BIGSERIAL PRIMARY KEY,
    did                 TEXT NOT NULL UNIQUE,
    org_id              UUID NOT NULL,
    instance_id         UUID NOT NULL,
    lineage_hash        TEXT,
    lineage_depth       INTEGER DEFAULT 0,
    current_score       INTEGER DEFAULT 700,
    status              TEXT DEFAULT 'probation',
    identity_type       TEXT DEFAULT 'jwt',
    first_seen          TIMESTAMPTZ DEFAULT NOW(),
    last_seen           TIMESTAMPTZ DEFAULT NOW(),
    probation_expires_at TIMESTAMPTZ
);

-- Behavioral events + feature vectors
CREATE TABLE IF NOT EXISTS agent_events (
    id              BIGSERIAL PRIMARY KEY,
    event_id        TEXT NOT NULL UNIQUE,
    agent_did       TEXT NOT NULL,
    event_type      TEXT NOT NULL,
    feature_vector  JSONB,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Async event queue
CREATE TABLE IF NOT EXISTS agent_event_queue (
    id           BIGSERIAL PRIMARY KEY,
    agent_did    TEXT NOT NULL,
    event_type   TEXT NOT NULL,
    payload      JSONB NOT NULL,
    privacy_tier INTEGER DEFAULT 1,
    processed    BOOLEAN DEFAULT FALSE,
    processed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

-- Dead letter queue
CREATE TABLE IF NOT EXISTS agent_event_dlq (
    id         BIGSERIAL PRIMARY KEY,
    payload    JSONB NOT NULL,
    error      TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Enforcement decisions (INSERT only — no UPDATE, no DELETE)
CREATE TABLE IF NOT EXISTS enforcement_decisions (
    id            BIGSERIAL PRIMARY KEY,
    agent_did     TEXT NOT NULL,
    decision      TEXT NOT NULL,
    score         INTEGER NOT NULL,
    score_delta   INTEGER,
    policy_fired  TEXT,
    reason_object JSONB,
    override      BOOLEAN DEFAULT FALSE,
    reason_code   TEXT,
    reviewer_id   TEXT,
    prev_hash     TEXT,
    this_hash     TEXT,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

-- Scoring explanations
CREATE TABLE IF NOT EXISTS scoring_explanations (
    id                      BIGSERIAL PRIMARY KEY,
    agent_did               TEXT NOT NULL,
    decision_id             BIGINT,
    score_before            INTEGER,
    score_after             INTEGER,
    trigger_events_json     JSONB,
    policy_fired            TEXT,
    reason_json             JSONB,
    peer_cluster            TEXT,
    peer_cluster_avg        INTEGER,
    deviation_from_cluster  INTEGER,
    created_at              TIMESTAMPTZ DEFAULT NOW()
);

-- FP metrics view
CREATE TABLE IF NOT EXISTS daily_fp_metrics (
    date      DATE PRIMARY KEY,
    fp_count  INTEGER DEFAULT 0,
    total_blocked INTEGER DEFAULT 0,
    fp_rate   NUMERIC(5,2) DEFAULT 0
);

-- E1-6 FIX: enforcement_decisions is INSERT-only at DB level
-- Tamper-evident audit log — no modifications permitted after write
-- Revoke from PUBLIC and from application user
REVOKE UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER
  ON enforcement_decisions FROM PUBLIC;
REVOKE UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER
  ON enforcement_decisions FROM are;
-- Verify: only INSERT and SELECT remain
-- SELECT privilege_type FROM information_schema.role_table_grants
-- WHERE table_name='enforcement_decisions' AND grantee='are';
-- Agent behavioral baselines (Task 4)
CREATE TABLE IF NOT EXISTS agent_baselines (
    id              BIGSERIAL PRIMARY KEY,
    agent_did       TEXT,
    org_id          UUID,
    cluster_id      TEXT DEFAULT 'default',
    feature_name    TEXT NOT NULL,
    mean            NUMERIC(12,4) DEFAULT 0,
    std_dev         NUMERIC(12,4) DEFAULT 1,
    sample_count    INTEGER DEFAULT 0,
    last_updated    TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(agent_did, feature_name)
);

CREATE TABLE IF NOT EXISTS cluster_baselines (
    id              BIGSERIAL PRIMARY KEY,
    cluster_id      TEXT NOT NULL DEFAULT 'default',
    feature_name    TEXT NOT NULL,
    mean            NUMERIC(12,4) DEFAULT 0,
    std_dev         NUMERIC(12,4) DEFAULT 1,
    sample_count    INTEGER DEFAULT 0,
    last_updated    TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(cluster_id, feature_name)
);

INSERT INTO cluster_baselines (cluster_id, feature_name, mean, std_dev, sample_count)
VALUES
    ('default', 'tool_call_rate_per_hour',       80,  60,  0),
    ('default', 'unique_endpoints_per_hour',      25,  30,  0),
    ('default', 'bulk_access_count_per_session',  400, 400, 0),
    ('default', 'pii_field_access_rate',          20,  25,  0),
    ('default', 'cross_tenant_probe_count',       0,   0.1, 0),
    ('default', 'permission_escalation_count',    0,   0.5, 0),
    ('default', 'sub_agent_spawn_depth',          0,   0.3, 0),
    ('default', 'token_refresh_rate',             1,   1,   0)
ON CONFLICT (cluster_id, feature_name) DO NOTHING;
-- Hash chain verification function (Task 7)
CREATE OR REPLACE FUNCTION verify_hash_chain(table_name TEXT)
RETURNS BOOLEAN AS $$
DECLARE
    rec RECORD;
    running_hash TEXT := '';
    valid BOOLEAN := TRUE;
BEGIN
    FOR rec IN
        SELECT id, created_at, decision, this_hash, prev_hash AS stored_prev
        FROM enforcement_decisions
        ORDER BY id ASC
    LOOP
        IF rec.stored_prev IS DISTINCT FROM running_hash THEN
            RAISE NOTICE 'Hash chain break at id=%', rec.id;
            valid := FALSE;
        END IF;
        running_hash := COALESCE(rec.this_hash, '');
    END LOOP;
    RETURN valid;
END;
$$ LANGUAGE plpgsql;