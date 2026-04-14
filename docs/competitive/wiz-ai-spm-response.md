# AgentRepEngine — Competitive Response: Wiz AI-SPM
**Document:** docs/competitive/wiz-ai-spm-response.md
**Date:** April 13, 2026
**Classification:** Internal — not for distribution
**Status:** UT-1 — untracked competitor, added to competitive registry

---

## Summary

Wiz is the fastest-growing cloud security company in history, with
approximately 35% Fortune 500 CSPM (Cloud Security Posture Management)
penetration as of Q1 2026. In Q1 2026, Wiz launched AI-SPM (AI Security
Posture Management), extending their platform into AI infrastructure security.

**Current Wiz AI-SPM scope:** Data exposure analysis, AI model inventory,
misconfiguration detection, and sensitive data flow mapping for AI pipelines.

**Current Wiz AI-SPM gap:** No per-agent behavioral baselines. No runtime
behavioral enforcement. No z-score anomaly detection against agent history.
No gateway-layer interception. No tamper-evident audit trail per agent action.

**Threat level:** HIGH. Not because of current product overlap — there is
minimal overlap today. Because of distribution: Wiz has direct relationships
with every CISO who already runs their CSPM. One product release from Wiz
could reframe the conversation from "ARE vs nothing" to "ARE vs Wiz."

**Competitive clock implication:** Be installed before Wiz ships behavioral
enforcement. A deployed ARE instance is defensible. A prospect considering
their first AI security purchase is not.

---

## What Wiz AI-SPM Does (Accurate, No Disparagement)

Wiz AI-SPM operates at the **posture layer** — it answers the question:
"What AI models, agents, and pipelines exist in my environment, and are
they configured securely?"

Specific capabilities (as of Q1 2026):
- AI asset inventory: discovers AI models, agents, and APIs across cloud environments
- Sensitive data exposure: identifies AI models with access to sensitive data
- Misconfiguration detection: flags AI services with overly permissive access
- Data flow mapping: traces how data moves through AI pipelines
- Supply chain risk: identifies vulnerable AI model dependencies

Wiz's detection surface: **configuration and posture** — what your AI
infrastructure looks like at rest, not what it is doing in real time.

---

## What ARE Does That Wiz Does Not

ARE operates at the **runtime enforcement layer** — it answers the question:
"What is each AI agent doing right now, and is it anomalous relative to
its own 30-day behavioral history?"

| Capability | ARE | Wiz AI-SPM |
|---|---|---|
| Per-agent behavioral baseline | ✅ 30-day EWMA per agent | ❌ No |
| Z-score anomaly detection vs agent history | ✅ Real-time | ❌ No |
| Gateway-layer enforcement (below app layer) | ✅ Kong plugin | ❌ No |
| Agent cannot see or route around enforcement | ✅ Architectural | ❌ No |
| Tamper-evident audit trail per decision | ✅ Hash-chained | ❌ No |
| Auto-rollback on FP spike | ✅ PostgreSQL trigger | ❌ No |
| GDPR Article 22 reason object per decision | ✅ Every block | ❌ No |
| Slow-walk attack detection (variance growth) | ✅ TW-6 wired | ❌ No |
| Deploy in 4 hours, zero SDK changes | ✅ Docker Compose | ❌ SaaS only |
| Customer-controlled audit data | ✅ Your PostgreSQL | ❌ Wiz cloud |

---

## The Architectural Distinction (Use This In Conversation)

**Never say "better than Wiz."** Say this instead:

> "Wiz tells you what your AI agents can access — their posture, their
> configuration, their data exposure. ARE tells you what your AI agents
> are doing right now, scored against their own 30-day behavioral history.
>
> Wiz sees a misconfigured AI service. ARE sees an AI agent that has been
> operating 4.2 sigma above its own baseline for 90 minutes.
>
> These are different detection surfaces. A well-configured agent can still
> go rogue. A misconfigured agent may never actually exfiltrate data.
>
> The question for your architecture is: does your environment need both
> a posture layer and a runtime enforcement layer? If you already run Wiz,
> ARE is the runtime enforcement layer Wiz does not provide."

**Key phrase:** "Wiz and ARE are complementary, not competing."
This positioning prevents the buyer from framing it as a choice.

---

## If A Buyer Raises Wiz Directly

**Buyer says:** "We already have Wiz. Do we need ARE too?"

**Response:**

> "Wiz gives you excellent visibility into your AI infrastructure posture.
> It does not give you per-agent behavioral enforcement at runtime.
>
> Here is the specific gap: an AI agent with a perfectly clean Wiz posture
> scan — correctly configured, no sensitive data exposure flags — can still
> systematically exfiltrate data over 11 days if its behavioral baseline
> drifts and nothing is watching the pattern.
>
> Wiz sees the configuration. ARE sees the behavior.
>
> Does your Wiz deployment currently include behavioral enforcement for
> AI agents at the gateway layer? If not, you have posture coverage
> without runtime enforcement. Those are two different problems."

---

## If Wiz Ships Behavioral Enforcement (Future Scenario)

If Wiz releases a behavioral enforcement feature, the response is:

> "Wiz enforces from the cloud posture layer — their data plane is SaaS,
> their enforcement requires routing traffic through Wiz infrastructure.
>
> ARE enforces from inside your Kong gateway — below your application layer,
> inside your network perimeter, with zero data leaving your environment.
>
> For regulated enterprises under HIPAA, DORA, and SOX: the question is
> whether enforcement data can leave your perimeter. ARE's architecture
> means it cannot. That is not a feature — it is a compliance requirement
> for your environment."

---

## Monitoring Cadence

Review Wiz AI-SPM product updates quarterly.
Trigger for immediate review: any Wiz announcement mentioning
"behavioral," "runtime," "agent scoring," or "enforcement."

Add to risk register: Kong open-source licensing (UT-2) and Wiz
behavioral enforcement (UT-1) are the two highest-probability threats
to ARE's competitive position in the next 12 months.

---

## Regulatory Anchor (Use With CISO)

Wiz AI-SPM does not produce:
- A per-decision reason object satisfying GDPR Article 22
- A hash-chained audit trail satisfying HIPAA §164.312(c)(1)
- A SQL-queryable FP rate satisfying NIST AI RMF MEASURE 2.5
- A customer-controlled enforcement record satisfying DORA Article 8(4)

ARE produces all four. In a regulated enterprise, these are not
optional features — they are audit requirements. Wiz's SaaS model
means these artifacts live in Wiz's cloud, not the customer's perimeter.

---

*docs/competitive/wiz-ai-spm-response.md | April 13, 2026*
*UT-1 closed — Wiz added to competitive registry*
*Review trigger: any Wiz product announcement mentioning behavioral enforcement*
