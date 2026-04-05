# ARE vs Gen Digital ADR — Competitive Response
# Built: April 5, 2026 | APEX v5.2
# Use: Any investor, buyer, or conference conversation where Gen Digital ADR is raised.
# Status: Gen Digital ADR launched March 2026. ~1,000+ installs. Developer-tool market.

---

## THE ONE-LINE ANSWER

> "Gen Digital ADR protects the developer's workstation. ARE protects the enterprise's
> production environment. They solve different problems for different buyers."

---

## WHAT GEN DIGITAL ADR ACTUALLY IS

Gen Digital ADR (Agentless Detection and Response) launched March 2026.
It is a client-side tool targeting developers using AI coding assistants
— Claude Code, Cursor, OpenClaw. It intercepts agent actions at the
developer's machine, applying 200+ detection rules for risky behaviors.

Key characteristics:
- **Layer**: Client-side (developer workstation)
- **Buyer**: Individual developer or dev team lead
- **Market**: Developer tools / productivity security
- **Install**: Per-developer installation
- **Scope**: One developer's AI agent session at a time
- **Regulatory fit**: Limited — no audit trail, no compliance mapping
- **Price point**: Consumer/SMB ($0–$50/month per developer)

---

## WHERE ARE SITS vs WHERE GEN DIGITAL ADR SITS

```
ENTERPRISE PRODUCTION ENVIRONMENT
─────────────────────────────────────────────────────────
  AI Agent ──→ Kong Gateway ──→ Production Systems
                    ↑
              ARE lives here
         (gateway layer, fleet-wide,
          regulated enterprise buyer,
          DORA/SEC/HIPAA audit trail)

DEVELOPER WORKSTATION
─────────────────────────────────────────────────────────
  Developer ──→ AI Coding Assistant ──→ Code/Files
                        ↑
            Gen Digital ADR lives here
         (client-side, per-developer,
          developer-tool buyer,
          no compliance mapping)
```

These are not competing products. They are different layers for different buyers.

---

## THE FIVE PRECISE DIFFERENCES

### 1. ENFORCEMENT LAYER
- **Gen Digital ADR**: Client-side. Runs on the developer's machine.
  An enterprise cannot centrally deploy, configure, or audit it.
  Agents can route around it by operating on a different machine.

- **ARE**: Gateway layer. Sits in the network path that every agent
  must traverse. No agent — regardless of framework, language, or
  deployment location — can bypass it. Not optional at the agent level.

**One-liner**: "ADR is a lock on a specific door. ARE is a lock on the
only road in and out of the building."

### 2. SCOPE: INDIVIDUAL vs FLEET
- **Gen Digital ADR**: One developer's session. No cross-agent visibility.
  Cannot detect coordinated behavior across multiple agents.

- **ARE**: Every agent in the enterprise simultaneously. Behavioral
  baselines are fleet-wide. Coordinated attacks — multiple agents
  deviating simultaneously — are detectable only with fleet visibility.

**One-liner**: "ADR sees one agent. ARE sees all of them at once."

### 3. BUYER: DEVELOPER vs CISO
- **Gen Digital ADR**: Developer installs it themselves. The CISO has
  no visibility into whether it's deployed, configured, or working.
  It doesn't appear in the CISO's security tooling inventory.

- **ARE**: CISO procurement. Central deployment. Appears in the security
  control register. Every enforcement decision is in the CISO's audit
  trail. The CISO controls enforcement mode, threshold configuration,
  and FP review — not individual developers.

**One-liner**: "ADR is a developer productivity tool. ARE is a CISO
security control."

### 4. REGULATORY COMPLIANCE
- **Gen Digital ADR**: No documented mapping to DORA, SEC Reg AI,
  HIPAA, or SOC2. No hash-chained audit trail. No tamper-evident
  enforcement log. A DORA auditor cannot rely on ADR as evidence.

- **ARE**: Explicit DORA Article 17/28/30 mapping. Cryptographically
  non-repudiable enforcement log. INSERT-only audit trail enforced at
  the PostgreSQL permission level. Customer-runnable chain verification.
  Built for regulated enterprise audit requirements from day one.

**One-liner**: "ADR has no compliance story. ARE is the compliance story."

### 5. BEHAVIORAL BASELINE
- **Gen Digital ADR**: Rule-based detection (200+ rules). Static.
  Rules do not adapt to the specific enterprise's agent behavior.
  An agent that violates no rules is invisible — even if its behavior
  is anomalous relative to its own history.

- **ARE**: Each agent scored against its own 30-day behavioral baseline.
  Z-score anomaly detection catches drift that violates no static rule.
  The slow-walk attack — an agent that looks clean on every individual
  call but systematically exfiltrates data over hundreds of sessions —
  is invisible to rule-based detection. ARE detects it.

**One-liner**: "ADR catches known-bad behavior. ARE catches unknown-anomalous
behavior — including attacks that have never been seen before."

---

## WHEN A PROSPECT RAISES GEN DIGITAL ADR IN A MEETING

**If they say "we're already using Gen Digital ADR":**

> "That's a developer productivity control — it protects your engineers'
> workstations while they're coding with AI assistants. ARE is a different
> layer entirely. It sits at the gateway and enforces behavioral boundaries
> on your production AI agents — the ones running 24/7 in your financial
> systems, not the ones helping your developers write code. Most enterprises
> that run ADR also need ARE, because they're solving different parts of
> the same problem."

**If they say "why wouldn't we just use Gen Digital ADR instead of ARE":**

> "Gen Digital ADR doesn't run in production. It runs on developer
> workstations. Your production AI agents — the LangChain workflows,
> the LlamaIndex pipelines running against your customer data — those
> never touch a developer's machine. ADR never sees them. ARE is the
> only tool in this layer."

**If an investor asks "how do you respond to Gen Digital ADR":**

> "Different market entirely. Gen Digital is selling to developers at
> $50/month per seat. We're selling to CISOs at $50K–$150K ACV per
> enterprise. Different buyer, different layer, different regulatory
> requirement. The presence of Gen Digital validates that the market
> recognizes AI agent security as a real problem. It doesn't compete
> with us — it sells the category to developers who then ask their
> CISO for a production-grade solution."

---

## WHAT GEN DIGITAL ADR DOES BETTER (honest assessment)

- **Distribution**: ~1,000+ installs in weeks. Developer-to-developer
  viral distribution is faster than enterprise sales cycles.
- **Awareness**: Developer-level awareness of AI agent risks drives
  CISO conversations. ADR's growth helps ARE's category awareness.
- **Rule coverage**: 200+ detection rules for known-bad patterns in
  developer workflows. Are doesn't target developer workflows.

ARE does not need to win the developer-tool market.
ARE needs to win the regulated enterprise production market.
These are not the same market.

---

## COMPETITIVE POSITIONING SUMMARY

| Dimension | Gen Digital ADR | AgentRepEngine |
|---|---|---|
| Layer | Client-side | Gateway (network) |
| Buyer | Developer | CISO |
| Scope | One developer session | Full agent fleet |
| Deployment | Per-developer install | Central enterprise deploy |
| Detection | 200+ static rules | Behavioral baseline + z-score |
| Slow-walk detection | No | Yes (100% on test corpus) |
| Compliance mapping | None documented | DORA/SEC/HIPAA/SOC2 |
| Audit trail | None | Hash-chained, tamper-evident |
| ACV | ~$0–$600/yr per dev | $50K–$150K per enterprise |
| Market | Developer tools | Regulated enterprise security |

---

## BOTTOM LINE FOR ANY CONVERSATION

Gen Digital ADR is not a threat to ARE's market.
It is a market educator that makes ARE's CISO conversations easier.

Every developer who installs ADR and starts thinking about AI agent
security is a future conversation that leads to a CISO who needs ARE.

---

## WHAT TO MONITOR

Gen Digital opened a GitHub PR on March 17 titled "session-level
behavioral baselines for anomaly detection" — not yet shipped.
If merged, GAP 5 (static rules vs behavioral baseline) narrows for
single-session detection only. Does not address gateway deployment,
audit trail, fleet scope, or compliance mapping.

Independent security analysis (Help Net Security, March 2026) noted:
"Independent performance benchmarks, latency overhead from real-time
interception, and any analysis of potential bypass techniques are
absent from the available materials." No FP rate has been published.

Check github.com/gendigitalinc/sage/pulls monthly for enterprise
deployment story. That is the only development that changes the
competitive picture materially.

The primary competitive clock remains Check Point–Lakera ($300M,
March 2026). Integration timeline 12–24 months. ARE's answer:
"Check Point bundles. We deploy in days."

---

Built: April 5, 2026 | Last updated: April 5, 2026 | APEX v5.2
File: docs/competitive/gen-digital-adr-response.md
