# ARE MASTER ROADMAP v2.0
# Built: April 5, 2026
# Sources: Full gap audit + FP Legitimacy Panel (E1–E5) + APEX v5.2
# Supersedes: EXECUTION_ROADMAP_v1.0
#
# HARD RULES:
# 1. Lloyd LoU is NOT a gate. Engineering runs on calendar.
# 2. Every task is additive. Zero modifications to existing passing tests.
# 3. G-FP gate: go test ./... green before AND after every engineering commit.
# 4. Any FP regression = revert immediately. No exceptions.
# 5. Never say "zero false positives" — always use the agreed Tier 1 claim.
#
# AGREED FP CLAIM (use verbatim everywhere):
# "Zero false positives across our 100-scenario internal validation corpus —
#  statistically bounding our FP rate below 3.6% with 95% confidence.
#  Before any enforcement activates, we run 30 days of observe mode on your
#  production traffic and measure your actual FP rate. You decide when
#  enforcement goes live."

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PHASE 0 — IMMEDIATE (April 5–6) | ~2 hrs | Verification + unblocking
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Priority: Run verification gates first. Send commercial messages second.

─── VERIFICATION GATES (30 min) ───────────────────────────────────────

[ ] TW-0: Feature vector storage check
    docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine -c \
    "SELECT column_name FROM information_schema.columns
     WHERE table_name = 'score_events' ORDER BY ordinal_position;"
    Expected: 8 columns including history_score, velocity_score, z_score,
    policy_violations, composite_score, band
    FAIL action: 3hr fix — add feature vector columns to score_events

[ ] TW-PRE-2: Org-scoped baseline check
    docker exec agentrepengine-redis-1 redis-cli \
    --no-auth-warning --user are_admin -a are_redis_dev KEYS "baseline:*" | head -20
    Expected: baseline:{org_id}:{agent_did}
    FAIL action: 2hr fix — prefix all baseline keys with org_id

─── FP CLAIM AUDIT (30 min) ────────────────────────────────────────────

[ ] FP-AUDIT: Grep all unqualified FP claims in docs
    cd /c/Users/rmaso/AgentRepEngine
    grep -r "0\.00%" docs/ README.md --include="*.md" | grep -v "internal corpus\|validation corpus\|internal validation"
    Every match: replace with agreed Tier 1 claim
    File: any doc showing bare "0.00% FP rate" without qualifier

─── COMMERCIAL MESSAGES (40 min) ───────────────────────────────────────

[ ] Paul Vann (Validia CEO) — LinkedIn TODAY
    Window closes Monday April 6. Send before midnight today.
    "Validia verifies the human. ARE enforces what the agent does after.
     Complete trust chain. 20 minutes this week?"

[ ] Andrew Gyamfi (Translucent AI Head of Security)
    Via Boardy message. Get direct email. Sitting buyer.

[ ] Unmukt Raizada (TrustEvals.ai) — LinkedIn (already connected)
    "FinServ validation conversation + co-sell exploration?"

[ ] Andy Watkin-Child — LinkedIn (warm Boardy intro)
    "Advisory board conversation + DORA/SEC buyer introductions?"

[ ] Sri Rajan — send drafted message
[ ] Rock Lambros — send drafted message
[ ] David Matousek — send drafted message

[ ] Turner Novak — verify domain (bana vs banana) then send email

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PHASE 1 — APRIL 5–7 | ~11 hrs | Pre-meeting hardening + FP foundation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

─── FP LEGITIMACY TRACK (new — from panel output) ──────────────────────

[ ] FP-1: Add fp_candidates table (migration)               [1 hr]
    File: migrations/002_fp_candidates.sql
    Purpose: Log every RESTRICTED/BLOCKED decision during observe mode
             for human review. Foundation of production FP measurement.
    Schema:
      CREATE TABLE fp_candidates (
        id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        agent_did       TEXT NOT NULL,
        decision_id     UUID REFERENCES enforcement_decisions(id),
        score           INTEGER NOT NULL,
        band            TEXT NOT NULL,
        reason_object   JSONB NOT NULL,
        flagged_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        reviewed_by     TEXT,
        confirmed_fp    BOOLEAN,
        reviewed_at     TIMESTAMPTZ,
        notes           TEXT
      );
    ADDITIVE ONLY. Zero impact on existing tables.
    Run go test ./... after migration applied.

[ ] FP-2: Add mode_change_log table (migration)             [1 hr]
    File: migrations/003_mode_change_log.sql
    Purpose: Audit trail for every enforcement mode change.
             Required for DORA Article 11 + SOC2 CC7.2 compliance.
    Schema:
      CREATE TABLE mode_change_log (
        id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        changed_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        from_mode             TEXT NOT NULL,
        to_mode               TEXT NOT NULL,
        trigger_reason        TEXT NOT NULL,
        fp_rate_at_trigger    DECIMAL(5,4),
        triggered_by          TEXT NOT NULL
      );
    Wire: update internal/enforcement/mode_controller.go to INSERT
          into mode_change_log on every mode change.
    ADDITIVE ONLY. Run go test ./... after.

[ ] FP-3: FP corpus boundary scenarios — 50 new             [2 hrs]
    File: tests/fp_scenarios/ — add 50 scenarios
    Rule: write behavioral SPEC in tests/fp_scenarios/specs/ FIRST,
          commit the spec, THEN build the scenario.
          Spec commit hash must be referenced in scenario file.
    Coverage: agents operating at z-score 2.5–3.5 (boundary zone)
    These are the scenarios that expose co-design bias if present.
    ADDITIVE. Do not modify existing 100 scenarios.
    Run full eval harness after. FP must remain 0.00%.

[ ] FP-4: FP corpus independence doc                        [1 hr]
    File: docs/enterprise/fp-corpus-independence.md
    Content:
      - Construction methodology (spec-first protocol)
      - List of behavioral categories covered
      - Independence verification statement
      - Boundary scenario rationale
      - Confidence interval disclosure: 95% CI [0%, 3.6%] on 100 scenarios
      - Path to Tier 2 claim (300 scenarios → CI [0%, 1%])

─── COMPETITIVE TRACK ──────────────────────────────────────────────────

[ ] GEN-ADR: Gen Digital ADR competitive response           [2 hrs]
    File: docs/competitive/gen-digital-adr-response.md
    Content:
      - Gateway-layer (ARE) vs client-side (Gen Digital ADR)
      - Fleet-wide enforcement vs per-developer tool
      - Regulated enterprise buyer vs developer tool market
      - ARE's answer when a prospect asks "how are you different from Gen Digital?"
    Must exist before any investor or buyer conversation this week.

─── ENGINEERING TRACK ──────────────────────────────────────────────────

[ ] TW-6: Variance growth rate trigger                      [2 hrs]
    File: internal/scoring/policy.go
    Add: VARIANCE_WINDOW_DAYS = 7 (named constant, auditable)
    Logic: if agent's variance growth rate > 2x weekly average
           → early warning signal before high-value action
    This makes slow-walk defense defensible at code layer.
    ADDITIVE ONLY. New policy check alongside existing violations.
    Run go test ./... before and after. Zero regressions.

─── LANGUAGE + REHEARSAL ───────────────────────────────────────────────

[ ] M1: Language upgrade — all customer-facing docs         [1 hr]
    Apply 13 replacements from 24X_STRATEGY_EXECUTION_PLAN.md M1 table
    Key replacements:
      "scoring formula"    → "behavioral attention engine"
      "observe mode"       → "zero-impact visibility mode"
      "pilot"              → "30-day zero-risk visibility deployment"
      "0.00% FP rate"      → agreed Tier 1 claim (verbatim)
    Verify with grep after.

[ ] TW-REHEARSAL: Say all 12 claims C1–C12 aloud            [20 min]
    Not skippable. Cognitive prep for Lloyd meeting.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PHASE 2 — APRIL 8–21 | ~49 hrs | Product deepening + FP legitimacy
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

─── FP LEGITIMACY TRACK (continued) ────────────────────────────────────

[ ] FP-5: FP corpus expansion 150 → 300 scenarios           [4 hrs]
    File: tests/fp_scenarios/ — add 150 more scenarios
    (50 boundary already added in Phase 1 → total 150+100=250 → add 50 more)
    Target distribution (200 new total from Phase 1+2):
      Batch job agents:               25 scenarios
      Cold-start agents:              20 scenarios
      Maintenance/recovery agents:    20 scenarios
      Cross-domain compliance agents: 20 scenarios
      High-PII authorized agents:     20 scenarios
      Monitoring/watchdog agents:     20 scenarios
      Multi-step research agents:     20 scenarios
      Boundary agents (z=2.5–3.5):    35 scenarios (25 added here)
      Legacy pattern agents:          20 scenarios
    SPEC FIRST protocol required for all. ADDITIVE ONLY.
    Result: "FP rate < 1% with 95% confidence" claim unlocked.

[ ] FP-6: FP measurement methodology doc                    [2 hrs]
    File: docs/enterprise/fp-measurement-methodology.md
    Content:
      - How the corpus was constructed (spec-first protocol)
      - All 9 behavioral categories with descriptions
      - How the eval harness runs
      - Confidence interval math (Clopper-Pearson)
      - Auto-rollback as compliance control (DORA/SOC2/HIPAA mapping)
      - Production measurement protocol (fp_candidates workflow)
      - 4-tier claim ladder with evidence requirements per tier

[ ] FP-7: Wire fp_candidates auto-population                [2 hrs]
    File: internal/store/score_store.go
    Logic: on every RESTRICTED or BLOCKED decision during observe mode,
           INSERT into fp_candidates with confirmed_fp = NULL
    ADDITIVE. Does not alter existing WriteScore() logic.
    Run go test ./... after.

─── ENGINEERING TRACK ──────────────────────────────────────────────────

[ ] M6: confidence_explanation field                        [12 hrs]
    File: internal/scoring/explainability.go
    Add: confidence_explanation string to reason object
    Format: "Score 513pts below band boundary. 3 policy violations.
             z-score 4.2σ above org baseline."
    CRITICAL: additive only. Do not alter existing fields.
    confidence_pct already live — add alongside it.
    Run go test ./... before and after. G-EXPLAIN gate must pass.

[ ] TW-3: Attack corpus 30 → 50 scenarios                  [2 hrs]
    File: tests/attack_corpus/ — add 20 new attack variants
    ADDITIVE. Do not modify existing 30 scenarios.
    TP rate will update — document new baseline.

[ ] TW-4: Kong payload validation                          [1 hr]
    File: kong/plugins/agent-reputation/handler.lua
    Validate all incoming payload fields before processing.
    ADDITIVE hardening only.

─── DOCUMENT TRACK ─────────────────────────────────────────────────────

[ ] M2: Adversarial baseline poisoning research note        [6 hrs]
    File: docs/research/adversarial-baseline-poisoning.md
    4 pages. Structure per 24X plan M2 section.
    After: upload to Zenodo → second DOI.
    Send to: Cornell Tech, NYU Tandon, IBM Research (Dr. Payel Das).

[ ] M3: SIEM integration story                             [8 hrs]
    docs/integrations/splunk-addon-spec.md          (2 hrs)
    docs/integrations/sentinel-connector-spec.md    (2 hrs)
    docs/enterprise/siem-enrichment-positioning.md  (4 hrs)
    Unlocks IT security budgets separate from CISO.

[ ] PL-SECURITY-ATTESTATION                                [3 hrs]
    File: docs/enterprise/security-attestation.md
    5 sections: key management, data residency, access controls,
    incident response, compliance mappings.

[ ] reason_object_v1.json schema                           [2 hrs]
    File: docs/specs/reason_object_v1.json
    Full JSON Schema including new confidence_explanation field.

─── COMMERCIAL TRACK ────────────────────────────────────────────────────

[ ] C-Corp conversion email                                [30 min]
[ ] Vendor package: MSA + pilot scope + data brief         [3 hrs]
[ ] Secondary investor list — 10 names for April 21–22     [2 hrs]
[ ] Sovren Software follow-up                              [15 min]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PHASE 3 — APRIL 22+ | ~70 hrs | The moat (code now, validate on pilot)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
NOTE: Build and test on synthetic data now.
      Production validation requires 14+ days of real pilot traffic.
      G-FP gate: full eval harness before AND after every commit.

[ ] M5-STEP-1: Peer cluster deviation scoring              [20 hrs]
    Files: internal/scoring/scorer.go + consumer.go
    Formula: V = f(own baseline) × 0.7 + f(cluster baseline) × 0.3
    cluster_baselines table already exists. Wire-up task.
    G-FP GATE: FP must not regress after this commit.

[ ] M5-STEP-2: Coordinated attack detection                [20 hrs]
    File: internal/scoring/policy.go
    Detect multiple agents in same org deviating simultaneously.
    New claim: "ARE detects both individual and coordinated attacks."

[ ] M7-STEP-1: Override tracking + threshold advisor       [15 hrs]
    Files: internal/audit/override.go + new threshold_advisor.go
    Foundation for monthly CISO intelligence report.

[ ] M7-STEP-2: A/B threshold comparison + human approval   [15 hrs]
    File: internal/enforcement/threshold_advisor.go
    Monthly output: "847 decisions, 23 TPs, 3 overrides.
    Suggested adjustment. F1: 0.9286→0.9401. Apply? [Yes/Review/Decline]"
    CFO renewal story. $150K ACV tier unlock.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
FP CLAIM UPGRADE LADDER
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Tier 1 — NOW (use today):
  "Zero false positives across 100 synthetic validation scenarios —
   statistically bounding FP rate below 3.6% with 95% confidence."

Tier 2 — After FP-3 + FP-5 (300 scenarios):
  "Zero false positives across 300 synthetic validation scenarios
   across 9 enterprise agent behavioral categories —
   bounding FP rate below 1% with 95% confidence."

Tier 3 — After 30-day pilot observe mode:
  "Zero false positives on [enterprise]'s production traffic —
   [N] agents, [M] requests, 30 days. Production FP rate: 0.00%."

Tier 4 — After M7 + 60+ days:
  "FP rate monitored continuously. Auto-rollback has never triggered
   in [N] days of operation."

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
EFFORT SUMMARY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Phase 0:  ~2 hrs    (verification + commercial messages)
Phase 1:  ~11 hrs   (FP foundation + competitive + engineering)
Phase 2:  ~49 hrs   (FP legitimacy + product deepening + docs)
Phase 3:  ~70 hrs   (peer cluster + self-improving thresholds)
─────────────────────────────────────────────────────────────
Total:    ~132 hrs engineering/docs + commercial actions

All tasks additive. No existing tests modified.
No new infrastructure required.
No scoring formula changes.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
END OF MASTER ROADMAP v2.0
Created: April 5, 2026 | APEX v5.2 + FP Panel E1–E5
Next update: after Phase 0 verification queries run
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
