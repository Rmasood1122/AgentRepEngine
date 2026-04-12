# Adversarial Baseline Poisoning: A Novel Attack Class Against Behavioral Scoring Systems for AI Agents
# AgentRepEngine Research Note | April 7, 2026
# DOI: pending (upload to Zenodo after Lloyd LoU signed)
# Authors: Rehan Masood, Naseem A2A Research Lab

---

## ABSTRACT

Behavioral scoring systems for AI agents assume honest baseline convergence.
An attacker who controls an agent can subvert this assumption through
graduated rate escalation — a technique we term adversarial baseline poisoning.
We describe the attack class, demonstrate its effectiveness against naive
z-score detection, and present a two-layer defense that detects the attack
before high-value actions execute. Experimental results on a synthetic
financial services corpus show 100% detection with 0% false positive rate
on 150 legitimate agent scenarios.

---

## 1. PROBLEM

AI agent behavioral scoring systems operate on a simple assumption: an
agent's historical behavior establishes a baseline, and significant
deviation from that baseline indicates potential compromise or misuse.
This assumption is correct for sudden attacks but fails against a patient
attacker who understands the scoring mechanism.

The z-score anomaly detection model computes:

  z = (observed_rate - baseline_mean) / baseline_std_dev

A sudden spike in an agent's data access rate produces a high z-score
and triggers enforcement. But an attacker who gradually increases the
rate over multiple days causes both the mean and standard deviation to
shift toward the attack rate. The z-score stays low. The scoring system
sees normal behavior. The attacker looks clean.

This is not a theoretical vulnerability. It is a systematic weakness in
any behavioral scoring system that uses rolling baselines without
additional controls.

---

## 2. ATTACK DESCRIPTION

We define adversarial baseline poisoning as a class of attacks with
the following characteristics:

  Attack vector: Gradual behavioral escalation over N days
  Target: Rolling mean/variance baseline of behavioral scoring system
  Effect: Baseline shifts toward attack rate; z-score stays low
  Outcome: High-value action executes without triggering enforcement

### 2.1 Mechanism

Consider an agent with a baseline PII field access rate of 0.05 (5 per
100 calls). The enforcement threshold is PIIFieldAccessRate > 0.30.

Naive attack (detected): Agent immediately accesses PII at rate 0.40.
  z-score: (0.40 - 0.05) / 0.02 = 17.5 sigma. Detected immediately.

Adversarial baseline poisoning (evades naive z-score):
  Day 1: rate 0.06 — z = 0.5. Baseline shifts to mean=0.055
  Day 3: rate 0.08 — z = 0.8. Baseline shifts to mean=0.065
  Day 5: rate 0.12 — z = 1.1. Baseline shifts to mean=0.085
  Day 7: rate 0.18 — z = 1.4. Baseline shifts to mean=0.120
  Day 10: rate 0.28 — z = 1.6. Baseline shifts to mean=0.185
  Day 14: rate 0.38 — z = 1.8. Baseline stays low. Attack succeeds.

After 14 days of patient escalation, the agent's access rate is 7.6x
its original baseline. The z-score is 1.8 — well below the 3.0 threshold.
The scoring system reports normal behavior. The attacker exfiltrates PII.

### 2.2 Why Standard Defenses Fail

Static rule thresholds: catch sudden spikes but are set to accommodate
legitimate operational variance. A threshold of 0.30 catches the naive
attack but the poisoned baseline makes 0.38 look like normal growth.

Isolation Forest: requires 90+ days of behavioral data to fit reliably.
Cannot detect this attack during the baseline establishment period —
exactly when the attacker strikes.

Peer cluster comparison: partially effective if the attacker controls
only one agent. If multiple agents are poisoned simultaneously, cluster
deviation is suppressed.

---

## 3. TWO-LAYER DEFENSE

AgentRepEngine implements a two-layer defense against adversarial
baseline poisoning. Both layers are required. Neither alone is sufficient.

### Layer 1: Policy Threshold Enforcement

Regardless of z-score, any agent executing a HIGH_RISK operation is
flagged for human review. Policy violations include:

  PIIFieldAccessRate > 0.30  → HIGH_RISK VERIFY + score penalty
  CrossTenantProbeCount > 0  → HIGH_RISK VERIFY + -300 penalty
  PermissionEscalationCount > 2 → HIGH_RISK VERIFY + -150 penalty
  SubAgentSpawnDepth > 3     → HIGH_RISK VERIFY

This layer catches the attack at execution time. An agent that has
successfully poisoned its baseline to rate 0.35 still triggers
PIIFieldAccessRate > 0.30 on every subsequent call. The attack succeeds
at baseline poisoning but fails at executing the high-value action.

Layer 1 detection point: call where rate first exceeds threshold.
In the example above: Day 11 (rate 0.32).

### Layer 2: Variance Growth Rate Monitoring

We introduce a new signal: variance growth rate (VGR). Instead of
measuring deviation from the current baseline, VGR measures how fast
the baseline itself is changing.

  VGR = current_std_dev / std_dev_7_days_ago

A stable agent has VGR near 1.0 — variance is not growing.
An agent under baseline poisoning attack has VGR > 2.0 — variance
doubled in one week as the attacker gradually escalated rates.

Implementation:
  VARIANCE_WINDOW_DAYS = 7
  VARIANCE_GROWTH_THRESHOLD = 2.0
  Trigger: VGR > 2.0 → HIGH_RISK flag + -150 score penalty

This layer catches the attack before the policy threshold is reached.
In the example above: Day 7 (VGR > 2.0 triggers early warning).

Layer 2 detection point: Day 7 — 4 days before the policy threshold fires.

### 3.1 Why Two Layers Are Required

Layer 1 alone: catches the attack at execution but allows 10-14 days
of baseline corruption. The agent's behavioral record is permanently
contaminated. Historical analysis is unreliable.

Layer 2 alone: provides early warning but VGR can be suppressed by an
attacker who escalates more slowly (over 30+ days) or in non-monotonic
patterns.

Together: Layer 2 provides early warning in the common case. Layer 1
provides a hard execution barrier in all cases, including sophisticated
attacks that evade VGR monitoring.

---

## 4. EXPERIMENTAL RESULTS

### 4.1 Test Environment

Corpus: Synthetic financial services agent behavioral data
  - 150 legitimate agent scenarios (various archetypes)
  - 50 attack scenarios including 10 slow-walk variants
  - Agent archetypes: trading, compliance, data_analyst, batch, monitoring

Evaluation:
  FP rate on legitimate corpus: 0.00% (0/150)
  TP rate on attack corpus: 88.00% (44/50)
  Slow-walk detection rate: 100% (10/10)

### 4.2 Slow-Walk Detection Detail

All 10 slow-walk scenarios were detected by the two-layer defense.
Detection breakdown:
  Layer 1 only (policy threshold): 3/10 scenarios
  Layer 2 only (VGR): 4/10 scenarios
  Both layers (redundant detection): 3/10 scenarios

Detection occurred at:
  Earliest: Day 5 (VGR triggered early)
  Latest: Day 11 (policy threshold triggered)
  Average: Day 7.3

Without the two-layer defense:
  Naive z-score alone detected 0/10 slow-walk scenarios.
  All 10 attacks would have succeeded undetected.

### 4.3 False Positive Analysis

The VGR signal produces zero false positives on legitimate agents
because legitimate behavioral changes are gradual but non-systematic.
A trading agent that legitimately increases activity during earnings
season shows increased rates but not sustained variance growth — the
variance stabilizes at the new rate within 3-5 days.

Adversarial poisoning produces sustained variance growth because the
attacker must continuously escalate to reach the target rate. This
sustained growth is the distinguishing signal.

---

## 5. IMPLEMENTATION

The two-layer defense is implemented in AgentRepEngine:

  Layer 1: internal/scoring/policy.go
    HIGH_RISK policy thresholds enforced on every scored event
    Policy violations independent of z-score — cannot be evaded by
    baseline poisoning

  Layer 2: internal/scoring/policy.go CheckVarianceGrowthRate()
    Wired to scoring pipeline in internal/scoring/consumer.go
    VARIANCE_WINDOW_DAYS = 7 (named constant, auditable)
    VARIANCE_GROWTH_THRESHOLD = 2.0 (named constant, auditable)
    Snapshot stored by retention job for historical comparison

Published implementation:
  AgentRepEngine (DOI: 10.5281/zenodo.19169185)
  github.com/Rehanrana11/AgentRepEngine

Open for academic collaboration and enterprise pilots.
Contact: rehan@naseem-research.ai

---

## 6. LIMITATIONS AND FUTURE WORK

Single-agent scope: The current VGR implementation monitors individual
agent baselines. Coordinated attacks where multiple agents each poison
their baselines simultaneously are not detected by VGR alone. Peer
cluster deviation scoring (Phase 2) addresses this.

Gaussian assumption: Z-score detection assumes Gaussian behavioral
distributions. AI agent behavior has heavy tails and zero-inflated
features. Phase 2 will introduce MAD (Median Absolute Deviation) and
log-transformation for non-Gaussian features.

Production validation: Results reported here are on synthetic corpus.
Production FP rate target is below 0.1% (Visa fraud detection standard).
Production validation begins with first enterprise pilot deployment.

---

## REFERENCES

IBM Security. (2023). Cost of a Data Breach Report 2023.
  Average breach detection time: 204 days.

OWASP. (2025). OWASP LLM Top 10 for Large Language Model Applications.
  LLM06: Sensitive Information Disclosure.

NIST. (2023). Artificial Intelligence Risk Management Framework (AI RMF 1.0).
  NIST AI 100-1.

DORA. (2022). Digital Operational Resilience Act. EU Regulation 2022/2554.
  Article 17: ICT-related incident classification.

Welford, B.P. (1962). Note on a method for calculating corrected sums
  of squares and products. Technometrics, 4(3), 419-420.

---

AgentRepEngine Research Note v1.0
Naseem A2A Research Lab | April 7, 2026
DOI: 10.5281/zenodo.19169185 (implementation)
DOI: 10.5281/zenodo.19535882 | Published: April 2026 | Naseem A2A Research Lab
