# AgentRepEngine — Deployment Guide
**Version:** 1.4.0 | **Date:** April 2026 | **Audience:** Customer Infrastructure / DevOps Team

This guide covers everything needed to deploy ARE in observe mode within 30 minutes.
No agent code changes required. No SDK. No vendor lock-in.

---

## PREREQUISITES

| Requirement | Minimum Version | Notes |
|---|---|---|
| Kong API Gateway | 2.8 | Plugin uses `kong.request`, `kong.response`, `resty.http` |
| Redis | 6.0 | AOF persistence recommended. ARE uses ACL user `are_admin`. |
| PostgreSQL | 13 | ARE uses port 5433 by default. Configurable. |
| Docker | 20.10 | Optional — ARE ships with docker-compose.yml |
| Go | 1.24 | Only required if building from source |

---

## STEP 1 — DEPLOY INFRASTRUCTURE (Docker)

If using Docker Compose:
```bash
git clone https://github.com/Rehanrana11/AgentRepEngine.git
cd AgentRepEngine
docker compose up -d
```

Verify all containers healthy:
```bash
docker compose ps
# Expected: scoring-service, postgres, redis, kong — all "Up"
```

Health check:
```bash
curl http://localhost:8080/health
# Expected: {"status":"ok","redis":"ok","postgres":"ok","mode":"observe"}
```

---

## STEP 2 — INSTALL KONG PLUGIN

Copy the plugin to your Kong plugins directory:
```bash
cp -r kong/plugins/agent-reputation /path/to/kong/plugins/
```

Add to Kong configuration (`kong.conf`):
plugins = bundled,agent-reputation
Enable the plugin on your target service or route:
```bash
curl -X POST http://localhost:8001/services/{your-service}/plugins \
  --data "name=agent-reputation" \
  --data "config.scoring_service_url=http://scoring-service:8080" \
  --data "config.enforcement_mode=observe" \
  --data "config.redis_host=redis" \
  --data "config.redis_port=6379"
```

Verify plugin active:
```bash
curl http://localhost:8001/plugins
# Expected: agent-reputation plugin listed with enforcement_mode=observe
```

---

## STEP 3 — CONFIGURE AGENT IDENTITY

Each AI agent must present a signed JWT in the `X-Agent-DID` header.

**Required JWT claims:**
```json
{
  "agent_did": "did:jwt:org:service:unique-id",
  "org_id":    "your-org-id",
  "instance_id": "agent-instance-uuid",
  "lineage_hash": "sha256-of-parent-agent-did-or-genesis",
  "exp": 1234567890,
  "iat": 1234567890
}
```

**Generate a test token:**
```bash
go run cmd/gentoken/main.go \
  --agent-did "did:jwt:your-org:service-name:agent-001" \
  --org-id "your-org-id" \
  --instance-id "inst-001" \
  --lineage-hash "genesis"
```

**Microsoft Entra / AGT users:**
ARE includes a bridge for Microsoft AGT identity tokens.
POST your AGT token to `/bridge/entra` — ARE maps it to ARE JWT format automatically.
```bash
curl -X POST http://localhost:8080/bridge/entra \
  -H "Content-Type: application/json" \
  -d '{"agt_token": "<your-AGT-JWT>"}'
# Returns: {"agent_did":"did:jwt:entra:tenant:object","org_id":"...","fail_open":false}
```

---

## STEP 4 — VERIFY OBSERVE MODE

Send a test request through Kong with the agent JWT:
```bash
curl -H "X-Agent-DID: <your-jwt>" http://localhost:8000/your-endpoint
```

Check response headers:
X-Agent-Score: 700        ← probation score for new agent
X-Agent-Band: MONITORED   ← correct for new agent
X-Gateway-Verified: true  ← JWT verified successfully
Check scoring service logs:
```bash
docker logs agentrepengine-scoring-service-1 --tail 20
# Expected: OBSERVE: agent=did:jwt:... score=700 band=MONITORED
```

---

## STEP 5 — VERIFY AUDIT TRAIL

Confirm hash chain is intact:
```bash
go run cmd/verify-chain/main.go \
  --dsn "postgres://are:are_dev@localhost:5433/agentrepengine?sslmode=disable"
# Expected: Chain intact: N records verified ✓
```

Generate DORA Article 8(4) compliance report:
```bash
go run cmd/dora-verify/main.go --format text \
  --dsn "postgres://are:are_dev@localhost:5433/agentrepengine?sslmode=disable"
# Expected: Full compliance report with chain verification
```

---

## STEP 6 — OBSERVE MODE OPERATION (Days 1–14)

During observe mode:
- All agents pass through — nothing is blocked
- Behavioral baselines build automatically over 30 days
- Every request is scored and logged
- FP rate is tracked in `daily_fp_metrics`

Monitor FP rate daily:
```bash
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine \
  -c "SELECT * FROM fp_rate_current;"
```

Monitor score distribution:
```bash
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine \
  -c "SELECT band, COUNT(*) FROM scoring_explanations GROUP BY band ORDER BY band;"
```

---

## STEP 7 — ENFORCE MODE ACTIVATION (Day 14+ CISO sign-off required)

**Enforce mode requires explicit CISO sign-off. ARE never activates enforcement automatically.**

Prerequisites before activating enforce mode (from `docs/enterprise/observe-to-enforce-criteria.md`):
- [ ] 14+ days of observe mode data
- [ ] FP rate verified ≤ 2% on production traffic
- [ ] CISO has reviewed baseline report
- [ ] Rollback tested (verify auto-rollback fires at FP > 2%)

Activate enforce mode:
```bash
curl -X POST http://localhost:8080/enforcement/override \
  -H "X-API-Key: <your-api-key>" \
  -H "Content-Type: application/json" \
  -d '{"mode": "enforce", "approved_by": "ciso-name", "reason": "14-day observe complete, FP rate 0.8%"}'
```

Verify:
```bash
curl http://localhost:8080/health
# Expected: {"mode":"enforce",...}
```

---

## SCORE BANDS — REFERENCE

| Score | Band | Enforcement Action |
|---|---|---|
| 700–1000 | TRUSTED | ALLOW — pass through |
| 500–699 | MONITORED | ALLOW + active audit logging |
| 200–499 | RESTRICTED | THROTTLE + human review on HIGH_RISK |
| 0–199 | BLOCKED | Synthetic response (never 403) |

---

## REDIS ACL — PRODUCTION NOTE

Default user is disabled. Use `are_admin`:
```bash
docker exec agentrepengine-redis-1 redis-cli \
  --no-auth-warning --user are_admin -a are_redis_dev KEYS "*"
```

Change `are_redis_dev` password before production deployment.

---

## POSTGRESQL — PRODUCTION NOTE

Default port is 5433 (not 5432). Connection string:
postgres://are:are_dev@localhost:5433/agentrepengine?sslmode=disable
Change `are_dev` password before production deployment.

---

## TROUBLESHOOTING

| Symptom | Cause | Fix |
|---|---|---|
| `X-Agent-Orphan: true` in response | No X-Agent-DID header | Agent not presenting JWT |
| `X-Agent-Invalid-JWT: true` | JWT signature invalid | Verify JWKS endpoint accessible |
| `X-Agent-Fail-Open: true` | Redis unavailable | Check Redis container health |
| Score stays at 700 | No behavioral history yet | Normal for first 24 hrs |
| Chain verify fails | Record tampered | Investigate PostgreSQL access logs |

---

## SUPPORT

Direct founder support during pilot.
All issues: open GitHub issue or contact pilot coordinator directly.
Response SLA: 4 hours during business hours.