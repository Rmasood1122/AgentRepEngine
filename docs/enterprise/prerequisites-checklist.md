# AgentRepEngine — Regulatory Accountability Checklist
# Version: 1.0 | Updated: March 31, 2026
# For: Enterprise Pilot Evaluation | NWN / Lloyd Lemish
# Confidential

---

## PURPOSE

This document maps each regulatory accountability requirement for
AI agent deployments to ARE's existing production capability.
Every item in this checklist is implemented, tested, and committed.
Your auditors can verify each claim independently.

---

## REGULATORY ACCOUNTABILITY CHECKLIST

### SECTION A — AUDIT TRAIL

| Requirement | ARE Capability | How to Verify |
|-------------|---------------|---------------|
| Every enforcement decision must be logged | Hash-chained enforcement_decisions table — every BLOCK, ALLOW, RESTRICT recorded | `SELECT COUNT(*) FROM enforcement_decisions;` |
| Audit log must be tamper-evident | INSERT-only PostgreSQL permissions — UPDATE/DELETE revoked at DB level | `SELECT privilege_type FROM information_schema.role_table_grants WHERE table_name='enforcement_decisions' AND grantee='are';` Must show INSERT, SELECT only |
| Audit log must be verifiable | Hash chain — each record contains this_hash and prev_hash, chain integrity verified on demand | `SELECT verify_hash_chain();` Must return true |
| Audit log must include agent identity | agent_did (JWT-signed, RS256) on every record | `SELECT agent_did FROM enforcement_decisions LIMIT 5;` |
| Audit log must include decision rationale | reason_object field — structured JSON with decision, score, confidence_pct, policy_fired, trigger_events | `SELECT reason_object FROM enforcement_decisions WHERE decision='BLOCKED' LIMIT 1;` |
| Audit log must include timestamp | created_at on every record, UTC | `SELECT created_at FROM enforcement_decisions ORDER BY created_at DESC LIMIT 5;` |
| Audit log must be exportable for regulators | SOC2-format export endpoint | `GET /audit/export?from=2026-01-01&to=2026-03-31` |

**GDPR Article 22 / DORA Article 45 / SEC AI Governance:** All seven audit trail requirements met. ✅

---

### SECTION B — INCIDENT RESPONSE

| Requirement | ARE Capability | How to Verify |
|-------------|---------------|---------------|
| Automated detection of anomalous behavior | Velocity + z-score scoring on every agent request, sub-millisecond | `GET /health` → scoring_engine: healthy |
| Alerting when incident detected | SIEM webhook fires on every BLOCKED decision | `grep -rn "SendBlocked" internal/ \| grep -v "siem.go"` Must return non-empty |
| Auto-rollback if false positive rate spikes | ModeController monitors FP rate, auto-rolls back to observe mode if >2% | `go test ./internal/enforcement/... -run TestAutoRollback -v` Must PASS |
| Incident details available immediately | reason_object includes agent_did, score, confidence_pct, contributing features, timestamp | Query enforcement_decisions |
| Slow-walk attack detection | Variance growth rate trigger — 2x weekly variance triggers early warning | `go test ./tests/attack_corpus/... -run TestSlowWalk -v` 100% detection |

**DORA Article 17 (ICT incident classification) / NIST AI RMF:** All five incident response requirements met. ✅

---

### SECTION C — HUMAN OVERSIGHT

| Requirement | ARE Capability | How to Verify |
|-------------|---------------|---------------|
| Human review before enforcement goes live | Observe mode — 30-day period where all decisions are logged but not enforced | Mode set in Redis, confirmed via `GET /health → enforcement_mode` |
| Human can override any enforcement decision | Override workflow — blocked decisions can be reviewed and overridden by security team | `SELECT * FROM enforcement_decisions WHERE override=true LIMIT 5;` |
| Override must be logged | override field + override reason stored on enforcement_decisions record | Same query above |
| Human sign-off before mode transition | Observe → Enforce transition requires explicit manual switch | ModeController.SetMode() called only on explicit instruction |
| Security team notified of all blocks | SIEM webhook delivers structured alert on every BLOCK | Verify SIEM endpoint receives payload on block event |

**GDPR Article 22 (right not to be subject to automated decision) / DORA / NIST AI RMF RC.2:** All five human oversight requirements met. ✅

---

### SECTION D — CORRECTIVE ACTION

| Requirement | ARE Capability | How to Verify |
|-------------|---------------|---------------|
| Enforcement can be rolled back without data loss | Auto-rollback to observe mode — no decisions deleted, all audit records preserved | TestAutoRollback: seeds 30% FP rate, verifies rollback fires, audit intact |
| Thresholds adjustable without redeployment | YAML policy packs — threshold changes require config reload only | Edit `config/policy_packs/*.yaml`, reload service |
| False positive rate monitored continuously | StartFPMonitor() runs in background, samples FP rate every 5 minutes | `GET /health → fp_monitor: running` |
| Corrective action documented | Every override logged with reason code and timestamp | `SELECT override_reason, created_at FROM enforcement_decisions WHERE override=true;` |

**DORA Article 17 / SOC2 CC7.4 (incident recovery):** All four corrective action requirements met. ✅

---

### SECTION E — IDENTITY AND ACCESS

| Requirement | ARE Capability | How to Verify |
|-------------|---------------|---------------|
| Every agent must have verified identity | JWT RS256 signed tokens, verified at Kong gateway layer | `POST /verify {"token": "..."}` → {valid: true} |
| Identity must be org-scoped | org_id in every JWT claim, baseline keyed on (org_id, agent_did, feature_name) | `SELECT DISTINCT org_id FROM agent_baselines;` |
| Forged tokens must be rejected | JWKS endpoint — Kong verifies RS256 signature on every request | Send unsigned token → 401 returned |
| Replayed tokens must be rejected | jti claim + Redis used-token cache, TTL = token expiry | Send same token twice → second rejected |
| Identity persists across sessions | agent_did is stable DID-format identifier, score persists in PostgreSQL | `SELECT agent_did, score FROM agent_scores WHERE agent_did='did:jwt:...'` |

Each agent must have a unique JWT identity | One unique `sub` claim per agent required — multiple agents sharing one JWT identity corrupt behavioral baselines and will produce invalid scores | Verify with your JWT issuer before install: each agent process must have a distinct `sub` claim |

**DORA Article 9 (ICT security) / Zero Trust / NIST SP 800-207:** All five identity requirements met. ✅

---

### SECTION F — DATA SOVEREIGNTY

| Requirement | ARE Capability | How to Verify |
|-------------|---------------|---------------|
| Behavioral data must not leave client network | ARE runs in client's Kong gateway — no data egress to vendor | Deployment: Docker Compose on client infrastructure |
| Data residency compliance | All data stored in client's PostgreSQL and Redis instances | `docker compose ps` — all containers on client host |
| Vendor cannot access client data | ARE has no phone-home, no telemetry, no external API calls | `grep -rn "http.Get\|http.Post" internal/ \| grep -v "_test.go"` — no external calls |
| Client owns all enforcement data | All data in client-controlled PostgreSQL — ARE has no proprietary data store | Client controls backup, retention, deletion |

**GDPR Article 5 (data minimization) / DORA Article 45 / HIPAA:** All four data sovereignty requirements met. ✅

---

## SUMMARY SCORECARD

| Section | Requirements | Met | Status |
|---------|-------------|-----|--------|
| A — Audit Trail | 7 | 7 | ✅ |
| B — Incident Response | 5 | 5 | ✅ |
| C — Human Oversight | 5 | 5 | ✅ |
| D — Corrective Action | 4 | 4 | ✅ |
| E — Identity and Access | 5 | 5 | ✅ |
| F — Data Sovereignty | 4 | 4 | ✅ |
| **TOTAL** | **30** | **30** | **✅ 100%** |

---

## INDEPENDENT VERIFICATION

Every item in this checklist can be verified by your team
without trusting ARE's attestation:

- Database queries run directly against your PostgreSQL instance
- Hash chain verification runs as a SQL function
- Test suite runs as `go test ./...` against your deployment
- SIEM webhook verified by your SOC team on first block event
- Token rejection verified by sending forged tokens to /verify

ARE does not ask you to trust our compliance claims.
We give your auditors direct access to verify them.

---

## REGULATORY FRAMEWORK COVERAGE

| Framework | Relevant Articles | ARE Coverage |
|-----------|------------------|--------------|
| DORA (EU) | Art. 9, 17, 45 | ✅ Full |
| GDPR (EU) | Art. 5, 22 | ✅ Full |
| NIST AI RMF | GV, MP, MG, MS | ✅ Full |
| NIST SP 800-207 | Zero Trust Architecture | ✅ Full |
| OWASP LLM Top 10 | LLM06, LLM08 | ✅ Full |
| SOC2 | CC6, CC7 | ✅ Full |

## API Key Rotation Procedure

### SCORING_API_KEY Rotation (required before pilot go-live)

The default SCORING_API_KEY (`are-internal-key-change-in-production`) must be
rotated before any enterprise pilot. Failure to rotate creates an open scoring
endpoint accessible to anyone who reads the default config.

**Rotation steps:**

1. Generate new key:
```bash
   openssl rand -hex 32
```

2. Update docker-compose.yml:
```yaml
   SCORING_API_KEY=your-new-key-here
```

3. Update Kong plugin config with new key

4. Restart scoring service:
```bash
   docker compose restart scoring-service
```

5. Verify old key rejected:
```bash
   curl -H "X-API-Key: are-internal-key-change-in-production" \
     http://localhost:8080/score/test
   # Expected: 401 Unauthorized
```

6. Verify new key accepted:
```bash
   curl -H "X-API-Key: your-new-key-here" \
     http://localhost:8080/health
   # Expected: 200 OK
```

### RS256 Key Rotation

See `docs/ops/key-management.md` for RS256 private key rotation procedure.
Full JWT reissuance required for all agents after rotation.

---

*AgentRepEngine v1.0 | Naseem A2A Research Lab*
*IP anchored: Zenodo DOI 10.5281/zenodo.19169185*
*For pilot evaluation only — not for public distribution*