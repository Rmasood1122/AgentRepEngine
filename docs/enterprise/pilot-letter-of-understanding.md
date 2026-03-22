# AgentRepEngine — Pilot Letter of Understanding
## 30-Day Observe-Mode Pilot Agreement
Version 1.0 | Confidential

---

### Parties

Vendor: Naseem A2A Research Lab (rehan@naseem-a2a.com)
Client: [Organization Name]
Date:   [Date]

---

### What This Is

This is not a contract. It is a mutual understanding of what
the 30-day pilot involves, what success looks like, and what
happens next. A formal agreement follows if both parties
choose to proceed after the pilot.

---

### Pilot Terms

**Duration:** 30 days from installation date
**Cost:** $0
**Mode:** OBSERVE only — no agent will be blocked
**Installation:** 4 hours on one environment
**Data:** All behavioral telemetry stays in your Kubernetes cluster

---

### What We Install

One Kong gateway plugin and one scoring service container.
No changes to your existing agents or infrastructure.
Full rollback in under 10 minutes at any time.

---

### What You Get

- Every agent request scored in real time
- Behavioral anomaly detection across sessions
- Full audit trail with tamper-evident hash chain
- SOC2-compatible export available at any time
- SIEM webhook on any BLOCKED-equivalent detection
- Daily summary of scored agent activity

---

### Success Definition

The pilot succeeds if within 30 days:

1. At least one behavioral anomaly is detected and logged
2. Zero production outages caused by the AgentRepEngine layer
3. Your security team can explain one enforcement decision
   without asking us

If none of these occur within 30 days, the pilot has failed
and we will tell you so directly.

---

### What Happens Next

If pilot succeeds and you choose to continue:

- Enforce mode available after your team reviews observe logs
- Pricing: $10,000/month per environment
- Annual: $100,000/year per environment (2 months free)
- Enterprise: volume pricing for 3+ environments

If pilot fails or you choose not to continue:
- We uninstall completely in under 10 minutes
- You keep all audit logs generated during the pilot
- No obligation, no invoice, no follow-up pressure

---

### Override Authority

You retain full override authority at all times.
ENFORCEMENT_MODE switch is yours to control.
We never flip it without your explicit instruction.
Auto-rollback to observe fires automatically if FP rate
exceeds 2% — protecting you from misconfiguration.

---

### What We Need From You

- One security engineer available for the 4-hour install
- Kong Gateway 3.6.x running in your environment
- PostgreSQL 16+ and Redis 7+ (or we deploy via Docker)
- Network access between Kong and scoring service container
- 30 days of patience while observe data accumulates

---

### Signatures

This letter of understanding becomes effective when both
parties confirm via email.

Vendor confirmation: rehan@naseem-a2a.com
Client confirmation: [contact email]

---

*This document is guidance, not a legal contract.
A formal MSA and DPA are available on request.*
