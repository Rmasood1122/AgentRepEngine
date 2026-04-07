# ARE EXECUTION ROADMAP v1.0
# Built: April 5, 2026 | Source: Full cross-session gap audit (APEX v5.2)
# Rule: Lloyd LoU is NOT a gate for engineering or documentation.
#        LoU affects commercial sequencing only.
# Rule: Every task is additive. Nothing modifies existing passing tests,
#        scoring formula, hash chain, or JWT identity model.
# Gate: G-FP — run go test ./... before and after every engineering commit.
#        Any FP regression = revert immediately.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PHASE 0 — TODAY (April 5) | ~70 min | No code
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

★ CRITICAL GATES (run first — 30 min)

[ ] TW-0: Feature vector storage verification
    Query:
      SELECT column_name FROM information_schema.columns
      WHERE table_name = 'score_events'
      ORDER BY ordinal_position;
    Expected: All 8 feature dimensions as individual columns
    (history_score, velocity_score, z_score, policy_violations,
     agent_type_modifier, time_context_modifier, composite_score, band)
    If only composite_score → 3hr fix before any pilot claim
    Effort: 15 min verify | 3 hrs fix if failing

[ ] TW-PRE-2: Org-scoped baseline verification
    Query:
      docker exec agentrepengine-redis-1 redis-cli \
        --no-auth-warning --user are_admin -a are_redis_dev \
        KEYS "baseline:*" | head -20
    Expected format: baseline:{org_id}:{agent_did}
    Failure format:  baseline:{agent_did} → cross-tenant contamination
    If wrong format → 2hr fix immediately
    Effort: 15 min verify | 2 hrs fix if failing

COMMERCIAL MESSAGES (send all today — 40 min total)

[ ] Paul Vann (Validia CEO) — LinkedIn
    He said "talk to me this week" at Newlab April 4.
    Window closes Monday April 6. One message today.
    Frame: Validia verifies human, ARE enforces agent behavior after.
    Complete trust chain. 20 minutes this week?

[ ] Andrew Gyamfi (Head of Security, Translucent AI)
    Via Boardy message — get his direct email.
    GV/NEA-backed healthcare fintech. Sitting buyer profile.
    Status: LinkedIn connection pending. Use Boardy to bypass.

[ ] Unmukt Raizada (TrustEvals.ai)
    LinkedIn direct. Already connected.
    Ask: FinServ validation conversation + co-sell exploration.
    TrustEvals = post-hoc audit. ARE = real-time enforcement. Complementary.

[ ] Andy Watkin-Child
    LinkedIn direct. Warm Boardy intro.
    Ask: Advisory board conversation + DORA/SEC buyer introductions.

[ ] Sri Rajan — send drafted message (already written)
[ ] Rock Lambros — send drafted message (already written)
[ ] David Matousek — send drafted message (already written)

[ ] Turner Novak (Banana Capital)
    Verify domain: bana vs banana → web search first
    Then send email. Do not send to wrong domain.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PHASE 1 — APRIL 5–7 | ~5 hrs | Pre-meeting hardening
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[ ] Gen Digital ADR competitive response doc           (2 hrs)
    File: docs/competitive/gen-digital-adr-response.md
    Content: gateway-layer vs client-side, fleet vs per-developer,
    regulated enterprise vs developer-tool market positioning.
    Must exist before any investor or buyer conversation.

[ ] TW-6: Variance growth rate trigger                 (2 hrs)
    File: internal/scoring/policy.go
    Add: VARIANCE_WINDOW_DAYS = 7 named constant
    Logic: if variance growth rate > 2x weekly average → early warning
    This makes slow-walk defense defensible at code layer, not just policy.
    ADDITIVE ONLY — new check alongside existing policy violations.
    Run go test ./... before and after. Zero regressions required.

[ ] M1: Language upgrade across all customer-facing docs (1 hr)
    Apply 13 phrase replacements (see 24X_STRATEGY_EXECUTION_PLAN.md M1 table)
    Key: "scoring formula" → "behavioral attention engine"
         "observe mode" → "zero-impact visibility mode"
         "pilot" → "30-day zero-risk visibility deployment"
    Verify: grep -r "0.00%" docs/ README.md | grep -v "internal corpus"
    Any unqualified 0.00% claim must add corpus qualifier.

[ ] TW-REHEARSAL: Say all 12 Lloyd claims C1–C12 aloud  (20 min)
    Not a doc task. Not skippable. Cognitive preparation.
    Lloyd meeting happens regardless of when — be ready.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PHASE 2 — APRIL 8–21 | ~43 hrs | Product deepening
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

FP LEGITIMACY TRACK (from April 5 audit — new priority)

[ ] FP corpus expansion: 100 → 300 scenarios              (4 hrs)
    File: tests/fp_scenarios/ — add 200 new legitimate scenarios
    Coverage must include:
      - Bursty-but-legitimate traffic (batch jobs, scheduled tasks)
      - Cold-start agents (new agent, no baseline yet, day 1 behavior)
      - Multi-tenant legitimate cross-references
      - Legacy behavioral patterns (slow, sequential, unusual endpoint order)
      - High-frequency legitimate agents (trading bots, monitoring agents)
    Result: "0.00% FP on 300-scenario internal corpus" — 3x evidence base
    RULE: do not modify existing 100 scenarios. Append only.
    Run full eval harness after. FP must remain 0.00%.

[ ] FP measurement methodology doc                        (2 hrs)
    File: docs/enterprise/fp-measurement-methodology.md
    Content:
      - How 100 (→300) scenarios were constructed
      - What agent behaviors each category represents
      - How the eval harness runs (go test ./tests/fp_scenarios/...)
      - What auto-rollback (2% threshold) means in production
      - Corpus independence statement: behaviors included that
        model was not specifically calibrated on
    Purpose: When a CISO asks "how did you measure that" — hand them this.

ENGINEERING TRACK

[ ] M6: confidence_explanation field                      (12 hrs)
    File: internal/scoring/explainability.go
    Add: confidence_explanation string to reason object
    Format: "Score 513pts below band boundary (187 actual vs 700 threshold).
             3 policy violations fired. z-score 4.2σ above org baseline."
    CRITICAL: additive only. Do not alter existing reason object fields.
    confidence_pct already live — add confidence_explanation alongside it.
    Run go test ./... before and after. G-EXPLAIN gate must pass.

[ ] TW-3: Attack corpus expansion 30 → 50 scenarios      (2 hrs)
    File: tests/attack_corpus/ — add 20 new attack variants
    Do not modify existing 30 scenarios. Append only.
    TP rate will update — document new baseline after run.

[ ] TW-4: Kong payload validation                         (1 hr)
    File: kong/plugins/agent-reputation/handler.lua
    Validate all incoming payload fields before processing.
    Additive hardening only.

DOCUMENT TRACK

[ ] M2: Adversarial baseline poisoning research note      (6 hrs)
    File: docs/research/adversarial-baseline-poisoning.md
    Title: "Adversarial Baseline Poisoning: A Novel Attack Class
            Against Behavioral Scoring Systems for AI Agents"
    4 pages. Structure in 24X_STRATEGY_EXECUTION_PLAN.md M2 section.
    After completion: upload to Zenodo → new DOI (second IP anchor).
    Send to: Cornell Tech (Ristenpart, Shmatikov), NYU Tandon, IBM Research.
    This defines the attack class before any competitor writes about it.

[ ] M3: SIEM integration story                            (8 hrs)
    Files:
      docs/integrations/splunk-addon-spec.md      (2 hrs)
      docs/integrations/sentinel-connector-spec.md (2 hrs)
      docs/enterprise/siem-enrichment-positioning.md (4 hrs)
    Purpose: Unlocks IT security budgets separate from CISO.
    Two budget owners = two entry points into the same enterprise.

[ ] PL-SECURITY-ATTESTATION                               (3 hrs)
    File: docs/enterprise/security-attestation.md
    5 sections: key management, data residency, access controls,
    incident response, compliance mappings.
    Required for enterprise procurement reviews.

[ ] reason_object_v1.json formal schema                   (2 hrs)
    File: docs/specs/reason_object_v1.json
    Full JSON Schema for the reason object including confidence_explanation.
    Required by enterprise architects in procurement reviews.

COMMERCIAL TRACK

[ ] C-Corp conversion email                               (30 min)
    Send to attorney. Required before any signed commercial agreement.

[ ] Vendor package build                                  (3 hrs)
    MSA draft + pilot scope document + data processing brief.
    Required before any enterprise can sign a pilot agreement.

[ ] Secondary investor list — 10 names                    (2 hrs)
    Build list for Leaders in AI Summit NYC (April 21–22)
    and Momentum AI NYC (April 27–28).

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
PHASE 3 — APRIL 22+ | ~70 hrs | The moat (pilot data required for validation)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

NOTE: Code can be built and tested on synthetic data now.
      Production validation requires 14+ days of real pilot traffic.

[ ] M5-STEP-1: Peer cluster deviation scoring             (20 hrs)
    File: internal/scoring/scorer.go + consumer.go
    Formula: V = f(agent vs own baseline) × 0.7
               + f(agent vs peer cluster baseline) × 0.3
    cluster_baselines table already exists in migrations/001_initial.sql
    peer_cluster_avg_score already exists in scoring_explanations
    This is a wire-up task, not a new build.
    GATE: run full eval harness before and after. FP must not regress.

[ ] M5-STEP-2: Coordinated attack detection               (20 hrs)
    File: internal/scoring/policy.go
    Logic: detect multiple agents in same org deviating simultaneously
    New claim: "ARE detects both individual and coordinated fleet attacks"
    New detection class no competitor has.

[ ] M7-STEP-1: Override tracking + threshold advisor      (15 hrs)
    File: internal/audit/override.go + new threshold_advisor.go
    Count confirmed TPs (human-approved blocks) vs overrides (FP candidates)
    Identify which policy fires most overrides.
    Monthly analysis foundation.

[ ] M7-STEP-2: A/B comparison + human approval gate       (15 hrs)
    File: internal/enforcement/threshold_advisor.go
    Monthly CISO output:
      "847 decisions, 23 confirmed TPs, 3 overrides.
       Suggested: lower pii_field_access_rate 0.30→0.25.
       Estimated: +2 TPs/month, 0 additional FPs.
       F1: 0.9286→0.9401. Apply? [Yes/Review/Decline]"
    This is the CFO renewal story. $150K ACV tier unlock.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
EFFORT SUMMARY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Phase 0 (today):          ~70 min
Phase 1 (April 5–7):      ~5 hrs
Phase 2 (April 8–21):     ~43 hrs  ← includes FP legitimacy track (new)
Phase 3 (April 22+):      ~70 hrs
─────────────────────────────────
Total:                    ~120 hrs engineering + docs
                          ~63 hrs commercial + verification
                          ~183 hrs total

Zero items require new infrastructure.
Zero items modify existing passing tests retroactively.
Zero items alter the scoring formula, hash chain, or JWT model.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
FP CLAIM TIERS — USE CORRECT TIER IN CORRECT CONTEXT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Tier 1 (now):
  "0.00% false positive rate on our 100-scenario internal validation
   corpus — zero legitimate agents blocked across all synthetic test
   cases. We expect below 1% on a well-configured production environment,
   validated during the 30-day observe mode pilot before any enforcement
   activates."

Tier 2 (after Phase 2 FP track):
  "0.00% false positive rate on our 300-scenario internal validation
   corpus across [N] agent behavior categories."

Tier 3 (after pilot observe data):
  "0.00% false positive rate on [enterprise name]'s production agent
   traffic over 30 days — [N] agents, [M] requests, zero legitimate
   agents incorrectly flagged."

Tier 4 (after M7):
  "FP rate monitored continuously in production. Auto-rollback fires
   if FP exceeds 2%. This has never triggered in [N] days of operation."

NEVER SAY: "zero false positives" without corpus qualifier.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
END OF EXECUTION ROADMAP v1.0
Created: April 5, 2026 | APEX v5.2 | Next update: after Phase 0 complete
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
