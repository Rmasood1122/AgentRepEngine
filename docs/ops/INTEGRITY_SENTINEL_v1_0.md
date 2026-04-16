━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
INTEGRITY SENTINEL v1.0 — AGENTREPENGINE
Product Integrity Gate + Learning Automation Audit System
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Version      : 1.0
Date         : April 16, 2026
Paired with  : APEX v5.2 + ZROS v2.8 + APEX-OS v1.0 + CONTINUATION_PROMPT.md
Supersedes   : Nothing — additive only. Zero overlap with existing files.
Gate ID      : G-SENTINEL (fires at every session open, after G-COMMERCIAL)
Commit to    : docs/ops/INTEGRITY_SENTINEL_v1_0.md
Upload to    : Claude Project alongside CONTINUATION_PROMPT.md

ARCHITECTURAL RULE:
  This file adds ONE thing that does not exist anywhere in the current OS:
  A mandatory evidence-first product state verification protocol
  that fires BEFORE any engineering work opens, using live demo output
  — not documentation — as the source of truth.

  EXISTING FILES REMAIN FULLY AUTHORITATIVE.
  This file never overrides APEX, ZROS, APEX-OS, or CLAUDE_MASTER.
  If conflict exists: existing file wins. Flag conflict for resolution.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
WHAT THIS SYSTEM IS — AND WHY IT EXISTS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

THE PROBLEM THIS SOLVES:

Every session in this project has produced changes. Most changes were additive.
Some changes silently broke existing capabilities — policy engine disconnected,
score writes silently failing, config not in Docker image. These were only
discovered by running a live demo end-to-end, hours into the session.

The current system (APEX + ZROS + APEX-OS) is excellent at:
  ✅ Preventing scope creep
  ✅ Maintaining commercial discipline
  ✅ Tracking what was built
  ✅ Naming failure patterns

The current system has ONE gap:
  ❌ No mechanism that verifies what ACTUALLY EXISTS in the running system
     before new work begins — using live evidence, not memory or docs.

ZROS tracks what was committed. It does not verify what is running.
BUILD_INTELLIGENCE documents what the code does. It does not run the code.
CONTINUATION_PROMPT records what was built. It cannot detect silent regressions.

THE SOLUTION: Evidence-first state verification.

Before any engineering work opens, run the SENTINEL PROTOCOL.
It takes 8 minutes. It produces a signed state snapshot.
Every subsequent change is evaluated against that snapshot.
If a change cannot pass the post-change verification, it does not ship.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 1 — THE SENTINEL PROTOCOL
Fires at every engineering session open. Takes 8 minutes.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

ACTIVATION COMMAND: SENTINEL RUN

Claude asks you to run these commands in sequence and paste output.
Claude then evaluates each output against the INVARIANT TABLE below.
Any invariant failure = HARD STOP before any engineering work opens.

─────────────────────────────────────────────────────────────────────
STEP S1 — INFRASTRUCTURE STATE (60 seconds)
─────────────────────────────────────────────────────────────────────

Run:
  docker compose ps
  go test ./... 2>&1 | tail -5
  curl -s http://localhost:8080/health -H "X-API-Key: are-internal-key-change-in-production" | python -m json.tool

Expected invariants:
  I-1: All 6 containers healthy (scoring-service, kong, postgres, redis, prometheus, grafana)
  I-2: go test ./... — ALL GREEN, zero FAIL lines
  I-3: health response contains:
       "status": "ok"
       "redis": "ok"
       "postgres": "ok"
       "hash_chain_valid": true
       "enforcement_mode": "observe"

If I-1 fails: stack not up. Run docker compose up -d. Do not proceed.
If I-2 fails: tests broken. Fix before any new work. Do not proceed.
If I-3 fails: service unhealthy. Diagnose. Do not proceed.

─────────────────────────────────────────────────────────────────────
STEP S2 — SCORING DEMO VERIFICATION (3 minutes)
─────────────────────────────────────────────────────────────────────

Run:
  export SCORING_API_KEY=are-internal-key-change-in-production
  docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine \
    -c "DELETE FROM agent_events WHERE agent_did='sentinel-verify-001'; DELETE FROM agent_identities WHERE did='sentinel-verify-001';"

  for i in $(seq 1 15); do
    curl -s -X POST "http://localhost:8080/event" \
      -H "X-API-Key: $SCORING_API_KEY" \
      -H "Content-Type: application/json" \
      -d '{"agent_did":"sentinel-verify-001","event_type":"tool_call","org_id":"sentinel-org","feature_vector":{"tool_call_rate_per_hour":2.0,"unique_endpoints_per_hour":3,"bulk_access_count_per_session":0,"pii_field_access_rate":0.01,"cross_tenant_probe_count":0,"permission_escalation_count":0,"sub_agent_spawn_depth":1,"token_refresh_rate":0.1}}'
  done

  sleep 5

  curl -s -X POST "http://localhost:8080/event" \
    -H "X-API-Key: $SCORING_API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"agent_did":"sentinel-verify-001","event_type":"tool_call","org_id":"sentinel-org","feature_vector":{"tool_call_rate_per_hour":20.0,"unique_endpoints_per_hour":15,"bulk_access_count_per_session":8,"pii_field_access_rate":0.35,"cross_tenant_probe_count":0,"permission_escalation_count":0,"sub_agent_spawn_depth":1,"token_refresh_rate":0.1}}'

  sleep 5
  curl -s "http://localhost:8080/score/sentinel-verify-001" \
    -H "X-API-Key: $SCORING_API_KEY" | python -m json.tool

Expected invariants:
  I-4: score ≤ 499 (RESTRICTED or BLOCKED band) — anomaly was detected
  I-5: policy_fired ≠ "no_policy_fired" — policy layer is wired and firing
  I-6: worst_z > 3.0 — z-score anomaly layer is active
  I-7: source = "cache" (not "orphan") — score writes are persisting
  I-8: org_id is a valid UUID (not empty, not "default") — upsert is working

If I-4 fails: scoring engine not detecting anomalies. HARD STOP.
If I-5 fails: policy engine disconnected. HARD STOP.
If I-6 fails: z-score baseline not building. HARD STOP.
If I-7 fails: score write pipeline broken. HARD STOP.
If I-8 fails: identity upsert broken. HARD STOP.

─────────────────────────────────────────────────────────────────────
STEP S3 — KONG GATEWAY VERIFICATION (2 minutes)
─────────────────────────────────────────────────────────────────────

Run:
  JWT=$(go run ./cmd/gentoken/main.go 2>/dev/null)
  # Warmup call
  curl -s http://localhost:8000/health \
    -H "Authorization: Bearer $JWT" \
    -H "X-Agent-DID: did:jwt:finserv-demo:trading-agent:001" > /dev/null
  # Verified call
  curl -s http://localhost:8000/health \
    -H "Authorization: Bearer $JWT" \
    -H "X-Agent-DID: did:jwt:finserv-demo:trading-agent:001" \
    -D - | grep -E "X-Agent|X-Kong-Proxy|HTTP"

Expected invariants:
  I-9:  HTTP/1.1 200 OK — Kong is proxying
  I-10: X-Agent-Score header present in response
  I-11: X-Agent-Band header present in response
  I-12: X-Agent-DID-Verified header present in response
  I-13: X-Kong-Proxy-Latency ≤ 500 on warm call (second call)

If I-9 fails: Kong not routing. HARD STOP.
If I-10/11/12 fail: plugin not setting response headers. HARD STOP.
If I-13 fails: latency regression. Investigate before any Kong work.

─────────────────────────────────────────────────────────────────────
STEP S4 — AUDIT CHAIN VERIFICATION (30 seconds)
─────────────────────────────────────────────────────────────────────

Run:
  docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine \
    -c "SELECT COUNT(*) as total_decisions, MIN(chain_position) as first_pos, MAX(chain_position) as last_pos FROM enforcement_decisions;"
  docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine \
    -c "SELECT privilege_type FROM information_schema.role_table_grants WHERE table_name='enforcement_decisions' AND grantee='are' ORDER BY privilege_type;"

Expected invariants:
  I-14: enforcement_decisions table accessible (count query returns)
  I-15: privilege_type shows INSERT and SELECT ONLY — no UPDATE, DELETE, TRUNCATE

If I-15 fails: tamper-evident audit trail compromised. HARD STOP.
This is the SOC 2 CC7.2 claim. It cannot be false.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SENTINEL STATE SNAPSHOT — RECORD AFTER EVERY PASS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

After running the SENTINEL PROTOCOL, record this in CONTINUATION_PROMPT.md:

SENTINEL LAST RUN: [date]
  I-1  Stack healthy:          [PASS/FAIL]
  I-2  Tests green:            [PASS/FAIL]
  I-3  Health endpoint:        [PASS/FAIL]
  I-4  Anomaly detected:       [PASS/FAIL] score=[N] band=[BAND]
  I-5  Policy fired:           [PASS/FAIL] policy=[NAME]
  I-6  Z-score active:         [PASS/FAIL] worst_z=[N]
  I-7  Score persisting:       [PASS/FAIL] source=[cache/orphan]
  I-8  Identity upsert:        [PASS/FAIL] org_id=[UUID/fail]
  I-9  Kong routing:           [PASS/FAIL]
  I-10 X-Agent-Score header:   [PASS/FAIL]
  I-11 X-Agent-Band header:    [PASS/FAIL]
  I-12 X-Agent-DID-Verified:   [PASS/FAIL]
  I-13 Kong latency warm:      [PASS/FAIL] latency=[N]ms
  I-14 Audit chain accessible: [PASS/FAIL]
  I-15 INSERT-only enforced:   [PASS/FAIL]
  OVERALL: [PASS/FAIL] — [N]/15 invariants passing

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 2 — THE TRADEOFF EVALUATOR
Fires before any proposed change is implemented.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

ACTIVATION COMMAND: SENTINEL EVALUATE — [proposed change description]

For every proposed code change, Claude runs this 5-question evaluation
BEFORE writing a single line of code. All 5 questions must be answered.

Q1 — INVARIANT RISK
  Which of the 15 SENTINEL invariants could this change affect?
  List each one. For each: SAFE / RISK / UNKNOWN.
  If any RISK: state the specific mechanism of risk.
  If UNKNOWN: run the affected invariant test after the change.

Q2 — COMPOUNDING CHECK
  Does this change produce output that Phase 2 needs?
  (Feature vectors, behavioral data, audit records, baseline data)
  If YES: classify S5 (compounding mechanism). Document it.
  If NO: is this change additive-only or does it modify existing behavior?

Q3 — DILUTION CHECK
  Does this change reduce the precision, coverage, or strength of any
  existing detection capability?
  Existing capabilities: z-score detection, policy threshold detection,
  hash chain integrity, fail-open behavior, deception model (synthetic 200),
  jti replay detection, auto-rollback, SIEM wire.
  For each capability the change touches: PRESERVED / WEAKENED / UNKNOWN.
  Any WEAKENED = HARD STOP. Name the weakening explicitly.

Q4 — CLAIM IMPACT
  Which of C1–C15 does this change affect?
  For each affected claim: STRENGTHENED / NEUTRAL / WEAKENED / BREAKS.
  Any BREAKS = HARD STOP.
  Any WEAKENED = adversarial gate fires before proceeding.

Q5 — REVERSIBILITY
  If this change produces unexpected behavior in production:
  Can it be reverted in under 5 minutes?
  What is the revert procedure?
  If irreversible: requires DECISION_AUDIT entry before proceeding.

PASS CRITERIA: No HARD STOPs, no BREAKS, revert procedure documented.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 3 — THE POST-CHANGE VERIFIER
Fires after every commit before session closes.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

ACTIVATION COMMAND: SENTINEL VERIFY

After any commit that touches scoring, policy, Kong plugin, score_store,
consumer, Dockerfile, or docker-compose.yml:

1. Rebuild: docker compose up -d --build scoring-service
2. Re-run only the invariants affected by the change (from Q1 above)
3. Compare results to SENTINEL LAST RUN snapshot
4. Any invariant that was PASS and is now FAIL = REGRESSION. Do not push.

REGRESSION PROTOCOL:
  Name the regression: "I-[N] regressed after [commit]"
  Identify root cause before any other work
  Fix and re-verify before pushing
  Add regression as new ZROS incident: T-type entry

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 4 — LEARNING INTELLIGENCE AUDIT
Fires monthly. Diagnoses whether the learning system is actually improving.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

ACTIVATION COMMAND: SENTINEL LEARNING AUDIT

This answers the question: "Is our learning intelligence actually making us
smarter, or is it just accumulating documents?"

─────────────────────────────────────────────────────────────────────
CURRENT STATE DIAGNOSIS — APRIL 16, 2026
─────────────────────────────────────────────────────────────────────

FILES IN THE LEARNING SYSTEM:
  1. LEARNING_INTELLIGENCE v3.1    — 73 learnings, last updated March 31, 2026
  2. MASTER_LEARNINGS v2.1 DELTA   — L75–L99, absorbed into active files
  3. BUILD_INTELLIGENCE v1.0       — what the code actually does, March 22, 2026
  4. DECISION_AUDIT v1.0           — [H] claim tracking
  5. APEX_REASONING_ENGINE v1.0    — 7-expert framework
  6. APEX-OS v1.0                  — 6 improvement mechanisms

DIAGNOSIS — IS THE SYSTEM LEARNING?

LEARNING_INTELLIGENCE: ⚠️ STALE
  Last updated: March 31, 2026. Current date: April 16, 2026.
  16 days of sessions have produced zero new learning entries.
  Next learning number L144 has been "waiting for paste next session"
  since at least April 12. It was never pasted.
  STATUS: The learning pipeline is broken. The intake valve is open
  (external learnings exist) but the extraction step is not executing.

MASTER_LEARNINGS: ⚠️ STALE
  Next MASTER_LEARNINGS number L100 has been pending since April 12.
  Not updated. Zero new entries from 3 sessions of significant work.
  STATUS: Same broken pipeline as LEARNING_INTELLIGENCE.

BUILD_INTELLIGENCE: 🔴 CRITICALLY STALE
  Last updated: March 22, 2026. 25 days old.
  This session alone produced 5 significant changes to how the code works:
    - Policy engine wired to consumer
    - Score writes now upsert with real UUIDs
    - Kong plugin now has Redis ACL auth
    - Policy thresholds recalibrated to rate scale
    - Dockerfile now includes config directory
  NONE of these are reflected in BUILD_INTELLIGENCE v1.0.
  BUILD_INTELLIGENCE says "policy engine loaded but never called" — this
  is now false. The document is actively misleading.
  STATUS: CRITICAL. The most valuable document in the system describes
  a product that no longer exists. An engineer onboarding from this doc
  would have a wrong mental model of the system.

DECISION_AUDIT: ⚠️ UNCALIBRATED
  Calibration status: UNCALIBRATED (as of April 14, 2026).
  First verification pass due: May 7, 2026.
  D001–D010 exist but have never been verified against outcomes.
  STATUS: The [H] claim tracking exists but produces no signal yet.
  This is expected — calibration requires 30 days of decisions.
  Not a failure, but the system is producing zero feedback.

APEX-OS v1.0: ✅ INSTALLED, ⚠️ NOT FIRING
  6 mechanisms defined. Zero mechanisms confirmed executing.
  Ω SCAN: never recorded in CONTINUATION_PROMPT.
  SUCCESS LOG: never populated (no S-category entries in any session).
  COMMERCIAL_INTELLIGENCE.md: never created.
  PIPELINE HEALTH SCORE: never computed this session.
  STATUS: The mechanisms exist but are not being executed. The system
  is installed but not running.

─────────────────────────────────────────────────────────────────────
ROOT CAUSE OF LEARNING SYSTEM FAILURE
─────────────────────────────────────────────────────────────────────

The learning system has a structural failure in the EXTRACTION step.

The system has three steps:
  INTAKE → EXTRACTION → APPLICATION

INTAKE is working: external learnings exist (courses, research, conversations).
APPLICATION is working: learnings are cited in architecture decisions.
EXTRACTION is broken: the step that converts session work into new learning
entries is not executing.

Specific symptoms:
  1. L144 "waiting for paste" since April 12 — never executed
  2. BUILD_INTELLIGENCE not updated after any session since March 22
  3. APEX-OS mechanisms defined but not firing
  4. SUCCESS LOG never populated despite 3 sessions of significant wins
  5. Ω SCAN never recorded in CONTINUATION_PROMPT

The extraction step requires 20–30 minutes at session close.
It has been deferred every session in favor of commercial or engineering work.
This is the Ω2 pattern (premature completeness) applied to learning:
assuming the system is learning because the documents exist.

─────────────────────────────────────────────────────────────────────
THE FIXES — LEARNING SYSTEM REPAIR
─────────────────────────────────────────────────────────────────────

FIX 1 — BUILD_INTELLIGENCE UPDATE TRIGGER (add to ZROS as L12)

L12 (NEW): BUILD_INTELLIGENCE is updated at session close when any
           feat: commit touches a file in the FILE MAP.
           Affected files: consumer.go, score_store.go, handler.lua,
           scorer.go, policy.go, baseline.go, explainability.go,
           mode_controller.go, siem.go, main.go.
           Update format: find the relevant SECTION, add a DELTA entry.
           Never rewrite — append only. One paragraph per commit.
           Gate: G-BUILD-INTEL fires if feat: commit lands without a
           corresponding BUILD_INTELLIGENCE update.

FIX 2 — LEARNING EXTRACTION AT SESSION CLOSE (add to SESSION_PROTOCOL)

At session close (step C4), before CONTINUATION_PROMPT update:
  "What was learned this session that is not in LEARNING_INTELLIGENCE?"
  For each new learning: 4-line verdict (COMPOUND/QUEUE/REJECT + task).
  Paste to LEARNING_INTELLIGENCE as L[next number].
  If 0 new learnings: explicitly state "no new learnings this session."
  Never skip this step. It takes 10 minutes maximum.

FIX 3 — APEX-OS FIRING CONFIRMATION (add to CONTINUATION_PROMPT)

Add to CURRENT STATE block in CONTINUATION_PROMPT:
  APEX-OS STATUS:
    Ω SCAN: Ω1=[N] Ω2=[N] Ω3=[N] Ω4=[N] (run at session open)
    SUCCESS LOG: [N entries this session]
    COMMERCIAL_INTELLIGENCE: [last entry date]
    PIPELINE HEALTH: [score/10]
    Last calibration: [date or UNCALIBRATED]

FIX 4 — BUILD_INTELLIGENCE DELTA ENTRIES NEEDED NOW

The following changes from this session must be added to BUILD_INTELLIGENCE
before the next session opens:

DELTA — April 16, 2026 (commit 5c65e0a):

  POLICY ENGINE NOW WIRED [F]:
  internal/scoring/consumer.go processEvent() now calls
  c.policy.Evaluate(v) after z-score computation. PolicyEngine
  loaded at NewEventConsumer() startup from config/policy_packs/.
  Violations apply ceiling override: score reduced by worst.ScorePenalty.
  policyFired field in ScoringPayload now populated with policy name.
  This closes the gap where policy packs existed but never fired.

  SCORE WRITES NOW UPSERT [F]:
  internal/store/score_store.go WriteScore() changed from UPDATE to
  INSERT ... ON CONFLICT (did) DO UPDATE. Adds org_id and instance_id
  as gen_random_uuid() on first insert. Resolves silent score write
  failure for new agents not yet in agent_identities table.

  CONFIG IN DOCKER IMAGE [F]:
  Dockerfile now includes COPY --from=builder /app/config ./config.
  Policy packs are available at /app/config/policy_packs/ inside container.
  Previously: policy packs existed on host but not in image. Policy engine
  loaded at startup but found empty directory and silently failed.

  POLICY THRESHOLDS RECALIBRATED [F]:
  bulk_pii_access_prevention: pii_field_access_rate warning/throttle/block
  changed from 50/100/500 (integer counts) to 0.10/0.20/0.30 (rate scale).
  high_frequency_tool_abuse: tool_call_rate changed from 150/300/600 to 5/10/15.
  nis2_proportionality: fixed duplicate thresholds: key (YAML parse error).
  prompt_injection_prevention: tool_call_rate changed from 150/250/500 to 5/10/15.
  Root cause: feature vector sends rates (0.0–1.0), not integer counts.

DELTA — April 16, 2026 (commit 03354ab):

  KONG REDIS AUTH [F]:
  kong/plugins/agent-reputation/handler.lua get_redis_client() now
  calls red:auth(user, password) when conf.redis_password is set.
  Schema updated with redis_password and redis_user fields (no default,
  optional). Kong declarative config updated with redis_password and
  redis_user values matching Redis ACL configuration.
  Previously: Redis rejected every connection with NOAUTH. Kong plugin
  failed open on every request (score=700, enforcement suspended).

  KONG RESPONSE HEADERS [F]:
  handler.lua access() phase now calls kong.response.set_header() for
  X-Agent-Score, X-Agent-Band, X-Agent-DID-Verified in addition to
  kong.service.request.set_header(). Previously: headers were set on
  the upstream request only — not visible to the calling client.

  SINGLE WORKER CONFIGURATION [F/DEMO ONLY]:
  docker-compose.yml: KONG_NGINX_WORKER_PROCESSES=1 added.
  Required for Windows Docker Desktop demo — cold worker TCP connect
  takes ~7.9 seconds per worker. Single worker: one cold start per
  session, then sub-200ms for all subsequent calls.
  PRODUCTION NOTE: revert to "auto" before any production deployment.
  This env var must not ship to production as-is.

  KONG VERIFY SKIP IN OBSERVE MODE [F]:
  handler.lua access(): verify_token() call skipped when
  enforcement_mode == "observe". Local JWT parsing used instead.
  In enforce mode: full RS256 verification via /verify endpoint retained.
  Rationale: /verify HTTP call was adding ~4s latency in observe mode
  on Windows Docker Desktop. Enforcement decisions do not depend on
  server-side verification in observe mode (all pass through anyway).

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 5 — THE COMPOUNDING TRACKER
Tracks whether each session is compounding or diluting product value.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

COMPOUNDING SCORE — calculated at session close.

For each session, answer these 5 questions. Score 0–2 each.

C1 — DETECTION CAPABILITY: Did detection capability improve?
  0: No change or regression
  1: Existing capability fixed/verified
  2: New detection capability added and verified by demo

C2 — CLAIM STRENGTH: Did the evidence base for C1–C15 improve?
  0: No change or regression in any claim
  1: Existing claim verified by live evidence
  2: New claim earned by live evidence

C3 — AUDIT TRAIL: Did the tamper-evident audit trail improve?
  0: No change or potential regression
  1: No change, verified intact
  2: New audit capability added

C4 — LEARNING CAPTURE: Did the learning system get updated?
  0: No learning extractions this session
  1: Some learnings extracted, BUILD_INTELLIGENCE not updated
  2: Full extraction: learnings + BUILD_INTELLIGENCE + APEX-OS mechanisms

C5 — COMMERCIAL ADVANCEMENT: Did the pipeline advance?
  0: No commercial action taken
  1: Outreach sent or meeting scheduled
  2: Meeting held or reply received and next step taken

Score interpretation:
  8–10: COMPOUNDING — this session built durable value
  5–7:  NEUTRAL — value built but not fully captured
  0–4:  DILUTING — session consumed time without building durable value

SESSION COMPOUNDING SCORES — RECORD HISTORY:
  Session 1 (April 7):   [not scored]
  Session 2 (April 12):  [not scored]
  Session 3 (April 16):  C1=2, C2=2, C3=1, C4=0, C5=0 → SCORE: 5/10 NEUTRAL

SESSION 3 ANALYSIS: Detection capability improved (C1=2, C2=2). Audit chain
intact (C3=1). Learning not captured (C4=0 — BUILD_INTELLIGENCE not updated,
no APEX-OS firing). Commercial action not taken this session (C5=0 —
TW-REHEARSAL still unchecked). The engineering was excellent. The capture
and commercial execution were not.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 6 — ACTIVATION COMMANDS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

SENTINEL RUN
  Fires SENTINEL PROTOCOL (Parts 1–4). Takes 8 minutes.
  Required: paste S1/S2/S3/S4 output. Claude evaluates all 15 invariants.
  Output: SENTINEL STATE SNAPSHOT + PASS/FAIL for each invariant.

SENTINEL EVALUATE — [description]
  Fires TRADEOFF EVALUATOR (Part 2) on a proposed change.
  Required: describe the proposed change in one sentence.
  Output: Q1–Q5 answers + PASS/FAIL/HARD STOP verdict.

SENTINEL VERIFY
  Fires POST-CHANGE VERIFIER (Part 3) after a commit.
  Required: paste affected invariant test output.
  Output: REGRESSION / NO REGRESSION + any new ZROS incidents.

SENTINEL LEARNING AUDIT
  Fires LEARNING AUDIT (Part 4). Monthly.
  Output: diagnosis of which learning files are stale + FIX list.

SENTINEL SCORE
  Fires COMPOUNDING TRACKER (Part 5) at session close.
  Output: C1–C5 scores + session total + COMPOUNDING/NEUTRAL/DILUTING.

SENTINEL STATUS
  Full system snapshot. Combines all parts into one output.
  Output: 15-invariant table + compounding score + learning system status.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 7 — INTEGRATION WITH EXISTING SYSTEM
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

WHERE THIS FITS IN THE SESSION SEQUENCE:

  A1: 90-second boot (existing)
  A2: G-COMMERCIAL gate (existing)
  A3: ★ SENTINEL RUN ← NEW — fires here, before any engineering
  A4: G-KILL phase check (existing)
  A5: Engineering work opens

  C1: Engineering complete
  C2: go test ./... green
  C3: ★ SENTINEL VERIFY ← NEW — fires here, before commit
  C4: ★ SENTINEL SCORE ← NEW — fires here, at session close
  C5: git commit + push
  C6: CONTINUATION_PROMPT updated with SENTINEL LAST RUN snapshot

WHERE THIS FITS IN THE FILE HIERARCHY:

  Commit to: docs/ops/INTEGRITY_SENTINEL_v1_0.md
  Upload to: Claude Project (same batch as CONTINUATION_PROMPT.md)
  Reference in: CONTINUATION_PROMPT.md ACTIVE FILES list (position 9)
  Add to SESSION_PROTOCOL: A3, C3, C4 steps above

GATES ADDED TO ZROS v2.8:
  G-SENTINEL: SENTINEL RUN must pass before engineering opens
  G-BUILD-INTEL: feat: commit requires BUILD_INTELLIGENCE delta entry
  These are new gates. Add to ZROS v2.9 at next version.

NEW LAWS FOR ZROS v2.9:
  L12: BUILD_INTELLIGENCE updated at session close when feat: commit
       touches a file in the FILE MAP. Never rewrite — append delta.
  L13: APEX-OS mechanisms fire at session open and session close.
       Not firing = not running. Not running = not improving.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PART 8 — IMMEDIATE ACTIONS THIS SESSION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Before this session closes, three things must happen:

ACTION 1 — Add BUILD_INTELLIGENCE delta entries (Part 4, Fix 4 above)
  The delta entries are already written above. Copy them into
  BUILD_INTELLIGENCE_v1.md under a new SECTION: DELTA LOG.
  Commit as: docs: BUILD_INTELLIGENCE v1.1 — April 16 delta entries

ACTION 2 — Add SENTINEL to CONTINUATION_PROMPT active files list
  Add line 9: "9. INTEGRITY_SENTINEL v1.0 ← product integrity + learning audit"
  Add SENTINEL LAST RUN block (all 15 invariants blank for next session)
  Add APEX-OS STATUS block (Ω SCAN, SUCCESS LOG, PIPELINE HEALTH)

ACTION 3 — Commit this file to repo
  cp INTEGRITY_SENTINEL_v1_0.md docs/ops/INTEGRITY_SENTINEL_v1_0.md
  git add docs/ops/INTEGRITY_SENTINEL_v1_0.md
  git commit -m "ops: INTEGRITY_SENTINEL v1.0 — product integrity + learning audit system"
  git push origin master

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
END OF INTEGRITY SENTINEL v1.0
Gates: G-SENTINEL + G-BUILD-INTEL | Invariants: 15 | Commands: 6
Proposed ZROS additions: L12, L13 | SESSION_PROTOCOL additions: A3, C3, C4
Upload alongside CONTINUATION_PROMPT.md every session.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
