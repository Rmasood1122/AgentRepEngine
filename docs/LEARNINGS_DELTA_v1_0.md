# LEARNINGS_DELTA v1.0 — AgentRepEngine
# Source: Full chat session extraction — April 5–6, 2026
# Sessions analyzed: 20 (all sessions in this project)
# New learnings: L103–L134 (32 new learnings)
# Cross-referenced against LEARNING_INTELLIGENCE v3.1 (L1–L102)
# Verdict system: COMPOUND | QUEUE | REJECT
# Build impact: COMPOUND + NEW | COMPOUND + EXISTS | DILUTES | NEUTRAL
#
# CRITICAL RULE FOR EVERY BUILD TASK IN THIS FILE:
# Before building: check if it EXISTS, DILUTES, or COMPOUNDS existing code.
# EXISTS = verify exact implementation, do not rebuild
# DILUTES = rejected unless there is a documented superior reason
# COMPOUNDS = build it, additive only, G-FP gate before and after
#
# Source sessions:
#   S1 = SWOT evaluation panel (April 5, chat 4d9e9661)
#   S2 = Cassandra Mack / scoring critique (April 5, chat f25225ef)
#   S3 = Expert panel critique framework (April 5, chat 39a5c0cd)
#   S4 = 24X product discovery (April 5, chat 43072162)
#   S5 = OMEGA operating system (April 5, chat cbe8b338)
#   S6 = Investor deck session (April 6, chat 5b3834da)
#   S7 = Gap audit + FP architecture (April 5, chat 52ce85bf)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SECTION A — FP ARCHITECTURE LEARNINGS (L103–L112)
Most critical category. FP rate is the enterprise unlock phrase.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

L103 — FP BENCHMARK COMPARISON CATEGORY IS WRONG | COMPOUND
Source: S7 (FP architecture session), S1 (SWOT panel)
Learning: ARE's correct benchmark comparison is NOT anomaly detection
(SIEM: 40–60% FP, EDR: 5–15%, WAF: 2–8%). It is financial transaction
fraud monitoring (1–2% FP). This is the industry that solved the same
problem: high-volume behavioral scoring with catastrophic FP consequences.
ARE's 0.00% on synthetic corpus bests even mature fraud systems.
Build impact: COMPOUNDS — changes how FP claims are framed in all docs.
Build task: Update all enterprise docs to benchmark against fraud systems,
not anomaly detection. File: docs/enterprise/fp-corpus-independence.md
Effort: 30 min doc update | Phase: Before Lloyd

L104 — TWO-SPEED BASELINE ADDRESSES COLD-START AND LEGITIMATE CHANGE | COMPOUND
Source: S2 (scoring critique), S1 (SWOT panel)
Learning: Phase 1 uses a single fixed-window z-score baseline. Two failure
modes exist: (1) cold-start — new agents have no baseline, default to cluster;
(2) legitimate behavioral change — a trader who changes strategy looks
anomalous for 30 days. The solution is two baselines:
  Fast baseline: 3-day EWMA (exponentially weighted moving average)
  Slow baseline: 30-day rolling (existing Phase 1)
  Signal = deviation from BOTH. Persistent deviation from fast but not slow
  = genuine behavioral shift, not attack. Override required.
Build impact: COMPOUNDS existing baseline. Does NOT replace. Additive.
Build task: internal/scoring/baseline.go — add FastBaseline struct and
  GetFastBaseline(). Consumer uses both; only flags if BOTH thresholds exceeded.
  G-FP gate: FP must not regress. Regression = revert immediately.
Effort: 4 hrs | Phase: After pilot LoU signed
Dependency: Requires 30+ days of production data to tune EWMA weight.
  On synthetic data it works but gains full value post-pilot.

L105 — INSIDER THREAT BLIND SPOT IN OVERRIDE LEARNING LOOP | QUEUE (Phase 2)
Source: S2 (scoring critique), S3 (expert panel)
Learning: The RLHF override loop (M7) has a structural blind spot: a malicious
insider who controls the override workflow can slowly train the system to
accept their own attack pattern as "authorized exception." This is the
equivalent of a fraud analyst approving their own fraudulent transactions.
The fix requires: (a) override reviewer != agent operator (separation of duties),
(b) override pattern analysis — if same reviewer approves same policy repeatedly,
flag for audit, (c) override rate per reviewer tracked separately from override
rate per policy.
Build impact: COMPOUNDS M7. Does not dilute. Adds reviewer-level anomaly detection.
Build task: internal/audit/threshold_advisor.go — add ReviewerPatternAlert struct.
  Flag if any single reviewer_id accounts for >30% of overrides on one policy.
Effort: 2 hrs | Phase: After M7 is live with real data

L106 — NON-GAUSSIAN BEHAVIORAL DISTRIBUTIONS MAKE Z-SCORE FRAGILE | QUEUE (Phase 2)
Source: S3 (expert panel E4 adversarial ML)
Learning: Z-score anomaly detection assumes Gaussian distribution of baseline
behavior. AI agent behavior is NOT Gaussian — it has heavy tails (bulk exports,
recursive spawns), multimodal distributions (trading agent peaks at open/close),
and zero-inflated features (cross_tenant_probe_count is almost always 0).
The correct statistical treatment for non-Gaussian distributions:
  MAD (Median Absolute Deviation) for heavy-tailed features
  Log-transformation before z-scoring for multiplicative features
  Separate null model for zero-inflated features
Build impact: COMPOUNDS scoring without replacing formula.
Build task: Phase 2 only — requires 90 days of production data to fit
  per-feature distributions. Phase 1 z-score is correct for synthetic data.
  Document the limitation in evaluation-harness-methodology.md NOW.
Effort: 30 min doc | 8 hrs build after 90 days data | Phase: Document now, build Phase 2

L107 — EQUAL WEIGHT 0.5H + 0.5V COMMODITIZES THE FORMULA | QUEUE (Phase 2)
Source: S2 (scoring critique), S3 (expert panel)
Learning: Equal weights H=0.5, V=0.5 is the correct Phase 1 starting point
(no data to justify anything else). But once production data exists, per-archetype
weight optimization produces significant accuracy gains. A trading agent should
weight V (velocity anomaly) more heavily during market hours. A batch agent
should weight H (historical trust) more heavily since it legitimately runs in bursts.
The A/B validation system (PL-13, now built) is exactly the right mechanism.
Build impact: PL-13 already built. This learning validates it. No new build.
Note: This is the Phase 2 payoff of PL-13. Document explicitly.
Effort: 0 (already built) | Phase: Activated post-pilot with real archetype data

L108 — FP CORPUS INDEPENDENCE REQUIRES EXTERNAL WITNESS | COMPOUND
Source: S7 (gap audit session), S1 (SWOT panel E12 regulatory)
Learning: "0.00% FP on self-authored corpus" is a credibility risk (L77 already
captured this). What was NOT captured: the path to a credible external witness.
Options in priority order:
  (1) Pilot data itself — customer's own production traffic is the gold standard
  (2) Academic collaboration — Cornell Tech or NYU Tandon co-authoring a
      validation study using ARE's corpus methodology
  (3) Zenodo pre-registration — publish corpus construction methodology BEFORE
      adding new scenarios (proves spec-first protocol, prevents hindsight bias)
Build impact: COMPOUNDS FP credibility. No code change.
Build task: docs/enterprise/fp-corpus-independence.md — add section on
  external validation path. Zenodo pre-registration for FP-5 boundary scenarios.
Effort: 1 hr doc | Phase: Before FP-5 build (300 scenario expansion)

L109 — 7-LAYER FP ELIMINATION STACK (APEX-FP-PRECISION) | QUEUE (Phase 2-3)
Source: S7 (5-expert FP panel)
Learning: A full FP elimination stack has 7 layers. Phase 1 implements layers 1-2.
The full stack:
  L1: Policy pre-filter (BUILT) — HIGH_RISK VERIFY independent of score
  L2: Ensemble enforcement (BUILT) — score AND policy must agree
  L3: Two-speed adaptive baseline (L104 above) — fast + slow windows
  L4: Per-archetype contextual weights (L107 above) — RLHF post-pilot
  L5: Isotonic/Platt calibration — post-hoc probability calibration
  L6: Active learning review queue — low-confidence decisions routed to human
  L7: RLHF threshold optimization — override pattern feeds weight updates (M7)
Build impact: L3-L7 are Phase 2-3. Document the roadmap. Do not build prematurely.
Build task: docs/enterprise/fp-measurement-methodology.md — add 7-layer stack
  section. Communicates technical depth to enterprise security reviewers.
Effort: 1 hr doc | Phase: Before pilot go-live (documentation only for L3-L7)

L110 — SELF-VERIFYING SQL PROOF SYSTEM IS A COMMERCIAL MOAT | COMPOUND
Source: S7 (FP architecture session)
Learning: "Here is the SQL query. Run it yourself. The result is your FP rate."
This is not just a demo technique — it is a commercial differentiation mechanism.
No SIEM vendor, no EDR vendor, no AI security platform can offer this because
they control the measurement tooling. ARE's FP rate is measured by the customer
on the customer's infrastructure. This is the enterprise trust unlock that cannot
be copied without changing architecture.
The query:
  SELECT COUNT(*) FILTER (WHERE override=true) / COUNT(*)::float AS fp_rate
  FROM enforcement_decisions
  WHERE created_at > NOW() - INTERVAL '30 days';
Build impact: COMPOUNDS. fp_candidates table (built) + override workflow (built)
  + this framing = complete self-verifying FP measurement system.
Build task: Add this exact SQL to docs/enterprise/fp-corpus-independence.md
  and Lloyd_meeting_prep.md as "C13 — The query your CISO runs themselves."
Effort: 30 min doc update | Phase: Before Lloyd meeting

L111 — FP RATE FRAMING UPGRADE: "STATISTICALLY BOUNDING" | COMPOUND
Source: S7 (gap audit — 5-expert panel output)
Learning: The current Tier 1 FP claim is: "0.00% FP on 100-scenario corpus —
bounding FP rate below 3.6% with 95% confidence (Clopper-Pearson)."
The 5-expert panel produced a mathematically stronger Tier 2 claim:
"Zero false positives across 150 synthetic validation scenarios including
50 boundary scenarios at z-score 4.0–5.5 — bounding FP rate below 2.0%
with 95% confidence."
This is stronger because boundary scenarios are harder — they specifically
test the cases most likely to produce FPs (legitimate agents near the threshold).
Build impact: COMPOUNDS. 50 boundary scenarios already built (FP-3, migration 003
  applied). The Tier 2 claim is already earned. Update the claim everywhere.
Build task: grep -r "3.6%" docs/ and replace with Tier 2 claim. 30 min.
Effort: 30 min | Phase: Immediate — before Lloyd meeting

L112 — PRODUCTION FP MEASUREMENT REQUIRES 3 PARALLEL TRACKS | COMPOUND
Source: S7 (gap audit)
Learning: Production FP measurement is not just one query. Three parallel tracks:
  Track 1: fp_candidates table (BUILT) — every RESTRICTED/BLOCKED logs for review
  Track 2: Override rate per policy (M7 BUILT) — override = confirmed FP proxy
  Track 3: Customer-reported FPs via /enforcement/override endpoint (BUILT)
All three built. What's missing: a daily_fp_metrics table populated by a
scheduled job that aggregates all three tracks into one FP rate number.
Build impact: COMPOUNDS. Fills the last gap in production FP measurement.
Build task: migrations/005_daily_fp_metrics_job.sql — scheduled job that
  reads fp_candidates + enforcement_decisions overrides and writes to
  daily_fp_metrics table. 2 hrs.
Effort: 2 hrs | Phase: Before enforce mode goes live

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SECTION B — SCORING ARCHITECTURE LEARNINGS (L113–L119)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

L113 — AGENT BEHAVIORAL PASSPORT SOLVES COLD-START + GENERATES COMPLIANCE EVIDENCE | COMPOUND
Source: S2 (Cassandra Mack session — full scoring critique)
Learning: Before any agent touches production, the CISO signs a mandate document
declaring: agent identity, authorized tools, authorized data scopes, expected
behavioral ranges, max spawn depth. This "behavioral passport" solves three problems:
  (1) Cold-start: passport defines expected behavior before baseline exists
  (2) Compliance evidence: signed mandate = documented human oversight per DORA Article 22
  (3) Enforcement threshold: passport thresholds override YAML policy for that agent
This is NOT Phase 1 infrastructure. It is Phase 2 governance layer.
But it is a CATEGORY-REDEFINING positioning move:
"Agent Governance Infrastructure" vs "Agent Runtime Security."
Build impact: COMPOUNDS — adds governance layer on top of enforcement layer.
  Phase 1 is the enforcement layer. Phase 2 adds the governance layer.
  The passport concept makes ARE the pre-deployment + runtime solution.
Build task: docs/architecture/agent-behavioral-passport-design.md — design doc only.
  Stubs exist: internal/trust/passport.go (already in repo). Build after pilot.
Effort: 2 hrs design doc | Phase: After Lloyd LoU signed

L114 — INTENT-AWARE BEHAVIORAL FINGERPRINTING (ASYNC LLM CLASSIFICATION) | QUEUE (Phase 2)
Source: S2 (scoring critique), S3 (expert panel Layer 4)
Learning: The 8-dimension feature vector captures WHAT agents do (rate, count, depth).
It does not capture WHY (intent). An async LLM classifier running post-request
on (tool_name, parameter_schema, output_schema) tuples produces a 64-dimension
intent vector per session. Deviation from intent baseline = more signal than
behavioral deviation alone. Combined score:
  Final = 0.5 * behavioral_score + 0.3 * intent_score + 0.2 * peer_cluster
Build impact: COMPOUNDS Phase 1. Does NOT replace. Requires Phase 2 infrastructure.
  Phase 1 feature vector is the training data for Phase 2 intent classifier.
  Every enforcement decision in Phase 1 is a labeled training example.
Build task: docs/architecture/phase2-scoring-design.md — add intent classification
  section. Design P2-10 (intent classification design, already in LEARNING_INTELLIGENCE)
  needs updating with 64-dimension intent vector spec.
Effort: 1 hr doc update | Phase: Document now, build after 90 days pilot data

L115 — CAUSAL CHAIN SCORING FOR MULTI-AGENT PIPELINES | QUEUE (Phase 3)
Source: S3 (expert panel), S5 (OMEGA — AADD feature)
Learning: In multi-agent pipelines (LangGraph, AutoGen), individual agent scores
miss coordinated attacks where each agent looks normal but the pipeline produces
harmful output. Causal chain scoring tracks: agent A calls agent B which calls
agent C. Scores propagate through the chain. If agent C's action is HIGH_RISK,
score penalty propagates backward through the call chain.
Build impact: COMPOUNDS M5-STEP-2 (coordinated attack detection, built).
  Coordinated detection catches simultaneous deviation. Causal chain catches
  sequential dependency attacks. Different problem, complementary solution.
Build task: Design doc only for Phase 3. Not before 3+ enterprise deployments.
Effort: 1 hr design doc | Phase: Phase 3 (after federation proven)

L116 — MICROSOFT AGT POSITIONING: COMPLEMENT NOT COMPETE | COMPOUND
Source: S7 (gap audit — 24X necessity analysis), S5 (OMEGA DEFENSE)
Learning: Microsoft AGT is application-layer, Azure-specific, requires SDK integration,
free and open source. ARE is gateway-layer, any cloud/on-prem, no SDK required.
The correct positioning is NOT "we're better than Microsoft" — it is:
"Microsoft governs Azure agents. ARE governs every other agent in your fleet.
Together: complete fleet coverage."
The Entra bridge endpoint (described in OMEGA DEFENSE) makes Microsoft AGT
a distribution channel for ARE — every Microsoft deployment creates a ARE need.
Build impact: COMPOUNDS positioning. Zero code change.
Build task: docs/competitive/microsoft-response.md — add "Entra bridge" section.
  3-line REST endpoint that reads AGT trust scores and maps to ARE score bands.
  The bridge makes ARE the "non-Azure" layer in every Microsoft AGT deployment.
Effort: 2 hrs endpoint + 1 hr doc | Phase: After pilot LoU — Week 2

L117 — SCORE BAND DEFINITIONS INCONSISTENT ACROSS DOCUMENTS | COMPOUND (FIX)
Source: S2 (34-upgrade roadmap verification, 94% accurate finding)
Learning: The expert panel identified a score band definition conflict.
APEX v5.2 spec defines: 800–1000 TRUSTED | 500–799 MONITORED | 200–499 RESTRICTED | 0–199 BLOCKED
CLAUDE_MASTER_v2_0 defines: 700–1000 TRUSTED | 500–699 MONITORED | 200–499 RESTRICTED | 0–199 BLOCKED
The APEX v5.2 spec is authoritative. The 700 threshold in CLAUDE_MASTER is a bug.
This matters: Lloyd_meeting_prep.md, any doc quoting "score above 700 = TRUSTED"
is wrong. The correct threshold is 800.
Build impact: FIX — not a dilution. Correcting an error, not changing architecture.
Build task: grep -r "700" docs/ and verify each instance is either the decay
  constant (e^-0.1 × days starting at 700 bootstrap) or a score band threshold
  (which should be 800). Fix mismatches.
Effort: 30 min | Phase: Before Lloyd meeting — URGENT

L118 — OBSERVE-TO-ENFORCE TRANSITION CRITERIA NOT DOCUMENTED | COMPOUND
Source: S2 (34-upgrade roadmap — C5 identified as not built)
Learning: The LoU staged curriculum doc describes phases but does not define
the quantitative criteria for advancing from observe → enforce mode.
The CISO will ask: "How do I know when I'm ready to go live?"
The answer must be specific and measurable:
  Gate 1: 14 days clean observe (no FP spikes, no auto-rollback events)
  Gate 2: Baseline established (agent_baselines table has ≥100 samples per agent)
  Gate 3: FP rate on production traffic < 2% (from fp_candidates table)
  Gate 4: At least 1 confirmed TP (attack detected, human-reviewed and confirmed)
  Gate 5: CISO sign-off documented (written, in audit trail)
Build impact: COMPOUNDS LoU. Not already built. Critical for pilot.
Build task: docs/enterprise/observe-to-enforce-criteria.md (new file)
  AND add section to pilot-letter-of-understanding.md.
Effort: 1 hr | Phase: Before Lloyd meeting — URGENT

L119 — X-AGENT-DID / JWT BINDING GAP IS A SECURITY FIX, NOT JUST VERIFICATION | COMPOUND
Source: S2 (34-upgrade roadmap — misclassified as verification item)
Learning: The expert panel flagged that X-Agent-DID header can be spoofed if
Kong does not verify that the DID in the header matches the DID in the JWT claims.
This is not just a "verification item" — it is a security fix. An agent that
can craft its own X-Agent-DID header can impersonate any other agent and inherit
its score. The fix: Kong Lua handler must compare X-Agent-DID value against
the sub or agent_did claim in the verified JWT. Reject if mismatch.
Build impact: COMPOUNDS security hardening. Should have been in CAT-1.
  Current code: handler.lua reads X-Agent-DID and calls /verify. If /verify
  passes but DID mismatch exists, the agent gets another agent's score.
Build task: kong/plugins/agent-reputation/handler.lua — after JWT verify,
  compare X-Agent-DID to jwt_claims.agent_did. Reject on mismatch.
  Add test: TestAgentDIDSpoofPrevention.
Effort: 1 hr code + 30 min test | Phase: Before enforce mode — URGENT SECURITY FIX

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SECTION C — ENTERPRISE + COMMERCIAL LEARNINGS (L120–L127)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

L120 — VENDOR RISK SUMMARY + DPA TEMPLATE NOT BUILT | COMPOUND
Source: S2 (34-upgrade roadmap — C6 identified as not built)
Learning: Before any enterprise pilot begins, the customer's legal/procurement
team will require: (a) a Vendor Risk Summary (2-page document: what data ARE
collects, where it's stored, who has access, incident response procedure);
(b) a Data Processing Agreement template (standard DPA covering GDPR Article 28,
DORA Article 30, CCPA). Without these, the pilot cannot start regardless of
CISO enthusiasm. Lloyd's IT security team will ask for both.
Build impact: COMPOUNDS commercial readiness. Zero code. Documentation only.
Build task: docs/enterprise/vendor-risk-summary.md + docs/enterprise/dpa-template.md
Effort: 3 hrs | Phase: Before enforce mode, ideally before Lloyd meeting

L121 — SOC2 TYPE II FOR ARE vs SOC2 EVIDENCE FOR CUSTOMERS — DIFFERENT PRODUCTS | COMPOUND
Source: S2 (Cassandra Mack session — SOC2 analysis)
Learning: There are two distinct SOC2 value propositions:
  (a) ARE obtains its own SOC2 Type II attestation — a 6-month parallel track
      that makes ARE credible as a vendor in enterprise procurement reviews.
  (b) ARE generates the SOC2 evidence customers need for their own audits —
      the hash-chained audit trail, the enforcement decision log, the FP rate
      measurement. This is available NOW and is the stronger commercial moat.
Correct framing: "ARE produces the SOC2 CC7.2 evidence your auditor needs.
  ARE's own SOC2 audit is planned for 2027 after first enterprise deployment."
Build impact: COMPOUNDS positioning. Zero code. Update all "SOC2-ready" claims.
Build task: Update docs to distinguish (a) vs (b). Current claim is correct
  ("SOC2-ready architecture") but needs the (b) framing added.
Effort: 30 min doc update | Phase: Before Lloyd meeting

L122 — CASSANDRA MACK (TENSORWAVE CISO) — CONFIRMED MEETING, EMAIL PROVIDED | COMMERCIAL
Source: S2 (Cassandra Mack session)
Learning: Cassandra Mack, CISO at TensorWave, confirmed a meeting and provided
her email. She is the host of The Cyber Breakfast Club and Founding Member of the
AI Security Council. TensorWave is a GPU cloud provider — they have AI agent
deployments at scale. This is a T8-equivalent opportunity separate from Lloyd.
Status: Meeting confirmed. Email provided. Next step: schedule and attend.
Build impact: Commercial action only. No code.
Pipeline note: Cassandra Mack should be added to pipeline tracking.

L123 — VICTOR OSAKWE REGULATORY FRAMING CORRECTION | COMPOUND (Positioning)
Source: S2 (Cassandra Mack session)
Learning: Victor Osakwe (EU AI Act compliance) publicly corrected a framing error:
Article 15 ≠ Article 17. The correct mapping for ARE's audit trail:
  Article 17 = "record-keeping obligations for high-risk AI systems" ← ARE's trail
  Article 15 = "accuracy, robustness and cybersecurity" ← different requirement
This is not a major error but it matters in EU DORA conversations. A DORA examiner
will notice the wrong article number. Fix it in all DORA compliance materials.
Build impact: COMPOUNDS accuracy of compliance docs.
Build task: grep -r "Article 15" docs/regulatory/ and verify each is correct.
  Update dora-examiner-protocol.md if needed.
Effort: 30 min | Phase: Before DORA conversation (Watkin-Child, BNY Mellon)

L124 — GIOVANNI DE LILLO (HYPEREUM CEO) — INDEPENDENT CONFIRMATION OF ARE THESIS | COMMERCIAL
Source: S2 (Cassandra Mack session)
Learning: Giovanni De Lillo, CEO of Hypereum Ltd (EU AI Act compliance platform),
independently articulated ARE's core thesis in LinkedIn threads TWICE without
prompting. This is independent external validation. He understands the category.
He is not a buyer (compliance software, not security enforcement) but could be:
  (a) A partnership for EU market entry (DORA + EU AI Act combined positioning)
  (b) A referral source to EU regulated enterprises
  (c) A co-author for the adversarial baseline poisoning research note
Build impact: Commercial action only. DM opportunity.
Pipeline note: Giovanni De Lillo — warm outreach candidate for EU market.

L125 — ENTERPRISE AGENT COUNT DATA (CRITICAL FOR DEMO SIZING) | COMPOUND
Source: S2 (Cassandra Mack session — enterprise agent count research)
Learning: Validated enterprise agent deployment scale:
  Small enterprise: 10–50 discrete agents
  Mid-market: 50–300 agents
  Large enterprise (JPMorgan): 450+ AI use cases → 200–600+ discrete agent identities
  66.4% of implementations use multi-agent architectures
This matters for: (a) demo sizing claim ("ARE manages 500+ agents per org in Phase 1"),
(b) Phase 1 capacity doc (≤500 agents confirmed as correct tier), (c) pricing model
justification ($50K ACV for up to 500 agents, $100K for 500–2000, $150K for 2000+).
Build impact: COMPOUNDS pricing model and capacity claims. No code change.
Build task: Update docs/ops/capacity.md with enterprise agent count benchmarks.
Effort: 30 min | Phase: Before any pricing conversation

L126 — DORA CMD/VERIFY BINARY IS A LEGALLY-DEFENSIBLE EVIDENCE GENERATOR | COMPOUND
Source: S4 (OMEGA output), S3 (expert panel E11 EU regulatory)
Learning: cmd/verify-chain already built (customer-runnable hash chain verifier).
What was NOT built: a DORA-specific wrapper that outputs the verification result
in the format a DORA examiner actually needs. The examiner needs:
  - Timestamp range (from/to)
  - Number of enforcement decisions in range
  - Hash chain verification status (valid/invalid)
  - Chain break location if invalid (with agent_did and timestamp)
  - Signing authority (who issued the enforcement decisions)
This is a 2-hour wrapper around the existing cmd/verify-chain binary.
Build impact: COMPOUNDS cmd/verify-chain without modifying it. Additive.
Build task: cmd/dora-verify/main.go — DORA-formatted output wrapper.
Effort: 2 hrs | Phase: After pilot LoU — Week 2

L127 — ARE SHOULD NOT COMPETE ON "AI GOVERNANCE" FRAMING | COMPOUND (Positioning)
Source: S1 (12-expert SWOT panel E8 enterprise CISO healthcare)
Learning: "AI Governance Infrastructure" is a crowded, vague category being claimed
by compliance software vendors, audit tools, policy frameworks, and monitoring
platforms. "Runtime enforcement" is specific and defensible. The framing battle:
  Vague (dilutes): "AI Governance Platform"
  Specific (compounds): "Runtime behavioral enforcement at the gateway"
The healthcare CISO expert made this point explicitly: CISOs respond to "enforcement"
(they buy it) not "governance" (they have a committee for that).
Build impact: COMPOUNDS positioning. Zero code.
Build task: Audit all docs for "governance" language. Verify it is never the
  primary framing. "Enforcement" is the primary framing. "Governance evidence"
  is what ARE produces — not what it is.
Effort: 30 min audit | Phase: Immediate

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SECTION D — ARCHITECTURE / MOAT LEARNINGS (L128–L134)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

L128 — BEHAVIORAL DATA ACROSS DEPLOYMENTS IS THE IRREPLACEABLE MOAT | COMPOUND
Source: S4 (24X product discovery), S5 (OMEGA — moat section)
Learning: The 24X product is not a new product — it is ARE's behavioral data
at scale. After 3+ enterprise deployments, ARE has the first cross-enterprise
behavioral baseline for AI agents. This baseline:
  (a) Cannot be replicated by any competitor without ARE's install base
  (b) Enables industry norms benchmarking ("your trading agents are in the
      top 10% safest for financial services")
  (c) Enables anomaly detection that uses cross-customer data, not just
      same-customer historical data
This is Stripe's network effect applied to agent behavioral trust.
Build impact: COMPOUNDS data moat. The Phase 1 feature vector storage (migration 004)
  is already collecting this data. The moat is being built right now.
Build task: docs/architecture/data-moat-roadmap.md — document the compound moat
  explicitly. This is the Series A story. Every pilot adds to the moat.
Effort: 1 hr | Phase: Before Series A conversations (Leaders in AI Summit April 21)

L129 — ARE IS STRIPE FOR AGENT TRUST, NOT SIEM FOR AI | COMPOUND (Positioning)
Source: S4 (24X product discovery — APEX NEXT output)
Learning: The category framing determines the valuation outcome.
  Wrong framing: "SIEM for AI agents" → security tool, $50M exit ceiling
  Wrong framing: "AI Governance Platform" → compliance tool, crowded, vague
  Correct framing: "Agent Trust Infrastructure" → network effects, data moat,
    every AI deployment needs it, Stripe-like infrastructure layer
The Stripe analogy is precise:
  Stripe: every payment needs a trust layer → Stripe built it once → $95B
  ARE: every AI agent deployment needs a trust layer → ARE builds it once → ?
Build impact: COMPOUNDS positioning and investor narrative. Zero code.
Build task: Update investor deck + pitch framing. "Agent Trust Infrastructure"
  is the category. "Runtime behavioral enforcement" is the mechanism.
Effort: 1 hr narrative update | Phase: Before Leaders in AI Summit April 21

L130 — FORENSIC REPLAY ENGINE UNLOCKS DORA ARTICLE 8(4) | QUEUE (Phase 2)
Source: S2 (34-upgrade roadmap C1), S3 (expert panel E11)
Learning: DORA Article 8(4) requires that financial entities can reconstruct
"any chain of events" related to an ICT incident. The replay endpoint
(/audit/replay, already built) produces a chronological event sequence.
What's missing: the DORA causal reconstruction format — a structured document
showing exactly which agent action triggered which enforcement decision,
with timestamps, scores at each step, and policy violations in sequence.
This is a 4-hour wrapper around the existing replay endpoint.
Build impact: COMPOUNDS existing replay endpoint. Additive.
Build task: cmd/dora-replay/main.go — DORA Article 8(4) causal reconstruction
  formatted output. Input: time range + agent_did. Output: causal chain document.
Effort: 4 hrs | Phase: After pilot LoU — Week 3

L131 — SDK LAYER UNLOCKS NON-KONG DEPLOYMENTS (E1 FROM UPGRADE ROADMAP) | QUEUE (Phase 2)
Source: S2 (34-upgrade roadmap E1), S3 (expert panel)
Learning: LangChain/LangGraph SDK (already built) allows non-Kong deployments.
But the full claim "works with any agent runtime, no Kong required" requires:
  - SDK that reports behavioral events to ARE scoring service directly
  - Kong remains optional (fleet-wide enforcement); SDK is per-runtime alternative
  - The SDK is the thin client; scoring engine remains centralized
This unlock changes the TAM dramatically: from "organizations running Kong"
to "any organization running any agent runtime."
Build impact: SDK partially built (36/36 tests). The full claim requires
  SDK documentation, deployment guide, and at least one non-Kong pilot.
Build task: docs/enterprise/sdk-deployment-guide.md — how to deploy ARE
  without Kong using the native SDK.
Effort: 2 hrs doc | Phase: After pilot LoU — Week 2

L132 — ARE NEEDS A PILOT SUCCESS CASE STUDY TEMPLATE NOW | COMPOUND
Source: S3 (expert panel E9 enterprise CISO financial services)
Learning: When Lloyd signs the LoU, ARE needs to produce a structured case study
at day 30 that his CISO can use internally to justify continued investment and
report to their own board/regulator. The format regulators expect:
  Section 1: Background (enterprise profile, agent count, deployment date)
  Section 2: Methodology (observe mode duration, baseline establishment, FP rate)
  Section 3: Results (enforcement decisions, confirmed TPs, overrides, FP rate)
  Section 4: Incidents (confirmed attacks detected, prevented, documented)
  Section 5: Regulator Readiness (audit trail verification, DORA Article mapping)
This template must exist before the pilot starts so ARE can collect the right data.
Build impact: COMPOUNDS pilot commercial value. No code. Documentation.
Build task: docs/enterprise/pilot-case-study-template.md — 5-section template.
Effort: 2 hrs | Phase: Before pilot starts — URGENT

L133 — GRADIO DEMO (HACKATHON) IS A PUBLIC PROOF ASSET, NOT JUST A DEMO | COMPOUND
Source: S2 (hackathon session), investor deck session (S6)
Learning: The Gradio URL was used at the hackathon but was never committed or
documented as a public proof asset. The demo shows: clean agent scoring 850 TRUSTED,
compromised agent dropping to 187 BLOCKED, hash chain verified, 0.00% FP rate.
This is the entire product in 90 seconds. It should be:
  (a) Committed to the repo as scripts/hackathon-demo.sh (M8, not yet built)
  (b) Linked from README.md as "Live demo"
  (c) Referenced in the investor deck
  (d) Sent to Lloyd along with the LoU as "here's what you'll see in 30 days"
Build impact: COMPOUNDS marketing and investor credibility. Low effort.
Build task: scripts/hackathon-demo.sh — real-time score visualization.
  README.md — add "Try the demo" link.
Effort: 2 hrs | Phase: Before Lloyd meeting

L134 — INVESTOR DECK EXISTS (ARE_InvestorDeck.html) — NEEDS METRICS UPDATE | COMPOUND
Source: S6 (investor deck session, chat 5b3834da)
Learning: A full 11-slide investor deck was built in the April 6 session
(ARE_InvestorDeck.html). It needs two updates:
  (1) F1 score: update from 0.9286 to 0.9362 (regression suite confirmed)
  (2) Tests: update from "60 tests / 14 files" to "17 packages / 63+ tests"
  (3) Add: peer cluster detection claim (M5-STEP-1 built)
  (4) Add: coordinated attack detection claim (M5-STEP-2 built)
The deck exists. The metrics are slightly stale. Update before any investor meeting.
Build impact: COMPOUNDS investor readiness. Zero engineering.
Build task: Update ARE_InvestorDeck.html with correct metrics.
Effort: 30 min | Phase: Before Leaders in AI Summit April 21

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
BUILD PLAN — RANKED BY PHASE AND COMPOUNDING IMPACT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

PHASE 0 — BEFORE LLOYD MEETING (April 7) | ~4 hrs | Zero risk to existing code

DANGER CHECK: None of these touch existing passing tests.
All are documentation or additive code with no regressions possible.

[ ] L117 — Fix score band threshold inconsistency (700 vs 800) in docs    [30 min]
[ ] L118 — docs/enterprise/observe-to-enforce-criteria.md (new)           [1 hr]
[ ] L110 — Add self-verifying SQL query to Lloyd_meeting_prep.md as C13   [30 min]
[ ] L111 — Update FP claim to Tier 2 ("150 scenarios, <2% at 95% CI")     [30 min]
[ ] L127 — Audit "governance" vs "enforcement" language in all docs        [30 min]
[ ] L121 — Add ARE SOC2 (b) framing to docs                               [30 min]
[ ] L132 — docs/enterprise/pilot-case-study-template.md (new)             [2 hrs]

PHASE 1 — AFTER LOU SIGNED | ~12 hrs | Additive engineering + docs

[ ] L119 — Kong DID spoofing fix: handler.lua DID≠JWT_claim rejection      [1.5 hrs] ⚠ SECURITY
[ ] L112 — migrations/005_daily_fp_metrics_job.sql                         [2 hrs]
[ ] L116 — Entra bridge endpoint (3-line REST mapping AGT→ARE)             [2 hrs]
[ ] L120 — Vendor risk summary + DPA template                              [3 hrs]
[ ] L126 — cmd/dora-verify/main.go (DORA-formatted audit output)           [2 hrs]
[ ] L131 — docs/enterprise/sdk-deployment-guide.md                        [2 hrs]
[ ] L133 — scripts/hackathon-demo.sh + README demo link                    [2 hrs]

PHASE 2 — PILOT RUNNING (needs production data) | ~20 hrs

[ ] L104 — Two-speed baseline (fast EWMA + slow rolling)                   [4 hrs]
[ ] L105 — Reviewer pattern alert in threshold_advisor.go                  [2 hrs]
[ ] L113 — Agent behavioral passport implementation                        [8 hrs]
[ ] L130 — cmd/dora-replay/main.go (DORA Article 8(4) causal reconstruction) [4 hrs]
[ ] L125 — Update capacity.md with enterprise agent count benchmarks       [30 min]

PHASE 3 — AFTER 90 DAYS PRODUCTION DATA | ~24 hrs

[ ] L106 — Per-feature distribution analysis (MAD, log-transform, zero-inflated) [8 hrs]
[ ] L114 — Intent-aware behavioral fingerprinting (async LLM + 64-dim vector)    [16 hrs]

NEVER BUILD (dilutes or premature):
[ ] L107 — Equal weight commoditization → PL-13 already handles this. No new build.
[ ] L115 — Causal chain scoring → Phase 3 only, after federation proven.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
DILUTION CHECK — WHAT COULD HURT EXISTING PRODUCT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

RISK 1 — L104 (two-speed baseline) introduced before production data
  Risk: Tuning EWMA weight on synthetic data introduces bias.
  Mitigation: Build with a feature flag (disabled by default). Enable after
  30 days of pilot data. Never replace Phase 1 baseline — additive only.
  Verdict: SAFE if feature-flagged.

RISK 2 — L119 (DID spoofing fix) breaks existing JWT tests
  Risk: The fix rejects valid JWTs if DID claim is not present (legacy agents).
  Mitigation: Fail-open on missing DID claim (log warning, allow). Only reject
  on explicit mismatch (DID header present AND JWT DID present AND mismatch).
  Verdict: SAFE with fail-open logic. Test before enabling.

RISK 3 — L116 (Entra bridge) creates dependency on Microsoft API
  Risk: If Microsoft changes AGT API, ARE's bridge breaks.
  Mitigation: Bridge is optional, not critical path. Fail-open if AGT unavailable.
  Verdict: SAFE if fail-open. Low coupling.

RISK 4 — L112 (daily_fp_metrics job) writes to daily_fp_metrics table
  Risk: Table already exists (migration 001). New job may conflict with existing
  data or duplicate rows.
  Mitigation: Check if daily_fp_metrics table schema matches what migration 005
  needs. Run `\d daily_fp_metrics` before writing the migration.
  Verdict: VERIFY FIRST before building.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SUMMARY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Total new learnings: 32 (L103–L134)
Verdict breakdown:
  COMPOUND (build now or soon):  24 learnings
  QUEUE (Phase 2-3):              6 learnings
  REJECT / never build:           2 learnings

Highest value items (impact per hour):
  1. L117 — Fix score band doc inconsistency (30 min, prevents CISO confusion)
  2. L119 — DID spoofing fix (1.5 hrs, real security vulnerability)
  3. L118 — Observe-to-enforce criteria doc (1 hr, Lloyd will ask for this)
  4. L110 — Self-verifying SQL as C13 claim (30 min, closes the trust loop)
  5. L132 — Pilot case study template (2 hrs, without it pilot data is lost)

Created: April 6, 2026
Source: 20 chat sessions + full project file audit
Next learning number: L135
Format: Compatible with LEARNING_INTELLIGENCE v3.1
To activate: commit to master, upload to Claude Project, run APEX ACTIVATE
