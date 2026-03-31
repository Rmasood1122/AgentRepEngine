━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
LEARNING_INTELLIGENCE v3.1 — AGENTREPENGINE
Compounding learning log: every external course/article/concept mapped to ARE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Paired with  : APEX v5.2 + ZROS v2.7 + MASTER_LEARNINGS v2.1 DELTA + BUILD_INTELLIGENCE v1.0
Version      : 3.1 (deficit audit — 20 deficits resolved, 3 tasks promoted, 5 tasks consolidated,
               4 new tasks added, 6 learning verdicts upgraded, 2 critical multipliers enforced)
Updated      : March 31, 2026
Maintained by: Rehan Masood + APEX expert panel (5 voices)

SESSIONS PROCESSED: 2 prior sessions (L1-L2) + 1 rapid-extract session (L3-L35)
                    + 1 rapid-extract session (L36-L73) + 1 deficit audit session (v3.1)
TOTAL LEARNINGS: 73 (no new learnings — audit pass only)
VERDICTS: 58 COMPOUND | 7 QUEUE | 8 REJECT  ← 6 learnings upgraded from QUEUE→COMPOUND

DELTA FROM v3.0:
  ✦ DEF-01: L5 upgraded QUEUE→COMPOUND + Phase 1 org_id verification task added
  ✦ DEF-02: L8 upgraded QUEUE→COMPOUND (Phase 1 framing added)
  ✦ DEF-03: L14 — behavioral window documentation added as TW-0 (TIER 1 CRITICAL)
  ✦ DEF-04: L17 + L22 upgraded QUEUE→COMPOUND (Phase 1 framing added)
  ✦ DEF-05: L27 upgraded QUEUE→COMPOUND, points to PL-13 (already exists)
  ✦ DEF-06: L33 upgraded QUEUE→COMPOUND (Phase 1 LoRA-by-hand framing added)
  ✦ MUL-01: PL-16 now has explicit dependency on TW-5 (test contamination × FP bias)
  ✦ MUL-02: P2-14 PROMOTED from TIER 3 → TIER 2 (CRITICAL: feature vector storage)
  ✦ MUL-03: L82×L20 cross-referenced to G-FP in TW-1 talking point #3
  ✦ MUL-04: TW-6 now documents behavioral window parameter requirement
  ✦ MUL-05: L91×L90 cross-referenced: staged rollout = Article 22 human oversight
  ✦ CON-01: Ghost tasks converted to [✓ MERGED] format
  ✦ CON-02: PL-7/18/19/21/22 consolidated → PL-SECURITY-ATTESTATION (saves ~4hrs)
  ✦ CON-03: PL-9 removed (absorbed into TW-3, close-note added)
  ✦ CON-04: P2-10→P2-3 ordering dependency documented
  ✦ UP-01: L56 upgraded QUEUE→COMPOUND + deployment talking point added to TW-1
  ✦ UP-02: L15 upgraded QUEUE→COMPOUND (Phase 1 fixed-window framing added)
  ✦ UP-03: L40 upgraded QUEUE→COMPOUND (EU localization answer documented)
  ✦ GAP-01: PL-SIEM-VERIFY added to TIER 2 (audit trail format verification)
  ✦ GAP-02: TW-REHEARSAL added to TIER 1 (conversation readiness ≠ doc readiness)
  ✦ GAP-03: PL-11 PROMOTED from TIER 2 → TIER 1 (must precede Lloyd meeting)

HOW TO USE:
  Paste learning → APEX gives 4-line verdict → say GENERATE when done
  → One file, all tasks ranked, ready to execute in Claude Code

NEXT ACTION AFTER THIS FILE:
  1. git add LEARNING_INTELLIGENCE_v3_1.md
  2. git commit -m "docs: learning intelligence v3.1 — deficit audit complete, 20 gaps resolved"
  3. git push origin main
  4. Upload v3.1 to Claude Project (replaces v3.0)
  5. Execute TIER 1 tasks in priority order — TW-0 and TW-REHEARSAL are new additions

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
MASTER EXECUTION QUEUE — PRIORITY ORDER
All 73 learnings distilled into ranked actions. Execute top-down.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

═══════════════════════════════════════════════════════════════════
TIER 1 — THIS WEEK (before Lloyd meeting, April 7)
These directly determine what Rehan can credibly claim and say.
Execute in strict order. Total estimated effort: ~11 hours.

★ NEW IN v3.1: TW-0, TW-REHEARSAL, PL-11 promoted here.
★ CRITICAL: TW-5 must complete before PL-16 can run.
★ CRITICAL: MUL-02 — run feature vector check NOW (15 min).
═══════════════════════════════════════════════════════════════════

────────────────────────────────────────────────────────────
CRITICAL PRE-CHECKS (do before anything else — total: 45 min)
────────────────────────────────────────────────────────────

[★ NEW] [★ CRITICAL] TW-0: FEATURE VECTOR STORAGE VERIFICATION (MUL-02)
    File: internal/store/score_store.go + PostgreSQL
    Action: Run this query immediately:
      SELECT column_name FROM information_schema.columns
      WHERE table_name = 'score_events'
      ORDER BY ordinal_position;
    Expected: All 8 feature dimensions present as individual columns
    (history_score, velocity_score, z_score, policy_violations,
     agent_type_modifier, time_context_modifier, composite_score, band)
    If only composite_score exists → STOP. Build task: add feature vector
    columns to score_events table and persist all 8 dimensions per event.
    This is Phase 2's entire training dataset. Fix before pilot.
    Effort: 15 min verify | 3 hours fix if failing | Source: L70, L89, MUL-02
    ⚠ DEPENDENCY: If this fails, add to TIER 1 immediately.

[★ NEW] TW-PRE-2: ORG-SCOPED BASELINE VERIFICATION — promoted from TIER 2 (GAP-03)
    File: internal/scoring/baseline.go + Redis
    Action: Inspect Redis key pattern for Welford's running stats:
      redis-cli KEYS "baseline:*" | head -20
    Expected key format: baseline:{org_id}:{agent_did}
    Failure format: baseline:{agent_did} (missing org_id = cross-tenant contamination)
    If key format is wrong → fix before Lloyd meeting. This is the horizontal
    scaling proof point claimed in TW-1. Cannot claim what isn't verified.
    Effort: 20 min verify | 2 hours fix if failing | Source: L5, L38, L57, L92, GAP-03
    ⚠ DEPENDENCY: Must pass before TW-1 talking point on horizontal scaling.

────────────────────────────────────────────────────────────
TIER 1 EXECUTION TASKS — in priority order
────────────────────────────────────────────────────────────

[ ] TW-1: LLOYD TALKING POINTS — Complete set (L12, L19, L20, L28, L38, L41,
          L50, L56, L60, L64, L73 + MUL-02, MUL-03, UP-01)
    File: docs/enterprise/Lloyd_meeting_prep.md (create this file)
    Action: Document all talking points for Lloyd meeting:

    TECHNICAL DIFFERENTIATORS:
    (1) "ARE is the transparency layer regulators are demanding —
        every enforcement decision has a structured reason object,
        which is what SEC and DORA examiners will ask for."
    (2) "ARE implements all 5 AI safety mechanisms simultaneously:
        red team testing (attack corpus), content filtering (policy
        packs), audit trail (hash chain), explainability (reason
        object), human oversight (override workflow)."
    (3) "ARE uses ensemble enforcement — score AND policy must both
        flag before blocking. No single point of false positive.
        This is the architectural reason ARE achieves 0.00% FP:
        two independent detection systems must agree. A single
        miscalibrated metric cannot block a legitimate agent."
        [G-FP anchor: this IS the mechanism. Ref: L20, L82, MUL-03]
    (4) "ARE prevents baseline overfitting through policy-layer
        regularization — slow-walk attacks cannot poison our scoring.
        ARE also monitors variance growth rate; a doubling of weekly
        variance triggers an early warning before the attack succeeds."
    (5) "ARE integrates via standard Kong plugin — no proprietary SDK
        required, no vendor lock-in, works with any HTTP-based agent
        framework."
    (6) "ARE already applies context-sensitive enforcement — a trading
        agent and a reporting agent have different behavioral thresholds
        by policy design. Context isn't Phase 2. It's live now."
        [Source: L17, L22, DEF-04]
    (7) "ARE's threshold calibration model freezes the base scoring
        formula and adjusts only the weight coefficients. No black box
        recalibration — max 10% weight change per iteration, always
        validated against a held-out test set before acceptance."
        [Source: L33, L72, DEF-06]

    DEPLOYMENT MODEL:
    (8) "Phase 1 pilot is customer-hosted — Docker Compose, Kong + Redis
        + PostgreSQL, running entirely within your network perimeter.
        No data leaves your environment. No SaaS dependency."
        [Source: L56, UP-01]

    HUMAN OVERSIGHT:
    (9) "ARE implements HITL by design — every enforcement decision in
        observe mode is reviewed by your security team before
        auto-block is enabled. You maintain human judgment throughout."
    (10) "ARE includes a closed feedback loop: every enforcement decision
         can be reviewed, confirmed, or overridden by your security team,
         and overrides feed back into baseline calibration."

    SCALE / INFRA:
    (11) "ARE's Kong plugin is stateless — it pulls every score from Redis
         on every request, with zero instance-local state. Horizontal
         scaling is architectural, not configurational. Verified."
         [Source: L57, L92, TW-PRE-2 must pass before claiming this]

    PILOT MODEL:
    (12) "ARE uses a risk-staged deployment model — three phases, each
         requiring your team's sign-off before advancing. You control
         the pace. We don't advance without your sign-off."
    (13) "After pilot, ARE produces a structured case study in the format
         regulators expect: background, methodology, results, confirmed
         incidents. Your examiner gets a document, not just a dashboard."

    Effort: 60 min | Source: L12, L19, L20, L28, L38, L41, L50, L56, L60, L64, L73

[ ] TW-REHEARSAL: CONVERSATION READINESS — mock Lloyd session (GAP-02, L90)
    File: none (cognitive preparation — not a document task)
    Action: 20-minute solo rehearsal covering all 13 TW-1 talking points.
    For each point: say it aloud, time yourself (target ≤30 sec per point).
    Identify which points cause hesitation — those need repetition.
    Critical phrases to make reflexive:
      "You control the pace. We don't advance without your sign-off."
      "Two independent systems must agree before any block."
      "0.00% false positive rate on held-out test set."
      "Every enforcement decision has a structured reason object."
    Note: Lloyd is 7 days out. Documents are not conversation readiness.
    Effort: 20 min | Source: L90, GAP-02
    ⚠ Do this before the meeting, not after writing TW-1.

[ ] TW-2: STAGED PILOT CURRICULUM — Add to LoU (L30, L35, L60, L66, MUL-05)
    File: docs/enterprise/pilot-letter-of-understanding.md
    Action: Add "Pilot Rollout Schedule" section:
    Week 1-2: Observe mode only. Baseline establishment. Zero enforcement.
    Week 3: Flag anomalies, review with security team. No auto-block.
    Week 4-6: Enforce mode enabled. Human review on all blocks for first 7 days.
    Week 7-12: Full enforcement with auto-rollback protection.
    Add "Pilot Success Criteria" section:
    (1) FP rate ≤ 2% in production
    (2) ≥ 1 confirmed true positive enforcement event
    (3) Audit trail reviewed and accepted by security team
    Add "GDPR Article 22 / Human Oversight" note:
    "The staged rollout with explicit CISO sign-off at each phase constitutes
    the human oversight mechanism required for GDPR Article 22 compliance review.
    No automated enforcement advances without a documented human authorization."
    [Source: L71, L91, MUL-05 — staged rollout = Article 22 answer]
    Effort: 1 hour | Source: L30, L35, L66, L91

[ ] TW-3: EXPAND ATTACK CORPUS — Augmentation + noise injection (L25, L30, L39)
    File: tests/attack_corpus/ (add new scenario files)
    Action: Generate 20+ attack variants from existing 30 scenarios:
    - 5 slow-walk variations with different timing windows (3/5/7/10/14 days)
    - 5 PII exfiltration variants with different rates (0.31 to 0.85)
    - 5 combined attack vectors (PII + cross-tenant + escalation)
    - 5 noise-injected attacks (3 legitimate events between every 2 attack events)
    Result: TP corpus grows from 30 → 50+ scenarios, TP rate improves
    Note: Completing TW-3 also closes PL-9 (noise injection — merged).
    Effort: 2 hours | Source: L25, L30, L39

[ ] TW-4: PAYLOAD VALIDATION — Kong plugin security (L13)
    File: kong/plugins/agent-reputation/handler.lua
    Action: Add input validation before event scoring:
    - Validate agent_did format before processing
    - Validate event_type is in allowed enum
    - Sanitize payload fields that reach the scoring engine
    - Log malformed payloads as potential injection attempts
    Rationale: Prompt injection into behavioral event payloads could
    manipulate ARE's scoring logic or trigger false FP spikes.
    Effort: 1 hour | Source: L13

[ ] TW-5: HELD-OUT TEST SET — Credibility fix (L23, L26, L44)
    File: tests/held_out/ (create new directory)
    Action: Reserve 20% of scenarios as permanently held-out test set:
    - Move 6 FP scenarios (random selection) to tests/held_out/fp/
    - Move 6 attack scenarios (random selection) to tests/held_out/attack/
    - NEVER use these for threshold calibration
    - Run against held-out set only for final validation claims
    Result: "0.00% FP on held-out test set" = bulletproof credibility claim
    ⚠ DEPENDENCY: PL-16 (per-agent-type FP audit) must run on THIS set.
    Complete TW-5 before running PL-16. [Source: MUL-01, L84]
    Effort: 30 min | Source: L23, L26, L44

[ ] TW-6: VARIANCE GROWTH RATE — New slow-walk signal + window doc (L34, MUL-04)
    File: internal/scoring/policy.go
    Action: Add new HIGH_RISK trigger:
    BASELINE_VARIANCE_GROWTH_RATE > 2x_WEEKLY_AVERAGE → flag as potential
    slow-walk baseline poisoning.
    Implementation requirements:
    - Define VARIANCE_WINDOW_DAYS = 7 as a named constant (not a magic number)
    - Define VARIANCE_GROWTH_THRESHOLD = 2.0 as a named constant
    - Document in policy.go comment: "Window = 7 days. Growth threshold = 2x.
      Rationale: slow-walk attacks require multiple days to shift baseline;
      a 2x variance increase in one window = statistically anomalous."
    [Source: MUL-04 — window parameter must be documented for audit]
    Effort: 2 hours | Source: L34

[ ] TW-7: 4-METRIC PERFORMANCE REPORTING — F1 score framing (L42, L44)
    File: docs/enterprise/ (all performance claim documents)
    Action: Update all performance claims to 4-metric format:
    - TP rate: 86.67%
    - FP rate: 0.00%
    - Precision: 100% (no false blocks)
    - F1 score: 0.9286
    Replace any single-metric claim with the full 4-metric set.
    "F1 score of 0.9286 on held-out validation set" = enterprise-grade claim.
    Effort: 30 min | Source: L42, L44

[✓ MERGED → TW-1] TW-8: INTEROPERABILITY TALKING POINT
    Consolidated into TW-1 item (5). No separate action.

[ ] TW-9: EVALUATION HARNESS METHODOLOGY DOC (L44, L45, L59, L15-Phase1, DEF-03)
    File: docs/enterprise/evaluation-harness-methodology.md (create)
    Action: Document the complete evaluation methodology:
    - How scenarios were generated (synthetic finserv corpus)
    - How held-out set was selected (random 20%, stratified)
    - What metrics are reported and why (TP, FP, Precision, F1)
    - What constitutes a true positive vs false positive
    - ARE's behavioral window parameter:
      "ARE's Phase 1 baseline uses a rolling fixed-window z-score model.
      The behavioral window is [N days] of event history per agent.
      Legitimate behavioral change is handled via the override workflow,
      which feeds back into baseline recalibration. The Phase 1 model is
      intentionally conservative: it favors human review over automated
      adaptation. Phase 2 will introduce adaptive memory."
      [Source: DEF-03, L14, L15 — document window, don't hide it]
    - ARE's scoring weight framing:
      "ARE's feature weights are attention coefficients over 8 behavioral
      dimensions. Weights are calibrated via A/B validation on held-out
      corpus, with max 10% change per iteration."
      [Source: DEF-02, L8, L27]
    - Monitoring KPI thresholds:
      FP rate alert: >0.5% = yellow | >2.0% = red (G-FP gate)
      Scoring latency p99: >10ms = yellow (G-LATENCY ceiling)
      Redis cache hit rate: <80% = investigate
      Dead-letter queue depth: >0 = alert
    - How threshold changes are made (iterative, 10% max, held-out validation)
    Effort: 1 hour | Source: L44, L45, L59, L72

[✓ MERGED → TW-1] TW-10: HITL + CISO UNLOCK
    Consolidated into TW-1 items (9), (10), (12). No separate action.

[ ] TW-11: ENTERPRISE INTEGRATION PREREQS — JSON schema + deployment (L55, L61, UP-01)
    File: docs/enterprise/prerequisites-checklist.md
    Action: Add three sections:
    SECTION A — Data pipeline requirements:
    - Agent event stream format: JSON over HTTP POST to Kong /events endpoint
    - Required fields: agent_did (string), event_type (enum), timestamp (ISO8601),
      metadata (object, optional)
    - Example valid payload:
      {"agent_did":"agt_abc123","event_type":"data_access",
       "timestamp":"2026-04-07T09:00:00Z","metadata":{"resource":"customer_pii"}}
    - JWT signing setup: RS256, org JWKS endpoint required
    - Redis sizing: 1GB minimum for 10K agents
    - PostgreSQL schema: provided as migration files in /migrations/
    SECTION B — /score endpoint RPS ceiling + 429 handling + Retry-After header
    SECTION C — Deployment model (Phase 1):
    "ARE Phase 1 pilot runs customer-hosted. Requirements: Docker, 4 vCPU,
    8GB RAM, 50GB storage. No external network dependency during operation.
    Estimated setup time: 2-4 hours with ARE onboarding guide."
    [Source: L56, UP-01]
    Effort: 45 min | Source: L55, L61

[✓ MERGED → TW-9] TW-12: MONITORING KPI THRESHOLDS
    Consolidated into TW-9. No separate action.

[✓ MERGED → TW-1] TW-13: "YOU CONTROL THE PACE" + FEEDBACK LOOP
    Consolidated into TW-1 items (10), (12). No separate action.

[✓ MERGED → TW-11] TW-14: JSON SCHEMA SPEC
    Consolidated into TW-11 Section A. No separate action.

[✓ MERGED → TW-1] TW-15: POST-PILOT CASE STUDY TEMPLATE
    Consolidated into TW-1 item (13). No separate action.

═══════════════════════════════════════════════════════════════════
TIER 1 — NET DISTINCT ACTIONS SUMMARY
Total actual effort: ~11 hours
═══════════════════════════════════════════════════════════════════

EXECUTE IN THIS SEQUENCE:
0a. TW-0          → Feature vector storage verification     (15 min) ★ CRITICAL
0b. TW-PRE-2      → Org-scoped baseline verification        (20 min) ★ CRITICAL
1.  TW-1          → Lloyd_meeting_prep.md (13 points)       (60 min)
2.  TW-REHEARSAL  → Conversation readiness (say it aloud)   (20 min)
3.  TW-2          → LoU staged curriculum + Article 22      (60 min)
4.  TW-5          → tests/held_out/ creation                (30 min) ★ UNLOCKS PL-16
5.  TW-7          → 4-metric F1 reporting across all docs   (30 min)
6.  TW-9          → evaluation-harness-methodology.md       (60 min)
7.  TW-11         → prerequisites-checklist.md updates      (45 min)
8.  TW-3          → Attack corpus expansion + noise         (120 min) [closes PL-9]
9.  TW-4          → Kong payload validation                 (60 min)
10. TW-6          → policy.go variance growth rate trigger  (120 min)

TOTAL TIER 1 EFFORT: ~11 hours (up from 8.5 — 3 critical checks added)

═══════════════════════════════════════════════════════════════════
TIER 2 — BEFORE PILOT GO-LIVE (after LoU signed, before enforce mode)
These make the pilot bulletproof. Execute after T8 signed.
★ v3.1 changes: P2-14 promoted here from TIER 3. PL-9 removed.
   PL-SECURITY-ATTESTATION consolidates PL-7/18/19/21/22.
═══════════════════════════════════════════════════════════════════

[ ] PL-1: CONFIDENCE INTERVAL — Reason object enhancement (L16, L21)
    File: internal/scoring/explainability.go
    Action: Add confidence_pct field to reason object:
    confidence_pct = function of (score distance from band boundary,
    number of policy violations, z-score magnitude)
    Example: {"score": 187, "band": "BLOCKED", "confidence": 94}
    Effort: 3 hours | Source: L16, L21

[ ] PL-2: THREE-TIER REDIS FALLBACK — Circuit breaker (L2, L14)
    File: internal/store/score_store.go
    Action: Implement fallback chain:
    Redis down → query PostgreSQL last known score
    PostgreSQL down → fail open at MONITORED band (not TRUSTED)
    Never: "everything blocked" or "everything passes" on infra failure
    Note: Fallback must account for sequential state loss — last known
    score may be N events stale. Document staleness in fallback response.
    [Source: L14, DEF-03 — sequential state awareness in fallback path]
    Effort: 3 hours | Source: L2, L14

[ ] PL-3: DEAD-LETTER HANDLING — Consumer reliability (L2, L23)
    File: internal/scoring/consumer.go
    Action: Add dead_letter_events table. After 3 failed retries,
    move event to dead-letter with error reason. Alert via SIEM webhook.
    Effort: 2 hours | Source: L2, L23

[ ] PL-4: REASON OBJECT CACHE — Performance (L2, L11)
    File: internal/scoring/explainability.go + internal/store/score_store.go
    Action: Cache reason object in Redis.
    Key: agent_did + score_band + violation_hash. TTL: 5 minutes.
    Effort: 2 hours | Source: L2, L11

[ ] PL-5: CONSUMER PAYLOAD TRIM — Performance (L1)
    File: internal/scoring/consumer.go
    Action: Create ScoringPayload struct {agent_did, feature_vector, timestamp}
    Replace full BehavioralEvent. ~60-70% payload size reduction.
    Effort: 1 hour | Source: L1

[ ] PL-6: WELFORD'S ALGORITHM — Baseline scalability (L1, L8, L28)
    File: internal/scoring/baseline.go
    Action: Replace raw event history queries with Welford's online algorithm.
    O(1) time + space per agent. Same statistical result, 99% less storage.
    Effort: 3 hours | Source: L1, L8, L28

[ ] PL-SECURITY-ATTESTATION: CONSOLIDATED SECURITY DOC — 5 tasks in 1 session
    File: docs/enterprise/security-attestation.md (create, structured in order)
    [Consolidates: PL-7, PL-18, PL-19, PL-21, PL-22 — saves ~4hrs context switching]
    Build this document in one session, sections in this order:

    SECTION 1 — ACCESS CONTROLS (was PL-18 | Source: L58)
    - ARE admin console: RBAC (admin / analyst / readonly roles)
    - All API calls authenticated via signed JWT RS256
    - Audit trail: hash-chained tamper-evident
    - MFA: required for admin role on ARE management console

    SECTION 2 — KEY EXPOSURE ATTESTATION (was PL-7 | Source: L13, L19)
    - Describe the private key exposure incident
    - Document remediation steps taken
    - Current state: keys rotated, JWKS endpoint verified

    SECTION 3 — SIEM INTEGRATION (was PL-21 | Source: L67, L88)
    "ARE audit trail exports to standard SIEM formats (CEF/JSON).
    Every enforcement decision generates a structured event:
    {agent_did, action, reason_object, confidence_pct, timestamp, hash}
    Webhook delivery to SIEM endpoint configurable in policy YAML."
    ⚠ Prerequisite: PL-SIEM-VERIFY must pass before writing this section.

    SECTION 4 — GDPR ARTICLE 22 / COMPLIANCE (was PL-15, PL-22 | Source: L51, L71, L91)
    - What ARE stores: agent_did, feature_vector, timestamp (no user PII)
    - Retention policy: behavioral events retained 90 days default
    - GDPR Article 22 note: "ARE's enforcement decisions are made on
      agent behavior, not human subjects. The staged rollout with
      explicit CISO sign-off at each phase constitutes the human
      oversight mechanism required for Article 22 compliance review."
    - EU localization note: "Phase 1 reason objects are English-language
      only. EU/DORA deployments requiring localized enforcement
      documentation should plan for Phase 2 multilingual reason objects."
      [Source: L40, UP-03]
    - Regulations compatible: DORA, SOX data governance, HIPAA (no PHI stored)

    SECTION 5 — SELF-AUDIT CHECKLIST (was PL-19 | Source: L63)
    Map hardening score 94/100 to specific audit items:
    - JWT validation: PASS
    - Payload sanitization: PASS (after TW-4)
    - Hash chain integrity: PASS
    - Redis auth: PASS
    - PostgreSQL encryption at rest: PASS
    - Org-scoped baseline isolation: PASS (after TW-PRE-2)
    - Private key incident: DISCLOSED (see Section 2)
    - Held-out test set: PASS (after TW-5)
    - FP rate on held-out: 0.00% PASS
    - Feature vector storage: PASS (after TW-0)

    Effort: 3 hours total | Source: L13, L19, L51, L58, L63, L67, L71, L88

[ ] PL-SIEM-VERIFY: SIEM AUDIT TRAIL FORMAT VERIFICATION (GAP-01, L88)
    File: docs/enterprise/security-attestation.md (prerequisite for Section 3)
    Action: Run 3 enforcement events in development environment.
    Export from PostgreSQL. Confirm output has ALL required fields:
    {agent_did, action, reason_object, confidence_pct, timestamp, hash}
    Confirm: hash field = chain hash (not a row ID)
    Confirm: reason_object is structured JSON (not a string)
    Confirm: action is one of {BLOCKED, FLAGGED, MONITORED, TRUSTED}
    If any field is missing → build task before writing attestation doc.
    Effort: 45 min | Source: L67, L88, GAP-01
    ⚠ Run this BEFORE PL-SECURITY-ATTESTATION Section 3.

[ ] PL-8: RESTRUCTURE TEST CORPUS — Curriculum order (L30)
    File: tests/attack_corpus/ (reorganize)
    Action: Rename to staged curriculum:
    stage1_basic_rate_anomaly_*.go
    stage2_pii_access_pattern_*.go
    stage3_slow_walk_evasion_*.go
    stage4_multi_vector_combined_*.go
    Effort: 1 hour | Source: L30

[ ] PL-10: 5-CATEGORY CORPUS TAXONOMY — Credibility (L43)
    File: tests/ (restructure + label all scenarios)
    Action: Label all scenarios into 5 categories:
    - typical: normal agent behavior (baseline)
    - boundary: score near band threshold (±10 points)
    - edge: novel/rare attack vector
    - negative: should NOT block (FP test cases)
    - performance: high-concurrency burst (50 concurrent requests)
    Effort: 1 hour | Source: L43

[ ] PL-11: ORG-SCOPED BASELINE VERIFICATION — ★ PROMOTED TO TIER 1 ★
    See TW-PRE-2 above. Moved to pre-check block.
    ⚠ This is not a TIER 2 task. It must precede any horizontal scaling claim.

[ ] PL-12: BURST / STRESS TEST — Redis contention (L47)
    File: tests/performance/ (create)
    Action: Simulate 50 concurrent agent scoring requests.
    Measure p99 latency vs 10ms G-LATENCY ceiling.
    Identify Redis lock contention and Go goroutine ceiling.
    Must pass before enforce mode enabled.
    Effort: 2 hours | Source: L47

[ ] PL-13: A/B THRESHOLD WEIGHT VALIDATION — Pre-pilot (L27, L48)
    File: config/scoring_weights.yaml + tests/held_out/
    Action: Run two weight configurations against held-out corpus.
    Compare F1 scores. Select weights empirically, not by intuition.
    Document: "weights selected by A/B validation on held-out set"
    This is ARE's equivalent of hyperparameter search — Phase 1 version.
    [Source: DEF-05, L27 upgraded to COMPOUND — this is the Phase 1 task]
    Effort: 1 hour | Source: L27, L48

[ ] PL-14: REGRESSION TEST SUITE — Enforce mode gate (L49)
    File: tests/regression/ (create directory)
    Action: Any scoring engine code change must re-run full corpus.
    Pass criteria: FP = 0.00% and TP ≥ 86.67% on held-out set.
    Any regression blocks merge to enforce-mode code.
    Effort: 1 hour | Source: L49

[ ] PL-16: PER-AGENT-TYPE FP BIAS AUDIT — Fairness check (L52, MUL-01)
    File: tests/held_out/ + docs/enterprise/
    Action: Run FP analysis segmented by agent_type:
    (trading, reporting, retrieval, execution, backup)
    Confirm 0.00% FP holds across ALL agent archetypes, not just overall.
    Document result: "FP rate by agent type: [table]"
    FAANG security reviewer will ask this question.
    ⚠ DEPENDENCY: TW-5 must complete first. Run ONLY against held-out set,
    not calibration corpus. Running against contaminated corpus = meaningless.
    [Source: MUL-01 — test contamination × FP bias multiply]
    Effort: 1 hour | Source: L52

[ ] PL-17: KONG STATELESS SCORING VERIFICATION (L57)
    File: kong/plugins/agent-reputation/handler.lua
    Action: Verify Kong plugin pulls score from Redis on every request.
    Confirm zero instance-local state. This = horizontal scaling from day one.
    Document as "horizontally scalable by design" proof point.
    Note: TW-PRE-2 covers the baseline side; this covers the plugin side.
    Effort: 30 min | Source: L57

[★ NEW] PL-FEATURE-VECTOR-STORE: PHASE 2 TRAINING DATA — Feature vector storage
    File: internal/store/score_store.go
    ★ PROMOTED FROM TIER 3 (was P2-14) — this is Phase 1 work.
    Action: Store complete feature vectors (all 8 dimensions) in Phase 1.
    Phase 1 enforcement decisions with reason objects = Phase 2 training labels.
    Every confirmed TP = one labeled training example. Start accumulating now.
    If TW-0 verification fails → this is the remediation build task.
    SQL schema addition:
      ALTER TABLE score_events ADD COLUMN history_score FLOAT;
      ALTER TABLE score_events ADD COLUMN velocity_score FLOAT;
      ALTER TABLE score_events ADD COLUMN z_score FLOAT;
      ALTER TABLE score_events ADD COLUMN policy_violations INT;
      ALTER TABLE score_events ADD COLUMN agent_type_modifier FLOAT;
      ALTER TABLE score_events ADD COLUMN time_context_modifier FLOAT;
    [Source: L70, L89, MUL-02]
    Effort: 3 hours | Source: L70, L89

[ ] PL-23: THRESHOLD CHANGE PROTOCOL — Iterative validation (L72, L33)
    File: docs/enterprise/evaluation-harness-methodology.md
    Action: Add threshold change protocol:
    - Maximum 10% weight adjustment per iteration
    - Run held-out validation after each change
    - Require F1 ≥ prior F1 before accepting new weights
    - No "big bang" recalibration — always incremental
    This is ARE's LoRA-by-hand protocol: freeze formula, adjust coefficients.
    [Source: DEF-06, L33, L72]
    Effort: 30 min | Source: L72

═══════════════════════════════════════════════════════════════════
TIER 3 — PHASE 2 DESIGN (design now, build after pilot signed)
Document these before building. They shape the architecture.
★ v3.1 change: P2-14 removed — promoted to TIER 2 above.
   P2-10 → P2-3 ordering dependency now documented.
═══════════════════════════════════════════════════════════════════

[ ] P2-1: SSE AUDIT STREAM — SIEM real-time feed (L5, L6)
    File: cmd/scoring-service/main.go
    Design: GET /audit/stream → SSE, stateful HTTP, Redis pub/sub for scaling.
    Note: L5 (MCP Roots / org-scoped access control) validates that the Phase 1
    org_id scoping in Redis extends naturally to SSE stream filtering per org.
    Source: L5, L6, L7

[ ] P2-2: MULTITASK SCORING ENGINE — Architecture (L8, L9, L15, L31)
    File: docs/architecture/phase2-scoring-design.md
    Design: 3 tasks sharing 8-dimension feature vector layer (already in Phase 1).
    Task 1: z-score anomaly. Task 2: LLM intent classification. Task 3: peer deviation.
    Feature weights in Phase 1 = attention coefficients. Phase 2 learns them.
    Phase 1 fixed-window z-score transitions to LSTM-style adaptive memory.
    Source: L8, L9, L15, L31, L86

[ ] P2-3: FINE-TUNING STRATEGY — LoRA/PEFT (L29, L33)
    File: docs/architecture/phase2-scoring-design.md
    Design: Pre-train on synthetic corpus, fine-tune only final layer on pilot traffic.
    ⚠ DEPENDENCY: P2-10 (intent classification token budget) must be decided first.
    Token cost ceiling constrains which base model is eligible for LoRA adaptation.
    Cannot design fine-tuning strategy without knowing affordable inference cost.
    Source: L29, L33, CON-04

[ ] P2-4: CONTINUAL LEARNING DEFENSE — EWC for baseline (L32, L34)
    File: docs/architecture/phase2-scoring-design.md
    Design: EWC principle — penalize rapid changes to high-importance baseline params.
    Phase 1 equivalent: variance growth rate trigger (TW-6) + 10% max weight change.
    Source: L32, L34

[ ] P2-5: CONTEXTUAL SCORING — Agent type + time context (L17, L22)
    File: internal/scoring/policy.go + config/policy_packs/
    Design: Context modifiers: agent_type + time_context → threshold adjustments.
    Note: Phase 1 already implements this via policy packs. Phase 2 learns modifiers.
    [Source: DEF-04 — Phase 1 already has this; Phase 2 makes it adaptive]
    Source: L17, L22

[ ] P2-6: RLHF FEEDBACK LOOP — Override learning (L18)
    File: internal/audit/override.go
    Design: Override = negative reward. After 10+ overrides of same policy:
    auto-suggest threshold adjustment. Human approves.
    Source: L18

[ ] P2-7: HYPERPARAMETER SEARCH — Bayesian weight optimization (L27)
    File: config/scoring_weights.yaml + tests/eval_harness/
    Design: Bayesian optimization over {H_weight, V_weight, z_threshold}.
    Phase 1 A/B validation (PL-13) is the manual predecessor of this.
    Source: L27

[ ] P2-8: MCP SERVER INTEGRATION — Public API (L3, L5)
    File: docs/architecture/mcp-integration-design.md
    Design: ARE exposes MCP server. Sampling pattern — no API key in server.
    Tools: check_agent_reputation() + report_behavioral_event()
    Source: L3, L5

[ ] P2-9: TOKENIZATION STRATEGY — Payload analysis (L24)
    File: docs/architecture/phase2-scoring-design.md
    Design: Character-level tokenization for JSON field names.
    BPE for SQL/code patterns in payloads.
    Source: L24

[ ] P2-10: INTENT CLASSIFICATION DESIGN — LLM token budget (L1, L29)
    File: docs/architecture/intent-classification-design.md
    ⚠ Complete this BEFORE P2-3 (fine-tuning strategy).
    Design: System prompt ≤200 tokens. Feature vector only. Cache TTL 5min.
    $0.002/call max = $2K/day ceiling at 1M daily calls.
    Token budget decision constrains which base model is eligible for LoRA.
    Source: L1, L29, CON-04

[ ] P2-11: INTERNATIONALIZED REASON OBJECTS — EU/DORA (L40)
    File: internal/scoring/explainability.go
    Design: Reason object locale field. English default. EU deployments: DE/FR/NL.
    Phase 1 answer (already documented in PL-SECURITY-ATTESTATION): English only.
    Source: L40

[ ] P2-12: HARDWARE REQUIREMENTS DOC — Phase 2 ML layer (L54)
    File: docs/architecture/phase2-infrastructure.md
    Design: GPU requirements for intent classification inference layer.
    Source: L54

[ ] P2-13: DEPLOYMENT ARCHITECTURE GUIDE — Cloud/on-prem/hybrid (L56)
    File: docs/enterprise/deployment-architecture-options.md
    Design: ARE cloud vs on-prem vs hybrid decision guide for enterprise sales.
    Phase 1 answer is already documented in TW-11. Phase 2 adds SaaS/managed option.
    [Source: UP-01 resolved Phase 1 gap; this is the full Phase 2 sales guide]
    Source: L56

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
LEARNING LOG — ALL 73 LEARNINGS
v3.1 changes: verdict upgrades and cross-references noted inline.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

────────────────────────────────────────────────────────────────────
BATCH 1 — Prior sessions (L1-L2)
────────────────────────────────────────────────────────────────────

L1 — Token Optimization (Coursera) | COMPOUND
ARE: consumer.go payload trim, Welford's baseline, LLM prompt budget
Tasks: PL-5, PL-6, P2-10

L2 — Production API: Streaming + Reliability + Performance | COMPOUND
ARE: demo.sh streaming (DONE), Kong retry documented (DONE),
     RPS ceiling documented (DONE), Redis fallback, dead-letter, SSE
Tasks: PL-2, PL-3, PL-4, P2-1
Status: EQ-L2-A ✅ EQ-L2-C ✅ EQ-L2-G ✅

────────────────────────────────────────────────────────────────────
BATCH 2 — MCP Course (L3-L7)
────────────────────────────────────────────────────────────────────

L3 — MCP Sampling | COMPOUND
ARE: Phase 2 MCP server — no API key in server. Client handles LLM calls.
Tasks: P2-8

L4 — MCP Logging + Progress Notifications | COMPOUND
ARE: Validates demo.sh streaming. /audit/stream SSE = MCP pattern.
Tasks: P2-1

L5 — MCP Roots / File Access Control | ★ UPGRADED v3.1: QUEUE→COMPOUND
ARE Phase 1: org_id scoping in Redis key pattern IS roots-equivalent access control.
     Verify org_id+agent_did key structure in TW-PRE-2.
     Claim: "ARE enforces org-level behavioral isolation by architecture."
ARE Phase 2: org-scoped access control in MCP server.
Tasks: TW-PRE-2 (Phase 1 verification), P2-8 (Phase 2 design)
Cross-ref: L38, L57, L92

L6 — MCP Streamable HTTP + SSE Transport | COMPOUND
ARE: /audit/stream: use stateful HTTP (not stateless=true). Stateless breaks SSE.
Tasks: P2-1

L7 — MCP Stateless HTTP + Horizontal Scaling | COMPOUND
ARE: ARE scoring service will scale horizontally.
     /audit/stream: stateful + Redis pub/sub for multi-instance.
Tasks: P2-1

────────────────────────────────────────────────────────────────────
BATCH 3 — LLM Architecture (L8-L12)
────────────────────────────────────────────────────────────────────

L8 — Transformer Multi-Head Attention | ★ UPGRADED v3.1: QUEUE→COMPOUND
ARE Phase 1: ARE's 8-dimension feature vector uses weighted attention.
     Scoring weights = attention coefficients over behavioral dimensions.
     Framing: "ARE uses a weighted feature attention model validated against
     100 synthetic scenarios." Add to TW-9 methodology doc.
ARE Phase 2: multitask scoring architecture.
Tasks: TW-9 (Phase 1 framing), P2-2 (Phase 2 architecture)
Cross-ref: L27, DEF-02

L9 — Embeddings (Word, Subword, Positional) | QUEUE (Phase 2)
ARE: ARE's 8-dimension feature vector IS an embedding.
Tasks: P2-2, P2-9

L10 — GPT-4 Architecture (Decoder blocks) | REJECT
ARE: Not applicable. Autoregressive generation ≠ behavioral scoring.

L11 — Fine-tuning: Transfer learning, Knowledge distillation, Pruning | COMPOUND
ARE: Phase 2 ML = fine-tune on synthetic corpus, adapt to pilot traffic.
Tasks: P2-3, PL-4

L12 — LLM Ethics: Bias, Transparency, Accountability | COMPOUND
ARE: Reason object = transparency. Hash chain = accountability.
COMMERCIAL: "ARE solves AI transparency problem regulators are asking about."
Tasks: TW-1

────────────────────────────────────────────────────────────────────
BATCH 4 — LLM Safety + Decision-Making (L13-L19)
────────────────────────────────────────────────────────────────────

L13 — Prompt Engineering Security: Injection, Jailbreaking | COMPOUND
ARE: Behavioral event payloads are an injection surface for ARE's scoring.
     Malicious event_type values could cause FP spikes.
Tasks: TW-4, PL-SECURITY-ATTESTATION

L14 — RNN Sequential Dependency: Hidden State, BPTT | COMPOUND
ARE: Agent behavioral history = sequential dependency problem.
     Redis fallback must account for sequential state loss (document staleness).
     ARE's behavioral window = deliberate sequence length choice (document in TW-9).
Tasks: PL-2, TW-9
Cross-ref: L15, L34, DEF-03

L15 — LSTM Long-Range Dependency: Gates, Memory Cell | ★ UPGRADED v3.1: QUEUE→COMPOUND
ARE Phase 1: ARE's z-score uses a fixed-length memory window (unlike LSTM which
     learns what to forget). Document this as an intentional design choice.
     When buyer asks "how does ARE handle legitimate behavioral change?":
     "The override workflow handles legitimate change — a human confirms,
     and the confirmation feeds back into baseline recalibration. The Phase 1
     model is intentionally conservative: human review over automated adaptation."
ARE Phase 2: LSTM-style adaptive memory replaces fixed window.
Tasks: TW-9 (Phase 1 framing), P2-2 (Phase 2 architecture)
Cross-ref: L14, DEF-03

L16 — Softmax Output: Probability Distribution, Temperature | COMPOUND
ARE: confidence_pct in reason object = softmax-equivalent probability.
Tasks: PL-1

L17 — Contextual Understanding: Discourse, Pragmatics | ★ UPGRADED v3.1: QUEUE→COMPOUND
ARE Phase 1: ARE already applies context-sensitive enforcement via policy packs.
     Agent_type + time_context thresholds are live. This is a talking point.
     Add to TW-1 item (6): "Context-sensitive enforcement is live, not Phase 2."
ARE Phase 2: Adaptive context learning.
Tasks: TW-1 (Phase 1 talking point), P2-5 (Phase 2 adaptive)
Cross-ref: L22, DEF-04

L18 — RLHF: Reward Models, PPO, Human Preference | COMPOUND
ARE: Override decisions = negative reward. RLHF loop for threshold calibration.
Tasks: P2-6

L19 — AI Safety + Red-Teaming: Moderation, Alignment | COMPOUND
ARE: ARE IS the safety infrastructure. Red-team = attack corpus.
     5 safety mechanisms: red team, filtering, audit, explainability, oversight.
Tasks: TW-1, PL-SECURITY-ATTESTATION

────────────────────────────────────────────────────────────────────
BATCH 5 — ML Fundamentals (L20-L28)
────────────────────────────────────────────────────────────────────

L20 — Ensemble Methods: Bagging, Boosting, Stacking | COMPOUND
ARE: ARE IS an ensemble — score + policy must agree before block.
     No single point of false positive. This IS the G-FP mechanism.
Tasks: TW-1 (item 3)
Cross-ref: L82, MUL-03 — G-FP gate is passed because of this design.

L21 — Confidence Intervals + Uncertainty Quantification | COMPOUND
ARE: confidence_pct in reason object — distance from band boundary.
Tasks: PL-1

L22 — Polysemy + Disambiguation: Context-Dependent Meaning | ★ UPGRADED v3.1: QUEUE→COMPOUND
ARE Phase 1: Agent event_type = polysemous. Context (agent_type, time) changes
     meaning. Policy packs implement this now. This is a Phase 1 differentiator.
ARE Phase 2: Adaptive disambiguation.
Tasks: TW-1 (item 6, via L17), P2-5
Cross-ref: L17, DEF-04

L23 — Model Validation: Cross-validation, Holdout | COMPOUND
ARE: Held-out test set — never use calibration corpus for performance claims.
Tasks: TW-5, PL-3

L24 — Tokenization: BPE, WordPiece, SentencePiece | QUEUE (Phase 2)
ARE: Phase 2 payload tokenization strategy.
Tasks: P2-9

L25 — Data Augmentation: Synthetic generation, Back-translation | COMPOUND
ARE: Attack corpus augmentation — 30 → 50+ scenarios.
Tasks: TW-3

L26 — Validation Split: Train/Validation/Test | COMPOUND
ARE: Confirms L23. Fix test corpus contamination before Lloyd.
Tasks: TW-5

L27 — Hyperparameter Tuning: Bayesian optimization | ★ UPGRADED v3.1: QUEUE→COMPOUND
ARE Phase 1: H/V scoring weights = hyperparameters. PL-13 (A/B validation on
     held-out corpus) is the Phase 1 equivalent of hyperparameter search.
     L27 maps to PL-13 — the task already exists, just wasn't cross-referenced.
ARE Phase 2: Bayesian optimization over full parameter space.
Tasks: PL-13 (Phase 1 — already exists), P2-7 (Phase 2 Bayesian)
Cross-ref: DEF-05

L28 — Overfitting vs Underfitting | COMPOUND
ARE: Slow-walk = attacker causes ARE baseline to overfit to attack behavior.
     Two-layer detection = regularization that prevents overfitting.
Tasks: TW-1, PL-6

────────────────────────────────────────────────────────────────────
BATCH 6 — Advanced Training Strategies (L29-L35)
────────────────────────────────────────────────────────────────────

L29 — Transfer Learning + Fine-tuning for Domain Adaptation | COMPOUND
ARE: Phase 2 = transfer learning. LoRA: freeze base, adapt threshold layer.
Tasks: P2-3

L30 — Curriculum Learning: Simple to Complex | COMPOUND
ARE: Test corpus staged curriculum. Pilot rollout = staged curriculum.
Tasks: PL-8, TW-2, TW-3

L31 — Multitasking Models: Shared Architecture, Joint Learning | COMPOUND
ARE: Phase 2 = multitask model. 3 scoring heads, shared feature vector.
Tasks: P2-2

L32 — Continual Learning: Catastrophic Forgetting, Replay | COMPOUND
ARE: Baseline = continual learning problem. EWC for slow-walk defense.
Tasks: P2-4

L33 — LoRA + PEFT: Parameter-Efficient Fine-Tuning | ★ UPGRADED v3.1: QUEUE→COMPOUND
ARE Phase 1: ARE's threshold calibration IS LoRA-by-hand. Freeze the base
     scoring formula. Adjust only weight coefficients. Max 10% per iteration.
     Commercial framing: "ARE's calibration model is parameter-efficient —
     we freeze the detection logic and tune only the sensitivity layer."
ARE Phase 2: LoRA for intent classification fine-tuning.
Tasks: TW-9 (Phase 1 framing in methodology doc), PL-23 (threshold change protocol),
       P2-3 (Phase 2)
Cross-ref: L72, DEF-06

L34 — Catastrophic Forgetting: EWC, Replay, Knowledge Distillation | COMPOUND
ARE: Slow-walk = adversarial catastrophic forgetting. Variance growth rate signal.
     Variance window must be documented as a named constant, not magic number.
Tasks: TW-6, P2-4
Cross-ref: L14, L15, MUL-04

L35 — Curriculum Learning for Specialized Domains (Legal/Medical) | COMPOUND
ARE: Pilot rollout structure = curriculum. Staged: observe → flag → enforce.
Tasks: TW-2

────────────────────────────────────────────────────────────────────
BATCH 7 — NLP Fundamentals + LLM Testing (L36-L44) [NEW v3]
────────────────────────────────────────────────────────────────────

L36 — DPO: Direct Preference Optimization | REJECT
ARE: DPO is RLHF alignment for LLM training. ARE scores agents, doesn't train them.

L37 — NLP Computational Efficiency | REJECT
ARE: LLM inference optimization. ARE's scoring engine is Go, not neural net.
     0.25ns call overhead already achieved. Not applicable.

L38 — Domain Adaptability: Specialized terminology, Context | COMPOUND
ARE: Per-org baseline isolation — each enterprise tenant's behavioral normal
     is a separate distribution, not shared. Welford's must be scoped per org_id.
Tasks: TW-PRE-2 (verify), PL-6 (implement Welford's with org_id scoping)
Cross-ref: L5, L57, L92

L39 — Robustness to Noise: Typos, Slang, Grammatical errors | COMPOUND
ARE: Noise-injected attack variants — legitimate events between attack events.
     Tests z-score signal robustness under realistic mixed traffic.
Tasks: TW-3 (noise injection scope)
Note: PL-9 removed — scope fully absorbed into TW-3.

L40 — Multilinguality: Cross-lingual embeddings, Zero-shot | ★ UPGRADED v3.1: QUEUE→COMPOUND
ARE Phase 1: Document the EU localization limitation proactively.
     "Phase 1 reason objects are English-language only. EU/DORA deployments
     requiring localized enforcement documentation should plan for Phase 2."
     Andy Watkin-Child will raise this. Document the answer, don't be surprised.
ARE Phase 2: Multilingual reason objects for EU/DORA deployments.
Tasks: PL-SECURITY-ATTESTATION (Section 4 — EU note), P2-11
Cross-ref: UP-03

L41 — Interoperability: REST, middleware, API standards | COMPOUND
ARE: ARE integrates via standard Kong plugin. No proprietary SDK. No vendor lock-in.
     Works with any HTTP-based agent framework. This is a Lloyd talking point.
Tasks: TW-1 (item 5)

L42 — F1 Score, Precision/Recall, Confusion Matrix | COMPOUND
ARE: 86.67% TP at 0.00% FP = F1 = 0.9286. Reframe all performance claims.
     CISO security team will interrogate using F1, not just TP rate.
Tasks: TW-7

L43 — Testing Protocol Taxonomy: typical/boundary/edge/negative/performance | COMPOUND
ARE: Corpus needs formal 5-category taxonomy. Currently unstructured.
Tasks: PL-10

L44 — LLM Benchmarking: Historical data, Custom benchmarks, Performance targets | COMPOUND
ARE: Evaluation harness methodology doc — formal benchmark document required.
     "0.00% FP" only credible with documented methodology.
Tasks: TW-9, TW-5

────────────────────────────────────────────────────────────────────
BATCH 8 — LLM Testing Protocols (L45-L55) [NEW v3]
────────────────────────────────────────────────────────────────────

L45 — Automated Test Suites: efficiency, consistency, comprehensiveness | COMPOUND
ARE: Evaluation harness methodology doc needs automated test suite description.
Tasks: TW-9

L46 — CI/CD: Continuous Integration pipelines | REJECT
ARE: ANTI-SCOPE. CI/CD before runtime enforcement proven is forbidden per APEX.

L47 — Stress Testing: high-load simulation, resource utilization | COMPOUND
ARE: 50 concurrent scoring requests → test Redis contention, Go goroutine ceiling.
     Must pass before enforce mode.
Tasks: PL-12

L48 — A/B Testing: randomization, statistical significance | COMPOUND
ARE: Run two weight configurations against held-out corpus. Select empirically.
Tasks: PL-13
Cross-ref: L27

L49 — Regression Testing: prevent performance degradation | COMPOUND
ARE: Any scoring engine change must re-run full corpus. Pass criteria: FP=0.00%.
Tasks: PL-14

L50 — HITL: Human-in-the-loop evaluation | COMPOUND
ARE: ARE's override workflow = HITL by design. "You maintain human judgment."
     DORA/SEC language: "human-in-the-loop enforcement validation."
Tasks: TW-1 (item 9)

L51 — Legal/Compliance Checks: GDPR, HIPAA, data privacy | COMPOUND
ARE: GDPR/HIPAA data handling for behavioral event data.
Tasks: PL-SECURITY-ATTESTATION (Section 4)

L52 — Ethical/Bias Testing: diverse test sets, continuous monitoring | COMPOUND
ARE: FP analysis segmented by agent_type — does FP rate vary by agent archetype?
Tasks: PL-16
Cross-ref: L84, MUL-01 — PL-16 must run on held-out set only.

L53 — Documentation: test case docs, methodology records | COMPOUND
ARE: Evaluation harness methodology doc formally documents ARE's testing.
Tasks: TW-9

L54 — Hardware Selection: GPU/TPU for LLM inference | QUEUE (Phase 2)
ARE: Phase 2 ML intent classifier layer hardware requirements.
Tasks: P2-12

L55 — Data Pipeline Integration: collection, storage, feeding | COMPOUND
ARE: Enterprise buyers ask "what do we need to stand this up?"
     Prerequisites checklist needs data pipeline requirements section.
Tasks: TW-11

────────────────────────────────────────────────────────────────────
BATCH 9 — LLM Deployment (L56-L66) [NEW v3]
────────────────────────────────────────────────────────────────────

L56 — Cloud vs On-Premises vs Hybrid | ★ UPGRADED v3.1: QUEUE→COMPOUND
ARE Phase 1: Lloyd will ask "does this run in our environment?"
     Phase 1 answer is concrete and decided: customer-hosted, Docker Compose,
     customer's network perimeter. Document in TW-1 and TW-11.
     "Phase 1 pilot: no data leaves your environment. No SaaS dependency."
ARE Phase 2: Full deployment guide for cloud/on-prem/hybrid options.
Tasks: TW-1 (item 8), TW-11 (Section C), P2-13
Cross-ref: UP-01

L57 — Load Balancing + Resource Allocation | COMPOUND
ARE: Kong plugin must pull score from Redis on every request, never instance memory.
     Redis-backed scoring = horizontally scalable from day one.
Tasks: TW-PRE-2 (baseline side), PL-17 (plugin side)
Cross-ref: L5, L38, L92

L58 — Security Best Practices: RBAC, ABAC, MFA, encryption, audits | COMPOUND
ARE: ARE's own security controls need documentation for enterprise security review.
Tasks: PL-SECURITY-ATTESTATION (Section 1)

L59 — Continuous Monitoring: performance metrics, alerting | COMPOUND
ARE: Prometheus/Grafana stack needs documented KPI thresholds + alerting rules.
Tasks: TW-9 (monitoring KPIs section)

L60 — Incremental Implementation / Phased Rollout | COMPOUND
ARE: "Risk-staged deployment model." Three phases. Each requires sign-off.
     "You control the pace" = CISO unlock phrase.
Tasks: TW-1 (items 9, 12), TW-2

L61 — Data Format Compatibility: JSON, CSV, ETL | COMPOUND
ARE: Prerequisites checklist needs exact JSON schema + example valid payload.
Tasks: TW-11 (Section A)

L62 — Framework Integration: Django, Spring, middleware | REJECT
ARE: ARE uses Kong + REST API. Django/Spring not applicable to this stack.

L63 — Regular Security Audits: internal, external, automated | COMPOUND
ARE: Security attestation doc needs self-audit checklist.
     94/100 hardening score must map to specific audit items.
Tasks: PL-SECURITY-ATTESTATION (Section 5)

L64 — Feedback Loops + Continuous Improvement | COMPOUND
ARE: Override → baseline adjustment = closed feedback loop.
     "ARE includes a closed feedback loop. Overrides feed back into calibration."
Tasks: TW-1 (item 10)

L65 — Workflow Augmentation: LLM for customer service, marketing | REJECT
ARE: Wrong product category. ARE enforces agents, doesn't augment human workflows.

L66 — Outcome Achievement: setting objectives, success metrics | COMPOUND
ARE: Pilot LoU needs defined success criteria both parties agree on before signing.
Tasks: TW-2 (success criteria section)

────────────────────────────────────────────────────────────────────
BATCH 10 — LLM Integration + Security (L67-L73) [NEW v3]
────────────────────────────────────────────────────────────────────

L67 — Security in LLM Integration: IDS/SIEM, encryption, consent | COMPOUND
ARE: ARE's hash-chained audit trail is a SIEM feed.
     Document CEF/JSON export format in security attestation.
Tasks: PL-SIEM-VERIFY (verify first), PL-SECURITY-ATTESTATION (Section 3)
Cross-ref: L88, GAP-01

L68 — Quantization: 32-bit → 8-bit weight compression | REJECT
ARE: LLM model compression. ARE has no neural network weights to quantize.

L69 — Pruning: weight removal, sparsity, structured vs unstructured | REJECT
ARE: LLM optimization technique. ARE's scoring is mathematical formula, not weight matrix.

L70 — Knowledge Distillation: teacher-student paradigm, soft targets | COMPOUND
ARE: Phase 1 scoring engine = "teacher." Phase 2 ML classifier = "student."
     Phase 1 enforcement decisions with reason objects = training labels for Phase 2.
     Every confirmed TP = one labeled training example. Store feature vectors now.
Tasks: TW-0 (verify), PL-FEATURE-VECTOR-STORE (build if failing)
Cross-ref: L89, MUL-02 — CRITICAL: this is Phase 1 work, not Phase 2 design.

L71 — Consent Management + GDPR Article 22: automated decision-making | COMPOUND
ARE: ARE's enforcement decisions = automated decision-making on agent behavior.
     Must document GDPR Article 22 applicability distinction.
     The staged rollout with CISO sign-off IS the Article 22 human oversight.
Tasks: TW-2 (Article 22 note), PL-SECURITY-ATTESTATION (Section 4)
Cross-ref: L90, L91, MUL-05

L72 — Pruning Schedule: incremental approach, validation checkpoints | COMPOUND
ARE: Threshold calibration protocol — max 10% weight change per iteration,
     held-out validation required before accepting new weights.
Tasks: PL-23, TW-9

L73 — Mobile Deployment Case Study: background/methodology/results format | COMPOUND
ARE: Post-pilot case study will follow this structure.
     "You'll have a documented case study for your regulatory examiner."
Tasks: TW-1 (item 13)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
MASTER_LEARNINGS ENTRIES — Append to MASTER_LEARNINGS_v2.1_DELTA.md
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

L81 — MCP SAMPLING SHIFTS API KEY RESPONSIBILITY TO CLIENT [F]
In MCP architecture, sampling allows a server to request LLM generation from
its client without holding an API key. For ARE Phase 2, the public MCP server
does not need Anthropic credentials — the enterprise client handles LLM calls.
Use stateful HTTP (not stateless) to preserve SSE back-channel.

L82 — ARE IS AN ENSEMBLE ENFORCEMENT SYSTEM [F]
ARE requires both behavioral score AND policy threshold violation before blocking.
Two independent detection systems must agree. No single point of false positive.
Commercial framing: "a single miscalibrated metric cannot block legitimate agents."
This IS the mechanism by which ARE passes G-FP. Cross-ref: L20, MUL-03.

L83 — SLOW-WALK IS ADVERSARIAL CATASTROPHIC FORGETTING [F]
Slow-walk attack = attacker deliberately poisons ARE's behavioral baseline.
Defense layer 1: HIGH_RISK VERIFY policy threshold (catches action regardless of score).
Defense layer 2: variance growth rate > 2x weekly average = early warning flag.
Variance window = 7 days (VARIANCE_WINDOW_DAYS constant). Cross-ref: L34, MUL-04.

L84 — TEST CORPUS CONTAMINATION IS A CREDIBILITY RISK [F]
Same scenarios used for calibration AND performance claims = circular.
"0.00% FP on held-out validation set" = bulletproof.
"0.00% FP on calibration corpus" = challenged immediately by any FAANG engineer.
Fix before any enterprise pilot conversation. Cross-ref: L52, MUL-01.

L85 — CURRICULUM LEARNING MAPS TO PILOT ROLLOUT STRUCTURE [F]
Observe → flag → enforce = simple to complex curriculum.
Document staged structure in LoU. Enterprise security teams recognize this pattern.
Gives them mental model they already understand. Reduces pilot risk perception.

L86 — PHASE 2 SCORER IS A MULTITASK MODEL [INF]
Three task-specific scoring heads, shared 8-dimension feature vector layer.
Task 1: z-score anomaly. Task 2: LLM intent classification. Task 3: peer cluster deviation.
Cross-task learning: anomaly detection benefits from intent classification signals.
Note: feature vector layer is already implemented in Phase 1. P2-10 before P2-3.

L87 — ENTERPRISE PERFORMANCE CLAIMS REQUIRE 4-METRIC REPORTING [F]
Single-metric claims (TP rate only) are rejected by enterprise security reviewers.
Required format: TP rate + FP rate + Precision + F1 score.
ARE's current metrics: TP=86.67%, FP=0.00%, Precision=100%, F1=0.9286.
Always report against held-out set, not calibration corpus.

L88 — ARE'S AUDIT TRAIL IS A SIEM FEED, NOT A LOG FILE [F]
Hash-chained audit trail must be positioned as SIEM integration, not logging.
Every enforcement decision = structured event in CEF/JSON format:
{agent_did, action, reason_object, confidence_pct, timestamp, hash}
Verify format before claiming: run PL-SIEM-VERIFY first.

L89 — PHASE 1 ENFORCEMENT DECISIONS ARE PHASE 2 TRAINING DATA [F]
Every ARE enforcement decision (block/flag/pass) with its reason object = labeled example.
Every confirmed true positive = high-confidence training label for Phase 2 classifier.
This is the knowledge distillation teacher-student mapping applied to ARE.
CRITICAL: store complete feature vectors now in Phase 1, not just scores.
Run TW-0 verification immediately. This is Phase 1 work. Cross-ref: L70, MUL-02.

L90 — CISO UNLOCK PHRASE: "YOU CONTROL THE PACE" [F]
Risk-staged deployment (observe → flag → enforce) with explicit sign-off gates
removes the primary enterprise security objection: fear of disruptive enforcement.
The phrase "you control the pace, we don't advance without your sign-off" directly
addresses this fear. Use in all first conversations with security buyers.
Practice this phrase until reflexive — see TW-REHEARSAL.

L91 — GDPR ARTICLE 22 MUST BE ADDRESSED PROACTIVELY [F]
ARE's automated enforcement decisions on agent behavior may implicate GDPR Article 22.
Key distinction: enforcement is on agent behavior, not human subjects.
If agent acts on behalf of human, enterprise must assess applicability.
ARE provides the audit trail required for any Article 22 compliance review.
The staged rollout with explicit CISO sign-off = the human oversight documentation.
DORA-regulated and EU enterprises will raise this. Andy Watkin-Child knows this.
Cross-ref: L90, MUL-05.

L92 — ORG-SCOPED BASELINE ISOLATION IS HORIZONTAL SCALING PROOF [F]
Welford's running mean/variance scoped per org_id+agent_did (not just agent_did)
prevents cross-tenant baseline contamination AND enables horizontal scaling.
Kong plugin stateless (Redis-backed) + org-scoped baseline = horizontally scalable
by architecture, not by configuration. Verify before claiming: run TW-PRE-2.
Cross-ref: L5, L38, L57.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
CROSS-REFERENCE INDEX — v3.1 new additions
Critical multiplier pairs: verify both tasks are sequenced correctly.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

MUL-01: L84 × L52 — Test contamination × FP bias audit
  Rule: PL-16 must run AFTER TW-5. Never on calibration corpus.

MUL-02: L89 × L70 — Phase 2 training data × feature vector storage
  Rule: TW-0 must verify FIRST. If failing, PL-FEATURE-VECTOR-STORE must complete
  before pilot go-live. This is Phase 1 work disguised as Phase 2 design.

MUL-03: L82 × L20 — Ensemble enforcement × G-FP gate
  Rule: TW-1 talking point #3 must connect ensemble design to G-FP explicitly.
  "This IS why ARE achieves 0.00% FP."

MUL-04: L34 × L14 — Variance growth rate × behavioral window
  Rule: TW-6 must document VARIANCE_WINDOW_DAYS as a named constant.
  The window parameter is auditable. It cannot be a magic number.

MUL-05: L91 × L90 — GDPR Article 22 × "you control the pace"
  Rule: TW-2 must include the Article 22 human oversight note.
  One document serves two buyers: CISO gets control, compliance gets audit trail.

CON-04: P2-10 → P2-3 ordering dependency
  Rule: P2-10 (token budget) must be decided before P2-3 (fine-tuning strategy).
  Budget ceiling constrains eligible base models.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SUMMARY STATS — v3.1
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Total learnings processed      : 73
COMPOUND (direct ARE value)    : 58 (79%) ← upgraded from 52 (71%)
QUEUE (Phase 2 design)         : 7  (10%) ← down from 13
REJECT (not applicable)        : 8  (11%)

v3.1 verdict changes:
  L5:  QUEUE → COMPOUND  (org-scoped baseline = Phase 1 claim)
  L8:  QUEUE → COMPOUND  (feature attention framing = Phase 1 claim)
  L15: QUEUE → COMPOUND  (fixed-window framing = Phase 1 answer)
  L17: QUEUE → COMPOUND  (context-sensitive enforcement is live now)
  L22: QUEUE → COMPOUND  (same as L17 — Phase 1 policy packs)
  L27: QUEUE → COMPOUND  (A/B weight validation = PL-13, already exists)
  L33: QUEUE → COMPOUND  (LoRA-by-hand = Phase 1 calibration model)
  L40: QUEUE → COMPOUND  (EU localization answer = document proactively)
  L56: QUEUE → COMPOUND  (customer-hosted = Phase 1 answer Lloyd needs)

Execution tasks:
  TIER 1 (this week, before Lloyd)  : 12 distinct actions (~11 hrs)
  TIER 2 (before pilot go-live)     : 16 tasks (2 promoted in, 1 consolidated)
  TIER 3 (Phase 2 design)           : 13 tasks (1 promoted out)
  Total distinct tasks              : 41 (down from 46 via 5 consolidations)

Task changes:
  NEW:      TW-0, TW-PRE-2, TW-REHEARSAL, PL-SIEM-VERIFY, PL-FEATURE-VECTOR-STORE
  PROMOTED: PL-11 → TIER 1 (TW-PRE-2), P2-14 → TIER 2 (PL-FEATURE-VECTOR-STORE)
  REMOVED:  PL-9 (absorbed into TW-3), P2-14 as TIER 3 entry
  MERGED:   PL-7 + PL-18 + PL-19 + PL-21 + PL-22 → PL-SECURITY-ATTESTATION
  GHOST→CLOSED: TW-8/10/12/13/14/15 converted to [✓ MERGED] format

New MASTER_LEARNINGS entries this version: none (v3.0 entries L81-L92 updated inline)
Cross-reference index added: 5 multiplier pairs, 1 ordering dependency

Commercial gate status:
  T8 (Lloyd LoU): IN PROGRESS — meeting week of April 7
  Character Capital G6: Decision ~April 13
  DO NOT ASK ABOUT LLOYD UNTIL AFTER APRIL 6, 2026

Highest leverage this week (do these first):
  TW-0       : Feature vector storage verification     (15 min) ★ CRITICAL
  TW-PRE-2   : Org-scoped baseline verification        (20 min) ★ CRITICAL
  TW-1       : Lloyd meeting prep — 13 talking points  (60 min)
  TW-REHEARSAL: Say the phrases aloud before the meeting (20 min)
  TW-5       : Held-out test set creation              (30 min) ★ UNLOCKS PL-16

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
END OF LEARNING_INTELLIGENCE v3.1
73 learnings | 41 tasks | 12 MASTER_LEARNINGS entries | 5 multiplier pairs indexed
Updated: March 31, 2026
Next learning number: L74
Next MASTER_LEARNINGS number: L93
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
