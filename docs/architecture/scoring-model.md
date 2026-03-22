# AgentRepEngine — Scoring Model
## Phase 1 Formula, Score Bands, and ATP State Mapping
Version 1.0 | For: Technical Review

---

### Phase 1 Scoring Formula
```
Score(t) = Clamp(0.5·H + 0.5·V, 0, 1000)
```

**H — Historical reputation component:**
```
H(t) = H(t-1) × e^(-0.1 × days_since_last_event)
```
Half-life: ~7 days. Recovery from BLOCKED to MONITORED: 3–4 weeks of clean behavior.
Rationale: 7-day half-life responds to behavior changes fast enough to be useful,
slow enough to prevent score gaming via brief periods of good behavior.

**V — Velocity anomaly component:**
```
V = 1000 × (1 - Clamp((observed_rate - μ) / σ, 0, 3) / 3)
```
- μ = per-agent individual baseline (updated hourly)
- σ = per-agent standard deviation
- z-score scope: per-agent individual (not global)
- Cold start: velocity counting only for first 48h (probation)
- z-score activates after minimum 10 events

**Weights:** config-driven, never hardcoded.
```yaml
# config/scoring_weights.yaml
scoring:
  history_weight: 0.5
  velocity_weight: 0.5
```
Weights must sum to 1.0 — validated on service startup.

**Phase 2 formula** (after 90 days of behavioral data):
```
Score(t) = Clamp(0.4·H + 0.3·V + 0.2·A + 0.1·P, 0, 1000)
```
A = Isolation Forest anomaly score
P = Peer cluster deviation score

---

### Score Bands and ATP State Mapping

Every numerical score maps to an ATP protocol state
from the Agentic Trust Architecture research corpus (Z1).

| Score | Band | ATP State | Enforcement Action |
|-------|------|-----------|-------------------|
| 700–1000 | TRUSTED | RESOLVE | ALLOW + passive log |
| 500–699 | MONITORED | VERIFY | ALLOW + active audit |
| 200–499 | RESTRICTED | VERIFY+ | THROTTLE + human review on HIGH_RISK |
| 0–199 | BLOCKED | DEFER | Synthetic response (not 403) |

**ATP State definitions (Z1 — Agentic Trust Protocol):**
- **RESOLVE:** Authority preconditions met. Action is authorized.
- **VERIFY:** Authority uncertain. Human verification required for HIGH_RISK.
- **DEFER:** Authority preconditions unmet. Action suspended pending review.

**Why synthetic response instead of 403:**
A 403 tells an adversarial agent it has been detected. A synthetic
cached response (indistinguishable from a slow legitimate response)
does not reveal the enforcement threshold. Adversaries cannot
calibrate their behavior against an unknown threshold.

---

### Conflict Severity Function Γ(C_o)

Γ maps behavioral conflicts to conflict severity.
Monotonic: adding conflicts never decreases severity.
```
Γ(C_o) = Σ(weight_i × unresolved_conflict_count_i)
```

Threshold → ATP state mapping:
- Γ < 0.3 → RESOLVE (score 700–1000)
- 0.3 ≤ Γ < 0.7 → VERIFY (score 400–699)
- Γ ≥ 0.7 → DEFER (score 0–399)

Conflict weights (config/scoring_weights.yaml):
```yaml
conflict_severity:
  bulk_pii_access:           0.40
  permission_escalation:     0.35
  cross_tenant_probe:        0.50
  recursive_spawn:           0.30
  credential_access:         0.45
  high_frequency_tool_abuse: 0.25
```

---

### Cold Start Policy

New agents cannot have a z-score baseline at N=1.
Policy solution (correct for Phase 1):

| Condition | Score | Status | z-score |
|-----------|-------|--------|---------|
| New provisioned agent | 700 | PROBATION | Suspended |
| Orphaned agent (unknown) | 500 | PROBATION | Suspended |
| Sub-agent (from parent) | min(parent_score, 700) | PROBATION | Suspended |
| Probation exit criteria | ≥700 + zero HIGH_RISK | MONITORED | Activates |

Probation duration: 48 hours minimum.
HIGH_RISK event during probation: resets clock by 48 hours.
z-score activates: after probation exit AND ≥10 events.

---

### HIGH_RISK Operation Override

Certain operations require human review regardless of score.
A TRUSTED agent (score 950+) cannot auto-execute HIGH_RISK actions.

HIGH_RISK operation classes:
- bulk_pii_export
- credential_access
- lateral_movement
- mass_deletion
- cross_tenant_data_access

When HIGH_RISK is triggered:
- Score ≥ 700: VERIFY state → human approval required
- Score < 700: DEFER state → blocked pending review
- Score ≥ 950: auto-RESOLVE available (configurable)

This closes the reputation attack: an adversary cannot build
a TRUSTED score over 30 days and then execute a HIGH_RISK
action freely.

---

### Score Update Latency

| Operation | Target | Measurement |
|-----------|--------|-------------|
| Score lookup (cache hit) | p99 ≤ 2ms | redis-cli --latency |
| Gateway enforcement overhead | p99 ≤ 10ms | wrk benchmark |
| Score update (async) | p95 ≤ 500ms | Prometheus metric |
| Full event pipeline | end-to-end ≤ 30s | integration test |

Score update is always async — never on the critical path.
Gateway reads from Redis cache only. Cache miss → PostgreSQL fallback.

---

*Scoring model current as of March 2026.
Phase 2 formula activates when 90 days of Phase 1 data exists.*