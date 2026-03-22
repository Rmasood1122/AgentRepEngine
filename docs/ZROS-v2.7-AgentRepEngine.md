━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ZROS v2.7 — ZERO REWORK OPERATING SYSTEM
AgentRepEngine Edition
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Paired with  : APEX v5.2 + MASTER_LEARNINGS v2.1 DELTA
Version      : 2.7
Date         : March 22, 2026
Supersedes   : ZROS v2.6

WHAT CHANGED IN v2.7:
  NEW 1   4 new incident classes (T21–T24) from March 22 build session
  NEW 2   3 new laws (L8–L10) — enforcement infrastructure reality
  NEW 3   G-HARDEN gate — pre-enterprise hardening checklist
  NEW 4   G-VERIFY gate — JWT verification chain completeness
  NEW 5   /verify endpoint + /jwks endpoint added to diagnostic stack
  NEW 6   APEX TEST protocol — expert panel live test (5 experts, 50+ cmds)
  NEW 7   Slow-walk detection documented — two-layer architecture
  NEW 8   INSERT-only enforcement: DB-level not application-level
  NEW 9   SIEM caller verification added to G-HARDEN
  NEW 10  Key volume mount — shared key pair between host and container

Rework target   : ≤10% fix: commits per session
Hard stops      : G-EXPLAIN ★ · G-FP ★ · G-IDENTITY ★ · G-DEPLOY ★ · G-HARDEN ★

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
WHAT ZROS v2.7 IS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

APEX v5.2   →  What to build, what gates must pass, what the product does
ZROS v2.7   →  How to build it, how to prevent rework, how to diagnose fast

The only metric that matters:

  Zone      Fix Rate   Action
  ────────  ─────────  ──────────────────────────────────────────────────────
  🟢 Green  < 10%      Continue building
  🟡 Yellow 10–20%     Audit root cause before next session
  🔴 Red    > 20%      STOP. Fix the system before building anything

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
THE TEN LAWS — IMMUTABLE ACROSS ALL VERSIONS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

L1  Every commit is a feature, never a fix for something just built.

L2  Fail closed on enforcement. Fail open on infrastructure.
    Enforcement failure → AUDIT not BLOCK (no unexplained blocks).
    Infrastructure failure → pass through (circuit breaker).
    These are two separate failure classes with two correct behaviors.
    Document both explicitly before any enterprise conversation.

L3  Clarify before build, not after failure.
    Name 3 failure modes before writing a single line.
    If you cannot name 3, you do not understand the component yet.

L4  The system gets lighter over time, not heavier.
    Every incident absorbed becomes a gate.
    Every gate that fires zero times for 30 days gets reviewed for removal.

L5  Never debug blind.
    State → Evidence → Hypothesis → Test. Always in that order.
    Blind patching is the #1 time multiplier of rework.

L6  Security tools that produce false positives get ripped out.
    Security tools without explainability never get trusted.
    Both are existential. Fix both first. Measure both always.

L7  The competitive clock is 12 months.
    Every hour of avoidable rework is a direct transfer of time to
    Check Point's integration roadmap. Rework is a competitive event.

L8  Auto-rollback is not optional before enforce mode goes live.
    ModeController + StartFPMonitor() must be initialized before
    any enterprise pilot switches from observe to enforce.
    Without it: FP spike = 2-4 hours of legitimate agents blocked
    = pilot ends. With it: FP spike detected in 5 minutes, auto-
    rollback fires, SIEM alerts, you know before the client does.
    This is Gate 0 before enforce mode. No exceptions.
    Implementation: internal/enforcement/mode_controller.go
    Test: TestAutoRollback must pass before enforce mode authorized.

L9  INSERT-only at database level, not application level.
    The application layer can be compromised. The DB cannot lie.
    enforcement_decisions must have INSERT+SELECT only at PostgreSQL
    permission level for the application user.
    Verify: SELECT privilege_type FROM information_schema.role_table_grants
            WHERE table_name='enforcement_decisions' AND grantee='are';
    Must show: INSERT, SELECT only. UPDATE/DELETE = CRITICAL VULNERABILITY.
    Fix: REVOKE UPDATE, DELETE, TRUNCATE ON enforcement_decisions FROM are;
    This must survive docker compose down/up — fix in migration file.

L10  SIEM built but unwired is worse than no SIEM.
     A SIEMWebhook that exists in code but has zero callers gives
     false confidence in enterprise conversations.
     Always verify before any SIEM claim:
       grep -rn "SendBlocked" internal/ | grep -v "siem.go"
     Must return non-empty. If empty → wire it before claiming SIEM.
     The caller is in score_store.go WriteScore() BLOCKED branch.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
FAILURE TAXONOMY — AGENTREPENGINE EDITION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Inherited from v2.6 (T13–T20), now extended:

  T13  FP Spike — enforcement threshold miscalibrated, >2% FP in production
       Consequence: Enterprise churn within 30 days. Existential.
       Gate: G-FP ★ + G-HARDEN (ModeController auto-rollback required)

  T14  Explainability Null — reason object missing or malformed
       Consequence: Security team distrust. Product pulled.
       Gate: G-EXPLAIN ★

  T15  Identity Drift — JWT claims inconsistent across services
       Consequence: Agent misidentified. Score applied to wrong entity.
       Gate: G-IDENTITY ★

  T16  Score Decay Bug — historical score decays incorrectly
       Consequence: Legitimate agents permanently restricted. FP spiral.
       Gate: G-SCORE

  T17  Latency Regression — gateway overhead exceeds 10ms p99
       Consequence: Engineers route around enforcement. Moat destroyed.
       Gate: G-LATENCY-IMPL

  T18  Tamper Chain Break — hash chain gap in enforcement_decisions
       Consequence: SOC2 audit failure. Compliance moat destroyed.
       Gate: G-TAMPER-IMPL

  T19  Feature Vector Missing — score stored without behavioral vector
       Consequence: Data moat not accumulating. Cannot recover.
       Gate: G-MOAT-DATA

  T20  Anti-Scope Creep — Phase 2-4 work before Phase 1 proven
       Consequence: 12-month clock wasted. Wedge never shipped.
       Gate: G-KILL

  T21  ModeController Missing — auto-rollback not in code before enforce mode
       Consequence: FP spike → 2-4 hours production outage → pilot ends.
       Root cause: Enforcement mode stored in env var (immutable at runtime).
       Fix: ModeController stores mode in Redis. StartFPMonitor() runs loop.
       Gate: G-HARDEN ★ (new — checks before any enforce-mode go-live)
       Commit: V6 closed March 22, 2026

  T22  SIEM Webhook Unwired — SendBlocked() defined but never called
       Consequence: Enterprise conversation claims SIEM, pilot day 1 reveals
       nothing fires. Trust destroyed before first incident.
       Root cause: siem.go written in audit package, never called from store.
       Fix: WriteScore() in BLOCKED branch calls s.siem.SendBlocked().
       Verify: grep -rn "SendBlocked" internal/ | grep -v "siem.go"
       Gate: G-HARDEN ★
       Commit: H5 closed March 22, 2026

  T23  Audit Log Mutable — enforcement_decisions has DELETE/UPDATE at DB level
       Consequence: SOC2 audit failure. "Tamper-evident" claim is false.
       Hash chain proves sequence, but DB permissions allow destruction.
       Root cause: PostgreSQL grants full permissions to table owner by default.
       REVOKE FROM PUBLIC does not affect the owning user.
       Fix: Explicit REVOKE UPDATE, DELETE, TRUNCATE FROM are in migration.
       Verify: SELECT privilege_type FROM information_schema.role_table_grants
               WHERE table_name='enforcement_decisions' AND grantee='are';
       Must show: INSERT, SELECT only.
       Gate: G-HARDEN ★ + G-TAMPER-IMPL
       Commit: E1-6 closed March 22, 2026

  T24  JWKS Missing — Kong cannot verify JWT signatures
       Consequence: Forged JWTs accepted by gateway. Identity security gap.
       Root cause: JWT verification in Go service, not at gateway layer.
       Kong has no built-in RSA verifier without JWKS endpoint.
       Fix: /jwks endpoint returns RS256 public key in JWKS format.
            /verify endpoint accepts raw token, returns {valid, agent_did}.
            Kong calls /verify on each request (500ms timeout, fail-open).
       Key sharing: docker-compose.yml volumes ./keys:/app/keys:ro
       Gate: G-VERIFY (new)
       Commit: E1-1 closed March 22, 2026

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
GATE SEQUENCE — TEMPORAL ORDER v2.7
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

★ = HARD STOP. No exceptions.

SESSION START:
  G-ENV         Environment + stack detection
  G-STATE       Repo + services + identity + pipeline + FP audit

BEFORE ANY CODE:
  G-KILL        Phase check — should we build this at all?
  G-READ ★      Read target file completely before touching it
  G1            Pre-code clarification + Security Pre-Mortem (3 failure modes)

IDENTITY + SECURITY:
  G-IDENTITY ★  JWT model integrity — HARD STOP if broken
  G-SECURITY    Enforcement security correctness
  G-VERIFY      JWT verification chain — /jwks + /verify endpoints exist

SCORING + ENFORCEMENT:
  G-SCORE       Scoring formula validation + idempotency check
  G-FP ★        False positive rate — HARD STOP if >2%
  G-EXPLAIN ★   Explainability object — HARD STOP if null or malformed

INFRASTRUCTURE:
  G-LATENCY-IMPL  Gateway latency baseline
  G-PIPELINE    Event pipeline integrity + privacy tiers
  G-INFRA       Redis/Postgres/services health
  G-TAMPER-IMPL Hash chain integrity check
  G-MOAT-DATA   Feature vector instrumentation check

OBSERVABILITY:
  G-OBSERVE     Structured logging + correlation IDs

BEFORE ENTERPRISE PILOT:
  G-HARDEN ★    Pre-enterprise hardening — HARD STOP (new v2.7)

BEFORE PUSH:
  G5            Pre-push: build + vet + test + FP + explain
  G-DEPLOY ★    Production verification 100%

AFTER DEPLOY:
  G7            Production verification within 48h

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
G-HARDEN ★ — PRE-ENTERPRISE HARDENING GATE (HARD STOP)
NEW IN v2.7 — Required before any enforce-mode go-live
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Run before switching any enterprise environment from observe to enforce.
All items must pass. No exceptions. No partial credit.

☐ ModeController initialized and StartFPMonitor() running
  grep -n "NewModeController\|StartFPMonitor" cmd/scoring-service/main.go
  # Must return both — not just defined, actually called in main()

☐ TestAutoRollback passing
  go test ./internal/enforcement/... -run TestAutoRollback -v
  # Must show: PASS — auto-rollback fires at >2% FP rate

☐ SIEM webhook wired — SendBlocked has callers
  grep -rn "SendBlocked" internal/ | grep -v "siem.go"
  # Must return non-empty — caller in score_store.go WriteScore()

☐ enforcement_decisions INSERT-only at DB level
  docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c \
  "SELECT privilege_type FROM information_schema.role_table_grants
   WHERE table_name='enforcement_decisions' AND grantee='are';"
  # Must show: INSERT, SELECT only. Nothing else.

☐ hash_chain_valid in /health endpoint
  curl -s http://localhost:8080/health | jq '.hash_chain_valid'
  # Must return: true

☐ SCORING_API_KEY set in docker-compose.yml
  grep "SCORING_API_KEY" docker-compose.yml
  # Must not be empty string

☐ Keys volume mounted — host and container share same key pair
  grep "keys:/app/keys" docker-compose.yml
  # Must be present

☐ /jwks endpoint returns RS256 public key
  curl -s http://localhost:8080/jwks | jq '.keys[0].alg'
  # Must return: "RS256"

☐ /verify endpoint rejects invalid tokens
  curl -s http://localhost:8080/verify -X POST \
    -H "Content-Type: application/json" \
    -d '{"token":"invalid.token.here"}' | jq '.valid'
  # Must return: false

☐ Operational safety architecture documented
  ls docs/enterprise/operational-safety-architecture.md
  # Must exist — fail-open/fail-closed distinction documented

☐ Pilot letter of understanding exists
  ls docs/enterprise/pilot-letter-of-understanding.md
  # Must exist before any pilot conversation

☐ Honest maturity statement current
  ls docs/enterprise/honest-maturity-statement.md
  # Must reflect actual score, not claimed score

HARD STOP conditions:
  ✗ ModeController not initialized — enforce mode not authorized
  ✗ TestAutoRollback failing — auto-rollback mechanism broken
  ✗ SIEM unwired — cannot claim SIEM integration in pilot
  ✗ enforcement_decisions has DELETE — tamper-evident claim is false
  ✗ Keys not mounted — /verify will reject all valid tokens

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
G-VERIFY — JWT VERIFICATION CHAIN GATE (NEW v2.7)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

☐ /jwks endpoint exists and returns valid JWKS
  curl -s http://localhost:8080/jwks | jq '.keys | length'
  # Must return: 1 or more

☐ /verify endpoint exists and validates RS256 signatures
  TOKEN=$(go run cmd/gentoken/main.go --did "did:jwt:test:verify:001" \
    --org "test-org" 2>/dev/null | head -1)
  curl -s http://localhost:8080/verify -X POST \
    -H "Content-Type: application/json" \
    -d "{\"token\":\"$TOKEN\"}" | jq '.valid'
  # Must return: true

☐ Keys mounted in scoring-service container
  grep "keys:/app/keys" docker-compose.yml
  # Must be present

☐ Replay detection active — second use of same token rejected
  go test ./internal/identity/... -run TestJTIReplayDetection -v
  # Must show: PASS

Note: Kong plugin calls /verify on each request with 500ms timeout.
Fail-open on verify service unavailable — infrastructure never blocks agents.
Full Kong-level RS256 verification is a Phase 2 hardening task.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
G-ENV: Environment Detection (EVERY SESSION)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# Services running?
docker compose ps

# Stack health
curl -s http://localhost:8080/health | jq .
# Must include: enforcement_mode, hash_chain_valid, redis, postgres

# JWKS endpoint live?
curl -s http://localhost:8080/jwks | jq '.keys[0].alg'
# Must return: "RS256"

# Gateway plugin active?
curl -s http://localhost:8001/plugins | jq '.data[].name'

# Redis authenticated?
docker exec agentrepengine-redis-1 redis-cli ping
# Must return: NOAUTH Authentication required

# Postgres INSERT-only enforced?
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c \
"SELECT privilege_type FROM information_schema.role_table_grants
 WHERE table_name='enforcement_decisions' AND grantee='are';"
# Must show: INSERT, SELECT only

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
G-STATE: Session State Audit (EVERY SESSION)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# 1. Git state
git log --oneline -10
git status

# 2. All tests passing?
go test ./... 2>&1 | grep -E "ok|FAIL"
# Must show: all ok, zero FAIL

# 3. FP rate
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c \
"SELECT ROUND(100.0 * COUNT(*) FILTER (WHERE override=true)
  / NULLIF(COUNT(*) FILTER (WHERE decision='BLOCKED'),0),2) as fp_rate
  FROM enforcement_decisions
  WHERE created_at > NOW() - INTERVAL '7 days';"
# Must be ≤ 2.00%

# 4. Feature vectors complete
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c \
"SELECT COUNT(*) - COUNT(feature_vector) as missing_vectors FROM agent_events;"
# Must be 0

# 5. Hash chain verified
curl -s http://localhost:8080/health | jq '.hash_chain_valid'
# Must be true

# 6. Explainability check
curl -s http://localhost:8080/score/did:jwt:finserv-demo:trading-agent:001 \
  -H "X-API-Key: are-internal-key-change-in-production" | jq '.reason'
# Must not be null

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SLOW-WALK DETECTION — TWO-LAYER ARCHITECTURE (NEW v2.7)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Phase 1 detects slow-walk attacks through two complementary layers:

LAYER 1 — Score (H+V formula):
  Catches: rate spikes, zero-tolerance events (cross-tenant probe)
  Misses:  ultra-slow baseline drift attacks
  Why it misses: agent baseline shifts toward attack behavior over time
                 z-score stays low because baseline drifts with attacker

LAYER 2 — HIGH_RISK VERIFY (score-independent):
  Catches: ALL policy threshold breaches regardless of score
  A TRUSTED agent (score 950) cannot silently execute:
    - bulk_pii_export (PIIFieldAccessRate > 0.3)
    - permission escalation (PermissionEscalationCount > 2)
    - cross-tenant access (CrossTenantProbeCount > 0)
    - deep spawn (SubAgentSpawnDepth > 3)
    - high frequency (ToolCallRatePerHour > 100)
  These require human VERIFY regardless of reputation score.

LAYER 3 — Phase 2 (Isolation Forest, pending):
  Requires 90 days of behavioral data.
  Adds pattern-level detection across multi-day sequences.
  Phase 1 cannot safely implement this without data.

CORPUS: tests/attack_corpus/slowwalk_test.go
  10 scenarios across 3/5/7-day windows
  Detection rate: 100% via combined L1+L2
  Score-only detection: 20% (documented gap)

Run before any enforce-mode go-live:
  go test ./tests/attack_corpus/... -run TestSlowWalk -v

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
APEX TEST PROTOCOL (NEW v2.7)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

5-expert live test protocol for product validation.
Full spec: docs/enterprise/expert-panel-test-protocol-v1.md

Activation: APEX TEST — [E1|E2|E3|E4|E5|FULL|QUICK]

Quick triage (10 minutes):
  cd ~/AgentRepEngine
  echo "=== STACK ===" && docker compose ps | grep -E "healthy|unhealthy"
  echo "=== HEALTH ===" && curl -s http://localhost:8080/health | jq .
  echo "=== TESTS ===" && go test ./... 2>&1 | grep -E "ok|FAIL"
  echo "=== FP ===" && docker exec agentrepengine-postgres-1 psql \
    -U are -d agentrepengine -c \
    "SELECT ROUND(100.0 * COUNT(*) FILTER (WHERE override=true)
     / NULLIF(COUNT(*) FILTER (WHERE decision='BLOCKED'),0),2) as fp_rate
     FROM enforcement_decisions WHERE created_at > NOW() - INTERVAL '7 days';"
  echo "=== CHAIN ===" && curl -s http://localhost:8080/health | \
    jq '.hash_chain_valid'
  echo "=== DEMO ===" && time bash scripts/demo.sh 2>&1 | tail -5

Composite scoring (March 22 baseline):
  E1 Security:       82/100 (Kong JWT wiring = Phase 2 task)
  E2 Infrastructure: 71/100 (Windows Docker latency — Linux production is fine)
  E3 ML/Scoring:     91/100
  E4 Compliance:     89/100
  E5 Product:        87/100
  Composite:         84/100 ENTERPRISE-READY

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
DIAGNOSTIC STACK — UPDATED v2.7
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

New diagnostic commands added in v2.7:

# JWT verification working?
TOKEN=$(go run cmd/gentoken/main.go --did "did:jwt:test:diag:001" \
  --org "test-org" 2>/dev/null | head -1)
curl -s http://localhost:8080/verify -X POST \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$TOKEN\"}" | jq '.valid'
# Must return: true

# JWKS public key available?
curl -s http://localhost:8080/jwks | jq '.keys[0]'
# Must return RSA key with kty, alg, n, e fields

# INSERT-only enforced?
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c \
"SELECT privilege_type FROM information_schema.role_table_grants
 WHERE table_name='enforcement_decisions' AND grantee='are';"
# Must show: INSERT, SELECT only

# ModeController running?
curl -s http://localhost:8080/health | jq '.enforcement_mode'
# Must return current live mode from Redis (not env var)

# SIEM wired?
grep -rn "SendBlocked" internal/ | grep -v "siem.go"
# Must return non-empty

# Auto-rollback functional?
go test ./internal/enforcement/... -run TestAutoRollback -v
# Must show: PASS

# Slow-walk detection functional?
go test ./tests/attack_corpus/... -run TestSlowWalk -v 2>&1 | tail -5
# Must show: 100% detection

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
DEFINITION OF DONE — L7 (unchanged from v2.6)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Done = L7. Levels 1–6 are NOT done.

L1  Code written
L2  Build passes: go build ./... exits 0, go vet ./... exits 0
L3  Unit tests pass: >90% coverage on scoring + enforcement
L4  FP test suite passes: go test ./tests/fp_scenarios/... ≤ 2%
L5  Explainability validated: reason object schema correct
L6  Latency verified: p99 ≤ 10ms gateway enforcement (measured)
L7  Regression suite passes: go test ./... — all green

For enforcement components, additionally:
L7+ Auto-rollback tested: TestAutoRollback passes
L7+ Hash chain verified: SELECT verify_hash_chain() = true
L7+ Observe mode run 48h before enforce mode enabled
L7+ Feature vectors confirmed: COUNT(*) = COUNT(feature_vector)
L7+ INSERT-only confirmed: privilege_type = INSERT, SELECT only
L7+ SIEM wired: grep SendBlocked callers non-empty
L7+ /verify endpoint: valid token returns {valid: true}

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
COMMIT CONVENTION (unchanged from v2.6)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

feat:      New capability — target ≥80% of all commits
fix:       Correcting something broken — if >10%: audit root cause
security:  Enforcement logic, identity model, tamper evidence
docs:      Documentation only
test:      Tests only
score:     Scoring formula changes, weight adjustments
moat:      Feature vector instrumentation, federation, integration
gtm:       Buyer materials, positioning, pricing updates
eval:      Evaluation harness, attack corpus, FP measurement
observe:   Observability, logging, Prometheus metrics
identity:  JWT library, JWKS, probation logic

🟢 fix: <10% | 🟡 fix: 10–20% | 🔴 fix: >20% → STOP AND AUDIT

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
END-OF-SESSION CHECKLIST — v2.7
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Date:
Phase worked:
Task worked:
Rework commits (fix:): ___ / total ___
Rate: ___% (🟢 <10% | 🟡 10–20% | 🔴 >20%)
Competitive clock: ___ months remaining

1. New incidents this session:
   Type (T13–T24) | Root cause | Time lost | Gate that catches it

2. FP rate: Measured: ___% | Target: ≤2% | Status: 🟢/🟡/🔴

3. Explainability: All decisions have reason objects: YES / NO / PARTIAL

4. INSERT-only confirmed: privilege_type = INSERT+SELECT only: YES / NO

5. SIEM wired: grep SendBlocked callers non-empty: YES / NO

6. Auto-rollback: TestAutoRollback passing: YES / NO

7. Hash chain: verify_hash_chain() = true: YES / NO

8. Feature vectors: missing_vectors = 0: YES / NO

9. /verify endpoint: valid token → {valid: true}: YES / NO

10. Phase discipline: Anti-scope requests: YES / NO | How handled:

11. G-HARDEN status: PASS / FAIL / NOT RUN

12. Next session — first 3 tasks:
    1.
    2.
    3.

13. CONTINUATION_PROMPT.md updated: YES / NO
14. ZROS + MASTER_LEARNINGS uploaded to Claude project: YES / NO

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
END OF ZROS v2.7
Laws: L1–L10 | Incidents: T13–T24 | Gates: 18 total | 4 HARD STOPS
Paired with: APEX v5.2 + MASTER_LEARNINGS v2.1 DELTA
Upload to Claude project as: ZROS-v2.7-AgentRepEngine.md
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━