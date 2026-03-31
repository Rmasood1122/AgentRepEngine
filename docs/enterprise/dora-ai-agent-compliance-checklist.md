# AgentRepEngine — DORA AI Agent Compliance Checklist
## For Financial Institutions Assessing AI Agent Compliance Posture
Version 1.0 | For: DORA Compliance Officer / CISO Review | Classification: Confidential

---

### Executive Summary

The Digital Operational Resilience Act (DORA) entered into force on January 17, 2025, imposing binding ICT risk management, incident reporting, and third-party oversight obligations on EU-regulated financial entities. AI agents operating within financial infrastructure are ICT third-party service components under DORA and must be governed accordingly — yet most institutions lack the behavioral telemetry and enforcement mechanisms to demonstrate compliance for autonomous software agents. This checklist maps AgentRepEngine's implemented capabilities to specific DORA articles, providing a structured path from gap identification to auditable compliance.

---

### DORA Compliance Mapping

| DORA Requirement | Article | What DORA Requires | How ARE Addresses It |
|---|---|---|---|
| **ICT security policies** | Article 9 | Financial entities shall establish ICT security policies covering risk management frameworks for all ICT services, including third-party components | ARE ships five OWASP LLM Top 10 policy packs (LLM01 prompt injection, LLM04 high-frequency abuse, LLM06 bulk PII exfiltration, LLM07 scope escalation, LLM08/09 recursive spawning) as enforceable, auditable ICT security policies applied to every AI agent request. Each policy defines graduated thresholds (warning, throttle, block) with structured reason objects documenting every enforcement decision. |
| **ICT business continuity** | Article 11 | Financial entities shall put in place ICT business continuity management ensuring continuity of critical functions during ICT disruptions | ARE's ModeController implements automatic rollback from enforce to observe mode when false-positive rate exceeds 2%, checked every 5 minutes. Infrastructure failure triggers fail-open behavior within 100ms via circuit breaker — agent traffic is never interrupted by scoring service unavailability. Full rollback to pre-ARE state completes in under 10 minutes. No restart required for mode transitions (Redis-backed, instant). |
| **ICT-related incident classification** | Article 17 | Financial entities shall classify ICT-related incidents according to criteria including data losses, duration, and criticality, with records maintained for supervisory reporting | ARE's `enforcement_decisions` table is INSERT-only at the PostgreSQL permission level (UPDATE, DELETE, TRUNCATE revoked). Every decision is hash-chained: `this_hash = sha256(prev_hash + id + timestamp + decision)`. Chain integrity is verified on every health check. This produces a cryptographically non-repudiable incident classification log that satisfies supervisory record-keeping requirements and is tamper-evident by construction. |
| **Third-party ICT risk management** | Article 28 | Financial entities shall manage ICT third-party risk as an integral component of ICT risk, including monitoring of outsourced or agent-mediated functions | ARE's behavioral scoring engine evaluates every AI agent across 8 dimensions (tool call rate, endpoint enumeration, bulk access volume, PII field access, cross-tenant probes, permission escalation, sub-agent spawn depth, token refresh rate) in real time. Each agent maintains an individual behavioral baseline. Deviations are scored, logged, and enforced per policy — providing continuous third-party ICT risk monitoring for autonomous software agents. |
| **ICT contractual arrangements** | Article 30 | Contractual arrangements with ICT third-party providers shall include service descriptions, performance targets, incident reporting, and exit strategies | ARE's Pilot Letter of Understanding includes: 30-day zero-cost visibility deployment, staged rollout gates with explicit CISO sign-off at each phase, defined success criteria (false-positive rate <2%, mean scoring latency <50ms, zero business interruption), full data residency within customer network, and 10-minute rollback exit strategy. Each gate transition is documented and requires human authorization. |
| **RTS on audit trail requirements** | Regulatory Technical Standards | DORA RTS mandates audit trails for ICT operations that are complete, tamper-resistant, and available for supervisory examination | ARE's enforcement log is cryptographically non-repudiable: SHA-256 hash chain across all enforcement decisions, INSERT-only at the database permission level, SOC2-compatible export with `chain_verified` field, and SIEM integration via structured CEF/JSON events (Splunk, Microsoft Sentinel compatible). The hash chain proves decision sequence and integrity without storing prompt content or user PII. |

---

### AI Agent DORA Readiness Self-Assessment

Answer each question yes or no. Each "no" identifies a compliance gap that DORA supervisors may examine.

| # | Question | DORA Article | ARE Capability |
|---|---|---|---|
| 1 | Do you have documented ICT security policies that specifically cover AI agent behavior, not just human user access? | Article 9 | OWASP LLM Top 10 policy packs with graduated enforcement thresholds |
| 2 | Can your AI agents continue operating if your governance or monitoring layer experiences a failure? | Article 11 | Fail-open circuit breaker (100ms), auto-rollback ModeController |
| 3 | Do you have an automated rollback mechanism that prevents your governance layer from disrupting business continuity? | Article 11 | ModeController auto-rollback at >2% false-positive rate, no restart required |
| 4 | Are all AI agent enforcement decisions (allow, audit, block) recorded in a tamper-evident log? | Article 17 | Hash-chained enforcement_decisions table, INSERT-only at DB level |
| 5 | Can you produce a complete, cryptographically verifiable audit trail of AI agent governance decisions for your supervisor on request? | RTS Audit Trail | SOC2-compatible export with chain_verified field, SIEM feed |
| 6 | Do you monitor AI agent behavioral patterns in real time across multiple risk dimensions? | Article 28 | 8-dimension behavioral scoring with per-agent individual baselines |
| 7 | Can you detect and respond to AI agent scope escalation, bulk data access, or cross-tenant probes within a single session? | Article 28 | Zero-tolerance cross-tenant policy (first probe = block), graduated bulk access thresholds |
| 8 | Does every automated enforcement action against an AI agent include a structured, human-readable explanation? | Article 9 | Structured reason objects on every decision; enforcement fails to AUDIT (not BLOCK) if explanation cannot be generated |
| 9 | Do your ICT contractual arrangements with AI agent providers include staged rollout gates, defined success criteria, and documented exit strategies? | Article 30 | Pilot LoU with phase gates, CISO sign-off, 10-minute rollback exit |
| 10 | Can you demonstrate that no AI agent governance decision has been retroactively modified or deleted? | RTS Audit Trail | PostgreSQL-level REVOKE on UPDATE/DELETE/TRUNCATE; SHA-256 hash chain verified on every health check |

**Scoring:**
- 10/10: Strong DORA posture for AI agent governance
- 7-9: Partial coverage — specific gaps require remediation before supervisory examination
- Below 7: Material gaps in AI agent DORA compliance; prioritize remediation

---

### How ARE Addresses Each Gap

**Gap 1 — No AI-agent-specific ICT security policies.**
Most institutions have ICT security policies written for human users and traditional software services. AI agents exhibit fundamentally different behavioral patterns — high-frequency tool calls, autonomous decision chains, recursive sub-agent spawning — that existing policies do not address. ARE ships five policy packs mapped to the OWASP LLM Top 10, each with graduated thresholds calibrated for AI agent behavior. These policies are version-controlled YAML files, auditable by your compliance team, and enforceable in real time.

**Gap 2 — Governance layer failure threatens business continuity.**
An enforcement layer that fails closed can become the single point of failure that DORA Article 11 is designed to prevent. ARE's architecture separates infrastructure failure (fail-open, agents continue) from enforcement failure (fail to AUDIT, not BLOCK). The circuit breaker fires within 100ms. Every unscored request is logged with timestamp for post-incident review. Your operations team receives an alert within 60 seconds.

**Gap 3 — No automated rollback from enforcement to observation.**
Manual rollback during a false-positive surge creates exactly the kind of operational disruption DORA seeks to prevent. ARE's ModeController monitors false-positive rate every 5 minutes and automatically rolls back to observe mode if the rate exceeds 2%. The mode change is instant (Redis-backed, no restart). Critically, ARE never auto-escalates to enforce mode — that decision always requires human authorization.

**Gap 4 — Enforcement decisions are mutable or incomplete.**
If enforcement records can be edited or deleted, they cannot satisfy DORA's incident classification requirements. ARE's enforcement_decisions table has UPDATE, DELETE, and TRUNCATE revoked at the PostgreSQL permission level for all roles including the application user. Every row includes prev_hash and this_hash fields forming a SHA-256 chain. This is not application-level protection — it is database-level immutability.

**Gap 5 — No cryptographically verifiable audit trail.**
DORA RTS requires audit trails that are complete, tamper-resistant, and available for supervisory examination. ARE's hash chain is verified on every health check. The SOC2-compatible export includes a chain_verified boolean. SIEM integration delivers structured CEF/JSON events to Splunk or Microsoft Sentinel. The hash chain proves decision sequence and integrity without storing prompt content, response payloads, or user PII.

**Gap 6 — No real-time behavioral monitoring for AI agents.**
Traditional APM and SIEM tools monitor infrastructure metrics, not agent behavioral patterns. ARE scores every agent request across 8 behavioral dimensions in real time, maintaining individual baselines per agent using Welford's online algorithm. Z-score velocity detection activates after 10 events, catching behavioral drift that static thresholds miss.

**Gap 7 — No session-level anomaly detection for AI agents.**
Cross-tenant probes, bulk PII extraction, and scope escalation are session-level attack patterns that per-request monitoring cannot detect. ARE's policy engine evaluates behavioral vectors per session, with zero-tolerance policies for cross-tenant probes (first probe triggers immediate block and mandatory human review) and graduated thresholds for bulk access patterns.

**Gap 8 — Enforcement actions lack structured explanations.**
An enforcement action without a documented reason is indefensible in a supervisory examination. ARE generates a structured reason object for every enforcement decision. If the reason object cannot be generated, enforcement defaults to AUDIT mode — the decision is logged for human review, but no agent is blocked without an explanation on record. This design choice prioritizes auditability over automation.

**Gap 9 — No staged rollout or exit strategy for AI governance.**
Deploying AI governance without staged validation creates the operational risk DORA Article 30 requires you to manage contractually. ARE's deployment follows a four-stage curriculum: zero-impact visibility, flagging with human review, enforcement with human review, and full enforcement. Each stage requires explicit CISO sign-off. The Pilot Letter of Understanding documents success criteria, and full rollback completes in under 10 minutes.

**Gap 10 — Cannot prove enforcement log integrity.**
Asserting that records are unmodified is not the same as proving it. ARE's hash chain is cryptographic proof: any modification to any row breaks the chain, and chain integrity is verified automatically on every health check. The verification function `verify_hash_chain('enforcement_decisions')` returns a boolean that is included in every SOC2 export. A broken chain is an auditable incident in itself.

---

### Next Step

30-day zero-risk visibility deployment. No procurement required. ARE deploys in observe-only mode within your network perimeter — no agent is blocked, no data leaves your environment, and full rollback completes in under 10 minutes. You receive a complete behavioral telemetry baseline of your AI agent fleet mapped to every DORA requirement in this checklist.

Contact: rehan@naseem-a2a.com

---

*DORA compliance mapping current as of March 2026.
This document is a technical compliance mapping, not legal advice.
Legal review recommended before regulatory submission.*
