# Maintenance Window Procedure
**AgentRepEngine — Operations**
Version: 1.0 | Date: April 2026 | Status: Production

---

## Purpose

Defines the procedure for scheduling, executing, and closing maintenance windows
for AgentRepEngine in production. Ensures enforcement continuity, audit trail
integrity, and DORA Art.11 operational resilience compliance during maintenance.

---

## Before You Start — Prerequisites

- [ ] ModeController confirms current mode: `curl -s http://localhost:8080/health | jq '.enforcement_mode'`
- [ ] Hash chain valid: `curl -s http://localhost:8080/health | jq '.hash_chain_valid'`
- [ ] FP rate ≤2%: check `/health` response `fp_rate_7d`
- [ ] SIEM notified (if production): send maintenance start event
- [ ] CISO sign-off obtained if switching from ENFORCE → OBSERVE during window

---

## Step 1 — Register Maintenance Window (Freeze Agent Scores)

For each agent that will be affected, register a freeze to prevent false positives
during maintenance activity:
```bash
curl -s -X POST http://localhost:8080/agent/freeze \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $SCORING_API_KEY" \
  -d '{
    "agent_did": "did:jwt:your-org:agent-id:001",
    "org_id": "your-org",
    "reviewer_id": "ops-engineer:your-name",
    "window_start": "2026-04-07T02:00:00Z",
    "window_end": "2026-04-07T04:00:00Z",
    "reason": "Scheduled maintenance — Kong plugin upgrade"
  }'
```

**Effect:** Agent score is frozen for the window duration. Behavioral events
during the window are logged but do not trigger enforcement decisions.
Freeze is recorded in `enforcement_decisions` with `decision=FREEZE` for audit trail.

---

## Step 2 — Switch to Observe Mode (if required)

If maintenance requires disabling enforcement entirely:
```bash
# Verify current mode
curl -s http://localhost:8080/health | jq '.enforcement_mode'

# Switch to observe (requires CISO sign-off)
curl -s -X POST http://localhost:8080/mode \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $SCORING_API_KEY" \
  -d '{"mode": "observe", "reason": "maintenance window", "authorized_by": "ciso:name"}'
```

**Note:** Mode is stored in Redis by ModeController. Survives service restart.
Auto-rollback monitor continues running in observe mode — FP rate still tracked.

---

## Step 3 — Execute Maintenance

Common maintenance tasks and their ARE implications:

### Kong plugin upgrade
```bash
# Stop Kong, upgrade, restart
docker compose stop kong
# ... upgrade steps ...
docker compose start kong

# Verify Kong is passing requests through
curl -s http://localhost:8001/plugins | jq '.data[] | select(.name=="agent-reputation")'
```

### Redis restart
```bash
# Redis uses AOF + RDB persistence — data survives restart
docker compose restart redis

# Verify SIR state persisted
docker exec agentrepengine-redis-1 redis-cli \
  --no-auth-warning --user are_admin -a are_redis_dev \
  KEYS "sir:*" | head -5

# Verify enforcement mode persisted
curl -s http://localhost:8080/health | jq '.enforcement_mode'
```

### PostgreSQL migration
```bash
# Always verify hash chain after any DB operation
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine \
  -c "SELECT verify_hash_chain('enforcement_decisions') AS valid;"
# Must return: t

# Verify INSERT-only permissions survived migration
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c \
  "SELECT privilege_type FROM information_schema.role_table_grants
   WHERE table_name='enforcement_decisions' AND grantee='are';"
# Must show: INSERT, SELECT only
```

### Scoring service restart
```bash
docker compose restart scoring-service

# Verify health
curl -s http://localhost:8080/health | jq '.'
# Check: status=ok, hash_chain_valid=true, enforcement_mode=<expected>
```

---

## Step 4 — Post-Maintenance Verification

Run all checks before closing the window:
```bash
# 1. Stack health
docker compose ps | grep -E "healthy|unhealthy"

# 2. Service health
curl -s http://localhost:8080/health | jq '{
  status: .status,
  enforcement_mode: .enforcement_mode,
  hash_chain_valid: .hash_chain_valid,
  fp_rate_7d: .fp_rate_7d
}'

# 3. JWT verification still working
TOKEN=$(go run cmd/gentoken/main.go --did "did:jwt:maintenance:check:001" \
  --org "maintenance" 2>/dev/null | head -1)
curl -s http://localhost:8080/verify -X POST \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$TOKEN\"}" | jq '.valid'
# Must return: true

# 4. Full test suite
SCORING_API_KEY=are-internal-key-change-in-production go test ./...
# Must be all green

# 5. Hash chain integrity
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine \
  -c "SELECT verify_hash_chain('enforcement_decisions') AS valid;"
# Must return: t
```

---

## Step 5 — Close Window and Restore Enforcement
```bash
# 1. Remove agent freezes (they expire automatically, but can be cleared early)
# Freezes expire at window_end — no manual action required

# 2. Restore enforce mode if it was changed
curl -s -X POST http://localhost:8080/mode \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $SCORING_API_KEY" \
  -d '{"mode": "enforce", "reason": "maintenance complete", "authorized_by": "ciso:name"}'

# 3. Notify SIEM
# Maintenance end event — send to SIEM webhook if configured

# 4. Log closure in ops log
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) maintenance-window-closed by $(whoami)" \
  >> logs/maintenance.log
```

---

## Rollback Procedure

If post-maintenance verification fails:
```bash
# 1. Immediately switch to observe mode
curl -s -X POST http://localhost:8080/mode \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $SCORING_API_KEY" \
  -d '{"mode": "observe", "reason": "maintenance rollback", "authorized_by": "ops:oncall"}'

# 2. Restore previous Docker image
docker compose down
git stash  # if config changes were made
docker compose up -d

# 3. Verify hash chain integrity
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine \
  -c "SELECT verify_hash_chain('enforcement_decisions') AS valid;"

# 4. Page CISO if enforce mode cannot be restored within 30 minutes
```

---

## DORA Compliance Note

This procedure satisfies DORA Art.11 (ICT Business Continuity) requirements:
- Maintenance windows are pre-registered with documented justification
- Enforcement continuity is maintained via observe-mode fallback
- All mode changes are logged to `enforcement_decisions` audit trail
- Hash chain integrity is verified before and after every maintenance event
- Auto-rollback (ModeController) provides automated recovery from FP spikes