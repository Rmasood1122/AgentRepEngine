# False Positive Corpus Independence Statement
## AgentRepEngine (ARE) | Naseem A2A Research Lab
## Version 1.0 | April 2026

---

## 1. Purpose

This document describes the construction methodology, behavioral coverage, and
statistical properties of the corpus used to validate ARE's false positive rate.
It is intended for security engineers, CISOs, and compliance reviewers conducting
technical due diligence on ARE's enforcement claims.

---

## 2. Construction Methodology

### 2.1 Spec-First Protocol

Every scenario in the FP validation corpus was constructed using a spec-first
protocol:

1. A behavioral specification was written describing a legitimate agent workflow
   (agent type, action sequence, rate parameters, org context).
2. The specification was committed to the repository before the scenario code
   was written. The spec commit hash is referenced in each scenario file.
3. The scenario was then implemented to match the spec — not to pass the scorer.

This protocol prevents co-design bias: scenarios cannot be tuned to match the
scorer's thresholds because the behavioral specification is locked before the
implementation begins.

### 2.2 Additive-Only Rule

The corpus is append-only. No scenario has been modified or removed after
initial commit. All 100 scenarios exist in their original form at their original
commit hashes. This is verifiable via `git log tests/fp_scenarios/`.

### 2.3 Scenario Authorship

All 100 scenarios were authored by the ARE engineering team. No external party
contributed scenarios. This is disclosed explicitly: the corpus is internally
constructed and has not been independently validated by a third party. External
validation on production traffic is available during the pilot engagement.

---

## 3. Behavioral Categories Covered

The 100-scenario corpus spans the following legitimate agent behavioral patterns:

| Category | Scenarios | Description |
|---|---|---|
| Normal operational baseline | 25 | Agents operating within 1σ of their established baseline |
| Bursty-but-legitimate | 20 | Batch jobs, scheduled tasks, end-of-period processing spikes |
| Cold-start agents | 15 | New agents with no prior baseline, day-1 behavior |
| Sequential workflow agents | 15 | Slow, ordered, non-concurrent API call patterns |
| High-frequency legitimate | 10 | Monitoring agents, trading bots, high-rate legitimate activity |
| Cross-reference legitimate | 10 | Agents with approved multi-system access patterns |
| Legacy behavioral patterns | 5 | Unusual endpoint ordering, non-standard but authorized sequences |

All 100 scenarios represent agents that ARE **must not block**. The corpus
tests for false positives only — it is not an attack detection corpus.

---

## 4. Statistical Properties

### 4.1 Observed Result

Zero false positives across all 100 scenarios. No legitimate agent was blocked,
throttled, or flagged for human review across any scenario in the corpus.

### 4.2 Confidence Interval Disclosure

A zero observed result on a finite corpus does not prove a zero population FP
rate. The correct statistical statement is:

> "Zero false positives on a 100-scenario internal validation corpus, bounding
> the FP rate below **3.6% with 95% confidence** (Clopper-Pearson exact interval)."

This is ARE's Tier 1 claim. It is the claim that is statistically defensible
today.

### 4.3 Path to Tier 2 Claim

Expanding the corpus to 300 scenarios (200 additional, additive-only, spec-first)
reduces the 95% CI upper bound to approximately 1.0%. This expansion is planned
for the post-LoU sprint. Upon completion, the defensible claim becomes:

> "Zero false positives on a 300-scenario internal validation corpus, bounding
> the FP rate below **1.0% with 95% confidence**."

### 4.4 Production Validation Path

Neither the Tier 1 nor Tier 2 claim is a production FP rate. Production
validation requires real agent traffic. The 30-day observe-mode pilot is
designed to produce the first production FP measurement:

> "Zero false positives on [enterprise]'s production traffic — [N] agents,
> [M] requests, 30 days."

This is Tier 3. It is the claim that closes the synthetic-to-production gap.

---

## 5. Boundary Scenario Coverage

The corpus includes 20 boundary scenarios covering agents operating at
z-score 2.5–3.5 — the zone immediately below the enforcement threshold.
These are the scenarios most likely to expose threshold sensitivity and
co-design bias.

All 20 boundary scenarios produce zero false positives at the current
threshold configuration (`z_score_threshold: 3.0` in `config/scoring_weights.yaml`).

---

## 6. Independent Verification

The full corpus is available for review at:
`tests/fp_scenarios/` (scenarios) and `tests/fp_scenarios/specs/` (specifications)

The eval harness that produced the reported result is at:
`tests/eval_harness/`

Any reviewer can reproduce the result by running:
```bash
go test ./tests/eval_harness/... -v
```

The test output includes per-scenario pass/fail, the aggregate FP count,
and the Clopper-Pearson confidence interval calculation.

---

## 7. What This Document Does Not Claim

- This document does not claim ARE has a zero FP rate in production.
- This document does not claim the corpus was constructed by an independent party.
- This document does not claim the synthetic scenarios represent the full
  distribution of legitimate enterprise agent behavior.
- The Zenodo DOI (10.5281/zenodo.19169185) establishes timestamped existence
  of ARE's architectural framework. It does not validate the FP methodology.

These gaps are closed by production pilot data. The 30-day observe-mode
deployment is the designed mechanism for closing them.

---

## 8. Version History

| Version | Date | Change |
|---|---|---|
| 1.0 | April 2026 | Initial release — 100-scenario corpus, Tier 1 claim |
| 2.0 | Post-LoU | 300-scenario corpus, Tier 2 claim (planned) |
| 3.0 | Post-pilot | Production FP rate, Tier 3 claim (planned) |

---

*AgentRepEngine is developed by Naseem A2A Research Lab / Call2leads Inc.*
*Contact: rehanrana@call2leads.com*
*IP anchor: DOI 10.5281/zenodo.19169185*
