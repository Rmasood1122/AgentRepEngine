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

-- Revoke UPDATE and DELETE on enforcement_decisions
REVOKE UPDATE, DELETE ON enforcement_decisions FROM PUBLIC;
