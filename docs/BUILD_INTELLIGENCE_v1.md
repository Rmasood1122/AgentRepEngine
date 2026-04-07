━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
BUILD_INTELLIGENCE v1.0 — AGENTREPENGINE
What a senior engineer who built this would tell you before you touch it
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Date        : March 22, 2026
Commits     : 58 (March 17–22, 2026)
Built by    : Rehan Masood + APEX v5.2
Stack       : Go · Kong (Lua) · Redis · PostgreSQL · JWT/RS256 · Docker

This document is not ZROS (operational gates) or APEX (strategy).
It is distilled product intelligence — what you learn by building
something that actually works, not by planning it.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SECTION 1 — WHAT THIS PRODUCT ACTUALLY IS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Not what the pitch says. What the code does.

A Kong Lua plugin intercepts every agent request.
It reads X-Agent-DID from the request header.
It looks up the agent's score from Redis (cache-first).
Based on score band, it either allows, throttles, or returns
a synthetic response (not a 403 — never a 403).
It logs the behavioral event asynchronously.

A Go scoring service runs in a Docker container.
It receives behavioral events from Kong via /event endpoint.
It stores them in PostgreSQL with a feature vector.
An event consumer processes the queue and updates scores.
The score update uses: Score = Clamp(0.5*H + 0.5*V, 0, 1000).
H decays with inactivity. V penalizes z-score anomalies.
Policy violations add instant penalties regardless of score.

A ModeController stores enforcement mode in Redis.
It checks FP rate every 5 minutes.
If FP > 2%, it auto-rolls back to observe mode.
No restart needed. Mode change is instant.

That is the entire product. 5,587 lines of Go. 31 files.
Everything else is documentation, tests, and configuration.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SECTION 2 — THE 5 THINGS THAT ACTUALLY MATTER
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

In order of business consequence. Not technical elegance.

1. THE REASON OBJECT
   Every enforcement decision produces a structured JSON object.
   Schema: docs/specs/reason_object_v1.json
   Without this, every blocked agent causes a support ticket.
   With it, the security engineer explains the block themselves.
   This is what makes the product self-defending in a pilot.
   File: internal/scoring/explainability.go

2. THE HASH CHAIN
   Every enforcement decision is linked to the previous via SHA-256.
   This is the SOC2 audit trail. It is what compliance teams cite.
   verify_hash_chain() is called on every health check.
   INSERT-only enforced at PostgreSQL permission level — not app level.
   The app layer can be compromised. The DB layer cannot lie.
   File: migrations/001_initial.sql (REVOKE statement)

3. THE AUTO-ROLLBACK
   ModeController stores enforcement mode in Redis.
   StartFPMonitor() checks FP rate every 5 minutes.
   If FP > 2%: rolls back to observe, fires SIEM alert.
   This is what keeps the pilot alive when thresholds are wrong.
   Without this, one miscalibration ends the pilot permanently.
   File: internal/enforcement/mode_controller.go
   Test: TestAutoRollback

4. THE HIGH_RISK VERIFY LAYER
   Policy threshold breaches trigger human review regardless of score.
   A TRUSTED agent (score 950) cannot silently execute bulk_pii_export.
   This is what catches slow-walk attacks that evade score-based detection.
   PIIFieldAccessRate > 0.3, PermissionEscalationCount > 2,
   CrossTenantProbeCount > 0, SubAgentSpawnDepth > 3 — all HIGH_RISK.
   File: internal/scoring/policy.go

5. THE ~60-SECOND STREAMING DEMO
   scripts/demo.sh runs end-to-end in ~60 seconds with streaming output.
   Shows: JWT identity → behavioral scoring → BLOCKED → reason object
          → hash chain verified → FP rate 0.00%
   Each step prints live with sleep 0.5 pacing — the viewer sees progress.
   This is the entire sales motion in ~60 seconds.
   Every claim is demonstrated live. Nothing is asserted.
   If this breaks, nothing else matters.
   March 31, 2026: Rewritten from batch output (12s) to streaming (60s).
   A demo that prints nothing for 12 seconds feels broken.
   A demo that streams results for 60 seconds feels fast.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SECTION 3 — THE 5 THINGS THAT LOOK IMPORTANT BUT AREN'T (YET)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. KONG JWT SIGNATURE VERIFICATION
   JWKS endpoint exists. /verify endpoint exists.
   Kong calls /verify on each request — but this is in the Lua plugin
   and has not been end-to-end verified in production.
   The security gap is real but mitigated by: port 8080 internal-only,
   SCORING_API_KEY required for direct access.
   Fix when: first enterprise asks about it.

2. ISOLATION FOREST
   Phase 2. Not Phase 1. Requires 90 days of behavioral data.
   Do not implement early. The slow-walk gap it closes is covered
   by HIGH_RISK VERIFY in Phase 1.

3. THE MANAGEMENT UI
   Grafana dashboards are Phase 1 monitoring.
   A polished UI is Phase 2, built after the pilot tells us
   what the security engineer actually needs to see.
   Do not build a UI before you have a user.

4. SOC2 TYPE II CERTIFICATION
   The architecture produces SOC2 evidence.
   The observation period has not started.
   Bring your auditor. We produce the evidence.
   Do not claim SOC2 Type II. Claim SOC2-ready architecture.

5. FEDERATION
   Phase 3. After 3+ paying enterprises with standalone proven value.
   The moat is not the federation protocol.
   The moat is the behavioral data that exists before federation.
   Do not discuss federation in Phase 1 sales conversations.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SECTION 4 — WHAT BROKE AND WHY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Every real bug found by the APEX TEST expert panel.
These are not theoretical. These were in the committed code.

BUG 1: enforcement_decisions had DELETE permission
  Found by: E1 Security, E1-6 command
  Impact: "Tamper-evident" claim was false. Any SQL injection or
          compromised credential could destroy the audit trail.
  Fix: REVOKE UPDATE, DELETE, TRUNCATE ON enforcement_decisions FROM are
       in migrations/001_initial.sql
  Lesson: Application-level INSERT-only is not enough.
          PostgreSQL permissions are the only real guarantee.

BUG 2: BlockedDecisionsTotal metric fired on TRUSTED decisions
  Found by: E2 Infrastructure, Prometheus metrics inspection
  Impact: Metric was useless. Grafana dashboard showed TRUSTED agents
          in the "blocked decisions" counter.
  Fix: Only increment BlockedDecisionsTotal when band == "BLOCKED"
  Lesson: Read your own metrics. If the label is wrong, the alert is wrong.

BUG 3: SCORING_API_KEY was empty string by default
  Found by: E1 Security, E1-4 command
  Impact: Score endpoint fully open in any deployment where the env var
          was not explicitly set. Any agent could probe its own score.
  Fix: Set default value in docker-compose.yml
  Lesson: Empty string is not "not set." It is "open."

BUG 4: SIEM webhook was defined but never called
  Found by: E1 Security, grep for SendBlocked callers
  Impact: Architecture review passes (siem.go exists). Pilot day 1
          reveals nothing fires. Trust destroyed before first incident.
  Fix: Wire SendBlocked() in score_store.go WriteScore() BLOCKED branch
  Lesson: grep for callers before claiming a feature exists.
          A function with zero callers does not exist in production.

BUG 5: pilot-letter-of-understanding.md was never committed
  Found by: E4 Compliance, ls docs/enterprise/
  Impact: The document Lloyd needs to sign was not in the repo.
          It existed in a previous session but was never git add'd.
  Fix: Recreate and commit
  Lesson: git add -A && git status before every session close.
          If it is not committed, it does not exist.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SECTION 5 — ARCHITECTURE DECISIONS AND WHY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Every decision that was not obvious.

DECISION: PostgreSQL async queue instead of Kafka
  Why: Kafka adds 4 hours to enterprise install time.
       A regulated financial services firm needs 3 months of Kafka
       procurement approval. PostgreSQL is already approved.
       Async queue in PostgreSQL is sufficient for Phase 1 volumes.
       Add Kafka in Phase 2 if the queue depth becomes a problem.

DECISION: YAML policy packs instead of OPA/Rego
  Why: A security engineer can read YAML in 5 minutes.
       OPA requires learning a new query language.
       Phase 1 thresholds are simple comparisons.
       YAML files are the right tool for simple comparisons.
       Add OPA in Phase 3 when policies become complex enough to need it.

DECISION: Synthetic response instead of 403 for BLOCKED agents
  Why: A 403 tells an adversarial agent it has been detected.
       A slow synthetic response does not reveal the enforcement threshold.
       Adversaries cannot calibrate their behavior against an unknown threshold.
       This is documented in the reason object — security team sees it.
       Enterprise users do not.

DECISION: Fail-open on infrastructure, fail-closed on enforcement
  Why: These are two different failure classes.
       Infrastructure failure (Redis down): agents keep working.
         Rationale: stopping all agent traffic because a scoring sidecar
         is unreachable is worse than scoring them later.
       Enforcement failure (reason object cannot be generated):
         fallback to AUDIT, never BLOCK.
         Rationale: blocking without explanation destroys trust permanently.
       Both behaviors are documented before any CISO conversation.
       Without documentation, a CISO reads both and sees contradiction.

DECISION: ModeController stores mode in Redis, not env var
  Why: Env vars are read once at startup and are immutable at runtime.
       Auto-rollback requires changing enforcement mode without restart.
       Redis is the only mutable shared state accessible to all services.
       Mode change takes effect on the next Kong plugin request — <1 second.

DECISION: Keys mounted as Docker volume, not baked into image
  Why: Baking keys into a Docker image means rotating keys = rebuilding image.
       A volume mount means key rotation = replace file + restart service.
       More importantly: gentoken runs on the host and needs the same key
       as the scoring service container. Volume mount is the only way to
       share keys between host and container without copying.

DECISION: /verify endpoint in scoring service instead of Lua RSA verifier
  Why: Writing a full RSA verifier in Lua is 200+ lines of error-prone code.
       The scoring service already has VerifyToken() with RS256 + jti replay.
       /verify is 40 lines of Go. It reuses battle-tested crypto libraries.
       Kong calls /verify with 500ms timeout. Fail-open if unavailable.
       This is cleaner, faster to build, and more maintainable.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SECTION 6 — THE SCORING MODEL IN PLAIN ENGLISH
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Score = Clamp(0.5 * H + 0.5 * V, 0, 1000)

H is reputation memory.
  Starts at 700 for new agents.
  Decays with inactivity: H(t) = H(t-1) * e^(-0.1 * days).
  7-day half-life: agent at 900 drops to ~405 after 3 weeks of silence.
  Recovers toward 700 with clean behavior over time.
  High H means the agent has a long history of normal behavior.

V is behavioral anomaly.
  Compares current behavior to the agent's own rolling baseline.
  z = (observed_rate - agent_baseline) / agent_std_dev
  At z > 3.0: velocity penalty applies, capped at -300 points.
  V = 1000 - velocity_penalty (floor at 0)
  High V means current behavior matches historical baseline.

Policy violations override the score.
  Cross-tenant probe: -300 immediate penalty.
  PII field access > 0.3 rate: -200 penalty + HIGH_RISK VERIFY.
  Permission escalation > 2: -150 penalty + HIGH_RISK VERIFY.
  These stack. An agent with all three at once: score drops fast.

Score bands:
  800–1000: TRUSTED   → ALLOW
  500–699:  MONITORED → ALLOW + active audit
  200–499:  RESTRICTED → THROTTLE + human review on HIGH_RISK
  0–199:    BLOCKED   → synthetic response

The slow-walk problem:
  An attacker who gradually increases their rate trains the baseline.
  After 7 days at 2x normal, the baseline has shifted ~50% toward attack rate.
  The z-score stays low because std_dev has also grown.
  Score barely moves. The attacker looks normal.
  Solution: HIGH_RISK VERIFY catches the policy threshold breach
  regardless of score. Slow-walk cannot silently execute high-value actions.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SECTION 7 — FILE MAP (what is where and why)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

internal/identity/
  jwt.go             RS256 sign + verify + jti replay detection
  identity.go        AgentClaims schema, probation logic
  identity_test.go   9 identity tests including replay detection

internal/scoring/
  scorer.go          Phase 1 formula: Score = 0.5*H + 0.5*V
  baseline.go        Rolling baseline store + exponential decay
  policy.go          YAML policy pack loader + violation evaluator
  explainability.go  Reason object generation — the most important file
  consumer.go        Async event queue processor
  features.go        FeatureVector struct — 8 behavioral dimensions
  config.go          ScoringConfig from scoring_weights.yaml

internal/store/
  score_store.go     Redis + PostgreSQL score read/write + SIEM wire

internal/enforcement/
  mode_controller.go ModeController + FP monitor + auto-rollback

internal/audit/
  siem.go            SIEM webhook client (Splunk HEC format)
  replay.go          Forensics replay + SOC2 export
  override.go        Human override workflow with reason code

cmd/scoring-service/
  main.go            HTTP server: /health /score /event /jwks /verify
                     + /audit/replay /audit/export /enforcement/override

kong/plugins/agent-reputation/
  handler.lua        Kong plugin: intercept → verify → score → enforce
  schema.lua         Plugin configuration schema

config/
  scoring_weights.yaml     H + V weights, decay rate, z-score threshold
  policy_packs/*.yaml      5 OWASP LLM Top 10 aligned enforcement rules
  redis/users.acl          Redis ACL — auth required, default user disabled

migrations/
  001_initial.sql    Schema + INSERT-only enforcement on enforcement_decisions

tests/
  eval_harness/      FP gate (0.00%) + TP gate (86.67%)
  fp_scenarios/      100 legitimate scenarios
  attack_corpus/     30 attack scenarios + 10 slow-walk scenarios

docs/
  enterprise/        8 CISO documents including pilot LOU
  compliance/        NIST AI RMF mapping
  architecture/      Scoring model + APEX Laws → ATP mapping
  specs/             reason_object_v1.json schema

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SECTION 8 — WHAT THE NEXT ENGINEER NEEDS TO KNOW
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

If you are picking this up for the first time:

1. Run this first:
   docker compose up -d && sleep 20 && bash scripts/demo.sh
   If the demo passes with streaming output (~60 seconds), the product works.
   If it fails, start with: docker compose logs scoring-service

2. The most dangerous file to touch:
   internal/scoring/explainability.go
   Breaking this breaks G-EXPLAIN gate.
   An unexplained block in an enterprise pilot ends the relationship.
   Always run: go test ./internal/scoring/... -run TestExplain after any change.

3. The most confusing part of the codebase:
   The event pipeline is async and the scoring is async.
   Kong sends events to /event → PostgreSQL queue → consumer processes
   queue → score updates → Redis cache. There is latency in this chain.
   A request scored immediately after an event may not reflect that event.
   This is by design. Do not "fix" it by making scoring synchronous.
   Synchronous scoring adds 500ms+ to every agent request.

4. The main.go tech debt:
   403 lines, touched 10 times, 8 endpoints.
   Before Phase 2: refactor into separate handler files.
   suggested split: handlers/health.go, handlers/score.go,
   handlers/event.go, handlers/audit.go, handlers/verify.go

5. The Kong plugin is Lua, not Go:
   kong/plugins/agent-reputation/handler.lua
   If you have never written Lua, read it carefully before touching it.
   The fail-open on verify service unavailable is intentional.
   The fallback to extract_jwt_claims_unverified is intentional.
   These protect availability when /verify is slow.

6. Never change enforcement_decisions DB permissions:
   migrations/001_initial.sql has REVOKE statements at the end.
   These must survive every docker compose down/up cycle.
   If you add a new migration, do not grant DELETE or UPDATE.
   Verify after every schema change:
     docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c \
     "SELECT privilege_type FROM information_schema.role_table_grants
      WHERE table_name='enforcement_decisions' AND grantee='are';"

7. The keys directory is critical:
   keys/private_key.pem + keys/public_key.pem
   These are mounted into the scoring-service container.
   gentoken runs on the host using these same keys.
   If you rotate keys: replace both files, restart scoring-service.
   Note: a previous private key was committed in error (commit 3048bbc)
   and removed (commit 492d015). The old key is in git history.
   It is not in production. It has been disclosed to enterprise contacts.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SECTION 9 — THE FOUR NUMBERS THAT MATTER
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

0.00%   FP rate on 100-scenario internal corpus
86.67%  TP rate on 30-scenario attack corpus
100%    Slow-walk detection via score + HIGH_RISK VERIFY (10 scenarios)
~60     Seconds for the full demo end-to-end (streaming output)

These are the four numbers you defend in every technical conversation.
Know how to reproduce each with a single command.

FP rate:
  go test ./tests/fp_scenarios/... -v | grep "FP rate"

TP rate:
  go test ./tests/eval_harness/... -v | grep "TP rate"

Slow-walk:
  go test ./tests/attack_corpus/... -run TestSlowWalk -v 2>&1 | tail -5

Demo:
  time bash scripts/demo.sh 2>&1 | tail -5

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
END OF BUILD_INTELLIGENCE v1.0
Built March 17–22, 2026 | 58 commits | 5,587 lines of Go
Upload to Claude project as: BUILD_INTELLIGENCE_v1.md
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━