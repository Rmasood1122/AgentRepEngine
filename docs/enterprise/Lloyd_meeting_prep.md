# Lloyd Meeting Prep — AgentRepEngine
# NWN Partner Conversation | Week of April 7, 2026
# CONFIDENTIAL

---

## ONE-SENTENCE PRODUCT DESCRIPTION

"ARE is a behavioral enforcement layer for AI agents — it sits in
your gateway, scores every agent action against its own baseline,
and blocks anomalous behavior before it commits. Like a firewall,
but for what AI agents do, not what packets they carry."

---

## 15 TALKING POINTS — KNOW COLD

Deliver in response to questions. Do not recite in sequence.

### DATA SOVEREIGNTY + AUDIT
(1) "ARE runs inside your Kong gateway — behavioral data never
    leaves your network, and every enforcement decision is in a
    tamper-evident log your auditors verify themselves."

(2) "ARE's audit trail is hash-chained and INSERT-only at the
    database level. Your auditors don't trust our attestation —
    they verify the chain themselves. That's audit-native
    architecture, not a compliance claim."

### DETECTION ARCHITECTURE
(3) "ARE implements the anomaly detection architecture NIST and
    OWASP now recommend for AI agent security — behavioral
    baseline plus deviation scoring. We built it before the
    standard was published."

(4) "ARE uses ensemble enforcement — score AND policy must both
    flag before blocking. Two independent detection systems must
    agree. That's the architectural reason for 0.00% false
    positive rate on our internal corpus. No single miscalibrated
    metric can block a legitimate agent."

(5) "ARE doesn't just say BLOCK or ALLOW. It tells you it's 94%
    confident this agent is anomalous, based on 30 days of its
    own behavioral baseline."

(6) "ARE detects slow-walk attacks that take 7 days to execute —
    the kind that score-based systems miss entirely. We monitor
    variance growth rate; a doubling of weekly variance triggers
    an early warning before the attack succeeds. 100% detection
    rate on our 10-scenario slow-walk corpus."

### DEPLOYMENT + BYPASS
(7) "Enforcement is at the Kong gateway layer, below the
    application. The agent has no visibility into it and no
    way to route around it."

(8) "Phase 1 pilot is customer-hosted — Docker Compose, Kong,
    Redis, PostgreSQL, running entirely within your network
    perimeter. No SaaS dependency. No data egress."

(9) "ARE is framework-agnostic — LangChain, LlamaIndex, custom
    agents. If it goes through Kong, ARE sees it. Zero changes
    to your agent code."

### RELIABILITY + HUMAN OVERSIGHT
(10) "ARE fails open — if the scoring service goes down, agents
     keep running, the audit trail is preserved, and your SOC
     sees the gap before you ask."

(11) "ARE implements human-in-the-loop by design — every
     enforcement decision in observe mode is reviewed by your
     security team before auto-block is enabled. You maintain
     human judgment throughout. Overrides feed back into
     baseline calibration."

### COMPLIANCE + REGULATORY
(12) "ARE passes every item on the regulatory accountability
     checklist — audit trail, incident response, human oversight,
     corrective action. Out of the box. DORA Article 45, GDPR
     Article 22, SEC AI governance — one architecture."

(13) "Every enforcement decision is human-readable — agent ID,
     score, confidence, contributing factors, timestamp. Your
     compliance team reads it without a data scientist. That's
     GDPR Article 22 right-to-explanation, built in."

### DATA MOAT + COMPOUNDING
(14) "The behavioral baseline updates on every transaction via
     Welford's online algorithm. Enforcement gets more precise
     the longer ARE runs — not less. Day 90 is significantly
     sharper than day 1. That's also your switching cost:
     90 days of org-specific behavioral data that no other
     vendor has."

### REGULATORY INEVITABILITY CLOSE
(15) "Regulators are about to require AI audit trails. ARE
     doesn't help you prepare for that requirement — ARE is
     the implementation of that requirement, already running."

---

## OBJECTION MAP

| Lloyd says...                         | Use...         |
|---------------------------------------|----------------|
| "Our data can't leave our network"    | (1) + (8)      |
| "How do we prove this to auditors?"   | (2) + (13)     |
| "Can it be bypassed?"                 | (7)            |
| "What if your system goes down?"      | (10)           |
| "Will it work with our stack?"        | (9)            |
| "Is this proven in finserv?"          | (4) + (6)      |
| "What's the ROI?"                     | Pilot offer ↓  |
| "How do we know it's accurate?"       | (4) + (5)      |
| "Will it catch sophisticated attacks?"| (6)            |
| "Will it get outdated?"               | (14)           |
| "Is this compliant with DORA/GDPR?"   | (3) + (12) + (15) |
| "What's the risk of a false block?"   | (4) + (10) + (11) |
| "Who's responsible if it goes wrong?" | (11) + (2)     |

---

## METRICS — NO NOTES

FP rate:              0.00% on 100-scenario internal corpus
TP rate:              86.67%
F1 score:             0.9286
Slow-walk detection:  100% (10/10 multi-day scenarios)
Call overhead:        0.25ns Linux — zero allocations
Demo runtime:         ~30 seconds (Windows Docker) — streaming output
Hardening score:      94/100

Correct framing for FP: "0.00% false positive rate on our
internal validation corpus. External validation on your
production traffic is what the 30-day pilot produces."

---

## PILOT OFFER — THE CLOSE

"30-day observe mode. Zero enforcement, zero risk. At day 30
you have the full audit trail, the ROI case — labor saved,
regulatory risk quantified, incidents documented — and the
decision is yours. We can have the letter of understanding
ready this week."

Pilot structure if asked:
- Week 1–2: Observe only. Baseline establishment. Zero enforcement.
- Week 3: Flag anomalies. Review with your security team. No auto-block.
- Week 4–6: Enforce mode. Human review on all blocks for first 7 days.
- Week 7–30: Full enforcement with auto-rollback protection.

You control the pace. We don't advance without your sign-off.

---

## REGULATORY INEVITABILITY CLOSE

Use this as the final sentence if the conversation is going well:

"Regulators are about to require AI audit trails. ARE doesn't
help you prepare for that requirement — ARE is the implementation
of that requirement, already running."

---

## WHAT NOT TO DO

- Never mention revenue projections unless Lloyd raises value first
- Never use the word "MVP"
- Never show the demo unsolicited — offer it
- Never pitch funding status, Character Capital, or investors
- Never mention Check Point by name unless Lloyd raises competition
- Never recite these points in sequence — respond, don't present

---

## SEQUENCE IF LLOYD GOES TECHNICAL

1. One-sentence description
2. "What agent frameworks are your clients deploying?" → listen
3. (9) framework-agnostic → (7) gateway layer → (3) NIST/OWASP
4. Offer demo: "I can show you a 12-second detection sequence
   against a slow-walk attack. Want to see it?"
5. Post-demo: (5) confidence → (13) reason object → (12) checklist
6. Close: Pilot offer → (15) regulatory inevitability

## SEQUENCE IF LLOYD GOES COMMERCIAL

1. One-sentence description
2. (1) data sovereignty → (8) customer-hosted
3. (12) accountability checklist → (2) audit-native
4. Pilot offer
5. (15) regulatory inevitability close

---

## REHEARSAL CHECKLIST — DO BEFORE THE MEETING

Say these aloud. Not in your head. Aloud.

[ ] One-sentence product description — under 15 seconds
[ ] Point (4) — ensemble enforcement / 0.00% FP explanation
[ ] Point (5) — confidence percentage claim
[ ] Point (6) — slow-walk / variance growth rate
[ ] Point (10) — fail-open
[ ] Point (14) — self-sharpening baseline
[ ] Point (15) — regulatory inevitability close
[ ] Pilot offer — under 30 seconds
[ ] All 7 metrics without looking

If you stumble → repeat until clean. The meeting moves fast.

---

*Updated: March 31, 2026 | Merged v3.1 + C1–C15*
*Commit this. Do not share externally.*