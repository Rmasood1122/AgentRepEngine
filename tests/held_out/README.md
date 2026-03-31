# Held-Out Validation Corpus

## What This Is

This directory contains a stratified 20% random sample of the full test
corpus, separated into `fp/` (legitimate agent scenarios) and `attack/`
(attack and slow-walk scenarios).

## Corpus Composition

| Set | Source | Total | Held Out | Remaining |
|-----|--------|-------|----------|-----------|
| Legitimate agents | tests/fp_scenarios/ | 100 | 20 | 80 |
| Attack scenarios | tests/attack_corpus/ (Phase1AttackCorpus) | 30 | 6 | 24 |
| Slow-walk scenarios | tests/attack_corpus/ (SlowWalkCorpus) | 10 | 2 | 8 |

## Selection Method

Stratified random sampling — scenarios were grouped by category (data
analyst, authorized export, research agent, new agent, recovering agent,
batch job, misc agent for FP; OWASP reference category for attacks), and
20% were selected from each stratum using a deterministic pseudo-random
method (every 5th scenario per stratum).

## Why This Exists

Published performance claims (TP rate, FP rate, Precision, F1 score) must
be validated on data that was never used to tune detection thresholds or
calibrate scoring weights. The held-out corpus provides this guarantee.

Without a held-out set, performance metrics are training accuracy — they
measure how well the system fits the data it was designed against, not how
well it generalizes. Held-out validation is the minimum standard for
credible performance claims.

## CRITICAL: Usage Restriction

**These scenarios MUST NEVER be used for threshold calibration.**

Specifically:
- Do NOT use held-out scenarios to tune scoring weight coefficients
- Do NOT use held-out scenarios to adjust policy pack thresholds
- Do NOT use held-out scenarios to set HIGH_RISK trigger levels
- Do NOT include held-out scenarios in A/B validation during calibration
- Do NOT reference held-out scenario behavior when making tuning decisions

**The held-out corpus exists for ONE purpose: final validation of published
performance claims (FP rate, TP rate, Precision, F1 score).**

If you contaminate the held-out set by using it for calibration, all
published metrics become training accuracy and are no longer independently
validated. This is not recoverable — you must generate a new held-out set
from fresh scenarios.

## Running Held-Out Validation

```bash
go test ./tests/held_out/fp/... -v
go test ./tests/held_out/attack/... -v
```

## Gate Thresholds

| Metric | Threshold | Gate |
|--------|-----------|------|
| FP rate | ≤ 2.0% | G-FP (hard stop) |
| TP rate | ≥ 85% | G-TP |
| Slow-walk detection | ≥ 70% | G-SW |
