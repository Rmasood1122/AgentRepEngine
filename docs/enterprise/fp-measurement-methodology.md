# Spec-First Corpus Construction for AI Agent Behavioral Enforcement: A False Positive Measurement Methodology

**Author:** Rehan Rana, Naseem A2A Research Lab / Call2leads Inc.

**Date:** April 2026

**DOI:** 10.5281/zenodo.19169185

**Repository:** github.com/Rehanrana11/AgentRepEngine

**Status:** Submitted for independent methodology review

---

## Abstract

Runtime behavioral enforcement systems for autonomous AI agents face a
measurement credibility problem: how does an operator verify that an
enforcement layer will not incorrectly block legitimate agent operations?
This paper presents a measurement methodology for false positive rates in
AI agent behavioral enforcement, consisting of three components: (1) a
spec-first corpus construction protocol that creates verifiable independence
between test design and system calibration, (2) a statistical framework
using Clopper-Pearson exact binomial confidence intervals for bounding FP
rates on finite corpora, and (3) a graduated claim ladder mapping evidence
strength to permissible assertions. Applied to AgentRepEngine (ARE), a
gateway-layer behavioral enforcement system, the methodology yields a
measured false positive rate of 0.00% (0/150 scenarios), bounded below
2.0% at 95% confidence. We describe the corpus structure, its 11
behavioral categories, known limitations including co-design bias, and
the production measurement protocol that extends synthetic validation to
live enterprise traffic.

---

## 1. Introduction

### 1.1 The Problem

AI agent deployment in regulated enterprises — financial services,
healthcare, government — increasingly requires behavioral enforcement:
the ability to detect and restrict anomalous agent actions at runtime.
Unlike traditional application security, AI agents exhibit emergent
behavioral patterns that resist static policy definitions. A trading
agent that normally accesses 20 PII fields per hour accessing 847 in
90 minutes may represent a breach — or an authorized end-of-quarter
reconciliation.

Behavioral enforcement systems score agent actions against historical
baselines and enforce access restrictions when anomalies are detected.
The critical failure mode is the false positive: a legitimate agent
action that is incorrectly flagged, throttled, or blocked. In
enterprise production, false positives are operationally more damaging
than false negatives. A missed anomaly is a security risk. A blocked
legitimate workflow is a business disruption that kills pilot adoption.

Despite this, no AI agent behavioral enforcement product has published
a formal false positive measurement methodology. This gap undermines
enterprise adoption: regulated buyers cannot assess enforcement quality
without reproducible, statistically grounded measurement.

### 1.2 Contributions

This paper makes four contributions:

1. **Spec-first corpus construction protocol** — a methodology for
   building behavioral validation scenarios where the behavioral
   specification is documented before scenario implementation,
   creating a traceable design-to-test chain.

2. **Behavioral category taxonomy** — an 11-category classification
   of legitimate enterprise agent behaviors covering the operational
   space where false positives are most likely.

3. **Confidence interval framework** — application of Clopper-Pearson
   exact binomial confidence intervals to FP rate measurement, providing
   statistically rigorous bounds for regulated enterprise reporting.

4. **Graduated claim ladder** — a 4-tier evidence framework mapping
   corpus size, data source, and statistical confidence to permissible
   performance assertions.

### 1.3 Scope and Limitations

This methodology is validated on AgentRepEngine (ARE), a gateway-layer
behavioral enforcement system implemented as a Kong API Gateway plugin
with a Go scoring service, Redis cache, and PostgreSQL audit store.
The validation corpus consists of 150 synthetic scenarios. The
methodology is designed to be applicable to other behavioral
enforcement systems, but the current validation is specific to ARE's
scoring architecture.

**Key disclosure:** The test corpus is synthetic and internally
constructed. It does not contain production enterprise traffic. All
performance claims are qualified by corpus type and size per the claim
ladder in Section 10.

---

## 2. System Under Test

### 2.1 Scoring Architecture

ARE implements per-agent behavioral baseline scoring. Each agent is
identified by a signed JWT (RS256, org-scoped JWKS). Behavioral events
are scored against the agent's own 30-day rolling baseline using z-score
anomaly detection with exponentially weighted moving averages (EWMA).

The composite score is computed as:

    Score = round(W_H * H + W_V * V)

Where:
- **H (History score):** Range 0-1000. Decays with inactivity
  (7-day half-life). Rewards sustained normal behavior.
- **V (Velocity score):** Range 0-1000. Computed as
  V = max(0, 1000 - penalty * 3), where
  penalty = min(100 * (worstZ - 3.0), 300) for worstZ > 3.0.
- **W_H, W_V:** Phase 1 weights: W_H = 0.5, W_V = 0.5.

Policy violations apply instant penalties regardless of H or V
(ceiling override model — behavioral history cannot shield active
policy violations).

### 2.2 Enforcement Bands

| Band | Score Range | Enforcement Action |
|------|------------|-------------------|
| TRUSTED | 700-1000 | ALLOW |
| MONITORED | 500-699 | ALLOW + active audit logging |
| RESTRICTED | 200-499 | THROTTLE + human review on HIGH_RISK |
| BLOCKED | 0-199 | Synthetic response (never HTTP 403) |

A false positive is defined as a legitimate agent action that receives
a RESTRICTED or BLOCKED band assignment (score < 500).

### 2.3 Behavioral Feature Vector

ARE scores agents across 8 behavioral dimensions:

| Feature | Baseline Mean | Baseline Std | Unit |
|---------|--------------|-------------|------|
| tool_call_rate_per_hour | 80 | 60 | calls/hr |
| unique_endpoints_per_hour | 25 | 30 | endpoints/hr |
| bulk_access_count_per_session | 400 | 400 | records/session |
| pii_field_access_rate | 20 | 25 | fields/hr |
| cross_tenant_probe_count | 0 | 0.1 | probes/hr |
| permission_escalation_count | 0 | 0.5 | escalations/hr |
| sub_agent_spawn_depth | 0 | 0.3 | depth |
| token_refresh_rate | 1 | 1 | refreshes/hr |

These baselines represent bootstrap values for new agents. In
production, each agent develops its own baseline from observed behavior.

---

## 3. Spec-First Corpus Construction Protocol

### 3.1 Motivation

The central threat to corpus validity is **co-design bias**: the
scenario author, knowing how the scoring system works, unconsciously
constructs scenarios that the system handles well. This produces an
artificially low FP rate that does not generalize to production traffic.

### 3.2 The Protocol

The spec-first protocol defines a four-step process:

**Step 1 — Behavioral Specification:** Write a specification describing
the agent type, expected behavioral pattern, business justification,
and the specific reason the behavior might trigger a false positive.

**Step 2 — Specification Lockdown:** Commit the specification to
version control. No modifications permitted after commit.

**Step 3 — Scenario Implementation:** Implement the test scenario
in a subsequent commit, referencing the specification commit hash.

**Step 4 — Independence Verification:** Verify that the specification
commit predates the implementation commit.

### 3.3 Current Implementation Status

The spec-first protocol is fully defined but partially implemented in
ARE's current 150-scenario corpus:

- **Base corpus (100 scenarios):** Constructed by behavioral category
  without formal specification lockdown files. Specifications exist as
  the category design rationale but not as per-scenario committed
  documents.

- **Boundary corpus (50 scenarios):** Contains inline behavioral
  specifications as structured source code comments, documenting agent
  type, z-score math, baseline calculations, and behavioral intent.
  These specifications are co-located with the implementation (not in
  separate files with independent commit timestamps).

- **Formal spec files** in a dedicated specifications directory do not
  exist for the current corpus.

**Implication:** The current corpus does not have verifiable temporal
separation between specification and implementation via commit history.
Co-design bias cannot be ruled out by audit trail alone. The protocol
is defined as the required standard for all future corpus expansion
(150 to 300 scenarios).

### 3.4 Co-Design Bias Mitigation

In the absence of formal spec-first provenance, the boundary corpus
(50 scenarios at z-score 4.0-5.5) serves as the primary co-design
bias control. These scenarios operate at the exact statistical
boundary where false positives would occur, with 22 of 50 scenarios
having MinScore = 500 (the enforcement threshold). A co-designed
corpus that deliberately avoids the boundary zone is materially
weaker; the inclusion of boundary scenarios at 33% of corpus
(50/150) raises the bar for co-design exploitation.

Full mitigation requires external corpus construction by a party
without access to the scoring algorithm (see Section 10, Tier 3/4).

---

## 4. Behavioral Category Taxonomy

The 150-scenario corpus spans 11 categories of legitimate enterprise
agent behavior. Categories are derived from the scenario definitions
in the codebase.

### 4.1 Base Corpus Categories (100 scenarios)

**Category 1: Data Analyst Operational Variants (20 scenarios)**
Normal through peak-load analyst workflows: quarter-end processing,
parallel reports, ad-hoc query bursts, model feature extraction,
cohort analysis. Tool call rates from 50 to 140/hr.

**Category 2: Authorized Bulk Export Agents (15 scenarios)**
Compliance exports, backup, GDPR subject access, billing
reconciliation, KYC data pulls, audit log extraction. Bulk access
counts from 300 to 2,000 records/session. PII access rates from
10 to 90 fields/hr.

**Category 3: Research and Discovery Agents (15 scenarios)**
High unique endpoint counts (40-110/hr). Competitor analysis, patent
search, regulatory lookup, threat intelligence, financial data
retrieval.

**Category 4: Cold-Start and New Agents (15 scenarios)**
First 48 hours through first week of operation. Light usage
(5-38 calls/hr), smoke tests, onboarding workflows, pilot
deployments. Tests bootstrap protocol behavior.

**Category 5: Recovering Agents (15 scenarios)**
Post-incident recovery from day 1 (15 calls/hr) through full
clearance (72 calls/hr). Gradual ramp, supervised operations,
probation exit.

**Category 6: Batch and Scheduled Job Agents (10 scenarios)**
Nightly syncs, weekend reports, monthly billing reconciliation,
cache warming, compliance scans. Burst patterns at 55-120 calls/hr.

**Category 7: Diverse Operational Agents (10 scenarios)**
Customer service, monitoring, code review, document summarization,
email triage, calendar scheduling, CRM updates, log analysis,
infrastructure provisioning, security scanning.

### 4.2 Boundary Corpus Categories (50 scenarios)

All boundary scenarios target z-score 4.0-5.5 on at least one
behavioral feature.

**Category 8: Batch Jobs at Threshold (10 scenarios)**
Tool call rates 310-360/hr (z = 3.8-4.7). Bulk access 1,500-2,400
(z = 2.75-5.0). The highest-load legitimate batch operations.

**Category 9: High-Endpoint Monitoring Agents (10 scenarios)**
Unique endpoints 145-175/hr (z = 4.0-5.0). Service mesh monitors,
SLA trackers, security posture scanners. These agents probe entire
service meshes by design.

**Category 10: High-PII Authorized Agents (10 scenarios)**
PII field access rates 118-145/hr (z = 3.9-5.0). Medical records,
clinical trials, insurance underwriting, GDPR requests, AML
monitoring, credit scoring.

**Category 11: Combined Multi-Feature High-Load (10 scenarios)**
Two or more features simultaneously elevated at z = 3.5-4.5. Data
pipelines, financial consolidation, risk aggregation, payroll,
customer onboarding.

**Category 12: Reactivation Agents (10 scenarios)**
Agents dormant 14-45 days returning to normal operation. H score
has decayed (7-day half-life); V score must compensate. Normal
behavioral vectors but low composite scores due to history decay.
These are the hardest FP scenarios.

### 4.3 Distribution Rationale

The corpus overweights boundary scenarios (50/150 = 33.3%) because
the enforcement threshold boundary is where FP risk concentrates.
22 of 50 boundary scenarios have MinScore = 500 — the exact
enforcement boundary.

---

## 5. Results

### 5.1 Aggregate Performance

| Metric | Value | Scope |
|--------|-------|-------|
| FP rate | 0.00% (0/150) | 150-scenario FP corpus |
| TP rate | 88.00% (44/50) | 50-scenario attack corpus |
| Precision | 100% | Full corpus (zero false blocks) |
| F1 score | 0.9362 | Full corpus |
| Held-out FP | 0.00% (0/20) | 20 held-out normal scenarios |
| Held-out TP | 100.00% (6/6) | 6 held-out anomalous scenarios |
| Slow-walk detection | 100% (10/10) | Single-agent scope |

### 5.2 Qualified FP Claim

**Stated claim:** "Zero false positives across 150 synthetic scenarios,
bounding FP rate below 2.0% at 95% confidence (Clopper-Pearson exact
method)."

This is the correct qualified claim. The bare "0.00%" is never used
without this qualification in external communication.

---

## 6. Confidence Interval Framework

### 6.1 Why Clopper-Pearson

The Clopper-Pearson method computes exact binomial confidence intervals
without normal approximation. For extreme proportions (0% or 100%),
normal approximation (Wald interval) produces nonsensical results
(negative bounds). The Clopper-Pearson method is conservative by
construction — it guarantees at least the stated coverage probability.

### 6.2 Calculation

For k = 0 observed false positives in n = 150 trials, the
Clopper-Pearson one-sided 95% upper bound is computed from the Beta
distribution:

    p_upper = Beta.InverseCDF(1 - alpha, k + 1, n - k)
            = Beta.InverseCDF(0.95, 1, 150)
            = 0.01974

### 6.3 Results by Corpus Partition

| Corpus | n | k (FPs) | 95% CI Upper | 99% CI Upper |
|--------|---|---------|-------------|-------------|
| Base corpus | 100 | 0 | 3.60% | 5.57% |
| Boundary corpus | 50 | 0 | 5.82% | 8.90% |
| Full FP corpus | 150 | 0 | 1.97% | 3.04% |
| Full + held-out | 170 | 0 | 1.74% | 2.69% |

### 6.4 What the CI Does Not Mean

The 95% CI upper bound of approximately 2.0% does not mean ARE will
produce 2.0% FPs in production. It means: given 0 observed FPs in
150 trials, we cannot statistically rule out a true rate above 2.0%
with 95% confidence.

Production FP rate will be measured during the first enterprise pilot.
ARE's auto-rollback mechanism enforces a hard 2% production FP ceiling
(see Section 8).

---

## 7. What the Corpus Does Not Cover

1. **Real production agent traffic:** All scenarios are synthetic.
   Production behavioral distributions may differ.

2. **Multi-tenant baseline contamination:** The corpus tests single-org
   deployments.

3. **Non-finserv agent archetypes:** The corpus is calibrated for
   financial services and general enterprise. Healthcare and government
   agents may require threshold recalibration.

4. **Long-horizon behavioral drift:** Scenarios test instantaneous
   behavioral vectors, not multi-day evolution. Drift over 90+ days
   is not covered.

5. **Co-design bias:** The corpus was constructed by the same team that
   built the scoring system (see Section 3.3).

These limitations define the scope of the 0.00% claim. The claim is
valid within its stated scope. It does not extend beyond it.

---

## 8. Auto-Rollback Circuit Breaker

ARE includes a production FP circuit breaker:

- StartFPMonitor() checks FP rate every 5 minutes
- If FP exceeds 2%: automatic rollback to observe mode + SIEM alert
- No restart required. Mode change is instant.
- No vendor intervention required.

Even if the production FP rate exceeds the corpus prediction, ARE
self-limits before operational impact.

---

## 9. Production Measurement Protocol

### 9.1 Observe-Mode Measurement

During the 30-day observe-mode pilot, every RESTRICTED or BLOCKED
decision is logged to an fp_candidates table for human review. The
customer's security team marks each candidate as confirmed FP or
confirmed TP. ARE does not control this measurement.

### 9.2 Customer-Verifiable FP Query

```sql
SELECT
  COUNT(*) FILTER (WHERE confirmed_fp = true) AS false_positives,
  COUNT(*) AS total_enforced,
  ROUND(
    100.0 * COUNT(*) FILTER (WHERE confirmed_fp = true)
    / NULLIF(COUNT(*), 0), 4
  ) AS fp_rate_pct
FROM fp_candidates
WHERE flagged_at >= NOW() - INTERVAL '30 days';
```

### 9.3 Audit Trail Integrity

All enforcement decisions are stored in a tamper-evident SHA-256 hash
chain. Chain integrity is independently verifiable via CLI or SQL.

---

## 10. Graduated Claim Ladder

| Tier | Evidence | Permissible Claim |
|------|----------|-------------------|
| Tier 1 (current) | 0 FPs on 150 synthetic scenarios | "0.00% FP, bounded below 2.0% at 95% CI" |
| Tier 2 (planned) | 0 FPs on 300 scenarios, spec-first verified | "0.00% FP, bounded below 1.0% at 95% CI" |
| Tier 3 (post-pilot) | 0 FPs on production traffic, 30 days | "Production FP: [measured]" |
| Tier 4 (post-90-days) | Continuous monitoring, auto-rollback never triggered | "FP monitored. Auto-rollback never fired in [N] days." |

Each tier requires all evidence from prior tiers. No tier permits the
bare claim "0.00% false positive rate" without qualification.

---

## 11. Comparison to Industry Practice

| System | FP Rate | Corpus | Published? |
|--------|---------|--------|-----------|
| ARE (this paper) | 0.00% (150 scenarios) | Synthetic | Yes |
| ARE production target | < 0.1% | Production | Pending |
| Visa fraud detection | ~0.1% | Production | Industry |
| Traditional SIEM | 40-60% | Production | Industry |
| ML-based UEBA | 5-15% | Production | Vendor |
| AI agent security vendors | Not disclosed | Unknown | No |

No AI agent behavioral enforcement product has published a measurement
methodology at this level of specificity.

---

## 12. Reproducibility

All results can be independently verified:

```bash
git clone https://github.com/Rehanrana11/AgentRepEngine.git
cd AgentRepEngine

# FP corpus (150 scenarios)
go test ./tests/fp_scenarios/... -v

# Full evaluation harness
go test ./tests/eval_harness/... -v

# Attack corpus (50 scenarios)
go test ./tests/attack_corpus/... -v

# Held-out sets
go test ./tests/held_out/... -v
```

---

## Acknowledgments

Independent methodology review by Dr. Momina Moetesum (NUST, IAPR-TC10)
is gratefully acknowledged. [Pending review completion.]

---

## References

1. Clopper, C.J. and Pearson, E.S. (1934). "The Use of Confidence or
   Fiducial Limits Illustrated in the Case of the Binomial."
   Biometrika, 26(4), pp. 404-413.

2. NIST AI Risk Management Framework (AI RMF 1.0). National Institute
   of Standards and Technology, January 2023.

3. NIST SP 800-207: Zero Trust Architecture. National Institute of
   Standards and Technology, August 2020.

4. OWASP Top 10 for LLM Applications. Open Worldwide Application
   Security Project, 2025.

5. Regulation (EU) 2022/2554 (DORA). Digital Operational Resilience
   Act, European Parliament, November 2022.

---

*AgentRepEngine FP Measurement Methodology v2.0*
*Naseem A2A Research Lab / Call2leads Inc.*
*April 2026*
*DOI: 10.5281/zenodo.19169185*
