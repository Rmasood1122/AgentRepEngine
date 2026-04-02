# AgentRepEngine — Competitive Positioning
# One-pager for enterprise conversations
# Last updated: March 31, 2026

---

## THE ONE SENTENCE

AgentRepEngine is the behavioral grounding layer for AI agents.
RAG solved hallucination for LLM outputs by anchoring responses
to verified data. ARE solves the same problem one layer deeper —
anchoring agent actions to org-specific behavioral baselines,
enforcing at the gateway, stopping anomalous behavior before
data leaves your environment.
---

## COMPETITIVE LANDSCAPE

### Lakera (now Check Point)
**What they do:** Prompt injection detection — scans LLM input/output
for malicious content before it reaches the model.

**What they do NOT do:**
- Runtime behavioral scoring across sessions
- Gateway-level behavioral enforcement with automatic human review escalation
- Cryptographically non-repudiable enforcement log of agent decisions
- Statistical behavioral deviation detection over time

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
- Per-request statistical behavioral deviation detection
- Agent identity model (JWT-based trust)

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
AgentRepEngine tells you the agent is trying to open it anyway —
and stops it before data leaves.

---

### SIEM / Splunk / QRadar
**What they do:** Log aggregation and alerting after events occur.

**What they do NOT do:**
- Stop actions in real time before data leaves
- Score behavioral patterns across sessions
- Produce structured reason objects per enforcement decision

**The relationship:** AgentRepEngine FEEDS your SIEM.
Every enforcement decision generates a structured CEF/JSON event
delivered to your SIEM within 500ms. Compatible with Splunk and
Microsoft Sentinel out of the box.
AgentRepEngine is not a SIEM replacement — it is the AI agent
behavioral intelligence layer your SIEM has never had.

---

### Traditional API Gateways (Kong, Apigee, AWS API GW)
**What they do:** Rate limiting, authentication, routing.

**What they do NOT do:**
- Behavioral scoring per agent identity across sessions
- Cross-session adversarial baseline poisoning defense
- Cryptographically non-repudiable behavioral enforcement log

**The relationship:** AgentRepEngine IS a Kong plugin.
It runs inside your existing gateway.
No new infrastructure. No new vendor procurement.
Installation time: 4 hours.

---

## THE MATRIX

| Capability | AgentRepEngine | Lakera | Palo Alto | SIEM |
|---|---|---|---|---|
| Behavioral attention engine (runtime scoring) | ✅ | ❌ | ❌ | ❌ |
| Gateway-level behavioral enforcement | ✅ | ❌ | ❌ | ❌ |
| Per-request blocking with human review escalation | ✅ | ❌ | ❌ | ❌ |
| Cryptographically non-repudiable enforcement log | ✅ | ❌ | ❌ | Partial |
| Structured SIEM feed (CEF/JSON reason objects) | ✅ | ❌ | ❌ | N/A |
| Agent identity model (JWT/RS256) | ✅ | ❌ | ❌ | ❌ |
| Adversarial baseline poisoning defense | ✅ | ❌ | ❌ | ❌ |
| DORA Article 17/28 audit trail | ✅ | ❌ | ❌ | Partial |
| Installation time | 4 hours | Days | Weeks | Weeks |
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
> AgentRepEngine does not replace your SIEM — it makes it work for
> AI agents for the first time. Every enforcement decision delivers
> a structured reason object to your SIEM within 500ms. Your SIEM
> currently receives logs. AgentRepEngine gives it decisions.

**Q: Do we need this if we already have Palo Alto?**
> Palo Alto secured your model before deployment.
> AgentRepEngine secures your agent after deployment.
> The attack surface is different. A clean model can still
> be exploited by a compromised agent runtime.

---

## THE NOVEL THREAT: ADVERSARIAL BASELINE POISONING

One attack class that no existing product defends against:

An attacker who controls an agent gradually escalates its activity
rate over 7-14 days. The behavioral scoring baseline shifts with it.
After 14 days, the agent's anomalous rate looks statistically normal.
The attacker strikes. The SIEM sees nothing. Lakera sees nothing.

AgentRepEngine defends against this with two independent layers:
1. OWASP LLM Top 10 enforcement policies catch high-value actions
   regardless of score — a "trusted" agent cannot silently exfiltrate
2. Variance growth rate monitoring flags the baseline shift as it
   begins — early warning before the attack succeeds

100% detection rate on 10 synthetic single-agent slow-walk scenarios.
Multi-agent coordinated evasion detection is a Phase 2 capability.
No other AI agent security product addresses this attack class.

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
| SIEM | "We enrich your SIEM with AI agent behavioral intelligence. Every decision ships a structured reason object within 500ms." |
| Kong | "We ARE a Kong plugin. 4-hour installation. No new vendor." |

---

## POSITIONING STATEMENT

**For:** Security and AI Platform teams at regulated enterprises
**Who have:** AI agents running in production with no behavioral visibility
**AgentRepEngine is:** Runtime trust and behavioral enforcement
                       infrastructure for AI agents
**That:** Scores every agent request using a behavioral attention engine,
          detects statistical behavioral deviation and adversarial
          baseline poisoning, and stops risky actions before data
          leaves — at the gateway, with cryptographic proof
**Unlike:** Prompt security tools that watch inputs,
            or cloud security tools that watch infrastructure
**We:** Watch what agents actually DO — and stop it in real time,
        with a reason object your DORA examiner can read

---

*Internal document — not for external distribution without review*
*Honest maturity statement: docs/enterprise/honest-maturity-statement.md*
