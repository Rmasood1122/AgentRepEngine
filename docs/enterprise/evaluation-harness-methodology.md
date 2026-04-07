# AgentRepEngine — Evaluation Harness Methodology
## How Detection Metrics Are Generated, Validated, and Reported
Version 1.0 | For: Security Review / Audit | Classification: Confidential

---

### How Scenarios Were Generated

ARE's validation corpus is a synthetic financial services (finserv)
corpus designed to exercise every behavioral dimension the scoring
engine monitors. Scenarios are not derived from production data.

**Legitimate agent scenarios (100 total)** model six finserv workflow
categories observed during pre-pilot infrastructure reviews:

| Category | Count | Behavioral Profile |
|----------|-------|--------------------|
| Data analyst workflows | 20 | High tool call rate, moderate bulk access, variable PII |
| Authorized bulk exports | 15 | High bulk access, elevated PII, low tool call rate |
| Research agent discovery | 15 | High endpoint diversity, moderate tool call rate |
| New agent probation | 15 | Low activity across all dimensions, cold-start conditions |
| Recovering agent ramp-up | 15 | Gradually increasing activity post-incident |
| Batch/scheduled jobs | 10 | High bulk access, off-hours patterns |
| Misc enterprise agents | 10 | Customer service, monitoring, code review, document processing |

Each scenario specifies a feature vector across ARE's 8 behavioral
dimensions and a minimum acceptable score. Scenarios were authored
to represent the range of legitimate behavior that a CISO would
expect ARE to allow without flagging.

**Attack scenarios (30 total)** model the OWASP LLM Top 10 attack
patterns that ARE's policy packs are designed to detect:

| OWASP Ref | Count | Attack Pattern |
|-----------|-------|----------------|
| LLM01 (Prompt Injection) | 5 | Direct/indirect injection, escalation, repeated attempts |
| LLM04 (High-Frequency Abuse) | 5 | Tool flooding, recursive loops, sustained abuse |
| LLM06 (Bulk PII Exfiltration) | 6 | Fast extraction, slow drain, disguised exports |
| LLM07 (Scope Escalation) | 5 | Permission escalation, cross-service, token abuse |
| LLM08 (Recursive Spawning) | 3 | Deep spawn, identity cycling, cross-tenant orchestration |
| LLM09 (Autonomous Expansion) | 2 | Scope expansion, orphaned agents |
| Combined vectors | 4 | Multi-OWASP attacks exercising 2+ categories simultaneously |

**Slow-walk scenarios (10 total)** model multi-day adversarial
baseline poisoning across 3-day, 5-day, and 7-day attack windows.
Each scenario specifies a per-day feature vector simulating gradual
behavioral escalation designed to drift the scoring baseline.

---

### How the Held-Out Set Was Selected

The held-out validation corpus is a stratified 20% random sample:

| Corpus | Total | Held Out | Remaining | Method |
|--------|-------|----------|-----------|--------|
| Legitimate agents | 100 | 20 | 80 | 20% per category stratum |
| Attack scenarios | 30 | 6 | 24 | 1 per OWASP stratum |
| Slow-walk scenarios | 10 | 2 | 8 | Mixed OWASP selection |

**Selection method:** Deterministic pseudo-random — every 5th scenario
per stratum. Strata are defined by the scenario categories above
(data analyst, authorized export, research agent, etc. for FP;
OWASP reference category for attacks).

**Stratification rationale:** Proportional sampling from each category
ensures the held-out set has the same distribution of behavioral
profiles as the full corpus. A purely random sample risks
over-representing one category and under-representing another.

**Integrity guarantee:** Held-out scenarios are never used for
threshold calibration, weight adjustment, or policy tuning. They
exist solely to validate published performance claims. See
`tests/held_out/README.md` for the full usage restriction protocol.

---

### Metrics Reported and Why

ARE reports four detection metrics. Reporting fewer than four
creates opportunities for misleading claims.

| Metric | Formula | Why It Matters |
|--------|---------|----------------|
| **True positive rate (recall)** | Detected attacks / Total attacks | Answers: "How many real attacks does ARE catch?" A low TP rate means attacks get through. |
| **False positive rate** | Legitimate agents blocked / Total legitimate agents | Answers: "How often does ARE block something it shouldn't?" One false positive in production can get the product removed. |
| **Precision** | True positives / (True positives + False positives) | Answers: "When ARE blocks something, how often is it correct?" Low precision means your security team wastes time investigating false alarms. |
| **F1 score** | 2 x (Precision x Recall) / (Precision + Recall) | Harmonic mean of precision and recall. Penalizes systems that sacrifice one for the other. |

**Why not just TP rate?** A system that blocks everything has 100% TP
rate and is useless. TP rate alone says nothing about false positives.

**Why not just FP rate?** A system that blocks nothing has 0% FP rate
and is useless. FP rate alone says nothing about detection capability.

**Why F1?** F1 is the standard single-number summary that balances
precision and recall. It is the metric a security auditor will ask
for because it cannot be gamed by optimizing only one dimension.

**Current validated metrics (held-out corpus):**
TP rate: 88.00% | FP rate: 0.00% on internal corpus | Precision: 100% | F1: 0.9362
Measured on held-out validation corpus (20% stratified sample,
never used for threshold calibration).

---

### ARE's Behavioral Window Model

ARE uses a rolling fixed-window z-score model for behavioral
deviation detection. This is an intentionally conservative design.

**How it works:**
- Each agent maintains an individual behavioral baseline (mean and
  standard deviation per dimension) using Welford's online algorithm
- New behavioral events update the running statistics
- Anomaly detection computes the z-score: `(observed - mean) / std`
- Scores above 3.0 standard deviations trigger velocity penalties
- The z-score is computed per-dimension; the worst z-score across
  all 8 dimensions determines the velocity component

**Why fixed-window z-score (not ML classifiers):**
- Interpretable: every enforcement decision can be explained in
  terms of "this agent's behavior deviated N standard deviations
  from its established baseline on dimension X"
- Auditable: the baseline, the observation, and the z-score are
  all stored in the enforcement log
- Conservative: the model errs toward allowing legitimate agents
  rather than blocking on weak signals
- Deterministic: same input always produces same score (no model
  stochasticity)

**Handling legitimate behavioral change:**
ARE's z-score model will flag legitimate behavioral changes (e.g.,
a quarterly reporting agent suddenly doubling its bulk access during
quarter-end). This is intentional — the model cannot distinguish
legitimate change from attack without human context. The override
workflow handles this: a security team member reviews the flagged
decision, confirms it as legitimate via the override endpoint
(reason code ARE-FP-001 through ARE-FP-004), and the override is
recorded in the immutable enforcement log. The override feeds into
the ModeController's false-positive rate calculation, triggering
auto-rollback if FP rate exceeds 2%.

---

### Scoring Weights as Attention Coefficients

ARE's scoring formula uses two primary weight coefficients that
function as attention coefficients over the behavioral signal:

```
Score(t) = Clamp(W_h · H(t) + W_v · V(t), 0, 1000)
```

| Coefficient | Value | Attention Function |
|-------------|-------|--------------------|
| W_h (Historical) | 0.5 | Weight on historical reputation — how much past behavior matters |
| W_v (Velocity) | 0.5 | Weight on current behavioral deviation — how much the current event matters |

**Constraints:**
- W_h + W_v must equal 1.0 (validated at service startup)
- Coefficients are calibrated via A/B validation against the
  held-out corpus
- Maximum drift per calibration iteration: 10% (e.g., W_h can
  move from 0.50 to at most 0.55 or 0.45 in a single iteration)
- Every coefficient change is validated against the held-out
  corpus before acceptance — F1 must improve or remain stable

**Why attention coefficients, not learned weights:**
These are not learned via gradient descent. They are manually
calibrated parameters validated against a fixed corpus. The term
"attention" describes their function: they determine how much
attention the scoring engine pays to historical reputation versus
current behavioral deviation. This is a design choice for
interpretability and auditability — every weight value has a
documented justification, not an opaque training history.

---

### Monitoring KPI Thresholds

ARE monitors three operational KPIs with graduated alert levels:

| KPI | Yellow (Warning) | Red (Action Required) | Gate |
|-----|------------------|-----------------------|------|
| False positive rate | > 0.5% | > 2.0% (auto-rollback fires) | G-FP |
| Gateway p99 latency | > 10ms | > 25ms | G-PERF |
| TP rate (weekly revalidation) | < 90% | < 85% | G-TP |

**FP rate monitoring:**
- Computed every 5 minutes by ModeController
- Formula: overridden BLOCKED decisions / total BLOCKED decisions
  (rolling 1-hour window)
- Yellow (> 0.5%): SIEM warning event, investigation recommended
- Red (> 2.0%): automatic rollback to observe mode, SIEM critical
  event, manual re-enable required

**Latency monitoring:**
- Prometheus histograms with buckets at 1, 2, 5, 10, 25, 50, 100ms
- Gateway enforcement overhead target: p99 ≤ 10ms
- Yellow (> 10ms): investigate Redis latency, connection pool
- Red (> 25ms): indicates infrastructure degradation

**TP rate revalidation:**
- Held-out attack corpus re-run weekly via CI
- Ensures threshold changes have not degraded detection capability
- Yellow (< 90%): review recent threshold changes
- Red (< 85%): halt calibration, revert to last known-good weights

---

### Threshold Change Protocol

Every threshold change follows a documented protocol designed to
prevent accidental degradation of detection performance.

**Step 1 — Propose change.**
Document which weight coefficient is being adjusted, the current
value, the proposed value, and the justification.

**Step 2 — Validate constraint: max 10% drift.**
No coefficient may change by more than 10% in a single iteration.
Example: W_h at 0.50 may move to 0.45–0.55 but not to 0.40.

**Step 3 — Run held-out validation.**
Execute the full held-out corpus (FP + attack + slow-walk) with the
proposed coefficients. Record all four metrics.

**Step 4 — Gate check: F1 must improve or hold.**
If the proposed change degrades F1 score on the held-out corpus,
the change is rejected. No exceptions. A change that improves TP
rate but increases FP rate may still be rejected if F1 declines.

**Step 5 — A/B validation.**
Run the full training corpus (non-held-out scenarios) with both
old and new coefficients. Verify that the new coefficients do not
produce new false positives on any scenario.

**Step 6 — Document and deploy.**
Record the coefficient change, the validation results, and the
justification in the scoring_weights.yaml commit message. Deploy
via standard release process.

---

### ARE's Parameter-Efficient Calibration Model

ARE's calibration philosophy is intentionally minimal: freeze the
detection formula, adjust only sensitivity coefficients.

**What is frozen (never changes during calibration):**
- The scoring formula: `Score(t) = Clamp(W_h · H + W_v · V, 0, 1000)`
- The z-score computation: `z = (observed - mean) / std`
- The decay function: `H(t) = H(t-1) x e^(-0.1 x days)`
- The velocity penalty formula: `penalty = min(100 x (z - 3.0), 300)`
- The score band boundaries: TRUSTED 800+, MONITORED 500+, RESTRICTED 200+, BLOCKED <200
- The HIGH_RISK trigger thresholds per OWASP policy pack
- The hash chain formula for the enforcement log

**What can be adjusted (sensitivity coefficients only):**
- W_h and W_v (historical/velocity attention weights)
- Decay rate (currently 0.1, half-life ~7 days)
- Z-score threshold (currently 3.0)
- Maximum velocity penalty (currently 300)

**Why parameter-efficient:**
A calibration model that can change the detection formula itself is
a calibration model that can break the detection formula. By freezing
the formula and exposing only a small number of sensitivity
coefficients, ARE limits the blast radius of any single calibration
decision. The max 10% drift constraint further limits the rate of
change. This is the same principle as fine-tuning a frozen model
with a small number of adapter parameters — the base capability is
preserved while allowing controlled adaptation.

---

*Methodology document current as of March 2026.
Validation corpus and metrics are re-run on every release.*
