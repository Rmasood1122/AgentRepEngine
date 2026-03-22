# AgentRepEngine — Competitive Positioning
# One-pager for enterprise conversations
# Last updated: March 22, 2026

---

## THE ONE SENTENCE

AgentRepEngine is the only product that enforces behavioral trust
at the gateway — scoring every AI agent request in real time and
blocking anomalous actions before data leaves.

---

## COMPETITIVE LANDSCAPE

### Lakera (now Check Point)
**What they do:** Prompt injection detection — scans LLM input/output
for malicious content before it reaches the model.

**What they do NOT do:**
- Runtime behavioral scoring across sessions
- Gateway-level enforcement on agent actions
- Tamper-evident audit trail of agent decisions
- Feature vector tracking over time

**The gap:** Lakera stops bad prompts going IN.
AgentRepEngine stops bad behavior coming OUT.
These are different problems. Different layers. Not competing.

**When Check Point bundles Lakera:** They will have prompt security.
They will not have runtime behavioral enforcement.
AgentRepEngine fills the gap Lakera cannot fill.

---

### Palo Alto Networks / Protect AI
**What they do:** AI supply chain security — model scanning,
dependency vulnerabilities, MLOps pipeline security.

**What they do NOT do:**
- Runtime enforcement on deployed agent behavior
- Per-request behavioral scoring
- Identity-based trust model for agents

**The gap:** Palo Alto secures the model before deployment.
AgentRepEngine secures the agent after deployment — in production,
on every request, in real time.
Complementary. Not competing.

---

### Wiz / Orca (Cloud Security)
**What they do:** Cloud infrastructure posture — misconfigured
S3 buckets, IAM permissions, network exposure.

**What they do NOT do:** Watch what AI agents actually DO with
the access they have been granted.

**The gap:** Wiz tells you the door is locked.
AgentRepEngine tells you the agent is trying to open it anyway.

---

### SIEM / Splunk / QRadar
**What they do:** Log aggregation and alerting after events occur.

**What they do NOT do:**
- Block in real time before data leaves
- Score behavioral patterns across sessions
- Structured reason objects per enforcement decision

**The relationship:** AgentRepEngine FEEDS your SIEM.
Every blocked decision ships a structured reason object via webhook.
AgentRepEngine is not a SIEM replacement — it is a SIEM enrichment layer.

---

### Traditional API Gateways (Kong, Apigee, AWS API GW)
**What they do:** Rate limiting, authentication, routing.

**What they do NOT do:**
- Behavioral scoring per agent identity
- Cross-session pattern detection
- Tamper-evident enforcement audit trail

**The relationship:** AgentRepEngine IS a Kong plugin.
It runs inside your existing gateway.
No new infrastructure. No new vendor procurement.
Install time: 4 hours.

---

## THE MATRIX

| Capability | AgentRepEngine | Lakera | Palo Alto | SIEM |
|---|---|---|---|---|
| Runtime behavioral scoring | ✅ | ❌ | ❌ | ❌ |
| Gateway-level enforcement | ✅ | ❌ | ❌ | ❌ |
| Per-request blocking | ✅ | ❌ | ❌ | ❌ |
| Tamper-evident audit trail | ✅ | ❌ | ❌ | Partial |
| Structured SIEM reason object | ✅ | ❌ | ❌ | N/A |
| Agent identity model (JWT) | ✅ | ❌ | ❌ | ❌ |
| Slow-walk detection | ✅ | ❌ | ❌ | ❌ |
| Install time | 4 hours | Days | Weeks | Weeks |
| Prompt injection detection | ❌ | ✅ | ❌ | ❌ |
| Model supply chain security | ❌ | ❌ | ✅ | ❌ |
| Cloud posture management | ❌ | ❌ | Partial | ❌ |

---

## THE THREE QUESTIONS EVERY BUYER ASKS

**Q: Do we need this if we already have Lakera?**
> Yes. Lakera watches what goes INTO your LLM.
> AgentRepEngine watches what your agent DOES after it gets a response.
> A compromised agent can exfiltrate data without ever triggering
> a prompt injection alert. That is the gap AgentRepEngine closes.

**Q: Do we need this if we already have a SIEM?**
> AgentRepEngine does not replace your SIEM — it makes it smarter.
> Every enforcement decision ships a structured reason object to your SIEM.
> Your SIEM currently receives logs. AgentRepEngine gives it decisions.

**Q: Do we need this if we already have Palo Alto?**
> Palo Alto secured your model before deployment.
> AgentRepEngine secures your agent after deployment.
> The attack surface is different. A clean model can still
> be exploited by a compromised agent runtime.

---

## THE CHECK POINT CLOCK

Check Point acquired Lakera in 2025. Integration timeline: 12 months.
When bundling is complete, every Check Point renewal conversation
will include "AI security" — meaning prompt injection protection.

It will NOT include runtime behavioral enforcement.
That is AgentRepEngine's permanent wedge.

**The window:** Be installed before the Check Point renewal conversation.
An installed product is 10x harder to displace than a competing pitch.

---

## ONE-LINER FOR EACH COMPETITOR

| If they mention... | Say this |
|---|---|
| Lakera | "Lakera stops bad prompts in. We stop bad behavior out. Different layer." |
| Check Point | "They have prompt security. We have behavioral enforcement. We feed their SIEM." |
| Palo Alto | "They secure the model. We secure the runtime. Complementary." |
| Wiz | "They check the door is locked. We watch the agent trying to open it." |
| SIEM | "We enrich your SIEM. Every decision ships a structured reason object." |
| Kong | "We ARE a Kong plugin. 4-hour install. No new vendor." |

---

## POSITIONING STATEMENT

**For:** Security and AI Platform teams at regulated enterprises
**Who have:** AI agents running in production with no behavioral visibility
**AgentRepEngine is:** Runtime trust and enforcement infrastructure for AI agents
**That:** Scores every agent request, detects anomalous behavior,
          and blocks risky actions before data leaves — at the gateway
**Unlike:** Prompt security tools that watch inputs,
            or cloud security tools that watch infrastructure
**We:** Watch what agents actually DO — and stop it in real time

---

*Internal document — not for external distribution without review*
*Honest maturity statement: docs/enterprise/honest-maturity-statement.md*