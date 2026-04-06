# Per-Archetype Adaptive Threshold Guide
**AgentRepEngine — Phase 1**
Version: 1.0 | Date: April 2026 | Status: Production

---

## Why Per-Archetype Thresholds Matter

A single global z-score threshold (default: 3.0) creates false positives for agents
whose legitimate behavior naturally runs hot. A data analyst running quarterly reports
will hit higher tool call rates than a monitoring agent. Without archetype awareness,
ARE would block legitimate high-volume agents while the z-score baseline catches up.

**Bias audit result (Phase 1):** 0.00% FP rate across all 7 archetypes on 100-scenario
corpus. This document records the threshold rationale that produces that result.

---

## The 7 Agent Archetypes

### 1. `data_analyst`
**Behavior profile:** High tool call rates, bulk access bursts, scheduled jobs, BI exports.
Quarter-end and annual report periods produce legitimate 3–5x traffic spikes.

| Feature | Normal Range | Alert Threshold | Block Threshold |
|---|---|---|---|
| tool_call_rate_per_hour | 20–80 | 150 | 300 |
| bulk_access_count_per_session | 100–500 | 2000 | 5000 |
| pii_field_access_rate | 10–50 | 100 | 500 |
| unique_endpoints_per_hour | 5–15 | 30 | 60 |

**Threshold note:** Bootstrap window is 100 samples. Quarter-end spikes should not
trigger enforcement until baseline reflects seasonal pattern (≥30 days of data).

**VRF jitter:** ±0.25 sigma applied. Attacker cannot probe exact block boundary.

---

### 2. `authorized` (privileged service agents)
**Behavior profile:** Pre-authorized bulk operations, system integrations, ETL pipelines.
May legitimately access large data volumes in short windows.

| Feature | Normal Range | Alert Threshold | Block Threshold |
|---|---|---|---|
| tool_call_rate_per_hour | 50–200 | 400 | 800 |
| bulk_access_count_per_session | 500–2000 | 8000 | 20000 |
| pii_field_access_rate | 50–200 | 500 | 1000 |
| cross_tenant_probe_count | 0 | 1 | 3 |

**Threshold note:** Cross-tenant probes are ZERO-TOLERANCE regardless of authorization
level. Any cross-tenant probe triggers immediate human VERIFY regardless of score band.

---

### 3. `research`
**Behavior profile:** Exploratory queries, broad endpoint enumeration for discovery,
model training data pulls. High unique endpoint counts are expected.

| Feature | Normal Range | Alert Threshold | Block Threshold |
|---|---|---|---|
| tool_call_rate_per_hour | 30–100 | 200 | 400 |
| unique_endpoints_per_hour | 15–40 | 80 | 150 |
| bulk_access_count_per_session | 200–1000 | 3000 | 8000 |
| pii_field_access_rate | 5–20 | 50 | 200 |

**Threshold note:** Unique endpoint counts alone do not justify blocking for research
agents. Require BOTH high endpoint count AND high PII access rate for block escalation.

---

### 4. `monitoring`
**Behavior profile:** Continuous low-volume polling, health checks, metric collection.
Very consistent — deviations are high-signal.

| Feature | Normal Range | Alert Threshold | Block Threshold |
|---|---|---|---|
| tool_call_rate_per_hour | 10–30 | 60 | 120 |
| unique_endpoints_per_hour | 3–8 | 15 | 30 |
| bulk_access_count_per_session | 1–10 | 50 | 100 |
| pii_field_access_rate | 0–2 | 10 | 30 |

**Threshold note:** Monitoring agents have the tightest baselines. A 2x deviation is
high-signal for this archetype. Z-score is most effective here — no threshold relaxation.

---

### 5. `batch`
**Behavior profile:** Scheduled bulk jobs, nightly ETL, periodic report generation.
Traffic is zero between runs, then spikes sharply at schedule time.

| Feature | Normal Range | Alert Threshold | Block Threshold |
|---|---|---|---|
| tool_call_rate_per_hour | 0–500 (burst) | 1000 | 2000 |
| bulk_access_count_per_session | 0–5000 (burst) | 15000 | 30000 |
| pii_field_access_rate | 0–100 (burst) | 300 | 800 |
| unique_endpoints_per_hour | 1–5 | 10 | 20 |

**Threshold note:** Two-speed baseline (ARE v1.1) is essential for batch agents.
ATTACK baseline (fast, 1-hour window) detects anomalies within the burst window.
LEGITIMATE_CHANGE baseline (slow, 30-day window) prevents false positives on
scheduled spikes. See `two_speed_baseline.go`.

---

### 6. `cold_start`
**Behavior profile:** New agents with <100 events — bootstrap period active.
No reliable personal baseline exists yet.

| Feature | Normal Range | Alert Threshold | Block Threshold |
|---|---|---|---|
| All features | Unknown | Cluster average + 2σ | Cluster average + 3σ |

**Threshold note:** Bootstrap uses peer cluster average (default: 700) until
`bootstrap_sample_threshold` (100 samples) is reached. Cold-start agents cannot
be blocked based on personal baseline — cluster baseline applies.
Enforcement in observe-mode only until 100 samples accumulated.

---

### 7. `recovery`
**Behavior profile:** Agents returning after incident or maintenance window.
Score may be suppressed; behavior may be unusually cautious or erratic during warmup.

| Feature | Normal Range | Alert Threshold | Block Threshold |
|---|---|---|---|
| All features | Degraded baseline | Personal baseline + 2σ | Personal baseline + 3σ |

**Threshold note:** Recovery agents are in SIR REMEDIATED or CLEARED state.
Probation window (7 days) applies — any relapse during probation returns to INCIDENT.
Threshold enforcement uses pre-incident personal baseline, not cluster average.

---

## Global Threshold Override Rules

1. **Cross-tenant probe: zero-tolerance** — any `cross_tenant_probe_count > 0` triggers
   human VERIFY regardless of archetype or score band.

2. **Sub-agent spawn depth > 3** — triggers human VERIFY regardless of archetype.

3. **NIS2 proportionality** — `confidence_pct < 71%` caps enforcement at THROTTLE
   regardless of archetype (see `nis2_proportionality.yaml`).

4. **VRF jitter** — all thresholds have ±0.25 sigma per-agent randomization applied
   at scoring time. Documented thresholds above are base values pre-jitter.

---

## Threshold Activation Timeline

| Phase | Archetype Support | Baseline Type |
|---|---|---|
| Phase 1 (now) | All 7 archetypes | Personal z-score + cluster fallback |
| Phase 2 (post-pilot) | All 7 + custom | RLHF weight adjustment per archetype |
| Phase 3 (90-day data) | All 7 + custom | Isolation Forest per archetype cluster |

---

## Bias Audit Status

Last run: April 2026
Corpus: 100 legitimate scenarios across 7 archetypes
Result: **0.00% FP rate — all archetypes PASS**

| Archetype | Scenarios | FPs | FP Rate |
|---|---|---|---|
| data_analyst | ~40 | 0 | 0.00% |
| authorized | ~15 | 0 | 0.00% |
| research | ~10 | 0 | 0.00% |
| monitoring | ~10 | 0 | 0.00% |
| batch | ~10 | 0 | 0.00% |
| cold_start | ~8 | 0 | 0.00% |
| recovery | ~7 | 0 | 0.00% |

Re-run before every enforce-mode go-live:
```bash
go test ./tests/fp_scenarios/... -run TestFPBiasAudit -v
```