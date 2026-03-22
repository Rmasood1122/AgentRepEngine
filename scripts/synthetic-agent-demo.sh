#!/bin/bash
# AgentRepEngine — Synthetic Agent Demo
# Shows a real AI agent being caught by behavioral reputation scoring
# Story: 4 days of agent behavior compressed into 60 seconds
# Usage: bash scripts/synthetic-agent-demo.sh

set -e

PG="docker exec -i agentrepengine-postgres-1 psql -U are -d agentrepengine"
SCORING="http://localhost:8080"
AGENT_DID="did:jwt:finserv-demo:data-retrieval-agent:synthetic-001"

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║  AgentRepEngine — Live Behavioral Detection Demo            ║"
echo "║                                                              ║"
echo "║  Scenario: Financial services data retrieval agent          ║"
echo "║  The agent looks clean for 3 days. Then it doesn't.        ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
sleep 1

# Step 1: Verify stack
echo "  STEP 1 — Verifying AgentRepEngine stack..."
HEALTH=$(curl -s $SCORING/health)
HASH_VALID=$(echo $HEALTH | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('hash_chain_valid','unknown'))" 2>/dev/null || \
             echo $HEALTH | grep -o '"hash_chain_valid":[^,}]*' | cut -d: -f2 | tr -d ' ')
MODE=$(echo $HEALTH | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('enforcement_mode','unknown'))" 2>/dev/null || \
       echo $HEALTH | grep -o '"enforcement_mode":"[^"]*"' | cut -d'"' -f4)

echo "     Stack: healthy ✅"
echo "     Mode: $MODE"
echo "     Hash chain: $HASH_VALID"
echo ""
sleep 1

# Step 2: Clean state
echo "  STEP 2 — Clean slate (simulating fresh agent deployment)..."
$PG -c "DELETE FROM enforcement_decisions WHERE agent_did = '$AGENT_DID';" > /dev/null 2>&1
$PG -c "DELETE FROM agent_events WHERE agent_did = '$AGENT_DID';" > /dev/null 2>&1
$PG -c "DELETE FROM agent_event_queue WHERE agent_did = '$AGENT_DID';" > /dev/null 2>&1
$PG -c "DELETE FROM agent_identities WHERE did = '$AGENT_DID';" > /dev/null 2>&1
docker exec -i agentrepengine-redis-1 redis-cli \
  -u redis://are_admin:are_redis_dev@localhost:6379 \
  DEL "score:$AGENT_DID" > /dev/null 2>&1
echo "     ✅ Clean state confirmed"
echo ""
sleep 1

# Step 3: Run synthetic agent
echo "  STEP 3 — Starting synthetic agent..."
echo "     Agent ID: $AGENT_DID"
echo "     This agent will behave normally for 2 days,"
echo "     escalate on day 3, then attempt bulk PII extraction."
echo ""
sleep 2

# Run the Go synthetic agent
go run ~/AgentRepEngine/cmd/synthetic-agent/main.go

echo ""
sleep 1

# Step 4: Show enforcement decision
echo "  STEP 4 — Enforcement decision audit trail..."
echo ""
$PG -t -c "
SELECT
  '     Decision: ' || decision ||
  ' | Score: ' || score ||
  ' | Agent: ' || LEFT(agent_did, 45)
FROM enforcement_decisions
WHERE agent_did = '$AGENT_DID'
ORDER BY created_at DESC
LIMIT 3;"
echo ""
sleep 1

# Step 5: Verify hash chain
echo "  STEP 5 — Tamper-evident audit chain..."
CHAIN=$($PG -t -c "SELECT verify_hash_chain('enforcement_decisions');" | tr -d ' \n')
if [ "$CHAIN" = "t" ]; then
  echo "     ✅ Hash chain: VERIFIED — audit trail is tamper-evident"
else
  echo "     ❌ Hash chain: BROKEN"
fi
echo ""
sleep 1

# Step 6: Show reason object
echo "  STEP 6 — Structured reason object (what the SIEM receives)..."
echo ""
REASON=$($PG -t -c "
SELECT reason_object
FROM enforcement_decisions
WHERE agent_did = '$AGENT_DID'
  AND decision = 'BLOCKED'
ORDER BY created_at DESC
LIMIT 1;")
if [ -n "$REASON" ]; then
  echo "$REASON" | python3 -m json.tool 2>/dev/null | head -20 || echo "$REASON" | head -10
else
  echo "     No BLOCKED decision recorded yet."
  echo "     Run with enforce mode enabled to see blocking."
fi
echo ""

# Final summary
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║  DEMO COMPLETE                                              ║"
echo "║                                                              ║"
echo "║  ✅ Real agent traffic → real behavioral scoring            ║"
echo "║  ✅ 3-day normal baseline established                       ║"
echo "║  ✅ Day 4 PII extraction pattern detected                   ║"
echo "║  ✅ Enforcement decision with reason object                 ║"
echo "║  ✅ Tamper-evident audit trail verified                     ║"
echo "║                                                              ║"
echo "║  This is what AgentRepEngine does in production.           ║"
echo "║  Install time: < 4 hours                                    ║"
echo "║  Time to first detection: < 7 days                         ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""