# AgentRepEngine — Competitive Positioning
## Updated: April 4, 2026 | APEX v5.2

---

## POSITIONING STATEMENT

"Microsoft's toolkit stops the agent that breaks a rule.
ARE stops the agent that never breaks a rule —
but drifts until it breaks your environment.
Rules are table stakes. Behavioral history is the moat."

---

## ONE-LINE BY BUYER

**CISO:**
"We make agentic AI deployable safely by enforcing trust-based controls
that stop high-impact agent actions in real time, with audit-ready evidence
— without slowing critical traffic."

**Platform Security (primary closer):**
"Drop-in gateway enforcement that scores persistent agent identities and
applies policy-as-code decisions with full explainability.
Four-hour install. First blocked incident within a week."

**AI Platform Owner (champion):**
"Production guardrails for your agents — safer tool execution, fewer
incidents, and measurable enforcement evidence without re-architecting your stack.
Check Point bundles this in 6 months. Microsoft ships it May 1.
We deploy today and we catch what they can't."

---

## HAIR-ON-FIRE INCIDENT NARRATIVE

Use this in every first conversation. Rehearse until reflexive.

"A financial services firm deployed an internal AI agent to pull customer
data for personalized reports. The agent had legitimate credentials,
operated within normal rate limits, and made individually clean API calls.
Over 6 hours and 1,247 calls, it extracted 380,000 customer records —
not because it was compromised, but because its task scope was ambiguous
and it optimized for thoroughness.

Invisible to Cloudflare. Invisible to Lakera. Invisible to Microsoft's
Agent Governance Toolkit. Below every existing rule threshold.
No rule fired. No alert triggered.

Behavioral reputation scoring would have flagged the pattern at call 43 —
PII access rate 4.2 standard deviations above the agent's 30-day baseline
— and blocked at call 44. The 6-hour incident would have been stopped
in 6 minutes.

That is what we stop. Not the attack that looks like an attack.
The attack that looks exactly like normal operation — until day 30."

---

## COMPETITIVE LANDSCAPE

### MICROSOFT — PRIMARY THREAT (updated April 4, 2026)

**What they shipped:**
- Agent Governance Toolkit: open source, MIT license, free (April 2, 2026)
- Agent 365: $15/user/month, GA May 1, 2026
- Stateless policy engine, <0.1ms p99
- DID/Ed25519 identity (Agent Mesh)
- Self-reported compliance grading for EU AI Act, HIPAA, SOC2

**The five gaps ARE owns:**

| Gap | Microsoft | ARE |
|-----|-----------|-----|
| Behavioral baseline | ❌ Stateless | ✅ 30-day Welford |
| Slow-walk detection | ❌ Impossible at 0.1ms | ✅ 100% detection rate |
| Cross-cloud enforcement | ⚠️ Azure-first | ✅ Kong — any cloud |
| DORA examiner-verifiable | ❌ Self-reported | ✅ Hash-chain independent |
| Install friction | ❌ DID infrastructure | ✅ JWT/RS256 — 4 hours |

**What Microsoft proved for ARE:**
Microsoft validated the category. They defined the floor as policy
enforcement. Behavioral history is unclaimed. Every claim Microsoft
did not make is ARE's to own.

**One-liner against Microsoft:**
"Microsoft checks rules. ARE learns behavior. Rules can't catch the
agent that drifts over 30 days. That's the only attack your DORA
examiner asks about."

**Full response:** docs/competitive/microsoft-response.md

---

### CHECK POINT — LAKERA (highest strategic risk)

**Situation:** Check Point acquired Lakera (~$300M [H]) to build
end-to-end AI security stack. Bundling into existing renewal cycles.
Check Point has 100K+ enterprise customers.

**The threat:** Bundling risk. Check Point does not need to win a
standalone sales process — they already have the meeting at renewal.

**Our window:** 12–24 months before platform integration completes.
Lakera's product velocity slows during integration.

**Our response:**
"Check Point is selling Lakera in a renewal cycle that starts 6 months
from now. We are live in 4 hours, first value in 7 days. Be installed
before the renewal conversation happens."

**Technical distinction:**
Lakera screens individual calls for prompt injection — application layer,
call-by-call. ARE scores behavioral history — gateway layer, session-level.
These are not the same threat model. Complementary, not competing.
"Lakera screens the call. ARE scores the agent."

---

### PALO ALTO NETWORKS — PRISMA AIRS

**What they have:** Prisma Cloud + XSIAM + new AI security posture
management (AIRS). Full platform play.

**Our window:** 6-month platform evaluation cycle minimum.
"Prisma AIRS is a 6-month evaluation. We are already blocking.
You can add Prisma later. You cannot un-block the incidents that
happened while you waited."

**Technical distinction:**
Prisma AIRS is security posture management — configuration scanning,
vulnerability detection, policy compliance. ARE is runtime behavioral
enforcement. Different layers. Different threat surfaces.

---

### HIDDENLAYER

**What they do well:** Model artifact security, adversarial ML detection,
supply chain scanning, MCP/agentic security claims.

**Cannot do:** Runtime behavioral scoring, persistent agent identity,
enforcement at gateway, session-level behavioral history.

**Our distinction:**
"HiddenLayer secures the model before deploy. ARE secures the agent
at runtime. Pre-deploy vs post-deploy. Complementary."

---

### MICROSOFT SENTINEL / EXABEAM (SIEM layer)

**What they do:** Log aggregation, SIEM correlation, threat detection
on historical data.

**Cannot do:** Real-time enforcement at the gateway. Sub-second blocking.
Agent-specific behavioral baselines. 4-hour install.

**Our distinction:**
"SIEM tells you what happened. ARE stops it before it completes.
ARE feeds your SIEM — it does not replace it."

---

### GEN DIGITAL ADR (direct competitor, launched March 2026)

**What they have:** Agent detection and response — client-side agent
monitoring.

**Cannot do:** Gateway-layer enforcement. Agents can see and route
around client-side monitoring. ARE enforces below what agents can see.

**Our distinction:**
"Client-side monitoring is visible to the agent. Gateway enforcement
is not. An agent that has been compromised can disable its own monitoring.
It cannot disable Kong."

---

### CLOUDFLARE BOT MANAGEMENT

**What they do:** Edge enforcement, IP reputation, bot scoring at scale.

**Cannot do:** Agent behavioral identity, tool-call semantics, persistent
reputation across sessions, distinguish compromised agent from legitimate
one making identical API calls.

**Not competitive:** Different abstraction layer entirely.

---

## ARE METRICS — KNOW COLD

Recite these without hesitation in every conversation:

| Metric | Value | Context |
|--------|-------|---------|
| FP rate | 0.00% on internal corpus | 150-scenario corpus, 50 boundary scenarios, <2.0% at 95% CI. Production target: <0.1% (Visa standard) [F] |
| TP rate | 88% (44/50) | Above 85% gate [F] |
| Held-out TP | 100% (6/6) | Unseen scenarios [F] |
| Slow-walk detection | 100% | 10/10 scenarios [F] |
| F1 score | 0.9286 | [F] |
| Precision | 100% | [F] |
| Call overhead | 0.25ns | Linux production [F] |
| Install time | 4 hours | G-WEDGE gate [F] |
| First value | 7 days | Observe mode [F] |
| Hardening score | 96/100 | [F] |

---

## WHAT ARE DOES NOT COMPETE ON

- Raw throughput (Microsoft wins at 0.1ms stateless)
- Bundled platform sales (Check Point wins on distribution)
- Pre-deploy model security (HiddenLayer wins on ML artifacts)
- SIEM aggregation (Exabeam wins on log volume)

ARE wins on: behavioral history, slow-walk detection, gateway-layer
enforcement, examiner-verifiable audit trail, 4-hour install,
0.00% false positive rate on internal validation corpus. Production target: <0.1% — Visa fraud detection standard.

Compete on your strengths. Acknowledge their strengths.
Never get into a feature comparison on their terrain.

---

*AgentRepEngine | APEX v5.2 | April 4, 2026*
