# AgentRepEngine Security Attestation
# Version: 1.0 — April 7, 2026
# Audience: Enterprise security teams, procurement reviewers, compliance officers
# Purpose: Five-section security attestation for enterprise procurement reviews

---

## SECTION 1 — KEY MANAGEMENT

### RS256 Key Pair (JWT Identity)

ARE uses RS256 asymmetric signing for all agent JWT tokens.

**Key storage:**
- Private key: `keys/private_key.pem` — gitignored, never committed
- Public key: `keys/public_key.pem` — mounted into scoring service container
- Keys are mounted as Docker volumes, never baked into container images

**Key generation:**
```bash
openssl genrsa -out keys/private_key.pem 2048
openssl rsa -in keys/private_key.pem -pubout -out keys/public_key.pem
```

**Key rotation procedure:** `docs/ops/key-management.md`
Full JWT reissuance required for all agents after rotation.

**Key incident disclosure:**
A prior private key was committed in error (commit 3048bbc) and removed
(commit 492d015). The old key is in git history and has been disclosed
to all enterprise contacts. It is not in production.

### SCORING_API_KEY

Symmetric key protecting the scoring service HTTP endpoints.
Default value must be rotated before any enterprise deployment.
Rotation procedure: `docs/enterprise/prerequisites-checklist.md`

### Redis Authentication

Redis requires ACL authentication. Default user is disabled.
Only `are_admin` ACL user has access. Key namespace restrictions enforced.
Config: `config/redis/users.acl`

---

## SECTION 2 — DATA RESIDENCY

### What ARE Stores

ARE stores behavioral metadata only. No payload content is stored.

| Data Type | Stored | Retention | Location |
|---|---|---|---|
| Agent behavioral feature vectors | YES | 90 days | PostgreSQL |
| Enforcement decisions | YES | Permanent | PostgreSQL (INSERT-only) |
| Agent baselines (mean/variance) | YES | Rolling | Redis + PostgreSQL |
| JWT token IDs (replay cache) | YES | TTL = token expiry | Redis |
| Request payload content | NO | Never | — |
| User PII | NO | Never | — |
| Agent prompts or responses | NO | Never | — |

### Data Never Leaves Customer Environment

ARE has no telemetry, no callbacks to ARE infrastructure, no analytics SDKs,
no external logging services. All data stays within the customer's deployment.

Verification: `grep -rn "http.Get\|http.Post" internal/ | grep -v "_test"` →
returns only internal service calls (Redis, PostgreSQL, /verify, /score).

### GDPR Compliance

ARE scores agent behavior, not human subjects.
Enforcement decisions are made on agent_did identifiers, not user identities.
Full GDPR position: `docs/enterprise/gdpr-position.md`

---

## SECTION 3 — ACCESS CONTROLS

### Authentication Layers

| Layer | Mechanism | Where |
|---|---|---|
| Agent identity | RS256 JWT per request | Kong plugin → /verify endpoint |
| Scoring API | SCORING_API_KEY header | All /score /event endpoints |
| Redis | ACL user + password | are_admin only |
| PostgreSQL | Username + password | are user, enforced in connection string |
| Audit trail | INSERT-only at DB level | REVOKE DELETE/UPDATE on enforcement_decisions |

### Audit Trail Tamper Protection

`enforcement_decisions` table has DELETE and UPDATE permissions revoked
at the PostgreSQL permission level — not just application level.

Verification:
```sql
SELECT privilege_type
FROM information_schema.role_table_grants
WHERE table_name = 'enforcement_decisions'
AND grantee = 'are';
```
Expected result: INSERT and SELECT only. No UPDATE. No DELETE.

### Port Exposure

| Port | Service | Exposure |
|---|---|---|
| 8080 | Scoring service | Internal only — never expose to internet |
| 5433 | PostgreSQL | Internal only |
| 6379 | Redis | Internal only |
| 8001 | Kong Admin API | Internal only — never expose to internet |

---

## SECTION 4 — INCIDENT RESPONSE

### FP Spike (False Positive Surge)

Trigger: FP rate exceeds 2% on any rolling 24-hour window.

Automatic response:
1. ModeController detects FP spike (checks every 5 minutes)
2. Auto-rollback to observe mode fires immediately
3. SIEM webhook fires: `{"event": "auto_rollback", "reason": "fp_spike"}`
4. All enforcement suspended until CISO reactivates
5. fp_candidates table logs all candidate FPs for human review

Manual escalation: `docs/ops/enforce-mode-gate.md`

### Redis Failure

Trigger: Redis becomes unreachable.

Response: Fail-open (agents keep working). See `docs/ops/redis-failover.md`
for full 4-scenario behavior matrix.

### Hash Chain Breach

Trigger: `verify_hash_chain()` returns false in /health endpoint.

Response:
1. /health returns degraded status immediately
2. SIEM alert fires
3. Enforcement suspended automatically
4. Human investigation required before enforcement reactivation

This is a non-recoverable automated event — human review mandatory.

### Private Key Compromise

Trigger: RS256 private key exposed or suspected compromised.

Response:
1. Rotate key immediately: `docs/ops/key-management.md`
2. Reissue all agent JWTs
3. Invalidate Redis used-token cache
4. Audit enforcement_decisions for decisions made with compromised key

---

## SECTION 5 — COMPLIANCE MAPPINGS

### DORA (Digital Operational Resilience Act)

| Article | Requirement | ARE Implementation |
|---|---|---|
| Art. 9 | ICT security policies | OWASP LLM Top 10 enforcement policies |
| Art. 10 | ICT change management | Raft quorum for ceiling changes (Phase 2) |
| Art. 11 | ICT business continuity | Auto-rollback + fail-open infrastructure |
| Art. 17 | ICT incident classification | Hash-chained enforcement_decisions table |
| Art. 28 | Third-party ICT risk | AI agent behavioral scoring + enforcement |
| Art. 30 | ICT contractual arrangements | Pilot LoU with staged rollout gates |

### SOC 2 Type II

ARE produces SOC 2 evidence but is not SOC 2 certified.
Claim: "SOC 2-ready architecture."

| Control | Evidence | Command |
|---|---|---|
| CC6.1 — Logical access | JWT RS256 per agent | /verify endpoint |
| CC6.6 — Data transmission | TLS in transit | Network config |
| CC7.2 — System monitoring | Hash-chained audit log | cmd/verify-chain |
| CC7.3 — Anomaly detection | Z-score + variance growth | internal/scoring/policy.go |
| A1.3 — Business continuity | Auto-rollback ModeController | internal/enforcement/ |

OSCAL evidence bundle: `cmd/oscal-generate --framework soc2`

### HIPAA

| Requirement | ARE Implementation |
|---|---|
| §164.312 — Access control | Agent identity via RS256 JWT |
| §164.312 — Audit controls | Hash-chained enforcement log |
| §164.528 — Access reporting | Merkle selective proof per incident |

Selective disclosure: `cmd/verify-decision [decision_id]`
Proves a specific enforcement decision without exposing other agents' data.

### NIST AI RMF

Full mapping: `docs/compliance/nist-ai-rmf-mapping.md`

Key controls:
- GOVERN 1.1: AI risk governance policies → OWASP policy packs
- MAP 5.1: AI risk likelihood → TP rate 88%, FP rate <0.1% target
- MEASURE 2.5: AI system monitoring → behavioral scoring + audit trail
- MANAGE 2.2: AI incident response → auto-rollback + SIEM integration

---

*Security Attestation v1.0 | AgentRepEngine | April 7, 2026*
*Next review: before each enterprise pilot go-live*