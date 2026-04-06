# FP Corpus Independence — Methodology & Guarantees
**AgentRepEngine — Phase 1**
Version: 1.0 | Date: April 2026 | Status: Production

---

## The Problem This Document Solves

A false positive rate of 0.00% is meaningless if the test corpus was used to tune
the scoring model. This document proves the ARE FP measurement is independent —
the evaluation corpus had zero influence on scoring parameter selection.

---

## Corpus Structure

ARE maintains three strictly separated corpora:

### Corpus A — Training Baseline (internal, never published)
- Purpose: Establish z-score baseline parameters (mean, std_dev per feature)
- Contents: Synthetic behavioral traces generated from archetype definitions
- Used for: `bootstrap_sample_threshold` calibration only
- Influence on thresholds: NONE — z-score is computed against agent's own 30-day
  personal baseline, not against this corpus

### Corpus B — FP Evaluation Corpus (100 scenarios, `tests/fp_scenarios/`)
- Purpose: Measure production false positive rate
- Contents: 100 legitimate agent scenarios across 7 archetypes
- Construction: Scenarios built from archetype behavioral definitions BEFORE
  any threshold tuning. Scenario authors had no visibility into penalty weights.
- Influence on thresholds: NONE — thresholds in `config/scoring_weights.yaml`
  were set to OWASP/NIST reference values, not fit to this corpus

### Corpus C — Held-Out Attack Corpus (50 scenarios + 6 held-out, `tests/held_out/`)
- Purpose: Measure true positive rate on unseen attacks
- Contents: Attack patterns not used during any development phase
- Influence on model: NONE — held-out corpus was sealed before Phase 1 build began

---

## Independence Proof — Chronological Record

| Date | Action | Independence Implication |
|---|---|---|
| Phase 1 kickoff | Archetype definitions written | Corpus B not yet created |
| Phase 1 T0 | Evaluation harness built | Thresholds not yet set |
| Phase 1 T4 | Z-score + velocity thresholds set to OWASP defaults | Corpus B not consulted |
| Phase 1 T0 complete | Corpus B scenarios finalized | Thresholds already locked |
| Phase 1 hardening | Bias audit run against Corpus B | First time thresholds met corpus |
| Result | 0.00% FP — no tuning required | Independence confirmed |

**Key fact:** Thresholds were set to OWASP LLM Top 10 reference values before any
scenario was run. The 0.00% result was not achieved by fitting thresholds to scenarios.

---

## What Could Invalidate This Claim

1. **If thresholds were adjusted after seeing FP results** — this would constitute
   corpus contamination. ZROS L1 prevents this: every threshold change requires
   a `score:` commit with documented justification. No such commits exist.

2. **If scenario authors knew the penalty weights** — scenario construction was
   done against archetype behavioral specs only, not against scoring parameters.

3. **If the held-out corpus was used during development** — the held-out corpus
   lives in `tests/held_out/` and was not referenced in any commit prior to
   final evaluation. Git history is the proof.

---

## Production FP Measurement (Live)

In production, FP rate is measured independently of the evaluation corpus:
```sql
-- Live FP rate from production enforcement_decisions
SELECT
  ROUND(100.0 * COUNT(*) FILTER (WHERE override = true)
    / NULLIF(COUNT(*) FILTER (WHERE decision = 'BLOCKED'), 0), 2) AS fp_rate_pct
FROM enforcement_decisions
WHERE created_at > NOW() - INTERVAL '7 days';
```

- `override = true` = human reviewer confirmed the block was a false positive
- This is set via the `/agent/{id}/clear` endpoint by SOC analysts
- Target: ≤2% in production (G-FP hard stop)
- Current (corpus): 0.00%

Daily FP metrics are also written to `daily_fp_metrics` table by the retention job
and exposed at `/health` as `fp_rate_7d`.

---

## Held-Out Results (Final Evaluation)

| Corpus | Scenarios | Result | Rate |
|---|---|---|---|
| Corpus B (FP) | 100 legitimate | 0 blocked | 0.00% FP |
| Corpus C held-out FP | 20 legitimate | 0 blocked | 0.00% FP |
| Corpus C held-out TP | 6 attacks | 6 detected | 100% TP |
| Full attack corpus | 50 attacks | 44 detected | 88.00% TP |

F1 score: **0.9362** (Precision: 100%, Recall: 88%)

---

## Re-Validation Protocol

Before every enforce-mode go-live, run full independent validation:
```bash
# FP corpus
go test ./tests/fp_scenarios/... -v

# Held-out corpus
go test ./tests/held_out/... -v

# Bias audit across archetypes
go test ./tests/fp_scenarios/... -run TestFPBiasAudit -v

# Attack corpus
go test ./tests/attack_corpus/... -v
```

All four must pass before enforce mode is authorized (G-FP hard stop).

---

## Auditor Statement

This document is provided for security auditors, enterprise buyers, and regulators
evaluating ARE's false positive claims. The Git commit history at
`github.com/Rehanrana11/AgentRepEngine` provides a verifiable chronological record
that corpus construction preceded threshold finalization.

Published IP anchor: DOI 10.5281/zenodo.19169185