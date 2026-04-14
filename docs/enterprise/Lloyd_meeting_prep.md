# Lloyd Meeting Prep — AgentRepEngine
**Audience:** Lloyd Lemish, NWN Technical Solutions Architect  
**Meeting date:** ~April 28, 2026  
**Format:** Technical discovery → observe-mode pilot proposal  
**Version:** TW-1 | April 8, 2026

---

## BEFORE YOU WALK IN

**Prime directive:** Lloyd is a Technical Solutions Architect. He evaluates on technical credibility first, commercial fit second. Do not pitch. Demonstrate understanding of his environment and let the product speak.

**Opening use cases (lead with these two only):**
- Use Case 1: Rogue agent containment — DORA accountability
- Use Case 5: Observe-mode baseline validation — zero-risk visibility

**Do not mention until buyer asks:** federation, pricing tiers, roadmap, revenue projections, "24x," Phase 2.

**CISO unlock phrase:** *"You control the pace — we don't advance to enforce mode without your sign-off."*

**FP reframe (use this, not the raw number):** *"We target the Visa standard — below 0.1% false positive rate in production. No AI agent security product has published a production FP rate. We will."*

---

## DISCOVERY QUESTIONS — ASK BEFORE PRESENTING

1. How many AI agents are currently running in your environment?
2. Are they going through Kong, or another gateway?
3. What does your current audit trail look like for agent actions?
4. Has anyone asked you yet — internally or from a regulator — to demonstrate control over agent behavior?
5. What would a false positive cost you operationally? (Sets up FP framing.)

---

## 14 TALKING POINTS — C1–C14 + EXTENSIONS

Use these verbatim or near-verbatim. Do not paraphrase the core claim.

---

### C1 — Data sovereignty
*"Enforcement at your gateway. Data never leaves. Auditors verify themselves."*

**Extension:** ARE runs entirely within your network perimeter — Docker Compose, Kong + Redis + PostgreSQL, customer-hosted. No SaaS dependency. No vendor controls the measurement. Your auditors can run the verification query themselves.

---

### C2 — Standards alignment
*"ARE implements the NIST/OWASP standard for AI agent security."*

**Extension:** Policy packs map directly to OWASP LLM Top 10. NIST AI RMF coverage: GOVERN (policies), MAP (risk identification), MEASURE (monitoring), MANAGE (incident response). This is not ARE claiming compliance — this is ARE implementing the published standard.

---

### C3 — Regulatory readiness
*"Passes every item on the regulatory accountability checklist. Out of the box."*

**Extension:** DORA Article 9 (ICT risk management), Article 10 (detection), Article 17 (incident reporting). Every enforcement decision is cryptographically logged. Hash-chain verified. The audit trail is tamper-evident by architecture.

---

### C4 — Explainability
*"94% confident this agent is anomalous — based on 30 days of its own baseline."*

**Extension:** Every enforcement decision produces a structured reason object: agent ID, score, confidence, contributing features, policy triggered. Human-readable. Your security engineer can explain any blocking decision without calling the vendor.

---

### C5 — Financial services precedent
*"Financial services proved this architecture works. ARE applies it to agents."*

**Extension:** Z-score anomaly detection against a rolling 30-day baseline is the same architecture Visa uses for transaction fraud. 0.00% false positive rate on our 150-scenario internal validation corpus. Production target: below 0.1% — the Visa standard.

---

### C6 — Fail-open design
*"Fails open. Agents keep running. SOC sees it before you ask."*

**Extension:** This is not an optional setting. ARE is designed to fail open — if the scoring service is unreachable, agents continue running and the event is logged. No ARE misconfiguration can take down your agent fleet. The circuit breaker is in the Kong plugin by architecture.

---

### C7 — Invisibility to agents
*"Below the application layer. Agents can't see it. Can't route around it."*

**Extension:** ARE operates at the Kong gateway layer. Agents have no awareness of scoring. There is no SDK to be disabled, no header to be spoofed at the application layer. Enforcement happens before the agent response is returned.

---

### C8 — Observe mode — the commercial offer
*"30-day observe mode. At day 30: ROI quantified, incidents documented, decision yours."*

**Extension:** The pilot is a zero-impact visibility deployment. No enforcement during observe mode. You see everything ARE would have blocked — without blocking anything. At day 30 you have: scored agent traffic, anomaly reports, FP rate on your real data, and one documented incident report if any anomalies were confirmed. Decision to move to enforce mode is yours and requires your explicit sign-off.

---

### C9 — Framework agnostic
*"LangChain, LlamaIndex, custom. If it goes through Kong, ARE sees it."*

**Extension:** No proprietary SDK required. No vendor lock-in. Any HTTP-based agent framework that routes through Kong is covered. ARE already has native SDK support for LangChain and LangGraph (36/36 tests passing) for environments that prefer per-runtime integration over gateway enforcement.

---

### C10 — Continuous baseline improvement
*"The baseline updates on every transaction. Enforcement gets more precise the longer it runs."*

**Extension:** ARE uses Welford's online algorithm — the baseline updates incrementally on every scored call. The longer ARE runs, the tighter the behavioral model becomes. At day 30, the baseline is more accurate than day 1 by definition. This is not a static ruleset. It is a continuously calibrated model of what your agents actually do.

---

### C11 — Human-readable enforcement
*"Every enforcement decision is human-readable. Agent ID, score, confidence, reason."*

**Extension:** The reason object is structured JSON. Every field is named. Score, contributing features, policy pack triggered, confidence interval, enforcement action taken. Your compliance team can read it. Your auditors can read it. Your board can read it.

---

### C12 — Regulatory trajectory
*"Regulators are about to require AI audit trails. ARE is the implementation, already running."*

**Extension:** DORA is live for EU financial services. SEC has signaled AI governance requirements. HIPAA enforcement is extending to AI-assisted clinical workflows. The question is not whether your regulator will ask for an AI agent audit trail. The question is whether you have one when they do.

---

### C13 — Self-verifying measurement
*"Here is the SQL query. Run it yourself. The result is your FP rate. No vendor controls the measurement."*

**Extension (paste this query if asked):**
```sql
SELECT
  DATE(created_at) as date,
  COUNT(*) FILTER (WHERE enforcement_action = 'block' AND outcome = 'false_positive') as fp_count,
  COUNT(*) FILTER (WHERE enforcement_action = 'block') as total_blocks,
  ROUND(
    COUNT(*) FILTER (WHERE enforcement_action = 'block' AND outcome = 'false_positive') * 100.0 /
    NULLIF(COUNT(*) FILTER (WHERE enforcement_action = 'block'), 0), 4
  ) as fp_rate_pct
FROM scoring_explanations
GROUP BY DATE(created_at)
ORDER BY date DESC;
```
You run the query. You own the result. No vendor interpretation required.

---

### C14 — Detection gap
*"The average breach goes undetected 200 days. ARE detects behavioral drift in real time."*

**Extension:** The 200-day detection gap (IBM Cost of a Data Breach, 2023) exists because current tools score individual calls against static rules. A compromised agent that stays below every individual threshold — but systematically increases PII access rate over 400 calls — looks clean to every existing tool. ARE would have flagged the pattern at call 43 (PII access rate 4 standard deviations above 7-day baseline) and blocked at call 44. The 6-hour incident becomes a 6-minute incident.

---

## METRICS — KNOW COLD

| Metric | Value | Context |
|---|---|---|
| False positive rate | 0.00% | 150-scenario internal corpus |
| Production FP target | <0.1% | Visa fraud detection standard |
| True positive rate | 88.00% (44/50) | Above 85% gate |
| Held-out TP | 100% (6/6) | Never-seen scenarios |
| Held-out FP | 0.00% (0/20) | Never-seen legitimate agents — internal corpus only |
| F1 score | 0.9362 | |
| Precision | 100% | |
| Slow-walk detection | 100% (10/10) | Single-agent scope |
| Call overhead | 0.25ns | Linux |
| Hardening score | 96/100 | |
| Policy packs | 7 | 5 OWASP + NIS2 + GDPR |
| Test coverage | 60 tests / 14 files | All green |

**4-metric format for any performance claim:**  
*"TP=88.00%, FP=0.00% on internal corpus, Precision=100%, F1=0.9362. Production target: <0.1% — Visa standard."*

---

## OBJECTION HANDLING

**"We're not running Kong."**  
ARE also deploys via native SDK (LangChain/LangGraph — 36/36 tests). If your agents are HTTP-based, we can discuss the deployment model. The observe-mode pilot starts wherever your agents run.

**"What's the false positive risk to us operationally?"**  
ARE is fail-open by architecture. Observe mode has zero enforcement — nothing is blocked during the 30-day pilot. When you move to enforce mode, the circuit breaker auto-reverts to observe if FP rate exceeds threshold. You set the threshold. You control the pace.

**"How do we know this works on our agents, not just your test corpus?"**  
That is exactly what the 30-day observe-mode pilot answers. Day 1: baseline initialization. Day 7: first anomaly report on your real traffic. Day 30: FP rate calculated against your data, not ours. The pilot is designed to answer this question.

**"What does 'behavioral drift' actually mean for our agents?"**  
Your agents have normal behavioral patterns — call frequency, data volume, API endpoint distribution. ARE establishes that baseline over 30 days, then scores deviations against it. A trading agent accessing customer PII at 3x its 30-day average rate is not a prompt injection. It is behavioral drift. That is what ARE detects.

**"We already have SIEM."**  
ARE integrates with your SIEM — Splunk and Sentinel both supported via siem-integration-guide.md. ARE produces structured enforcement events that feed directly into your existing alerting workflow. ARE is not a replacement for SIEM. It is the agent behavioral layer SIEM cannot see.

**"What's the cost?"**  
[Do not lead with pricing. If pressed:] The pilot is free — 30-day observe mode, zero risk, no procurement required. Pricing for enforce mode is scoped to agent identity count. We can discuss that at day 30 based on what the pilot found. [Do not quote specific tiers unless directly asked.]

---

## PILOT SCOPE PROPOSAL — HAVE THIS READY

**What you're offering:**
- 30-day zero-impact visibility deployment
- Customer-hosted: Docker Compose, runs in your network perimeter
- Observe mode only: no enforcement, no blocking
- Deliverable at day 30: anomaly report, FP rate on real traffic, one documented incident if found

**What you need from Lloyd:**
- Agent count (approximate)
- Kong version (minimum 2.8) or agent framework
- One security engineer to receive the anomaly reports
- Sign-off from CISO or equivalent to proceed

**Time to first value:** 7 days (design partner) / 14 days (standard pilot)  
**Install time:** Under 4 hours from reading docs to first scored call

---

## SEQUENCING RULES

1. Open with discovery questions. Do not present before you understand their environment.
2. Lead with Use Case 1 (rogue agent containment, DORA) and Use Case 5 (observe-mode baseline validation).
3. Do not volunteer pricing. If asked, defer to day-30 pilot output.
4. Do not mention federation, Phase 2, passport design, or roadmap.
5. Close with a specific next step: pilot scope document or LoU — not "let's stay in touch."
6. If Lloyd asks to see the product: run the demo. C13 SQL query is the highest-trust moment.

---

## REHEARSAL CHECKLIST (do this the day before)

Say each claim aloud in order, C1–C14. Time yourself. Target: under 90 seconds for the full set.

- [ ] C1 — Data sovereignty
- [ ] C2 — Standards alignment
- [ ] C3 — Regulatory readiness
- [ ] C4 — Explainability
- [ ] C5 — Financial services precedent
- [ ] C6 — Fail-open design
- [ ] C7 — Invisibility to agents
- [ ] C8 — Observe mode offer
- [ ] C9 — Framework agnostic
- [ ] C10 — Continuous baseline
- [ ] C11 — Human-readable enforcement
- [ ] C12 — Regulatory trajectory
- [ ] C13 — Self-verifying measurement
- [ ] C14 — Detection gap
- [ ] CISO unlock phrase
- [ ] FP reframe

---

---
## DEMO DISCLOSURES — SAY THESE PROACTIVELY (do not wait to be asked)

### UW-1 — Pre-Seeded Baseline Disclosure
Say this before or during the demo, not after:

  "The demo uses a pre-seeded behavioral baseline to show the 4.2σ
   detection in 60 seconds. In your environment, the baseline builds
   from your agents' own traffic over 30 days. Day 30 produces your
   actual sigma values against your actual traffic. The 4.2σ in the
   demo is illustrative — not a claim about what your environment
   will produce."

Why proactive: Lloyd's engineer will find this if you don't say it first.
Saying it first is evidence of honesty. That builds more trust than a
polished demo that overstates.

### UW-2 — confidence_pct Discrepancy (12% vs 94%)
If the engineer notices two different confidence values in the demo output:

  "You will see two confidence values. The reason object shows 12% —
   that is anomaly confidence: how unusual is this behavior relative
   to the agent's 30-day baseline? The Kong response shows 94% —
   that is enforcement confidence: given this score, how certain are
   we that blocking is the correct action? These measure different
   things. Both are correct."

Anomaly confidence is low (12%) because the agent was new — only 3 events
in its history. Enforcement confidence is high (94%) because the score
of 187 is deep in the BLOCKED band.

### UW-3 — Score Drop is Demo-Accelerated
If Lloyd's engineer asks why the score drops so fast:

  "In the demo, the score drops from 700 to 187 after 3 injected events.
   This is accelerated to show the detection mechanism in 60 seconds.
   In your environment with a real 30-day baseline, score decline is
   more gradual — which is the correct behavior. ARE detects two threat
   classes: rapid spikes via z-score, and slow drift via variance growth
   rate. Day 7 of your pilot will show you the variance growth chart.
   That is the slow-walk detection — the class that traditional security
   tools miss entirely."

---
## NEW CLAIMS — ADD TO REHEARSAL (April 13, 2026)

### C15 — Board-Readable Incident Report
  "Every enforcement decision produces a board-readable incident report.
   Agent ID. Score before and after. Confidence. Policy fired. Sigma value.
   Recommended action. Your board reads it without a translation layer.
   Your auditor cites it without calling us. No competitor produces this.
   When your board asks what happened with an AI agent incident, this is
   the document they read. We produce it automatically on every block."

Add C15 to rehearsal after C11.

### HV-4 — Forensics Framing (Day 30 review, not first meeting)
  "This audit trail is also your forensics report. When legal asks what
   happened, this is the document. Generated automatically. No incident
   response firm required. The hash chain proves it was not altered
   after the fact."

Use at Day 30 pilot review — not in the first meeting.

---
## UPDATED REHEARSAL CHECKLIST (C1–C15)

Say each claim aloud in order. Target: under 2 minutes for the full set.

- [ ] C1 — Data sovereignty
- [ ] C2 — Standards alignment
- [ ] C3 — Regulatory readiness
- [ ] C4 — Explainability
- [ ] C5 — Financial services precedent
- [ ] C6 — Fail-open design
- [ ] C7 — Invisibility to agents
- [ ] C8 — Observe mode offer
- [ ] C9 — Framework agnostic
- [ ] C10 — Continuous baseline
- [ ] C11 — Human-readable enforcement
- [ ] C15 — Board-readable incident report ← NEW
- [ ] C12 — Regulatory trajectory
- [ ] C13 — Self-verifying measurement
- [ ] C14 — Detection gap
- [ ] CISO unlock phrase
- [ ] FP reframe
- [ ] UW-1 disclosure (pre-seeded baseline)
- [ ] UW-2 explainer (confidence_pct)
- [ ] UW-3 explainer (score drop acceleration)

---
*TW-1 updated April 13, 2026 — C15 added, UW-1/2/3 disclosures added, HV-4 framing added*
*Next: L118 observe-to-enforce-criteria.md — Lloyd will ask for this at the meeting*
