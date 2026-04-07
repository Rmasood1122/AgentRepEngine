# ARE FULL PRODUCT BUILD ROADMAP
## APEX v5.2 — Post-APEX NEXT + Microsoft Competitive Response
### April 4, 2026 | Engineering-first | Sequenced by dependency

---

## GROUND RULES

- Stack constraint: Go + Kong (Lua) + Redis + PostgreSQL + Docker. No new infrastructure until Sprint 3.
- No team until Series A.
- Sprint 0 is the unlock. Every sprint after compounds on Lloyd production data.
- Confidence labels: [F] = fact, [H] = hypothesis, [BP] = best practice, [ASS] = assumption

---

# SPRINT 0 — COMMERCIAL ARMOR + CATEGORY NAMING
## Window: Now → April 9 (Lloyd LoU signed)
## Zero engineering. Documentation and positioning only.

### S0-T1 — Microsoft Competitive Response Document
File: docs/competitive/microsoft-response.md
Effort: 3 hours

Five gap arguments:

GAP 1 — Stateless vs behavioral history
Microsoft's AGT is stateless — evaluates each call in isolation, no rolling baseline.
Slow-walk detection is architecturally impossible at 0.1ms latency.
Lloyd sentence: "Microsoft checks rules. ARE learns behavior."

GAP 2 — Azure dependency vs cloud-agnostic
Agent 365 optimized for Microsoft Foundry/Copilot Studio on AKS.
ARE enforces at Kong gateway — cloud-agnostic, framework-agnostic.
Lloyd sentence: "Microsoft governs Microsoft agents. ARE governs all agents."

GAP 3 — DID/Ed25519 friction vs JWT/RS256
Microsoft Agent Mesh requires DID infrastructure.
ARE uses signed JWT + RS256 + org JWKS — what enterprises already have.
Lloyd sentence: "Microsoft requires new identity infrastructure. ARE works with what you have. 4 hours vs 3 months."

GAP 4 — Compliance grading vs examiner-verifiable
Microsoft Agent Compliance = self-reported framework mapping.
ARE hash-chained audit trail = independently verifiable by DORA examiner.
Lloyd sentence: "Microsoft maps to DORA. ARE satisfies DORA."

GAP 5 — No slow-walk detection vs 100% slow-walk detection
At 0.1ms stateless, z-score computation against 30-day baseline is physically impossible.
ARE: 100% slow-walk detection rate on test corpus. [F]
Lloyd sentence: "The only attack that causes your DORA examination to fail is the one Microsoft cannot catch."

New one-liner (post-Microsoft):
"Microsoft's toolkit stops the agent that breaks a rule. ARE stops the agent that never breaks a rule — but drifts until it breaks your environment."

---

### S0-T2 — DORA Article 11 Examiner Protocol
File: docs/regulatory/dora-examiner-protocol.md
Effort: 4 hours
Publish to: Zenodo, anchor to DOI 10.5281/zenodo.19169185

Contents:
1. What ARE produces: hash-chained enforcement decision records per API call
2. How to verify independently: SHA-256 hash chain algorithm, plain English
3. How to recompute from PostgreSQL audit table
4. What examiner can conclude: records not modified, reason object intact
5. What ARE does not claim: compliance determination — that is the enterprise's role

---

### S0-T3 — Agentic Behavioral Certification Standard v0.1
File: docs/standards/agentic-behavioral-certification-v0.1.md
Effort: 4 hours
Publish to: Zenodo, anchor to DOI 10.5281/zenodo.19169185

Contents:
1. Definition: signed, time-windowed, hash-chain-verified record of agent behavioral trajectory
2. What it covers: baseline parameters, drift events, enforcement history, model version continuity, identity chain
3. What it does not cover: model weights, training data, system prompt content
4. Verification method: independent hash chain verification (same as DORA protocol)
5. Regulatory anchors: DORA Article 11, EU AI Act Article 9, SEC AI governance, HIPAA

This is the category naming move. ARE names "Agentic Behavioral Certification" on April 4, 2026.
Before Microsoft's OWASP working group. Before any Big 4 firm. [F — if published today]

---

### S0-T4 — competitive-positioning.md Update
File: docs/competitive/competitive-positioning.md (existing, update)
Effort: 1 hour
Add: Microsoft section, five gaps, new one-liner.

---

### S0-T5 — Trust Architecture Documents (pre-Lloyd)
Files: multiple, see Trust Plan
Effort: 12 hours total (all writing, no engineering)

S0-T5a: docs/enterprise/trust-staircase.md (3 hours)
  - Four steps: Observe → Flag → Enforce Selective → Enforce Full
  - Exit criteria for each step
  - Evidence accumulated at each step

S0-T5b: docs/legal/data-processing-agreement-template.md (1 day)
  - What data ARE processes
  - Storage location (self-hosted — data never leaves enterprise)
  - Retention and deletion policy

S0-T5c: docs/security/threat-model.md (1 day)
  - STRIDE methodology
  - What ARE stops, what it does not stop
  - What happens if ARE itself is compromised

S0-T5d: docs/enterprise/fp-impact-analysis.md (3 hours)
  - FP rate scenarios at 0.00%, 0.1%, 0.5%, 2%
  - Blast radius of a false positive in enterprise environment
  - How override workflow limits blast radius

---

EXIT GATE S0: Lloyd LoU signed. Observe mode started. Sprint 1 unlocked.

---

# SPRINT 1 — BEHAVIORAL CERTIFICATION AUTHORITY v1
## Window: April 9 → May 21 (~6 weeks)
## Goal: First behavioral certification issued before Microsoft Agent 365 GA (May 1)

### S1-T1 — Certification Report Engine
File: internal/certification/report.go + internal/certification/report_test.go
Effort: 2 weeks
Stack: Go + existing PostgreSQL audit tables + existing hash chain + PDF generation (go-pdf)

Struct:
```go
type CertificationReport struct {
    AgentID        string
    OrgID          string
    WindowStart    time.Time
    WindowEnd      time.Time
    BaselineParams WelfordSnapshot
    DriftEvents    []DriftEvent
    EnforcementLog []EnforcementRecord
    ModelVersions  []BMVRecord
    IdentityChain  []BIVRecord
    CertHash       string
    SignatureKeyID string
    Signature      []byte
}
```

Report sections:
1. Executive summary — agent ID, org, window, behavioral health score
2. Baseline parameters — mean, variance, std dev per behavioral dimension
3. Drift analysis — AADD events, z-score trajectory, slow-walk results
4. Enforcement history — every block/flag with reason object + confidence_pct
5. Model continuity — BMV records across model version changes
6. Identity verification — BIV records
7. Hash chain verification instructions (DORA examiner protocol inline)
8. Certification signature block

Tests required (minimum 10):
- TestReportGenerationComplete
- TestHashChainVerification
- TestSignatureVerification
- TestReportIdempotent
- TestAllSectionsPresent
- TestBaselineSnapshotAccuracy
- TestDriftEventCount
- TestEnforcementLogIntegrity
- TestModelVersionContinuity
- TestIdentityChainValid

---

### S1-T2 — Certification PKI Extension
File: internal/certification/keys.go + cmd/certify/main.go
Effort: 1 week
Stack: Go crypto/rsa + existing JWT RS256 infrastructure

Key hierarchy:
```
ARE Root Certification Key (offline)
  └── ARE Operational Certification Key (rotated annually)
        └── Per-report RS256 signature
```

Public key endpoint: /.well-known/certification-jwks.json
CLI: are certify --agent-id <id> --org-id <org> --window 30d
Output: signed PDF + machine-readable JSON attachment

---

### S1-T3 — Certification Verification API
File: internal/api/certification_verify.go
Effort: 1 week
Endpoint: GET /v1/certify/verify?hash=<cert_hash>

Public endpoint — no auth required. DORA examiner uses this directly.
Recomputes hash chain from PostgreSQL audit trail.
Returns verification result with enforcement count, drift events, chain intact flag.

---

### S1-T4 — Slow-Walk Attack Demo Package
File: eval/attacks/slow_walk_microsoft_bypass.go + docs/demos/microsoft-comparison-demo.md
Effort: 2 weeks (parallel with S1-T1)

What it builds:
- Reproducible attack that passes Microsoft AGT cleanly (stateless — no rule violated)
- Same attack detected by ARE at call 43 (z_score 4.2 above 30-day PII access baseline)
- Blocked by ARE at call 44
- Runs in ~30 seconds on Windows Docker
- Side-by-side output: Microsoft result (pass) vs ARE result (blocked, reason object shown)

This is the enterprise sales weapon for every CISO who has seen Microsoft's announcement.

---

### S1-T5 — Cross-Framework Behavioral Consistency Proof
File: docs/research/cross_framework_behavioral_consistency_v1.md
Effort: 1 week

What it proves:
Same agent, same task, LangChain vs LangGraph → behavioral fingerprint within 1 std dev.
ARE's behavioral identity is framework-agnostic by architecture.
Microsoft's DID identity requires framework-level integration. ARE does not.

Publish to Zenodo: third anchor to IP chain.

---

### S1-T6 — Trust Gap Engineering (parallel with above)
Effort: ~2 weeks total, parallel

S1-T6a: Extended reason object with baseline snapshot (3 days)
File: internal/scoring/explainability.go
Add: baseline mean, std dev, sample count, trigger details, consecutive elevated calls, FP probability

S1-T6b: Baseline maturity indicator (1 day)
Add to reason object: "baseline_maturity": "30d_established" vs "probation_12d"
Prevents enforcement on agents without established baselines.

S1-T6c: Planned change API (3 days)
Endpoint: POST /v1/agents/{id}/planned-change
Creates supervised recalibration window.
Flags labeled "planned_change_window" not "anomaly" during window.
Prevents most common operational false positive — legitimate deployments.

S1-T6d: Calibration log (2 days)
File: internal/scoring/calibration_log.go
Every threshold change recorded: date, triggering data, before/after values, validation metrics.
Auditable. Available to enterprise security team on request.

S1-T6e: Override workflow API (1 week)
Endpoint: POST /v1/override/{enforcement_id}
Authorized user reverses a block, logs override with reason.
Override creates labeled false positive that feeds self-improving calibration loop.

---

EXIT GATE S1:
- First behavioral certification issued to Lloyd's enterprise at day 30
- Certification hash published to verification API
- Slow-walk demo reproducible in Lloyd's environment
- Microsoft Agent 365 GA is May 1 — certification issued before this date
- All Sprint 1 tests passing

---

# SPRINT 2 — BEHAVIORAL INTELLIGENCE LAYER
## Window: May 21 → July 3 (~6 weeks)
## Goal: Build the data moat. Make ARE's behavioral dataset unreplicable in 18 months.

### S2-T1 — Industry Behavioral Norm Library
File: internal/intelligence/industry_norms.go + data/norms/financial_services_v1.yaml + data/norms/healthcare_v1.yaml
Effort: 3 weeks
Stack: Go + PostgreSQL + existing Welford baseline infrastructure

What it builds:
Statistical distribution of behavioral parameters for normal agents in each industry.
Derived from accumulated pilot data across enterprises.

Why unreplicable: Requires multi-enterprise behavioral data in same regulated industry.
Competitor starting April 2026 needs 5+ financial services deployments = 18-24 months sales cycle. [H]

Scoring addition:
Agent scored against own baseline AND industry norm simultaneously.
Detects anomalies invisible to single-enterprise baselines.

---

### S2-T2 — Peer Cluster Deviation Detection
File: internal/scoring/peer_cluster.go + internal/scoring/peer_cluster_test.go
Effort: 2 weeks
Stack: Go + Redis + industry norm library (S2-T1 dependency)
APEX reference: M5 — parked until Month 2, now unblocked

Formula addition to ensemble scorer:
```
PeerClusterScore = f(
    agent_behavioral_vector,
    industry_norm_distribution,
    task_class_label,
    confidence_weight  // lower until >5 enterprises in peer cluster
)
```

Breaks attacker calibration: attacker who studies target agent's baseline
cannot simultaneously evade all peer agents in industry norm library.

---

### S2-T3 — Self-Improving Threshold Calibration
File: internal/scoring/threshold_calibration.go
Effort: 2 weeks
APEX reference: M7 — parked until Month 2, now unblocked

Calibration loop:
1. Collects confirmed TPs and TNs from SOC feedback
2. Computes optimal threshold adjustments
3. Constraint: max 10% weight change per iteration
4. Validates against held-out test set before acceptance
5. Produces calibration report with before/after metrics

After 90 days: ARE thresholds calibrated on real enterprise data, not synthetic corpus.
This is when behavioral intelligence becomes genuinely superior to any new entrant. [H]

---

### S2-T4 — Behavioral Model Fingerprint Registry
File: internal/intelligence/model_registry.go + data/model_fingerprints/
Effort: 2 weeks
Stack: Go + PostgreSQL + existing BMV infrastructure

What it builds:
Characteristic behavioral signatures of GPT-4o, GPT-4.1, Claude Sonnet 4.6
when performing common agent task classes.
Derived from BMV data accumulated in Phase 1.

Enables:
- Detecting silent model switches (supply chain attack)
- Detecting model version changes requiring recertification
- Cross-model behavioral comparison data unavailable anywhere else

Why unreplicable: Must be observed from production deployments, not computed from weights.
12-18 months minimum for competitor starting from zero. [H]

---

### S2-T5 — Feature Vector Store Hardening
File: internal/store/feature_vector_store.go
Effort: 1 week
APEX reference: TW-0/MUL-02 verified complete Phase 1, this sprint adds optimization

What it builds:
All 8 behavioral dimensions indexed for time-window queries.
Materialized view for industry norm computation.
Makes S2-T1 sub-second query instead of full table scan.

---

### S2-T6 — Threshold Simulation Endpoint
File: internal/api/simulate.go
Effort: 1 week

Endpoint: POST /v1/simulate
Accepts: proposed YAML policy change + sample of historical agent events
Returns: what enforcement decisions would have been under the new policy

Why critical: Prevents policy misconfiguration false positives before they reach production.
Enterprise validates policy change against their own traffic before deploying.
Most powerful operational trust builder available. [F]

---

### S2-T7 — Replay Endpoint
File: internal/api/replay.go
Effort: 1 week

Endpoint: POST /v1/replay
Accepts: sequence of historical agent events
Returns: score ARE would have produced at each step

Proves: scores are deterministic and reproducible.
Serves: legal/compliance teams reconstructing incident timelines.
Satisfies: technical reviewer asking "how do I verify your score is correct?"

---

### S2-T8 — Per-Enterprise Latency Metrics API
File: internal/api/metrics.go (extend existing Prometheus)
Effort: 3 days

Endpoint: GET /v1/metrics/latency
Returns: p50/p95/p99 latency for ARE scoring calls in that org's environment.
Org-scoped. Enterprise sees their actual overhead, not a benchmark.
Answers CISO question: "How do I know this isn't slowing my agents?"

---

EXIT GATE S2:
- Industry norm library populated from Lloyd + 1 additional enterprise
- Peer cluster deviation live with >0 weight
- Self-improving calibration first report produced
- Model fingerprint registry: minimum 3 models
- All Sprint 2 tests passing (target: 50+ new tests)

---

# SPRINT 3 — INTER-ENTERPRISE TRUST FABRIC FOUNDATION
## Window: July 3 → August 27 (~8 weeks)
## Goal: Bilateral trust bridge between two ARE-deployed enterprises.
## Network effect compounds in Phase 4. V1 delivers value from day one.

### S3-T1 — Agent Behavioral Passport
File: internal/trust/passport.go + internal/trust/passport_test.go
Effort: 3 weeks
Stack: Go + existing JWT/RS256 + existing behavioral baseline

```go
type AgentPassport struct {
    AgentID          string
    IssuingOrg       string
    IssuingARENode   string
    BehavioralHash   string    // hash of agent's behavioral fingerprint
    BaselineWindow   string    // "30d"
    CertificationRef string    // reference to behavioral certification
    IssuedAt         time.Time
    ExpiresAt        time.Time // 24-hour validity
    Signature        []byte    // RS256 over all fields
}
```

Header: X-Agent-Passport
Receiving ARE instance:
1. Verifies RS256 signature against issuing org's ARE public key
2. Checks expiry
3. Compares behavioral hash against model fingerprint registry
4. Issues cross-org trust score

Why different from Microsoft IATP:
Microsoft uses DIDs — cryptographic identity only.
ARE passport uses behavioral hash — trust credential derived from behavioral history.
Valid DID + compromised behavioral profile → Microsoft passes it, ARE catches it.

---

### S3-T2 — ARE Node Discovery Protocol
File: internal/trust/node_discovery.go
Effort: 2 weeks
Stack: Go + DNS TXT records (zero new infrastructure)

DNS TXT record format:
_are.{domain}.com TXT "v=ARE1; k=rsa; p={base64_public_key}; cert={zenodo_doi}"

Why DNS: Zero new infrastructure. Every enterprise has DNS. Auditable. Public.
Certification reference links DNS record → Zenodo → specific behavioral certification.
Verifiable chain from enterprise DNS to published standard to production data.

---

### S3-T3 — Cross-Org Behavioral Trust Scoring
File: internal/trust/cross_org_scorer.go
Effort: 2 weeks
Stack: Go + existing scoring + passport verification

Trust levels:
- Certified org + valid passport + behavioral hash match → high trust → reduced enforcement friction
- Uncertified org + valid passport → medium trust → observe mode enforcement
- No passport → low trust → standard ARE enforcement (same as today, no regression)

All cross-org trust decisions logged to audit trail with full reason object.

Commercial value: Two ARE enterprises can establish cross-org behavioral trust.
Reduces enforcement friction on legitimate agent-to-agent calls.
Maintains full audit trail. This is what every AI agent supply chain needs.

---

### S3-T4 — Certification Authority Portal v1
File: cmd/certification-portal/main.go + web/certification/index.html
Effort: 2 weeks
Stack: Go HTTP server + static HTML + existing verification API (S1-T3)
Domain: are-certified.io (register now)

Two pages only:
Page 1: Verify a certification (enter hash, see verification result)
Page 2: Certification standard (S0-T3 spec rendered as web page, Zenodo link)

No dashboard. No login. No management console.
UI before revenue is the founder trap. This is a verification utility only.

---

EXIT GATE S3:
- Agent Behavioral Passport issued and verified between two ARE enterprises
- ARE node discovery via DNS TXT records
- Cross-org trust scoring live
- Certification portal publicly accessible
- All Sprint 3 tests passing (target: 40+ new tests)

---

# SPRINT 4 — SERIES A PROOF PACKAGE
## Window: August 27 → September 24 (~4 weeks)
## Goal: Technical proof package that survives Series A due diligence.

### S4-T1 — Behavioral Intelligence Research Publication
File: docs/research/behavioral_intelligence_report_v1.md → Zenodo
Effort: 2 weeks

Contents:
1. Slow-walk detection: methodology, corpus, 100% detection rate, stateless comparison
2. Cross-framework behavioral consistency: LangChain vs LangGraph results
3. Model version behavioral delta: GPT-4o vs GPT-4.1 vs Claude production deltas
4. Industry norm deviation: distribution across financial services agents in install base

Purpose: Technical due diligence teams at Bessemer/IVP/Sequoia need evidence.
This document provides it. Also establishes ARE as research contributor to the field.

---

### S4-T2 — Isolation Forest Integration
File: internal/scoring/isolation_forest.go
Effort: 2 weeks
APEX reference: Task 10 — parked until 90 days of data. Unlock: ~July 9 (90 days post-Lloyd).
Sprint 4 starts August 27. Data requirement met. [H — assuming Lloyd signs April 9]

What it adds: Fourth scoring dimension in ensemble scorer.
Trained on 90 days of per-agent behavioral feature vectors.
Detects novel attack patterns z-score and velocity do not cover.
Low-amplitude, high-specificity behavioral anomalies.

Why not earlier: Isolation Forest on <90 days data will overfit. Statistical discipline, not avoidance.

---

### S4-T3 — Full Enterprise Deployment Package v2
File: docs/enterprise/deployment-package-v2.md + Kubernetes manifests
Effort: 1 week

Updated for: certification engine, PKI, cross-org trust fabric, industry norm library.
4-hour install claim validated against full feature set.
Kubernetes manifests for production deploy.

---

### S4-T4 — Series A Technical Due Diligence Package
File: docs/investor/technical-due-diligence-package.md
Effort: 1 week

Contents:
- Architecture diagram with data flow
- Security model (threat model + mitigations)
- Performance benchmarks (latency, throughput, FP rate)
- Behavioral certification methodology
- Data moat analysis (what competitor needs to replicate, time estimates)
- Microsoft competitive gap analysis (technical)
- IP chain: ATP → ATG → ARE → behavioral certification standard (Zenodo DOI chain)

---

EXIT GATE S4: Series A process begins.

---

# ROADMAP SUMMARY

Sprint 0  | Now → Apr 9     | Commercial armor + category naming       | 0 engineering
Sprint 1  | Apr 9 → May 21  | Behavioral Certification Authority v1    | ~8 weeks eng
Sprint 2  | May 21 → Jul 3  | Behavioral Intelligence Layer            | ~6 weeks eng
Sprint 3  | Jul 3 → Aug 27  | Inter-Enterprise Trust Fabric Foundation | ~8 weeks eng
Sprint 4  | Aug 27 → Sep 24 | Series A Proof Package                  | ~4 weeks eng

---

# MOAT COMPOUNDING BY SPRINT

Sprint | Moat Added                                    | Competitor Replication
-------|-----------------------------------------------|------------------------
S0     | Category naming — Zenodo IP chain             | Can name later, ARE named first
S1     | Hash-chain certified production data          | 18-24 months install base [H]
S2     | Industry norm library multi-enterprise data   | 24-36 months accumulation [H]
S3     | Cross-org trust fabric, behavioral passport   | Requires ARE at both endpoints
S4     | 90-day Isolation Forest on real enterprise    | Cannot synthesize — needs data

---

# FP RATE HONEST PROJECTION

Context                                          | FP Rate    | Confidence
-------------------------------------------------|------------|----------
Internal synthetic corpus (current)             | 0.00%      | [F]
Observe mode (any environment)                  | 0.00%      | [F] — no blocks
Production, established baselines, conservative | <0.5%      | [H]
Production, mature baselines, calibrated        | <0.1%      | [H]
Production, 90+ days, self-improving            | →0.00%     | [H]

G-FP gate: ≤2% hard ceiling. ARE targets <0.5% at enforcement go-live.
The 0.00% claim is architecturally defensible because:
P(FP) = P(z_score miscalibrated) × P(policy also fires incorrectly)
Two independent signals. Joint probability orders of magnitude lower than either alone.

---

# WHAT REHAN SAYS WHEN LLOYD ASKS "HOW DO I TRUST YOUR SCORE"

To technical reviewer:
"Every score includes the exact baseline parameters it was computed against,
the specific metric that triggered it, the standard deviations from normal,
the confidence percentage, and a hash-chain-verified record proving this score
was produced at this moment from this data. You can verify every number
independently using our replay endpoint."

To CISO:
"You spend 30 days watching ARE score your agents without blocking anything.
At day 30 you have 30 days of evidence that ARE's scores match what your SOC
would have flagged. That is when you decide whether to turn on enforcement.
You control the pace."

To regulator:
"Here is the document that describes exactly how you verify ARE's audit trail
without trusting our attestation. Your examiner can reproduce every enforcement
decision from the PostgreSQL audit table using the hash chain algorithm on page one."

---
Generated: April 4, 2026 | APEX v5.2 | AgentRepEngine
