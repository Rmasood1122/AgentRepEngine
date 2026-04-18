#!/usr/bin/env bash
set -euo pipefail

# ══════════════════════════════════════════════════════════════
# AgentRepEngine — Blast Radius Analysis
# Seeds synthetic agent interaction graph, then traverses it
# from an anomalous agent to map affected downstream agents.
# Self-contained: seeds its own data, no persistent state changes.
# ══════════════════════════════════════════════════════════════

POSTGRES_CONTAINER="agentrepengine-postgres-1"
DB_USER="are"
DB_NAME="agentrepengine"

psql_cmd() {
    docker exec "$POSTGRES_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -c "$1"
}

psql_file() {
    docker exec -i "$POSTGRES_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -f "$1"
}

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║  AgentRepEngine — Blast Radius Analysis                     ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Step 1: Run migration
echo "STEP 1 — Creating agent_interactions table..."
docker cp migrations/007_agent_interactions.sql "$POSTGRES_CONTAINER":/tmp/007.sql
MSYS_NO_PATHCONV=1 docker exec "$POSTGRES_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -f /tmp/007.sql
echo "  ✅ Table ready"
echo ""

# Step 2: Seed synthetic interaction graph
echo "STEP 2 — Seeding synthetic agent interaction graph..."
psql_cmd "
DELETE FROM agent_interactions WHERE source_agent LIKE 'blast-demo-%';

INSERT INTO agent_interactions (source_agent, target_agent, interaction_type, data_classification) VALUES
    ('blast-demo-agent-A', 'blast-demo-agent-B', 'api_call', 'PII'),
    ('blast-demo-agent-A', 'blast-demo-agent-C', 'data_share', 'financial'),
    ('blast-demo-agent-B', 'blast-demo-agent-D', 'api_call', 'internal'),
    ('blast-demo-agent-C', 'blast-demo-agent-D', 'delegation', 'PII'),
    ('blast-demo-agent-D', 'blast-demo-agent-E', 'api_call', 'financial');
"
echo "  5 interactions seeded (A→B, A→C, B→D, C→D, D→E)"
echo ""

# Step 3: Run blast radius analysis from Agent A (the anomalous agent)
echo "STEP 3 — Blast radius from anomalous agent 'blast-demo-agent-A'..."
echo "─────────────────────────────────────────────────────"
psql_cmd "
WITH RECURSIVE blast AS (
    -- Seed: the anomalous agent
    SELECT
        source_agent,
        target_agent,
        interaction_type,
        data_classification,
        1 AS depth,
        source_agent || ' → ' || target_agent AS path
    FROM agent_interactions
    WHERE source_agent = 'blast-demo-agent-A'

    UNION ALL

    -- Traverse: follow downstream interactions
    SELECT
        ai.source_agent,
        ai.target_agent,
        ai.interaction_type,
        ai.data_classification,
        b.depth + 1,
        b.path || ' → ' || ai.target_agent
    FROM agent_interactions ai
    JOIN blast b ON ai.source_agent = b.target_agent
    WHERE b.depth < 5
)
SELECT
    depth AS hop,
    target_agent AS affected_agent,
    interaction_type,
    data_classification,
    path
FROM blast
ORDER BY depth, target_agent;
"
echo "─────────────────────────────────────────────────────"
echo ""

# Step 4: Summary
echo "STEP 4 — Blast radius summary..."
psql_cmd "
WITH RECURSIVE blast AS (
    SELECT target_agent, 1 AS depth
    FROM agent_interactions
    WHERE source_agent = 'blast-demo-agent-A'
    UNION ALL
    SELECT ai.target_agent, b.depth + 1
    FROM agent_interactions ai
    JOIN blast b ON ai.source_agent = b.target_agent
    WHERE b.depth < 5
)
SELECT
    COUNT(DISTINCT target_agent) AS total_affected_agents,
    MAX(depth) AS max_propagation_depth
FROM blast;
"
echo ""

# Step 5: Cleanup demo data
echo "STEP 5 — Cleaning up demo data..."
psql_cmd "DELETE FROM agent_interactions WHERE source_agent LIKE 'blast-demo-%';"
echo "  ✅ Demo data cleaned"
echo ""

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║  BLAST RADIUS COMPLETE                                      ║"
echo "║                                                              ║"
echo "║  When agent A goes rogue, ARE maps every downstream agent    ║"
echo "║  that touched A's data — by hop depth, interaction type,     ║"
echo "║  and data classification. One query. Real-time.              ║"
echo "╚══════════════════════════════════════════════════════════════╝"