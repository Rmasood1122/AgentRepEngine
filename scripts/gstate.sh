#!/bin/bash
PG="docker exec -it agentrepengine-postgres-1 psql -U are -d agentrepengine"
REDIS="docker exec -it agentrepengine-redis-1 redis-cli"

echo "=== SERVICES ===" && docker compose ps --format "table {{.Name}}\t{{.Status}}"
echo "=== FP RATE (7d) ===" && $PG -c "SELECT COUNT(*) FILTER (WHERE decision='BLOCKED' AND override=true) as fp, COUNT(*) FILTER (WHERE decision='BLOCKED') as blocks, ROUND(100.0 * COUNT(*) FILTER (WHERE decision='BLOCKED' AND override=true) / NULLIF(COUNT(*) FILTER (WHERE decision='BLOCKED'),0),2) as fp_rate FROM enforcement_decisions WHERE created_at > NOW() - INTERVAL '7 days';"
echo "=== EVENTS + VECTORS (24h) ===" && $PG -c "SELECT COUNT(*) as events, COUNT(DISTINCT agent_did) as agents, COUNT(*) FILTER (WHERE feature_vector IS NOT NULL) as vectors FROM agent_events WHERE created_at > NOW() - INTERVAL '24 hours';"
echo "=== HASH CHAIN ===" && $PG -c "SELECT verify_hash_chain('enforcement_decisions') AS valid;"
echo "=== IDENTITY ===" && $PG -c "SELECT identity_type, COUNT(*) FROM agent_identities GROUP BY identity_type;"
echo "=== REDIS ===" && $REDIS keys "score:*" 2>/dev/null && $REDIS info | grep -E "connected_clients|used_memory_human"
echo "=== EXPLAINABILITY ===" && curl -s http://localhost:8080/score/did:jwt:test:agent-1 | jq .
echo "=== KONG PLUGINS ===" && curl -s http://localhost:8001/plugins | jq '.data | length'
echo "=== SCORING HEALTH ===" && curl -s http://localhost:8080/health | jq .
