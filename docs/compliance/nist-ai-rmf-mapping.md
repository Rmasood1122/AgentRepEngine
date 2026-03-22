# AgentRepEngine — NIST AI RMF Mapping
## Alignment with NIST AI Risk Management Framework 1.0
Version 1.0 | For: Compliance Review | Classification: Confidential

Reference: NIST AI RMF 1.0 (January 2023)
URL: https://www.nist.gov/itl/ai-risk-management-framework

---

### Executive Summary

AgentRepEngine maps directly to all four core functions of the
NIST AI RMF: GOVERN, MAP, MEASURE, and MANAGE. This document
provides the specific mapping for compliance teams, auditors,
and security architects evaluating AgentRepEngine against their
AI governance obligations.

For organizations subject to EO 14110 (Safe, Secure, and
Trustworthy AI), NIST AI RMF compliance is a procurement
requirement. This mapping demonstrates how AgentRepEngine
satisfies the behavioral monitoring and governance requirements
of the framework.

---

### GOVERN Function

*Establish and enforce AI risk management policies.*

| NIST AI RMF Subcategory | AgentRepEngine Implementation |
|------------------------|-------------------------------|
| GOVERN 1.1 — Policies for AI risk management | 5 YAML policy packs (OWASP LLM Top 10 aligned). Each policy defines thresholds, enforcement actions, and escalation paths. Config-driven, auditable. |
| GOVERN 1.2 — Roles and responsibilities | Override workflow requires authenticated human reviewer + reason code. Every enforcement decision has an attributable human in the loop for HIGH_RISK actions. |
| GOVERN 1.4 — Organizational teams are committed to AI risk management | Pilot Letter of Understanding defines pilot success criteria, override authority, and rollback procedure before deployment begins. |
| GOVERN 2.2 — Explainability of AI decisions | Every enforcement decision produces a structured reason object: policy fired, threshold breached, baseline deviation, recommended action. Nothing blocks without explanation. |
| GOVERN 4.1 — Organizational risk policies | ENFORCEMENT_MODE=observe default. No blocking without explicit human decision to enable enforce mode. Auto-rollback to observe if FP rate exceeds 2%. |
| GOVERN 6.1 — Policies for AI transparency | Honest Maturity Statement published. Current limitations documented. No overclaiming. |

---

### MAP Function

*Identify and categorize AI risks in context.*

| NIST AI RMF Subcategory | AgentRepEngine Implementation |
|------------------------|-------------------------------|
| MAP 1.1 — Context established for AI risk | Agent identity model: JWT/RS256 with org_id, agent_did, lineage_hash. Every agent is contextually identified before scoring begins. |
| MAP 1.5 — Organizational risk priorities | 5 OWASP LLM Top 10 policy packs map to specific risk categories: LLM01 (prompt injection), LLM04 (tool abuse), LLM06 (bulk PII), LLM07 (scope escalation), LLM09 (recursive spawn). |
| MAP 2.1 — Scientific findings are applied | Behavioral scoring via velocity + z-score anomaly detection. Per-agent individual baseline. Peer cluster deviation in every reason object. Exponential decay model with 7-day half-life. |
| MAP 2.3 — AI system risks are documented | Honest Maturity Statement. Known gaps documented with closure timelines. No hidden limitations. |
| MAP 3.5 — Risk to third parties | Data residency: all behavioral telemetry stays in customer Kubernetes cluster. No cross-tenant data sharing in Phase 1. Tier 1 metadata-only telemetry by default. |
| MAP 5.1 — Likelihood of AI risks | Attack corpus: 20 attack scenarios across OWASP LLM Top 10. TP rate: 86.67%. FP rate: 0.00% on 100-scenario legitimate corpus. |

---

### MEASURE Function

*Analyze and assess AI risks quantitatively.*

| NIST AI RMF Subcategory | AgentRepEngine Implementation |
|------------------------|-------------------------------|
| MEASURE 1.1 — Metrics for AI risks | Primary KPIs: FP rate (target ≤2%), TP rate (target ≥85%), gateway latency p99 (target ≤10ms), override rate (target <5%). All measured continuously. |
| MEASURE 2.1 — AI system performance | Prometheus metrics: blocked_decisions_total, fp_rate_7d, score_update_duration_ms, event_queue_depth, override_rate_30d. All visible in Grafana dashboard. |
| MEASURE 2.5 — AI system effectiveness | FP rate monitored every 5 minutes via ModeController. Auto-rollback to observe mode if FP rate exceeds 2% threshold. |
| MEASURE 2.6 — Fairness and bias | Per-agent individual baseline prevents population-level bias. Agent scored against its own historical baseline, not a global threshold. Peer cluster context is informational only — not used in enforcement decisions in Phase 1. |
| MEASURE 2.10 — Privacy risks | Tier 1 telemetry: metadata only (event_type, timestamp, endpoint, count, duration). No prompt or completion content. No PII by default. GDPR position documented. |
| MEASURE 3.3 — Effectiveness of risk treatments | Hash chain verification: SELECT verify_hash_chain('enforcement_decisions'). Tamper evidence confirmed on every health check. SOC2-compatible export available. |

---

### MANAGE Function

*Prioritize and address AI risks.*

| NIST AI RMF Subcategory | AgentRepEngine Implementation |
|------------------------|-------------------------------|
| MANAGE 1.1 — Responses to AI risks | Four enforcement bands: TRUSTED (700-1000) → ALLOW, MONITORED (500-699) → ALLOW + audit, RESTRICTED (200-499) → THROTTLE, BLOCKED (0-199) → synthetic response. |
| MANAGE 1.3 — Responses to emergent risks | HIGH_RISK operation class: bulk_pii_export, credential_access, lateral_movement, mass_deletion → VERIFY state regardless of score. Human review required before any HIGH_RISK action proceeds. |
| MANAGE 2.2 — Mechanisms for incident response | Override workflow: POST /enforcement/override with reviewer_id + reason_code. Forensics replay: GET /audit/replay. SOC2 export: GET /audit/export?format=soc2. |
| MANAGE 2.4 — Risk treatments documented | Every enforcement decision produces structured reason object stored in tamper-evident log. Decision rationale preserved indefinitely. |
| MANAGE 3.1 — Risks are prioritized | Score bands map to risk priority: BLOCKED (imminent risk) → synthetic response; RESTRICTED (elevated risk) → throttle + human review; MONITORED (watch list) → active audit. |
| MANAGE 4.1 — Residual risks are documented | Honest Maturity Statement. Current gaps: auto-rollback verified, slow-walk evasion testing in progress, SOC2 Type II observation period not yet initiated. |

---

### NIST AI RMF Profile: AI Agent Governance

AgentRepEngine can serve as a Profile contribution to the
NIST AI RMF for the "AI Agent Governance" use case.

**Profile name:** Agentic AI Runtime Enforcement Profile
**Use case:** Governing autonomous AI agents in regulated enterprise environments
**Primary functions addressed:** GOVERN, MAP, MEASURE, MANAGE (all four)

**Key profile characteristics:**
- Runtime behavioral monitoring (not static analysis)
- Cryptographic agent identity (JWT/RS256)
- Tamper-evident audit trail (SHA-256 hash chain)
- Human override authority preserved at all times
- Fail-open on infrastructure, fail-closed on enforcement
- Metadata-only telemetry by default (privacy-preserving)

---

### Regulatory Cross-References

| Regulation | Relevant Requirement | AgentRepEngine Coverage |
|------------|---------------------|------------------------|
| EO 14110 (US) | Section 4.2: AI safety testing | Eval harness: 100 FP scenarios + 20 attack scenarios |
| EU AI Act Art. 9 | Risk management system for high-risk AI | YAML policy packs + override workflow + audit log |
| EU AI Act Art. 13 | Transparency and provision of information | Reason object on every decision + Honest Maturity Statement |
| SOC2 CC7.2 | System monitoring | Prometheus metrics + Grafana + hash chain verification |
| NIST SP 800-207 | Zero Trust architecture | JWT-based identity, never implicit trust, verify every request |
| OWASP LLM Top 10 | AI application security | 5 policy packs covering LLM01, LLM04, LLM06, LLM07, LLM09 |

---

### How to Use This Document

**For compliance teams:**
Use this mapping to demonstrate AI governance controls in
internal audits and regulatory submissions. AgentRepEngine
provides documented, measurable controls for all four NIST
AI RMF functions.

**For security architects:**
Each subcategory mapping links to a specific, verifiable
implementation. Commands for verification are available
in the FAANG Enterprise Audit (docs/enterprise/faang-enterprise-audit-72q.md).

**For procurement teams:**
This mapping satisfies AI governance requirements for
organizations following NIST AI RMF as a procurement standard
under EO 14110 and related federal guidance.

---

### Verification Commands
```bash
# GOVERN: Verify policy packs exist and are OWASP-aligned
ls ~/AgentRepEngine/config/policy_packs/

# MAP: Verify agent identity model
go test ./internal/identity/... -v -run TestRS256Only

# MEASURE: Verify FP rate
docker exec agentrepengine-postgres-1 psql -U are -d agentrepengine \
  -c "SELECT fp_rate FROM daily_fp_metrics ORDER BY date DESC LIMIT 1;"

# MANAGE: Verify hash chain integrity
curl -s http://localhost:8080/health | jq '.hash_chain_valid'
# Must return: true

# MANAGE: Verify auto-rollback mechanism
go test ./internal/enforcement/... -run TestAutoRollback -v
```

---

*NIST AI RMF mapping current as of March 2026.
Updated when framework versions change or new controls are added.
Legal review recommended before regulatory submissions.*