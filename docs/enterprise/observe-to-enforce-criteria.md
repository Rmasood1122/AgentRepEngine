# Observe-to-Enforce Transition Criteria
**AgentRepEngine — Pilot Governance Document**  
**Version:** 1.0 | April 8, 2026  
**Audience:** CISO, Security Engineering Lead, Compliance Officer

---

## Purpose

This document defines the exact criteria that must be satisfied before AgentRepEngine advances from observe mode to enforce mode. No transition occurs without explicit CISO sign-off. Every gate is measurable. Every measurement is self-verifiable by the customer.

**The rule:** ARE never advances enforcement mode without your written authorization. This document is the checklist you sign.

---

## The Five Gates

All five gates must pass. Any single failure holds the transition.

---

### Gate 1 — Minimum Observe Duration
**Requirement:** 14 consecutive days of clean observe mode with no FP spike events and no auto-rollback triggers.

**Why 14 days:** Behavioral baselines require sufficient data to be statistically meaningful. Welford's algorithm converges to a stable mean/variance after approximately 100 samples per agent. For agents making 10+ calls per day, 14 days exceeds this threshold. For lower-frequency agents, the baseline establishment period extends accordingly (see Gate 2).

**Verification query:**
```sql
SELECT
  MIN(created_at) as observe_start,
  MAX(created_at) as observe_current,
  DATE_PART('day', MAX(created_at) - MIN(created_at)) as days_elapsed,
  COUNT(*) FILTER (WHERE event_type = 'AUTO_ROLLBACK') as rollback_events,
  COUNT(*) FILTER (WHERE event_type = 'FP_SPIKE') as fp_spike_events
FROM mode_change_log
WHERE from_mode = 'observe'
  AND created_at > NOW() - INTERVAL '30 days';
```

**Pass condition:** `days_elapsed >= 14` AND `rollback_events = 0` AND `fp_spike_events = 0`

---

### Gate 2 — Baseline Established
**Requirement:** Every active agent has ≥100 scored samples in the `agent_baselines` table.

**Why 100 samples:** Below 100 samples, z-score calculations have high variance and the confidence interval on anomaly detection is too wide for reliable enforcement. At 100+ samples, the baseline is statistically stable.

**Verification query:**
```sql
SELECT
  agent_did,
  sample_count,
  CASE WHEN sample_count >= 100 THEN 'READY' ELSE 'NOT READY' END as baseline_status,
  last_updated
FROM agent_baselines
ORDER BY sample_count ASC;
```

**Pass condition:** Every agent in scope has `sample_count >= 100`. Any agent below 100 samples remains in observe mode for that agent specifically — enforce mode can be activated per-agent, not fleet-wide.

**Note for Lloyd:** If NWN has high-frequency agents (100+ calls/day), Gate 2 passes in under 24 hours. If agents are lower frequency (10 calls/day), Gate 2 takes 10 days. The 14-day Gate 1 minimum is designed to accommodate this.

---

### Gate 3 — Production FP Rate Within Threshold
**Requirement:** False positive rate on production traffic < 2% over the observe period.

**Definition:** A false positive is any RESTRICTED or BLOCKED decision that a human reviewer confirms was a legitimate agent action. ARE logs all candidate FPs in the `fp_candidates` table during observe mode. Your security team reviews and marks each one.

**Verification query:**
```sql
SELECT
  DATE(flagged_at) as date,
  COUNT(*) as total_candidates,
  COUNT(*) FILTER (WHERE confirmed_fp = true) as confirmed_fps,
  COUNT(*) FILTER (WHERE confirmed_fp = false) as confirmed_tps,
  COUNT(*) FILTER (WHERE reviewed_by IS NULL) as pending_review,
  ROUND(
    COUNT(*) FILTER (WHERE confirmed_fp = true) * 100.0 /
    NULLIF(COUNT(*) FILTER (WHERE confirmed_fp IS NOT NULL), 0), 4
  ) as fp_rate_pct
FROM fp_candidates
GROUP BY DATE(flagged_at)
ORDER BY date DESC;
```

**Pass condition:** No single day exceeds 2% FP rate. Rolling 7-day average < 1%.

**Production target:** We target below 0.1% — the Visa fraud detection standard. The 2% gate is the minimum for enforcement activation. The 0.1% target is the steady-state goal at 90 days.

**No vendor controls this measurement.** You run the query. You own the result.

---

### Gate 4 — At Least One Confirmed True Positive
**Requirement:** At least one enforcement decision during the observe period has been reviewed by a human and confirmed as a genuine anomaly — an agent behaving outside its normal pattern in a way that represents real risk.

**Why this matters:** Gate 3 measures false positives. Gate 4 measures that the system is actually detecting something. An observe period with zero anomalies flagged could mean your environment is clean — or it could mean the baselines are too loose. Gate 4 distinguishes between these two cases.

**Verification query:**
```sql
SELECT
  fpc.agent_did,
  fpc.score,
  fpc.band,
  fpc.reason_object,
  fpc.confirmed_fp,
  fpc.reviewed_by,
  fpc.reviewed_at,
  fpc.notes
FROM fp_candidates fpc
WHERE fpc.confirmed_fp = false
  AND fpc.reviewed_by IS NOT NULL
ORDER BY fpc.reviewed_at DESC
LIMIT 10;
```

**Pass condition:** At least one row with `confirmed_fp = false` and `reviewed_by IS NOT NULL`. This is a human-reviewed, human-confirmed true positive.

**What to do if Gate 4 fails:** Zero confirmed TPs after 14 days means either (a) your agents are operating normally — which is the best possible outcome and should be documented, or (b) thresholds require calibration. ARE's threshold calibration process is documented separately. Do not advance to enforce mode without resolving this.

---

### Gate 5 — CISO Written Sign-Off
**Requirement:** Written authorization from the designated CISO or security lead to activate enforce mode. This authorization must be logged in the ARE audit trail.

**This is not a courtesy.** It is a regulatory requirement. DORA Article 9 requires documented human authorization for significant changes to ICT risk management controls. GDPR Article 22 requires documented human oversight of automated decision systems. ARE's sign-off gate is the implementation of both requirements.

**The sign-off records:**
- Date and time of authorization
- Name and role of authorizing officer
- Scope of enforcement activation (all agents, specific agent types, HIGH_RISK operations only)
- Enforcement level (HIGH_RISK only vs. full enforcement)
- FP threshold for auto-rollback

**The phrase:** *"You control the pace — we don't advance to enforce mode without your sign-off."*

---

## Enforce Mode Activation Sequence

Once all five gates pass:

**Step 1 — HIGH_RISK operations only (Days 1–7 of enforce mode)**
```
ENFORCEMENT_MODE=enforce
ENFORCEMENT_SCOPE=HIGH_RISK
```
Only operations classified HIGH_RISK in your policy configuration are subject to auto-block. All other operations continue in observe mode. Human review required on every block for the first 7 days.

**Step 2 — Full enforcement (Day 8+ if HIGH_RISK phase passes)**
```
ENFORCEMENT_MODE=enforce
ENFORCEMENT_SCOPE=ALL
```
Full enforcement across all policy-configured risk levels. Auto-rollback protection remains active.

**Auto-rollback trigger:**
```
FP_RATE_THRESHOLD=2.0
```
If production FP rate exceeds 2% on any rolling 24-hour window, ARE automatically reverts to observe mode and logs the rollback event. You are notified immediately. No manual intervention required to protect your agents.

---

## Enforcement Mode Status — Always Visible

```sql
SELECT
  current_mode,
  enforcement_scope,
  activated_at,
  activated_by,
  fp_threshold_pct,
  auto_rollback_enabled
FROM enforcement_config
ORDER BY activated_at DESC
LIMIT 1;
```

---

## Summary Checklist — Print This Page

Before any enforcement activation, confirm:

| Gate | Requirement | Query | Status |
|---|---|---|---|
| G1 | 14 days clean observe | mode_change_log | ☐ |
| G2 | ≥100 samples per agent | agent_baselines | ☐ |
| G3 | FP rate < 2% production | fp_candidates | ☐ |
| G4 | ≥1 confirmed TP | fp_candidates | ☐ |
| G5 | CISO written sign-off | enforcement_config | ☐ |

All five checked → activation authorized.  
Any single gate open → activation blocked.

---

## Related Documents

- `docs/enterprise/Lloyd_meeting_prep.md` — meeting talking points
- `docs/enterprise/pilot-letter-of-understanding.md` — LoU staged curriculum
- `docs/enterprise/pilot-case-study-template.md` — day-30 deliverable format
- `docs/enterprise/prerequisites-checklist.md` — pre-install requirements

---

*L118 complete | APEX v5.2 | April 8, 2026*  
*Next: L132 — pilot-case-study-template.md (must exist before pilot starts)*
