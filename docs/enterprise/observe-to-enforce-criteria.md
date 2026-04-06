# Observe-to-Enforce Transition Criteria
# AgentRepEngine — Enterprise Pilot Gate Document
# Version: 1.0 | Date: April 2026

## Purpose

This document defines the five quantitative gates that must pass before
AgentRepEngine advances from observe mode (monitoring only) to enforce mode
(active blocking). No transition occurs without CISO written sign-off.

---

## The Five Gates

### Gate 1 — Observation Duration
**Requirement:** 14 consecutive days of clean observe mode.
**Measurement:** Zero auto-rollback events. Zero anomalous FP spikes (>5% daily rate).
**Data source:** `enforcement_decisions` table, `auto_rollback_events` log.
**Status check:**
```sql
SELECT COUNT(*) FROM enforcement_decisions
WHERE created_at > NOW() - INTERVAL '14 days'
  AND auto_rollback = true;
-- Must return 0
```

### Gate 2 — Baseline Established
**Requirement:** ≥100 behavioral samples per agent identity.
**Measurement:** `agent_baselines` table row count per agent_did.
**Rationale:** Z-score anomaly detection requires sufficient samples for statistical validity.
**Status check:**
```sql
SELECT agent_did, COUNT(*) as sample_count
FROM agent_baselines
GROUP BY agent_did
HAVING COUNT(*) < 100;
-- Must return 0 rows
```

### Gate 3 — Production FP Rate < 2%
**Requirement:** False positive rate on live production traffic below 2%.
**Measurement:** Override rate from `fp_candidates` table over the observe window.
**This is the customer's own data. No claims required — the query proves it.**
**Status check:**
```sql
SELECT
  COUNT(*) FILTER (WHERE override = true)::float / NULLIF(COUNT(*), 0) AS fp_rate
FROM enforcement_decisions
WHERE created_at > NOW() - INTERVAL '14 days'
  AND decision IN ('RESTRICTED', 'BLOCKED');
-- Must return < 0.02
```

### Gate 4 — At Least One Confirmed True Positive
**Requirement:** At least one enforcement decision confirmed as a genuine attack
by the security team.
**Measurement:** `fp_candidates` record with `override = false` and `reviewed = true`.
**Rationale:** Validates that ARE detects real threats in your environment before
enforcement goes live.
**Status check:**
```sql
SELECT COUNT(*) FROM fp_candidates
WHERE override = false
  AND reviewed = true
  AND created_at > NOW() - INTERVAL '30 days';
-- Must return >= 1
```

### Gate 5 — CISO Written Sign-Off
**Requirement:** CISO (or designated security lead) provides written authorization
to activate enforce mode.
**Format:** Email or signed document referencing Gates 1–4 status.
**Stored in:** ARE audit trail as an `enforcement_mode_change` event with
  `authorized_by`, `authorization_timestamp`, and `gate_evidence_json`.

**This gate cannot be automated. It is a human decision.**

---

## Transition Process

1. Security team runs Gates 1–4 SQL queries (self-service, no ARE involvement required)
2. Results reviewed in joint session (ARE team + CISO + SOC lead)
3. CISO provides Gate 5 written authorization
4. ARE team activates enforce mode via configuration change (YAML policy update)
5. First 48 hours: enhanced monitoring, ARE team on call
6. Rollback available at any time: single config change, no data loss

---

## Rollback Guarantee

Enforce mode can be deactivated in < 60 seconds by changing the policy YAML:
```yaml
enforcement_mode: observe  # change from: enforce
```
The change is hot-reloaded at the Kong gateway. No restart required.
All enforcement decisions remain in the audit trail.

---

## "You Control the Pace"

No timeline is imposed. The 14-day observe window is a minimum, not a maximum.
Some customers observe for 30–90 days before transitioning. The gates exist
to give you objective criteria — the decision to proceed is always yours.

---

*Document version: 1.0*
*Maintained by: AgentRepEngine team*
*Related: docs/enterprise/pilot-letter-of-understanding.md*
*Related: docs/enterprise/pilot-case-study-template.md*