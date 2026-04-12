# CLAUDE_MASTER v2.0 — AgentRepEngine (ARE)
# Unified Project Intelligence | Built from APEX v5.2 + ZROS v2.7 + BUILD_INTELLIGENCE v1.0 + LEARNING_INTELLIGENCE v3.1 + MASTER_LEARNINGS v2.1
# Last updated: April 5, 2026

---

## ⚠ READ THIS FIRST — OPERATING CONTRACT

This file is the single authoritative context layer for all Claude Code sessions
on AgentRepEngine. It supersedes CLAUDE.md v1.0 entirely.

It is paired with — not a replacement for — these files, all of which must be
loaded in the same project:
  - APEX v5.2 (strategy, gates, modes)
  - ZROS v2.7 (operational discipline, failure taxonomy, ten laws)
  - BUILD_INTELLIGENCE v1.0 (what broke, what the code actually does)
  - LEARNING_INTELLIGENCE v3.1 (73 learnings → 41 execution tasks)
  - MASTER_LEARNINGS v2.1 DELTA (L75–L80 + hardening score tracker)
  - CONTINUATION_PROMPT.md (live state — current commit, open gates, pipeline)

**CONTINUATION_PROMPT.md is the live source of truth for current state.**
This file contains architecture, principles, and rules.
CONTINUATION_PROMPT.md contains the current commit hash, test status, and open gates.
When they conflict: CONTINUATION_PROMPT.md wins on current state.

---

## PRODUCT IDENTITY — WHAT THIS ACTUALLY IS

AgentRepEngine is a runtime behavioral trust enforcement layer.
It sits at the Kong API gateway, scores AI agents against their
own 30-day behavioral baseline using velocity counting and z-score
anomaly detection, and blocks or flags risky agent actions in real time
— before they complete — for CISOs at regulated enterprises (DORA, SEC, HIPAA).

**What the code actually does (from BUILD_INTELLIGENCE, not the pitch):**

A Kong Lua plugin intercepts every agent request. It reads X-Agent-DID
from the request header. It looks up the agent's score from Redis (cache-first).
Based on score band, it allows, throttles, or returns a synthetic response
— never a 403. It logs the behavioral event asynchronously.

A Go scoring service runs in Docker. It receives behavioral events from Kong
via /event. It stores them in PostgreSQL with a feature vector. An event
consumer processes the queue and updates scores using:

  `Score = Clamp(0.5*H + 0.5*V, 0, 1000)`

H = historical trust. Decays with inactivity: H(t) = H(t-1) × e^(-0.1 × days).
7-day half-life. Agent at 900 drops to ~405 after 3 weeks of silence.

V = behavioral anomaly. z = (observed_rate - agent_baseline) / agent_std_dev.
At z > 3.0: velocity penalty applies, capped at -300 points. V = 1000 - penalty.

Policy violations override the score instantly:
- Cross-tenant probe: -300 immediate
- PII field access rate > 0.3: -200 + HIGH_RISK VERIFY
- Permission escalation > 2: -150 + HIGH_RISK VERIFY

Score bands:
  800–1000 → TRUSTED → ALLOW
  500–699  → MONITORED → ALLOW + active audit
  200–499  → RESTRICTED → THROTTLE + human review on HIGH_RISK
  0–199    → BLOCKED → synthetic response (never 403)

A ModeController stores enforcement mode in Redis.
StartFPMonitor() checks FP rate every 5 minutes.
If FP > 2%: auto-rolls back to observe mode + fires SIEM alert.
No restart needed. Mode change is instant.

That is the entire product. 5,587 lines of Go. 31 files.
Everything else is documentation, tests, and configuration.

---

## THE ONE TRUTH THAT OVERRIDES EVERYTHING ELSE

**The codebase looks like a scoring engine.
It is an FP management system with enforcement as a side effect.
Optimize for FP elimination first. Detection second. Always.**

The 0.00% FP rate is the enterprise unlock phrase — not the architecture,
not the compliance alignment, not the OWASP mapping.

Correct framing (never deviate from this):
"0.00% false positive rate on our 150-scenario internal validation corpus.
External validation on your production traffic is available during the pilot
— we expect <1% on a well-configured environment."

Never say "zero false positives." Say "0.00% on internal corpus."
The number stays. The framing protects credibility (L77).

---

## THE FOUR NUMBERS — KNOW COLD. REPRODUCE ON DEMAND.

```
0.00%   FP rate on 100-scenario internal corpus
88.00%  TP rate on 50-scenario attack corpus (TP=88.00%, Precision=100%, F1=0.9362)
100%    Slow-walk detection via score + HIGH_RISK VERIFY (10 scenarios, single-agent)
~60s    Full demo end-to-end (scripts/demo.sh — streaming output, not batch)
```

Reproduction commands (run before any enterprise conversation):

```bash
# FP rate
go test ./tests/fp_scenarios/... -v | grep "FP rate"

# TP rate (4-metric format required by enterprise reviewers — L87)
go test ./tests/eval_harness/... -v | grep -E "TP rate|FP rate|Precision|F1"

# Slow-walk
go test ./tests/attack_corpus/... -run TestSlowWalk -v 2>&1 | tail -5

# Demo integrity
time bash scripts/demo.sh 2>&1 | tail -5
```

**If scripts/demo.sh fails: stop all other work until it passes.**
The demo is the entire sales motion in 60 seconds.
Every claim is demonstrated live. Nothing is asserted.
If it breaks, nothing else matters. (BUILD_INTELLIGENCE Section 2 item 5)

---

## CURRENT PHASE

**Phase 1. T8 is the sole open gate.**

All T0–T7 complete. Hardening 96/100. Tests green.
Sprint 1 is next: `internal/certification/report.go Generate()`
(only when Lloyd LoU is signed — do not implement until in sprint)

See CONTINUATION_PROMPT.md for current commit hash and live test status.

Do not implement Phase 2 or Phase 3 stubs under any circumstance.
Do not suggest expanding scope. The engineering is done.
The open gate is commercial, not technical.

---

## TECH STACK

| Layer | Technology | Version / Notes |
|---|---|---|
| Language | Go | 1.24 |
| Gateway plugin | Kong (Lua) | min 2.8 — enforced in plugin, not at deploy time |
| Score cache | Redis | 512mb, noeviction, AOF+RDB, ACL user `are_admin` |
| Event store | PostgreSQL | port 5433 (NOT 5432) |
| Runtime | Docker Desktop on Windows | MINGW64 / Git Bash |
| Identity | RS256 JWT + org JWKS | Per-request, /verify endpoint |
| Observability | Prometheus + Grafana | Configured, not Phase 1 critical path |
| Test container | golang:1.24-alpine | Linux parity |
| Shell | Git Bash (MINGW64) | MSYS_NO_PATHCONV=1 required for docker run |
| Python | 3.14 | Utility scripts only — C:/Users/rmaso/AppData/Local/Python/pythoncore-3.14-64/python.exe |
| Editor | VS Code | `code <filename>` to open before editing |
| Repo | github.com/Rehanrana11/AgentRepEngine | branch: master (NOT main) |
| IP anchor | Zenodo DOI 10.5281/zenodo.19169185 | Published March 22, 2026 |

PostgreSQL connection: `docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c "..."`
Redis connection: `docker exec agentrepengine-redis-1 redis-cli --no-auth-warning --user are_admin -a are_redis_dev KEYS "*"`

---

## DIRECTORY STRUCTURE — ONE RESPONSIBILITY PER DIRECTORY

Every directory has exactly one responsibility. If generated code
doesn't fit cleanly into one directory's responsibility, it is
in the wrong place.

```
cmd/
  scoring-service/      HTTP entrypoint. :8080. /health /score /event /jwks /verify
                        /audit/replay /audit/export /enforcement/override
                        ⚠ main.go is 403 lines — known tech debt. See TECH DEBT section.
  verify-chain/         Customer-runnable audit log verifier. No logic — calls internal/audit only.
  certify/              STUB. Phase 2. Do not implement.
  certification-portal/ STUB. Phase 2. Do not implement.

internal/
  identity/             JWT library. RS256 sign + verify + jti replay detection.
                        Files: jwt.go, identity.go (AgentClaims, probation logic), identity_test.go
  scoring/              ALL scoring logic. Phase 1 formula only. Nothing else.
                        Files: scorer.go, baseline.go, policy.go, explainability.go,
                               consumer.go (async queue processor), features.go (FeatureVector, 8 dims),
                               config.go (ScoringConfig from scoring_weights.yaml)
                        ⚠ explainability.go is the most dangerous file in the codebase. See FRAGILE section.
  gateway/              Kong Lua plugin. Intercept → cache lookup → allow/block.
                        No scoring logic here. No policy logic here.
                        Files: handler.lua, schema.lua
                        ⚠ Most fragile component. See FRAGILE section.
  store/                Redis cache layer + PostgreSQL event store + retention job + SIEM wire.
                        Files: score_store.go (Redis+PG read/write + SendBlocked() caller)
  enforcement/          ModeController + FP monitor + auto-rollback.
                        Files: mode_controller.go
                        ⚠ StartFPMonitor() must be called before any enforce-mode go-live. (L75, L8)
  policy/               YAML policy pack loader + violation evaluator. ALL policy logic lives here.
  explainability/       Reason object construction. Called once per enforcement decision. Never inline.
  replay/               Forensics mode. Event replay. SOC2 export. Read-only operations.
  audit/                Hash chain construction + tamper detection + SIEM webhook + override workflow.
                        Files: siem.go, replay.go, override.go
                        ⚠ Never modify hash logic without explicit approval. Breaks all existing logs.
  certification/        STUB. Sprint 1 Task 1 — implement report.go Generate() ONLY when in sprint.
  trust/                STUB. Phase 3. Do not touch.
  intelligence/         STUB. Phase 2. Do not touch.
  api/                  STUB. Phase 2. Do not touch.

tests/
  evaluation/           Attack corpus. READ ONLY — never modify retroactively. (See NEVER-TOUCH)
  fp_scenarios/         100 legitimate scenarios — the FP gate corpus.
  attack_corpus/        30 attack scenarios + 10 slow-walk scenarios.
  eval_harness/         Combined FP + TP gate runner.
  integration/          pilot_readiness_test.go + integration tests.

config/
  scoring_weights.yaml  H + V weights, decay rate, z-score threshold.
  policy_packs/*.yaml   5 OWASP LLM Top 10 aligned enforcement rules.
  redis/users.acl       Redis ACL — auth required, default user disabled.

migrations/
  001_initial.sql       Schema + INSERT-only REVOKE on enforcement_decisions.
                        ⚠ REVOKE statements must survive every docker compose down/up.

docs/
  enterprise/           8 CISO documents including pilot LoU, operational-safety-architecture.md
  compliance/           NIST AI RMF mapping
  architecture/         Scoring model + APEX Laws → ATP mapping
  specs/                reason_object_v1.json schema
  competitive/          microsoft-response.md, competitive-positioning.md
  regulatory/           dora-examiner-protocol.md
  ops/                  key-management, redis-failover, chain-verification, kong-compatibility,
                        enforce-mode-gate, capacity

scripts/
  demo.sh               ~60-second streaming end-to-end demo. Streaming output, not batch.
                        ⚠ If this breaks, stop everything else until it passes.

web/                    STUB. Phase 2. Do not touch.
keys/                   Gitignored. RS256 key pair. NEVER committed.
                        ⚠ A prior private key was committed in error (commit 3048bbc)
                           and removed (commit 492d015). The old key is in git history.
                           It is not in production. It has been disclosed to enterprise contacts.
```

---

## LAYER RULES — HARD RULES, NOT GUIDELINES

If generated code violates these rules, reject it and ask for a corrected version.
These are not style preferences. Violations create untestable, unsafe code.

| Logic type | Correct location | Prohibited locations |
|---|---|---|
| Agent scoring formula, conditionals, weights | `internal/scoring/` | handlers, cmd, gateway, anywhere else |
| Policy pack evaluation, violation scoring | `internal/policy/` | inline in handlers, scoring, gateway |
| Reason object construction | `internal/explainability/` | inline inside any handler or middleware |
| Hash chain writes | `internal/audit/` | nowhere else, ever |
| JWT validation, replay detection | `internal/identity/` | anywhere else |
| Redis reads/writes | `internal/store/` | direct in handlers or cmd |
| Enforcement mode control | `internal/enforcement/` | env vars, handler logic |
| Error logging | structured fields at handler boundary | silent swallowing, anywhere |
| Demo output | `scripts/demo.sh` | Go test output, cmd main |

**The test:** If you are writing a conditional that determines agent behavior,
it belongs in `internal/scoring/` or `internal/policy/`.
If it is in a handler, cmd, or gateway file — it is wrong. Move it.

**Fail-open vs fail-closed — two distinct behaviors, both correct (L78, ZROS L2):**
- Infrastructure failure → FAIL OPEN. Always pass traffic. Never become the
  reason an agent pipeline goes down. Circuit breaker in Kong plugin.
- Enforcement failure → FAIL CLOSED. No unexplained blocks. Every block
  has a reason object. A block without a reason object does not ship.

These are documented separately in `docs/enterprise/operational-safety-architecture.md`.
Any CISO reading both behaviors needs the rationale for each. Both are correct.
The distinction: infrastructure problems never stop business; enforcement
problems never produce unexplained blocks.

---

## STATE OWNERSHIP — ABSOLUTE RULES

```
Redis       = hot state only.
              Current scores. Velocity windows. TTL-based.
              CRITICAL PATH. p99 ≤2ms cache hit required.
              Can be wiped and rebuilt from PostgreSQL.
              NEVER the source of truth.
              Also: enforcement mode (ModeController), used-token cache (replay detection).

PostgreSQL  = durable state. Source of truth for everything.
              Full behavioral event log. Hash chain. Policy decisions.
              Enforcement_decisions (INSERT-only at DB permission level — L9).
              Queue (agent_event_queue). Baselines. Feature vectors.
              CANNOT be reconstructed from Redis.
              NEVER reverse this dependency.

YAML files  = policy state. Loaded at startup. No runtime mutation in Phase 1.

JWT         = identity state. Per-request validation via JWKS. No server-side session.
              jti claim required. Redis used-token cache prevents replay.

Baseline    = Welford's running mean/variance per org_id+agent_did (NOT just agent_did).
              Key format: baseline:{org_id}:{agent_did}
              Missing org_id = cross-tenant baseline contamination. (L92, TW-PRE-2)
```

**The dependency rule is absolute:**
Redis depends on PostgreSQL. PostgreSQL never depends on Redis.
Any code that writes to Redis before PostgreSQL is wrong.

---

## ASYNC PATTERN

- PostgreSQL async queue: `agent_event_queue` → background goroutine → score update → Redis cache
- Kong plugin: synchronous on critical path (Redis cache hit only — no queue on hot path)
- Score updates: async and off critical path (p95 ≤500ms target)
- No Kafka. No callbacks. No channels as primary event transport.
- Go goroutines + PostgreSQL queue only.

**Known intentional mixed pattern — do not "fix":**
`/score` is synchronous but triggers an async queue write internally.
This is the fail-open design. Making scoring synchronous adds 500ms+
to every agent request. Do not change this without explicit instruction.

**The async pipeline has latency — do not change it to synchronous:**
Kong sends event to /event → PostgreSQL queue → consumer processes queue
→ score updates → Redis cache. A request scored immediately after an event
may not reflect that event. This is by design. (BUILD_INTELLIGENCE Section 8 item 3)

---

## ERROR HANDLING

| Location | Behavior | Rule |
|---|---|---|
| Gateway plugin | FAIL OPEN always | Infrastructure unreachable → pass through + audit log |
| Kong /verify unreachable | Fallback to extract_jwt_claims_unverified | Intentional — protects availability |
| Scoring service | Structured log + handler-level panic recovery | Nothing silently swallowed |
| Audit chain integrity failure | Hard error | Surfaces in /health. Not auto-recoverable. Human required. |
| Queue processing failures | Log + continue (no dead-letter in Phase 1) | Exponential backoff retry |
| Agent-facing responses | Allow/block only | Agent NEVER sees ARE reason objects |
| Operator-facing errors | Structured JSON with reason codes | |
| All endpoints | Timeout enforced | /verify: 500ms, /score: 5s (L80) |

**The fail-open rule is non-negotiable.**
ARE never becomes the reason an agent pipeline goes down.
This is documented in `docs/enterprise/operational-safety-architecture.md`
and must be explained to every CISO with its rationale — not just stated as a fact.

---

## THE FIVE COMPONENTS THAT ACTUALLY MATTER
(In order of business consequence. From BUILD_INTELLIGENCE Section 2.)

**1. THE REASON OBJECT** — `internal/scoring/explainability.go`
Every enforcement decision produces a structured JSON object.
Schema: `docs/specs/reason_object_v1.json`
Without this: every blocked agent causes a support ticket.
With it: the security engineer explains the block themselves.
This is what makes the product self-defending in a pilot.
⚠ Breaking explainability.go breaks G-EXPLAIN gate.
After any change: `go test ./internal/scoring/... -run TestExplain`

**2. THE HASH CHAIN** — `internal/audit/` + `migrations/001_initial.sql`
Every enforcement decision linked to the previous via SHA-256.
This is the SOC2 audit trail. INSERT-only enforced at PostgreSQL
permission level — not app level. The app can be compromised.
The DB cannot lie. (ZROS L9)
After any schema change: verify REVOKE statements survived.

**3. THE AUTO-ROLLBACK** — `internal/enforcement/mode_controller.go`
ModeController stores enforcement mode in Redis.
StartFPMonitor() checks FP rate every 5 minutes.
If FP > 2%: rolls back to observe, fires SIEM alert.
No restart needed. Mode change is instant.
This is what keeps the pilot alive when thresholds are wrong.
Without it: one miscalibration ends the pilot permanently.
GATE: TestAutoRollback must pass before enforce mode authorized. (L75, L8)

**4. THE HIGH_RISK VERIFY LAYER** — `internal/scoring/policy.go`
Policy threshold breaches trigger human review regardless of score.
A TRUSTED agent (score 950) cannot silently execute bulk_pii_export.
This is what catches slow-walk attacks that evade score-based detection.
Thresholds: PIIFieldAccessRate > 0.3, PermissionEscalationCount > 2,
CrossTenantProbeCount > 0, SubAgentSpawnDepth > 3 — all HIGH_RISK.
The slow-walk problem: attacker who gradually increases rate trains
the baseline. After 7 days at 2x normal, baseline has shifted ~50%.
Z-score stays low because std_dev has grown. Score barely moves.
HIGH_RISK VERIFY catches the policy threshold breach regardless of score.
Also: variance growth rate monitoring — doubling of weekly variance
triggers early warning before attack succeeds. (L34, MUL-04)

**5. THE ~60-SECOND STREAMING DEMO** — `scripts/demo.sh`
Shows: JWT identity → behavioral scoring → BLOCKED → reason object
       → hash chain verified → FP rate 0.00%
Streaming output with sleep 0.5 pacing. Viewer sees progress.
A demo that prints nothing for 12 seconds feels broken.
A demo that streams results for 60 seconds feels fast. (L80)
Rewritten March 31, 2026. Do not revert to batch output.

---

## SECURITY REQUIREMENTS — MUST BE TRUE AT ALL TIMES

Every feature must pass all of these before it is done.

1. Every agent request carries a unique RS256-signed JWT validated
   against org JWKS endpoint. No shared identities. No unsigned requests scored.
   jti claim required. Redis used-token cache prevents replay.

2. Private keys in `keys/`. Gitignored. Never committed. Rotation procedure:
   `docs/ops/key-management.md`. Full JWT reissuance required on rotation.
   ⚠ Prior private key committed in error (commit 3048bbc), removed (492d015).
   Old key is in git history. Not in production. Disclosed to enterprise contacts.

3. Redis: default user is OFF. Only `are_admin` ACL user. Do not add users.
   Do not enable default user. Do not change ACL without explicit instruction.

4. Enforce mode requires 14 consecutive days of clean observe mode.
   Gate documented in `docs/ops/enforce-mode-gate.md`.
   Do not weaken or remove this gate without documented decision.
   Weakening creates regulatory liability under DORA/SEC.

5. Audit log is hash-chained. Historical modification breaks chain
   and surfaces in `/health`. This is a feature, not a bug.
   INSERT+SELECT only on enforcement_decisions at PostgreSQL permission level.
   Verify after any schema change or docker compose down/up.

6. `cmd/verify-chain` allows customer-independent verification.
   The customer must never need to trust ARE's own verification output.

7. Data never leaves the customer environment. No telemetry. No callbacks
   to ARE infrastructure. No analytics SDKs. No external logging services.

8. All agent event payloads validated against schema before queue write.
   Malformed events are rejected and logged — never silently dropped.

9. SIEM webhook must be wired, not just defined.
   Before any SIEM claim: `grep -rn "SendBlocked" internal/ | grep -v "siem.go"`
   Must return non-empty. If empty: wire before claiming. (L76, ZROS L10)

10. Org-scoped baseline isolation required. Key format: `baseline:{org_id}:{agent_did}`
    Missing org_id = cross-tenant contamination + false horizontal scaling claims. (L92)

---

## PROHIBITED LIBRARIES — HARD FENCES

Do not suggest, import, or reference these. Reject immediately if proposed.

| Library / Pattern | Reason | Revisit when |
|---|---|---|
| Kafka | Ops complexity kills 4-hour install target | Phase 2 only if queue is bottleneck |
| OPA (Open Policy Agent) | YAML threshold config is the policy engine | Phase 2+ if policy complexity demands |
| Isolation Forest / scikit-learn / ML inference | Requires 90 days of behavioral data first | After 90-day baseline exists |
| DID / ledger / blockchain identity | JWT RS256 is the identity model | Never in Phase 1 |
| CI/CD that writes to master | Runtime enforcement not yet proven in production | After production proof |
| Any external analytics SDK | Data must never leave customer environment | Never |
| Multi-tenant architecture | Single-tenant not yet validated | After Phase 1 single-tenant proven |
| Federation protocol | Phase 3 moat — standalone value must be proven first | Phase 3 only |

---

## PERFORMANCE CONSTRAINTS — HARD CEILINGS

| Metric | Target | Hard ceiling | Notes |
|---|---|---|---|
| Gateway score lookup | p95 ≤5ms | p99 ≤10ms | Cache hit is critical path |
| Cache hit | — | p99 ≤2ms | Redis miss triggers async refresh |
| Score update (async) | p95 ≤500ms | — | Off critical path |
| Scoring engine call overhead | 0.25ns Linux | Zero allocations | Measured in tests |
| Demo runtime | ~60 seconds | — | Streaming. Never revert to 12s batch. |
| Phase 1 capacity | ≤500 agents | ≤1000 req/min | Documented in docs/ops/capacity.md |
| Redis memory | 512mb | noeviction | New writes fail. No evictions. |
| /verify endpoint timeout | — | 500ms | (L80) |
| /score endpoint timeout | — | 5s | (L80) |

---

## KONG GATEWAY PLUGIN — SPECIAL HANDLING

`internal/gateway/` and `kong/plugins/agent-reputation/handler.lua`
is the most fragile component in the codebase.

**Why it's fragile:**
1. Lua has no type safety. A change to the scoring service response format
   will not be caught at compile time — only at runtime when a request hits the plugin.
2. Fail-open means a broken plugin silently passes all traffic.
   The failure mode is INVISIBLE to the agent pipeline.
3. The Kong version check (min 2.8) is enforced in the plugin itself,
   not at deploy time — version mismatch produces confusing runtime errors,
   not a clean startup failure.
4. The fallback to `extract_jwt_claims_unverified` when /verify is slow
   is intentional for availability. Do not remove it.

**Rule:** Any change to the scoring service response schema requires
a manual test of the Kong plugin against that new schema before commit.
No automated guard exists here. This check is manual. Always.

---

## THE FIVE BUGS THAT ALREADY HAPPENED — DO NOT REPEAT
(From BUILD_INTELLIGENCE Section 4)

**BUG 1: enforcement_decisions had DELETE permission**
Claim "tamper-evident" was false. Any SQL injection or compromised
credential could destroy the audit trail.
Fix applied: REVOKE in migrations/001_initial.sql.
Lesson: Application-level INSERT-only is not enough. DB permissions are the guarantee.
Verify: `SELECT privilege_type FROM information_schema.role_table_grants WHERE table_name='enforcement_decisions' AND grantee='are';`

**BUG 2: BlockedDecisionsTotal metric fired on TRUSTED decisions**
Grafana dashboard showed TRUSTED agents in the "blocked" counter.
Metric was useless. Alert would have been wrong.
Lesson: Read your own metrics. If the label is wrong, the alert is wrong.

**BUG 3: SCORING_API_KEY was empty string by default**
Score endpoint fully open in any deployment where env var was not set.
Any agent could probe its own score.
Fix applied: Default value set in docker-compose.yml.
Lesson: Empty string is not "not set." It is "open."

**BUG 4: SIEM webhook defined but never called**
Architecture review passes (siem.go exists). Pilot day 1: nothing fires.
Trust destroyed before first incident.
Fix applied: SendBlocked() wired in score_store.go WriteScore() BLOCKED branch.
Lesson: `grep -rn "SendBlocked" internal/ | grep -v "siem.go"` before any SIEM claim.

**BUG 5: pilot-letter-of-understanding.md never committed**
The document Lloyd needs to sign was not in the repo.
Fix applied: Recreated and committed.
Lesson: `git add -A && git status` before every session close.

---

## NEVER-TOUCH LIST

These require explicit human review before any modification.
State what you are about to change and why. Wait for approval.

| Item | Why it's sacred |
|---|---|
| `keys/` | RS256 private keys. Rotation = full JWT reissuance for all agents. |
| `internal/audit/` hash logic | Any change to how events are hashed breaks ALL existing audit logs. |
| `tests/evaluation/` corpus | Retroactive modification invalidates the 0.00% FP claim. Corpus is a historical artifact. |
| `docs/ops/enforce-mode-gate.md` | Weakening the 14-day rule creates regulatory liability under DORA/SEC. |
| `docker-compose.yml` | Redis noeviction + AOF/RDB persistence config. Wrong change = silent data loss. |
| `migrations/001_initial.sql` REVOKE statements | Removing these re-opens the tamper vulnerability. Must survive every down/up. |
| `scripts/demo.sh` | The sales proof artifact. Breaking it means nothing can be demonstrated live. |
| `go push -f` to master | Never. |
| Any push to master without green tests | Never. |

---

## KNOWN TECH DEBT — DO NOT FIX UNLESS ASKED

These are known, accepted, and documented. Surfacing them uninvited
wastes session time and violates scope discipline.

- **Redis dead-letter not implemented.** Failed queue events log and continue.
  Acceptable for Phase 1 (≤500 agents).
- **main.go is 403 lines.** Known. Before Phase 2: refactor into handlers/health.go,
  handlers/score.go, handlers/event.go, handlers/audit.go, handlers/verify.go.
  Do not refactor now.
- **Helm chart missing.** Blocked on Lloyd confirming Kubernetes environment.
  Do not build until confirmed.
- **Phase 2/3/4 stubs compile but have zero callers and no logic.**
  `internal/trust/`, `internal/intelligence/`, `internal/certification/`,
  `internal/api/`, `cmd/certify/`, `cmd/certification-portal/`, `web/`.
  These exist to anchor Phase 2/3 architecture without blocking Phase 1.
  They compile. They do not run. Do not implement.
- **Slow-walk detection is single-agent scope only.** Multi-agent coordinated
  slow-walk not detected. Caveat documented in 3 customer-facing docs. Do not remove.
- **Hardening score 96/100.** 4 remaining points are known gaps. Not Phase 1 targets.
- **`peer_cluster.go` and `threshold_calibration.go` are Phase 2.** Do not implement.
- **JWKS endpoint + /verify end-to-end not verified in Kong production path.**
  Security gap is real but mitigated by port 8080 internal-only + SCORING_API_KEY.
  Fix when first enterprise asks. Do not panic-fix before Lloyd meeting.
- **Prior private key in git history** (commit 3048bbc). Not in production.
  Disclosed. Do not attempt to purge git history without explicit instruction.

---

## FRAGILE DEPENDENCIES MAP

```
Kong Lua plugin → scoring service response schema
  (no type safety — schema change requires manual plugin test before commit)

Redis → PostgreSQL
  (Redis is never source of truth — wipe Redis, rebuild from PG, never reverse)

Audit chain → event hash order
  (any reordering of events breaks the chain — hash logic in internal/audit/ is sacred)

Enforce mode → 14-day observe gate
  (regulatory liability if weakened — documented in enforce-mode-gate.md)

0.00% FP claim → tests/evaluation/ corpus integrity
  (retroactive modification = credibility destroyed — corpus is immutable)

SIEM claim → SendBlocked() callers
  (grep before every enterprise SIEM conversation)

"Horizontally scalable" claim → baseline:{org_id}:{agent_did} key format
  (missing org_id = false claim + cross-tenant contamination)

scripts/demo.sh integrity → entire sales motion
  (demo broken = nothing demonstrable = no pilot conversation)
```

---

## PRE-SESSION VERIFICATION CHECKLIST

Run at the start of every session before any code work:

```bash
# 1. Services running
docker ps | grep -E "redis|postgres|kong|agent-rep"

# 2. Tests green
go test ./...

# 3. Scoring service health
curl http://localhost:8080/health

# 4. Kong plugin active
curl http://localhost:8001/plugins | jq '.data[].name'

# 5. INSERT-only on enforcement_decisions (ZROS L9)
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c \
  "SELECT privilege_type FROM information_schema.role_table_grants \
   WHERE table_name='enforcement_decisions' AND grantee='are';"
# Expected: INSERT, SELECT only. Any UPDATE/DELETE = CRITICAL VULNERABILITY.

# 6. SIEM webhook wired (ZROS L10)
grep -rn "SendBlocked" internal/ | grep -v "siem.go"
# Must return non-empty.

# 7. Baseline key format (L92 — horizontal scaling claim integrity)
docker exec agentrepengine-redis-1 redis-cli --no-auth-warning \
  --user are_admin -a are_redis_dev KEYS "baseline:*" | head -5
# Expected: baseline:{org_id}:{agent_did} format
# Failure: baseline:{agent_did} only = cross-tenant contamination risk
```

---

## NAMING CONVENTIONS

| Item | Convention | Status |
|---|---|---|
| Files | snake_case.go | Consistent |
| Exported functions | PascalCase | Consistent |
| Unexported functions | camelCase | Consistent |
| Variables | camelCase | Consistent |
| DB tables | snake_case (agent_event_queue, agent_scores, audit_log, enforcement_decisions) | Consistent |
| API routes | lowercase, no versioning (/health, /score, /events, /verify, /jwks) | Consistent |
| Redis keys | namespace:id format (baseline:{org_id}:{agent_did}, score:{agent_id}) | Consistent |
| Test files | *_test.go | Consistent |
| Commit prefixes | feat/fix/refactor/chore/docs/test/score/security/moat/gtm/eval | Convention not tooling |

Do not introduce new naming patterns without explicit instruction.
Consistency is the invariant. When in doubt, match existing files.

---

## DEFINITION OF DONE

A feature is done when ALL of these are true. Not before.

```
[ ] go build ./... passes with zero errors
[ ] go test ./... passes green
[ ] No new layer violations (scoring logic in handlers, etc.)
[ ] No prohibited library imported
[ ] Every error path has structured logging — nothing silently swallowed
[ ] Fail-open behavior preserved on gateway path
[ ] Security requirements checklist passed (see SECURITY section above)
[ ] INSERT-only on enforcement_decisions verified if schema was touched
[ ] SIEM wiring verified if enforcement code was touched
[ ] If internal/audit/ touched — explicit human approval received first
[ ] If tests/evaluation/ touched — explicit human approval received first
[ ] scripts/demo.sh still passes (~60 seconds, streaming output)
[ ] CONTINUATION_PROMPT.md updated with new commit hash
[ ] CONTINUATION_PROMPT.md uploaded to Claude Project
[ ] Committed to master with descriptive prefix (feat/fix/etc.)
```

---

## PROHIBITED SCOPE — TRIGGER IMMEDIATE CHALLENGE

If any of these appear in a session, name the violation and stop:

| What appears | Name to say | Why it's blocked |
|---|---|---|
| Kafka | ANTI-SCOPE: Kafka | Phase 1 locked stack |
| OPA | ANTI-SCOPE: OPA | Phase 1 locked stack |
| Isolation Forest | ANTI-SCOPE: Isolation Forest | Needs 90 days of behavioral data |
| Federation protocol | ANTI-SCOPE: Federation | Phase 3 — standalone value first |
| Multi-tenant | ANTI-SCOPE: Multi-tenant | Single-tenant not yet validated |
| DID / blockchain / ledger | ANTI-SCOPE: DID | Phase 1 locked stack |
| CI/CD automation to master | ANTI-SCOPE: CI/CD | Production not yet proven |
| trust/, intelligence/ stubs | STUB CREEP | They compile. They don't run. |
| New CISO documents when pipeline unsent | SCOPE CREEP SUBSTITUTION | Name the gate |
| Engineering work when T8 is open | SCOPE CREEP SUBSTITUTION | Name the gate |

---

## FAILURE PATTERNS — NAMED AND ENFORCED

Say the name mid-session to immediately reset behavior.

**SCOPE CREEP SUBSTITUTION**
When commercial action feels uncertain, engineering scope expands.
More features, more hardening, more research — all feel productive.
None advance the open gate (T8).
Watch for: Suggesting new engineering work when T8 is the only open gate.
Fix: Name the gate. "T8 is the prime directive. What advances T8 right now?"

**SILENT FAIL-OPEN REGRESSION**
A change to error handling in the gateway path introduces non-fail-open behavior.
Watch for: Any `return error` in the Kong plugin critical path without pass-through.
Fix: Gateway always passes traffic on infrastructure failure. Always.

**LAYER BLEED**
Scoring conditional written inside a handler or cmd file.
Watch for: Business logic conditionals outside `internal/scoring/` or `internal/policy/`.
Fix: Move to correct layer. Handler calls internal package. Never reverse.

**REDIS DEPENDENCY INVERSION**
Code writes to Redis before PostgreSQL, or treats Redis as source of truth.
Watch for: Any score write that doesn't touch PostgreSQL first.
Fix: PostgreSQL always first. Redis is always the derived cache.

**CORPUS CONTAMINATION**
Modifying `tests/evaluation/` attack scenarios retroactively.
Watch for: Any suggested edit to existing evaluation corpus files.
Fix: Never. The corpus is a historical artifact. Immutable.

**STUB CREEP**
Implementing logic inside a Phase 2/3 stub because it "seems simple."
Watch for: Any code generation touching `internal/trust/`, `internal/intelligence/`,
`internal/certification/` outside an explicitly scoped sprint task.
Fix: Stop. Confirm sprint scope. Stubs compile. They do not run.

**SIEM GHOST**
Claiming SIEM capability without verifying the caller exists.
Watch for: Any addition of enforcement code without checking SendBlocked() callers.
Fix: `grep -rn "SendBlocked" internal/ | grep -v "siem.go"` before every claim.

**METRIC OVERCONFIDENCE**
Stating "0.00% false positives" without the corpus qualifier.
Watch for: Any unqualified "zero" or "0.00%" FP claim in materials or conversation.
Fix: "0.00% on our 100-scenario internal validation corpus. External validation
available during pilot."

---

## SESSION OPERATING PROTOCOL

**Every session starts with this exact sequence:**

```
1. Claude reads CONTINUATION_PROMPT.md (live state — commit hash, open gates)
2. Claude runs: go test ./...
3. Claude confirms: tests green / not green + current commit hash
4. Claude states: current phase, open gates, Sprint 1 status
5. Claude runs pre-session verification checklist (see above)
6. Rehan confirms summary is accurate or corrects it
7. Work begins only after step 6
```

**During every session:**
- Run G-KILL before any build work (phase alignment check)
- Plan Mode before every non-trivial feature
- Every code change ends with: `go build ./... && go test ./...`
- One feature at a time — verify before advancing
- If same tool error repeats twice — start new session
- Monitor context window — new session at 40-50%
- Complete file replacements preferred over diffs for any file >50 lines
- No partial implementations — if building a function, build it to production
  quality with tests in the same session

**Paste protocol:**
- Rehan pastes only code blocks into terminal — never explanation text
- `[200~` prefix = bracketed paste error — stop immediately, fix terminal state
- Never chain pipes in a single exec call — write to temp file, then execute

**Every session ends with:**
```
1. go build ./... && go test ./... green
2. Committed to master with descriptive prefix (feat/fix/etc.)
3. CONTINUATION_PROMPT.md updated (new commit hash, test status, any new decisions)
4. CONTINUATION_PROMPT.md uploaded to Claude Project
5. README.md implementation status updated if applicable
```

---

## SHELL OPERATION RULES

Runtime: Git Bash on Windows (MINGW64).

```bash
# MSYS_NO_PATHCONV=1 required for ALL docker run commands
MSYS_NO_PATHCONV=1 docker run --rm \
  -v "//c/Users/rmaso/AgentRepEngine:/app" \
  -w /app golang:1.24-alpine go test ./...

# Standard build + test
go build ./... && go test ./...

# Push (master, not main — always)
git push origin master

# Never force push
# Never: git push -f
# Never: git push origin main
```

---

## COMPETITIVE CONTEXT — WHAT CLAUDE CODE NEEDS TO KNOW

**Primary threat: Gen Digital ADR** (launched March 2026)
Direct live competitor. No competitive response prepared as of April 2.
This is an open item. Do not ignore if referenced in session.

**Check Point–Lakera acquisition** (March 2026, ~$300M estimated)
Lakera is now inside Check Point's platform integration roadmap.
Product velocity slows during integration (12–24 months to ship behavioral
history as a Check Point feature). Check Point has 100K+ enterprise customers.
Bundling risk replaces velocity risk as primary competitive threat.
ARE's answer: "Check Point bundles. We deploy in days. That is the contrast."
Competitive clock: 12 months from March 2026.

**ARE is positioned as:** Agent Trust Infrastructure.
NOT endpoint security. NOT API security. NOT prompt injection detection.
The wedge: Kong plugin. No SDK required. No vendor lock-in.
The differentiator: behavioral history per agent, not per call.
"Every competitor scores the call. We score the agent. That difference is everything."

**Andy Watkin-Child** (DORA/SEC specialist) and **Unmukt Raizada** (ex-Goldman/JPMorgan)
are warm Boardy intros not yet contacted. Both are pipeline assets for the commercial gate.

---

## ENTERPRISE CONVERSATION RULES

These are non-negotiable in any enterprise conversation:

1. Never mention "MVP," revenue projections, or the "24x roadmap"
   unless the other party raises value first. Lead with technical value.

2. Lead with Use Cases 1 (rogue agent containment, DORA) and 5
   (observe-mode baseline validation) for Lloyd.

3. CISO unlock phrase: **"You control the pace."**
   Full phrase: "You control the pace — we don't advance to enforce mode
   without your sign-off after 14 days of clean observe."

4. FP framing (never deviate): "0.00% on our 100-scenario internal
   validation corpus. External validation available during pilot."

5. 4-metric format required for any performance claim (L87):
   "TP=86.67%, FP=0.00%, Precision=100%, F1=0.9286"

6. Do not discuss federation in Phase 1 sales conversations.
   Do not claim SOC2 Type II — claim "SOC2-ready architecture."

7. GDPR Article 22 (for DORA/EU buyers): "ARE's staged rollout with
   explicit CISO sign-off is the human oversight documentation." (L91)

8. Fail-open narrative must always include the rationale, not just the fact.
   The CISO will perceive contradiction without it. (L78)

---

## CLASSIFICATION REGISTER — WHAT TO REVIEW WEEKLY

| Concept | Class | Why it matters |
|---|---|---|
| FP elimination as primary optimization | M | Every feature decision flows from this |
| Scoring formula: Score = Clamp(0.5*H + 0.5*V, 0, 1000) | M | Core algorithm — any drift is a defect |
| Fail-open on infrastructure / fail-closed on enforcement | M | Regulatory + trust requirement |
| Redis never source of truth | M | Silent data loss risk if inverted |
| Layer rules (scoring/policy location) | M | Violated constantly without enforcement |
| Hash chain integrity = hard error | M | The enterprise trust artifact |
| Kong Lua has no type safety | M | Most fragile point — manual test required |
| 14-day enforce gate | M | Regulatory liability if weakened |
| Corpus integrity (0.00% FP) | M | Product claim integrity |
| Phase 2/3 stubs do not run | M | Scope discipline |
| Auto-rollback before enforce mode | M | Pilot survival requirement (L75, L8) |
| SIEM caller verification | M | Enterprise trust on day 1 (L76, L10) |
| Org-scoped baseline key format | M | Horizontal scaling claim integrity (L92) |
| Demo integrity (scripts/demo.sh) | M | Sales proof artifact |
| T8 is the prime directive | M | Everything else is subordinate |
| Async pipeline latency is by design | U | Do not "fix" to synchronous |
| MSYS_NO_PATHCONV=1 | U | Windows Docker shell requirement |
| Prometheus/Grafana status | U | Configured, not critical path |
| Prohibited library list | U | Reference before any new dependency |
| Performance ceiling numbers | U | Reference when designing new endpoints |
| Dead-letter debt | R | Known, accepted, documented |
| Helm chart status | R | Blocked on Lloyd confirmation |
| main.go refactor debt | R | Before Phase 2, not now |

---

## THE RULE THAT COMPOUNDS EVERYTHING

After every session:
"Do I understand this well enough to explain every decision made
here to someone who has never seen it?"

If yes → advance.
If no → that gap is today's only study target.

The 0.00% FP rate is the enterprise unlock phrase.
Everything else — the architecture, the compliance alignment, the OWASP mapping,
the hash chain, the auto-rollback — serves one conversation:

"Our FP rate is 0.00% on our internal corpus.
Here is the harness you can run yourself.
You control when enforcement goes live."

Never let engineering drift from that truth.

---

## SYSTEM VERSION REGISTER

This file works with:
- APEX v5.2 (March 17, 2026)
- ZROS v2.7 (March 22, 2026)
- BUILD_INTELLIGENCE v1.0 (March 22, 2026)
- LEARNING_INTELLIGENCE v3.1 (March 31, 2026)
- MASTER_LEARNINGS v2.1 DELTA (March 31, 2026)
- CONTINUATION_PROMPT.md (live — update every session)

CLAUDE_MASTER version: 2.0
Built: April 5, 2026
Supersedes: CLAUDE.md v1.0
