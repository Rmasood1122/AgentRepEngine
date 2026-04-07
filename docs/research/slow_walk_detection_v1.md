# Slow-Walk Attack Detection in AI Agent Behavioral Scoring
# Technical Note v1.0 | April 1, 2026
# Naseem A2A Research Lab | DOI: 10.5281/zenodo.19169185
# Author: Rehan Masood

---

## ABSTRACT

Slow-walk attacks represent a class of adversarial evasion
techniques where an AI agent incrementally expands its
behavioral footprint over days or weeks — staying below
single-point detection thresholds while accumulating
unauthorized access or capability. This note documents
AgentRepEngine's two-layer slow-walk detection methodology,
empirical results across 10 multi-day attack scenarios, and
the statistical foundation underlying detection.

**Key results:**
- 100% slow-walk detection rate (10/10 scenarios)
- 0.00% false positive rate on 150-scenario internal validation corpus
- Detection latency: sub-second at gateway layer
- Evasion scenarios tested: 3-day to 7-day attack windows

---

## 1. THE SLOW-WALK PROBLEM

Standard anomaly detection systems trigger on point-in-time
deviations. An agent making 1,000 API calls per hour when its
baseline is 10 will trigger immediately. An agent making 11
calls per hour — growing by 10% daily — will never trigger
a threshold designed for acute attacks.

This is the slow-walk attack pattern:
```
Day 1:  baseline = 10 calls/hour | actual = 11 calls/hour
Day 2:  baseline = 10 calls/hour | actual = 12 calls/hour
Day 3:  baseline = 10 calls/hour | actual = 14 calls/hour
Day 7:  baseline = 10 calls/hour | actual = 20 calls/hour
Day 14: baseline = 10 calls/hour | actual = 40 calls/hour
```

If the baseline itself is naive (e.g., rolling average without
variance tracking), the slow-walk contaminates the baseline
and becomes invisible. The agent normalizes its own anomaly.

This pattern maps directly to documented APT techniques in
traditional network security — low-and-slow exfiltration,
credential stuffing over extended windows, lateral movement
that stays below SIEM alert thresholds.

---

## 2. ARE'S TWO-LAYER DETECTION ARCHITECTURE

AgentRepEngine detects slow-walk attacks through two
independent mechanisms that must both be evaded simultaneously.

### Layer 1 — Welford's Online Algorithm with Variance Tracking

ARE maintains per-agent behavioral statistics using Welford's
online algorithm — a numerically stable, single-pass method
for computing running mean and variance without storing the
full history window.

For each behavioral dimension d and agent a:
```
μ_d(t) = μ_d(t-1) + (x_t - μ_d(t-1)) / n
σ²_d(t) = σ²_d(t-1) + (x_t - μ_d(t-1)) * (x_t - μ_d(t))
```

The z-score for each new observation:
```
z_d(t) = (x_t - μ_d(t)) / sqrt(σ²_d(t))
```

A slow-walk attack that increments gradually will produce
consistently elevated z-scores — small individually, but
sustained over time. ARE scores against the running baseline,
not a fixed threshold, detecting drift that accumulates.

### Layer 2 — Variance Growth Rate Trigger

The key insight: a slow-walk attack causes the variance of
behavioral measurements to grow faster than legitimate
behavioral evolution.

ARE monitors the weekly variance growth rate:
```
VGR = σ²(this_week) / σ²(last_week)
```

Legitimate agents: VGR ≈ 1.0 — 1.2 (stable or slowly growing)
Slow-walk attacks: VGR > 2.0 (variance doubling = attack signal)

**VARIANCE_WINDOW_DAYS = 7** (named constant, auditable)
**VGR_THRESHOLD = 2.0** (named constant, policy-configurable)

When VGR exceeds threshold, ARE fires an early warning before
the z-score accumulates to enforcement level. This provides
detection days before a naive threshold system would trigger.

### Why Two Layers

A slow-walk attacker who understands z-score detection can
modulate their increment rate to keep individual z-scores
below threshold. But modulating the increment rate requires
changing the rate of change — which immediately manifests
as variance growth.

An attacker who understands variance growth rate detection
must keep their increment perfectly constant — which produces
a sustained non-zero z-score that accumulates to enforcement.

Both evasions are mutually exclusive. The two layers create
a detection envelope that cannot be simultaneously evaded.

---

## 3. EMPIRICAL RESULTS

### Test Corpus — 10 Slow-Walk Scenarios

All scenarios run against ARE's held-out validation corpus.
No scenario was used in calibration. Results reflect true
out-of-sample detection performance.

| Scenario | Attack Window | Detection Method | Detected |
|----------|--------------|-----------------|---------|
| SW-001 | 3-day permission escalation | Layer 1 (z-score accumulation) | ✅ |
| SW-002 | 5-day PII access ramp | Layer 2 (VGR trigger) | ✅ |
| SW-003 | 7-day credential probing | Layer 1 + Layer 2 | ✅ |
| SW-004 | 3-day permission escalation buildup | Layer 1 | ✅ |
| SW-005 | 4-day tool call frequency growth | Layer 2 | ✅ |
| SW-006 | 6-day cross-tenant drift | Layer 1 (z-score) | ✅ |
| SW-007 | 5-day sub-agent spawn escalation | Layer 1 + Layer 2 | ✅ |
| SW-008 | 7-day ultra-slow PII harvest | Layer 2 (VGR trigger) | ✅ |
| SW-009 | 3-day token refresh anomaly | Layer 1 | ✅ |
| SW-010 | 4-day lateral movement preparation | Layer 1 + Layer 2 | ✅ |

**Slow-walk detection rate: 100% (10/10)**
**False positive rate on internal validation corpus: 0.00% (0/150 scenarios)**

### Held-Out Validation — Key Scenarios
```
SW-004: 3-day permission escalation buildup
  Score at detection: 750 | Band: TRUSTED→RESTRICTED
  Detection: Day 2, z-score accumulation across 3 dimensions

SW-008: 7-day ultra-slow PII harvest
  Score at detection: 713 | Band: TRUSTED→MONITORED
  Detection: Day 5, VGR trigger fired before z-score threshold
  Note: VGR detected this 2 days before z-score would have triggered
```

SW-008 is the critical result. A pure z-score system would
have detected this attack on Day 7. ARE's variance growth
rate layer detected it on Day 5 — two days earlier, before
significant data exfiltration could complete.

---

## 4. BEHAVIORAL DIMENSIONS MONITORED

ARE tracks 8 behavioral dimensions per agent per request:

| Dimension | Slow-Walk Relevance |
|-----------|-------------------|
| pii_field_access_rate | Direct exfiltration signal |
| permission_escalation_count | Scope expansion signal |
| tool_call_rate_per_hour | Activity volume signal |
| sub_agent_spawn_depth | Lateral movement signal |
| unique_endpoints_per_hour | Reconnaissance signal |
| token_refresh_rate | Credential abuse signal |
| cross_tenant_access_attempts | Isolation violation signal |
| recursive_call_depth | Amplification signal |

All 8 dimensions are monitored simultaneously. Slow-walk
attacks that spread across dimensions (e.g., gradual PII
access + gradual permission escalation) are detected through
composite z-score scoring, not single-dimension thresholds.

---

## 5. ORG-SCOPED BASELINE ISOLATION

All baselines are scoped per `org_id + agent_did`. Redis key
format: `baseline:{org_id}:{agent_did}`.

This prevents cross-tenant baseline contamination — a
multi-tenant slow-walk attack that spreads across org
boundaries cannot dilute any individual org's baseline.
Each org's agents are measured against their own population.

---

## 6. COMPARISON TO EXISTING APPROACHES

| Approach | Slow-Walk Detection | False Positive Rate |
|----------|--------------------|--------------------|
| Static threshold rules | ❌ Evaded by gradual ramp | Low |
| Rolling average | ❌ Baseline contaminated | Low |
| Pure z-score (point-in-time) | ⚠️ Partial — detects late | Medium |
| SIEM correlation rules | ⚠️ Requires manual tuning | High |
| ARE (z-score + VGR) | ✅ 100% detection | 0.00% on internal corpus |

The combination of Welford's online algorithm with variance
growth rate monitoring provides detection that neither
approach achieves alone.

---

## 7. IMPLEMENTATION NOTES

**Language:** Go
**Baseline storage:** Redis (per-org, per-agent keys)
**Persistence:** PostgreSQL (enforcement_decisions table)
**Gateway integration:** Kong Lua plugin
**Variance window:** VARIANCE_WINDOW_DAYS = 7 (configurable)
**VGR threshold:** VGR_THRESHOLD = 2.0 (policy-configurable)
**Detection latency:** <1ms at gateway layer

Full implementation available at:
github.com/Rehanrana11/AgentRepEngine

---

## 8. CONCLUSION

Slow-walk attacks represent the primary evasion technique
for AI agents operating in regulated environments. Point-in-
time detection systems are insufficient. ARE's two-layer
architecture — Welford's online variance tracking combined
with variance growth rate monitoring — achieves 100% detection
across a diverse set of multi-day attack scenarios with zero
false positives on legitimate traffic.

The variance growth rate trigger provides detection 2+ days
earlier than pure z-score systems on ultra-slow attacks,
reducing the window for unauthorized data access before
enforcement fires.

---

## REFERENCES

- Welford, B.P. (1962). Note on a method for calculating
  corrected sums of squares and products. Technometrics, 4(3).
- OWASP Top 10 for LLM Applications (2025)
- NIST AI Risk Management Framework (2023)
- DORA Technical Standards for ICT Risk (2025)

---

*Naseem A2A Research Lab*
*AgentRepEngine v1.0*
*Zenodo DOI: 10.5281/zenodo.19169185*
*GitHub: github.com/Rehanrana11/AgentRepEngine*
*Contact: rehanrana@call2leads.com*
