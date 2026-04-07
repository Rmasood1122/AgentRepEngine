━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
QUICK MODE — 10-MINUTE TRIAGE
Run when you need a fast health check, not a full audit
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```bash
# Run all at once — paste full output
cd ~/AgentRepEngine

echo "=== STACK ===" && docker compose ps | grep -E "healthy|unhealthy|Exit"
echo "=== HEALTH ===" && curl -s http://localhost:8080/health | jq .
echo "=== TESTS ===" && go test ./... 2>&1 | grep -E "ok|FAIL|---"
echo "=== FP RATE ===" && docker exec agentrepengine-postgres-1 \
  psql -U are -d agentrepengine -c \
  "SELECT ROUND(100.0 * COUNT(*) FILTER (WHERE override=true) /
   NULLIF(COUNT(*) FILTER (WHERE decision='BLOCKED'),0),2) as fp_rate
   FROM enforcement_decisions WHERE created_at > NOW() - INTERVAL '7 days';"
echo "=== VECTORS ===" && docker exec agentrepengine-postgres-1 \
  psql -U are -d agentrepengine -c \
  "SELECT COUNT(*) - COUNT(feature_vector) as missing_vectors FROM agent_events;"
echo "=== CHAIN ===" && docker exec agentrepengine-postgres-1 \
  psql -U are -d agentrepengine -c \
  "SELECT verify_hash_chain('enforcement_decisions') as chain_valid;"
echo "=== DEMO ===" && time bash scripts/demo.sh 2>&1 | tail -5
```

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
CONFIDENCE LABELS — REQUIRED ON EVERY FINDING
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[F]   = Confirmed from command output this session
[OBS] = Observed in code/docs (not live verified)
[H]   = High confidence inference from evidence
[INF] = Inferred — not directly confirmed
[ASS] = Assumption — needs command to verify
[GAP] = Not implemented — confirmed from command output
[VULN]= Confirmed vulnerability from command output

RULE: Never use [F] without showing the command output that confirms it.
      A claim without evidence is [ASS] until proven.

╔══════════════════════════════════════════════════════════════════════════════╗
║  APEX TEST — ACTIVATION SUMMARY                                            ║
║                                                                             ║
║  5 expert evaluators. 50+ commands. Real output only.                     ║
║  Every claim verified by running system.                                   ║
║  No assumptions. No self-certification. Evidence only.                    ║
║                                                                             ║
║  Start with: APEX TEST — QUICK                                             ║
║  Full audit: APEX TEST — FULL                                              ║
║  One expert: APEX TEST — E1 through E5                                    ║
╚══════════════════════════════════════════════════════════════════════════════╝