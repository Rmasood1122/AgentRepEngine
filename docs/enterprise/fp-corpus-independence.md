# False Positive Corpus Independence Statement
## AgentRepEngine (ARE) | Naseem A2A Research Lab
## Version 2.0 | April 2026

---

## 1. Purpose

This document describes the construction methodology, behavioral coverage, and
statistical properties of the corpus used to validate ARE's false positive rate.
It is intended for security engineers, CISOs, and compliance reviewers conducting
technical due diligence on ARE's enforcement claims.

---

## 2. Construction Methodology

### 2.1 Corpus Structure

The FP validation corpus consists of 150 synthetic scenarios across two sets:

- **Base corpus (100 scenarios):** LEG-001 through LEG-100 in
  `tests/fp_scenarios/scenarios.go`. These cover 7 categories of legitimate
  agent behavior at normal-to-elevated operating levels.

- **Boundary corpus (50 scenarios):** BND-001 through BND-050 in
  `tests/fp_scenarios/boundary_scenarios.go`. These target the z-score 4.0–5.5
  range — the statistical boundary where legitimate high-activity agents are
  most likely to be misclassified. The boundary corpus is the most stringent
  test in the validation suite.

### 2.2 Construction Protocol

Each scenario defines a behavioral feature vector representing a legitimate
agent workflow. The vector specifies rates for up to 8 behavioral dimensions
(tool call rate, unique endpoints, bulk access count, PII field access rate,
cross-tenant probes, permission escalations, sub-agent spawn depth, token
refresh rate). Each scenario also specifies a minimum expected score (MinScore)
— the lowest score the agent should receive while remaining above the
enforcement threshold.

The boundary corpus includes inline behavioral specifications documenting:
the agent type, why the behavior is legitimate, which feature(s) are elevated,
and the z-score math showing where the scenario falls relative to the
enforcement threshold. These specifications are embedded in the source file
as structured comments preceding each scenario group.

### 2.3 Spec-First Protocol — Status and Limitations

The spec-first protocol — where a behavioral specification is committed to a
separate file in `tests/fp_scenarios/specs/` before the scenario implementation
— is defined as the required protocol for all future corpus expansion (150→300).

**Current state:** The base corpus (LEG-001 to LEG-100) was constructed by
category without formal specification lockdown files. The boundary corpus
(BND-001 to BND-050) has inline specifications in the source file documenting
behavioral intent, z-score calculations, and baseline math. Separate spec files
in `tests/fp_scenarios/specs/` do not exist for the current 150 scenarios.

**Implication:** The current corpus does not have a verifiable temporal
separation between specification and implementation via commit history. This
is a known limitation. Co-design bias cannot be ruled out by audit trail alone.
The mitigation is the boundary corpus itself: scenarios at z-score 4.0–5.5 are
the hardest possible test for the scoring model, and their inclusion at 33% of
the corpus (50/150) makes co-design bias materially harder to exploit.

### 2.4 Additive-Only Rule

The corpus is append-only. No scenario has been modified or removed after
initial commit. All 150 scenarios exist in their original form at their original
commit hashes. This is verifiable via `git log tests/fp_scenarios/`.

### 2.5 Scenario Authorship

All 150 scenarios were authored by the ARE engineering team. No external party
contributed scenarios. This is disclosed explicitly: the corpus is internally
constructed and has not been independently validated by a third party. External
validation on production traffic is available during the pilot engagement.

---

## 3. Behavioral Categories Covered

The 150-scenario corpus spans 11 categories of legitimate enterprise agent
behavior. Categories were derived from the actual scenario names and behavioral
vectors in the source code.

### 3.1 Base Corpus Categories (100 scenarios)

| # | Category | Scenarios | IDs | Description |
|---|----------|-----------|-----|-------------|
| 1 | Data analyst operational variants | 20 | LEG-001–020 | Normal through peak-load analyst workflows — quarter-end, parallel reports, ad-hoc queries |
| 2 | Authorized bulk export agents | 15 | LEG-021–035 | Compliance exports, backup, GDPR subject access, billing, KYC, audit log pulls |
| 3 | Research and discovery agents | 15 | LEG-036–050 | High endpoint counts — competitor analysis, patent search, regulatory lookup, threat intel |
| 4 | Cold-start and new agents | 15 | LEG-051–065 | First 48 hours through first week — light usage, smoke tests, onboarding, pilot deployment |
| 5 | Recovering agents (post-incident) | 15 | LEG-066–080 | Day 1 recovery through full clearance — gradual ramp, supervised ops, probation exit |
| 6 | Batch and scheduled job agents | 10 | LEG-081–090 | Nightly syncs, weekend reports, billing reconciliation, cache warming, compliance scans |
| 7 | Diverse operational agents | 10 | LEG-091–100 | Customer service, monitoring, code review, email triage, CRM, log analysis, security scanning |

### 3.2 Boundary Corpus Categories (50 scenarios)

| # | Category | Scenarios | IDs | Z-score Range | Description |
|---|----------|-----------|-----|---------------|-------------|
| 8 | Batch job agents at boundary | 10 | BND-001–010 | tool_call z=4.0–4.7 | ETL bursts, billing close, audit extracts at enforcement threshold |
| 9 | High-endpoint monitoring agents | 10 | BND-011–020 | endpoints z=4.0–5.0 | Service mesh monitors, SLA trackers, security scanners touching 145–175 endpoints/hr |
| 10 | High-PII authorized agents | 10 | BND-021–030 | pii z=4.0–5.3 | Medical records, clinical trials, insurance, AML, fraud investigation agents |
| 11 | Combined multi-feature high-load | 10 | BND-031–040 | Multiple features z=3.5–4.5 | Data pipeline, compliance, financial consolidation with 2+ elevated features |
| 12 | Reactivation agents | 10 | BND-041–050 | Low H, normal V | Agents dormant 14–45 days, returning to normal operation with decayed history scores |

### 3.3 Category Distribution Rationale

The corpus overweights boundary scenarios (50/150 = 33.3%) because the scoring
threshold boundary is where FP risk concentrates. A corpus that underweights
boundary scenarios produces an artificially favorable FP rate. The boundary
corpus includes scenarios with MinScore=500 — the exact enforcement threshold
— which is the hardest possible test.

---

## 4. Statistical Properties

### 4.1 Observed Result

Zero false positives across all 150 scenarios. No legitimate agent was scored
into the RESTRICTED (200–499) or BLOCKED (0–199) band in any scenario.

### 4.2 Confidence Interval Disclosure

A zero observed result on a finite corpus does not prove a zero population FP
rate. The correct statistical statement uses the Clopper-Pearson exact binomial
method:

> "Zero false positives on a 150-scenario internal validation corpus, bounding
> the FP rate below **2.0% with 95% confidence** (Clopper-Pearson exact interval)."

Precise Clopper-Pearson 95% upper bound for 0/150: **1.97%**

This is ARE's Tier 1 qualified claim. It is the claim that is statistically
defensible today.

### 4.3 Path to Tier 2 Claim

Expanding the corpus to 300 scenarios (150 additional, additive-only, spec-first
with separate specification files) reduces the 95% CI upper bound to
approximately 1.0%. This expansion is planned for the post-LoU sprint. Upon
completion, the defensible claim becomes:

> "Zero false positives on a 300-scenario internal validation corpus, bounding
> the FP rate below **1.0% with 95% confidence**."

### 4.4 Production Validation Path (Tier 3)

Neither the Tier 1 nor Tier 2 claim is a production FP rate. Production
validation requires real agent traffic. The 30-day observe-mode pilot is
designed to produce the first production FP measurement:

> "Zero false positives on [enterprise]'s production traffic — [N] agents,
> [M] requests, 30 days. Production FP rate: [measured]."

This is Tier 3. It is the claim that closes the synthetic-to-production gap.

---

## 5. Boundary Scenario Coverage

The 50 boundary scenarios (BND-001 to BND-050) cover agents operating at
z-score 4.0–5.5 — the zone immediately below and at the enforcement threshold.

Score boundary math (from scoring formula):
- Score = round(0.5 × H + 0.5 × V)
- V = max(0, 1000 − penalty × 3)
- penalty = min(100 × (worstZ − 3.0), 300), only when worstZ > 3.0
- Score = 500 requires V = 300, penalty = 233, worstZ = 5.33

All 50 boundary scenarios produce scores at or above 500 (MONITORED band) at
the current threshold configuration. 22 of 50 boundary scenarios have
MinScore = 500 — the exact enforcement boundary. These are the scenarios most
likely to expose threshold sensitivity and co-design bias.

---

## 6. Independent Verification

The full corpus is available for review at:

- `tests/fp_scenarios/scenarios.go` — 100 base scenarios
- `tests/fp_scenarios/boundary_scenarios.go` — 50 boundary scenarios (with inline specs)
- `tests/fp_scenarios/fp_test.go` — scoring function and FP gate test
- `tests/fp_scenarios/boundary_fp_test.go` — boundary-specific FP test
- `tests/fp_scenarios/fp_bias_audit_test.go` — bias audit checks
- `tests/eval_harness/harness_test.go` — full evaluation harness

Any reviewer can reproduce the result by running:

```bash
go test ./tests/fp_scenarios/... -v
go test ./tests/eval_harness/... -v
```

The test output includes per-scenario pass/fail and the aggregate FP count.

---

## 7. What This Document Does Not Claim

- This document does not claim ARE has a zero FP rate in production.
- This document does not claim the corpus was constructed by an independent party.
- This document does not claim the corpus has verifiable spec-first temporal
  separation via commit history (see Section 2.3).
- This document does not claim the synthetic scenarios represent the full
  distribution of legitimate enterprise agent behavior.
- The Zenodo DOI (10.5281/zenodo.19169185) establishes timestamped existence
  of ARE's architectural framework. It does not validate the FP methodology.

These gaps are closed by production pilot data. The 30-day observe-mode
deployment is the designed mechanism for closing them.

---

## 8. Version History

| Version | Date | Change |
|---------|------|--------|
| 1.0 | April 12, 2026 | Initial release — 100-scenario corpus, Tier 1 claim |
| 2.0 | April 16, 2026 | Updated to 150 scenarios (100 base + 50 boundary), 11 categories from code, corrected CI bounds (3.6%→2.0%), honest spec-first status disclosure |
| 3.0 | Post-LoU | 300-scenario corpus, Tier 2 claim, formal spec-first provenance (planned) |
| 4.0 | Post-pilot | Production FP rate, Tier 3 claim (planned) |

---

*AgentRepEngine is developed by Naseem A2A Research Lab / Call2leads Inc.*
*Contact: rehanrana@call2leads.com*
*Repository: github.com/Rehanrana11/AgentRepEngine*
*IP anchor: DOI 10.5281/zenodo.19169185*
