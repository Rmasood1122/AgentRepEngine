**Stated claim:** "Zero FPs across 150 synthetic scenarios, bounding FP below
2.4% at 95% CI (Clopper-Pearson exact method)."

This is the correct qualified claim. The bare "0.00%" is not used without
this qualification in any external communication.

### 5.2 What the CI Does Not Mean

The 95% CI upper bound of 2.4% does not mean ARE will produce 2.4% FPs in
production. It means: given 0 observed FPs in 150 trials, we cannot statistically
rule out a true rate above 2.4% with 95% confidence.

Production FP rate will be measured during the first enterprise pilot and
published. ARE's auto-rollback mechanism enforces a hard 2% production FP
ceiling — if FP rate exceeds 2% in any 5-minute window, the system automatically
reverts to observe mode and fires a SIEM alert.

---

## 6. WHAT THE CORPUS DOES NOT COVER

**Not covered — known limitations:**

1. Real production agent traffic: All scenarios are synthetic.

2. Multi-tenant baseline contamination: The corpus tests single-org deployments.

3. Non-finserv agent archetypes: The corpus is calibrated for financial services.
   Healthcare and government agent archetypes may require threshold recalibration.

4. Long-horizon behavioral drift: Scenarios simulate up to 30-day windows.
   Agent behavioral drift over 90+ days is not covered.

These limitations define the scope of the 0.00% FP claim. The claim is valid
within its stated scope. It does not extend beyond it.

---

## 7. THE AUTO-ROLLBACK CIRCUIT BREAKER

ARE includes a live FP circuit breaker:

```go
// internal/enforcement/mode_controller.go
// StartFPMonitor() — checks FP rate every 5 minutes
// If FP > 2%: rolls back to observe mode + fires SIEM alert
// No restart required. Mode change is instant.
```

Even if the production FP rate exceeds the corpus prediction, ARE self-limits
before the CISO needs to act. The system enforces the FP ceiling mechanically.

---

## 8. VERIFICATION COMMANDS FOR AUDITORS

**SQL query — FP rate from production enforcement_decisions:**
```sql
SELECT
  COUNT(*) FILTER (WHERE decision = 'BLOCKED' AND human_override = true) AS false_positives,
  COUNT(*) FILTER (WHERE decision = 'BLOCKED') AS total_blocks,
  ROUND(
    100.0 * COUNT(*) FILTER (WHERE decision = 'BLOCKED' AND human_override = true)
    / NULLIF(COUNT(*) FILTER (WHERE decision = 'BLOCKED'), 0),
    2
  ) AS fp_rate_pct
FROM enforcement_decisions
WHERE created_at > NOW() - INTERVAL '30 days';
```

**Hash chain integrity check:**
```bash
go run cmd/verify-decision/main.go --from $(date -d '30 days ago' +%Y-%m-%d)
```

---

## 9. COMPARISON TO INDUSTRY BENCHMARKS

| System | FP rate | Corpus | Published? |
|--------|---------|--------|-----------|
| ARE (synthetic corpus) | 0.00% (150 scenarios) | Synthetic finserv | Yes — this document |
| ARE (production target) | <0.1% | Production | Pending pilot |
| Visa fraud detection | ~0.1% | Production | Industry reported |
| Traditional SIEM rules | 40–60% | Production | Industry reported |
| ML-based UEBA | 5–15% | Production | Vendor reported |

No AI agent behavioral enforcement product has published a methodology document
at this specificity level. This document is the first.

---

## 10. CONTACT AND COLLABORATION

**Repository:** github.com/Rehanrana11/AgentRepEngine
**Published IP:** DOI 10.5281/zenodo.19169185
**Research contact:** rehanrana@call2leads.com

Independent validation of this corpus and methodology is welcomed.
Academic reviewers may request the full scenario specification for reproduction.

---

AgentRepEngine FP Measurement Methodology v1.0
Naseem A2A Research Lab | April 2026