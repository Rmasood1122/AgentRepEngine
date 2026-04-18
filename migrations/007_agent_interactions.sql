-- 007_agent_interactions.sql
-- Agent interaction graph for blast radius analysis.
-- NEW table only — no changes to existing tables.
-- Self-contained: does not ALTER or GRANT on existing objects.

CREATE TABLE IF NOT EXISTS agent_interactions (
    id SERIAL PRIMARY KEY,
    source_agent TEXT NOT NULL,
    target_agent TEXT NOT NULL,
    interaction_type TEXT NOT NULL,
    data_classification TEXT,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    session_id TEXT,
    metadata JSONB
);

CREATE INDEX IF NOT EXISTS idx_agent_interactions_source ON agent_interactions(source_agent);
CREATE INDEX IF NOT EXISTS idx_agent_interactions_target ON agent_interactions(target_agent);