# AgentRepEngine — Product Maturity Statement
## What We Have, What We Don't, and What's Coming
Version 1.1 | March 2026

This document exists because security vendors who oversell maturity
create surprises. We would rather you know exactly what you're
getting before we install anything.

---

### Current Product Maturity Score

Independent adversarial audit (March 2026) scored AgentRepEngine
at 67/100 against a FAANG-grade enterprise rubric.

Week 1 hardening target: 80/100 (Enterprise-Ready)
Week 3 hardening target: 88/100 (FAANG-Grade)

Gaps being closed in parallel with pilot conversations:
- Auto-rollback on FP spike: in progress
- Slow-walk evasion test corpus: in progress
- SIEM webhook confirmation: this week
- GDPR tombstone procedure: documented in gdpr-position.md

We do not claim enterprise-ready until the rubric confirms it.

---

### What Is Production-Ready Today

✅ **Runtime enforcement engine**
Gateway-level behavioral scoring and enforcement via Kong plugin.
Tested. Gate-verified. 0.00% false positive rate on our
100-scenario internal validation corpus. External validation
on production traffic available during pilot — expected < 1%
on well-configured environments. Latency under 10ms p99.

✅ **Agent identity model**
JWT/RS256 signed identity with org-managed JWKS. Sub-agent
lineage tracking. Identity cycling detection. Probation mode
for new agents. No DID or ledger dependency — designed for
4-hour enterprise install.

✅ **JWT replay detection**
Implemented. jti claim on every token. Redis used-token cache
with TTL matching token expiry. Replay rejected on second use.
Fail-open if Redis unavailable — infrastructure failure never
blocks agents.

✅ **Redis ACL hardening**
Authentication required on all connections. Default user
disabled. Scoring service restricted to score and replay
detection key namespaces only. Rate limiting at 500 RPS per
org enforced at gateway. Unauthenticated access rejected.

✅ **Explainability on every decision**
Every enforcement decision produces a structured reason object:
what triggered, what threshold, what baseline deviation,
what policy fired. Nothing blocks without explanation on record.

✅ **Tamper-evident audit log**
SHA-256 hash chain on enforcement_decisions. INSERT-only at DB
permission level. Verifiable: SELECT verify_hash_chain() = true.
SOC2-compatible export endpoint available.

✅ **5 OWASP-aligned policy packs**
Pre-built enforcement rules covering prompt injection (LLM01),
bulk PII access (LLM06), scope escalation (LLM07),
high-frequency tool abuse (LLM04), and recursive agent
spawning (LLM09). All configurable via YAML — no code changes
required for threshold adjustments.

✅ **Observe mode + override workflow**
48-hour observe window standard. Human auth + reason code
required for every override. Full audit trail on all override
decisions.

✅ **Two failure modes, both correct**
Infrastructure failure → fail-open (agents keep working).
Enforcement failure → fail-closed (AUDIT not BLOCK).
Both behaviors are documented, tested, and configurable.

---

### What Is Not Yet Production-Ready

⚠️ **Auto-rollback on FP spike**
Specified in architecture. Implementation verification in
progress. Target: TestAutoRollback passing before pilot
enforce-mode go-live. Mitigation: 48-hour observe window
prevents enforce mode until manually enabled.

⚠️ **Slow-walk evasion test corpus**
Multi-day distributed attack patterns not yet in test suite.
HIGH_RISK override is primary mitigation — blocks critical
actions regardless of historical score. Full corpus: Week 2.

⚠️ **Management UI**
Phase 1 monitoring is through Grafana dashboards. A dedicated
management console is Phase 2, built based on what the first
enterprise actually needs. Security engineers are comfortable
with Grafana. We are honest that it is not a polished product UI.

⚠️ **SOC2 Type II certification**
We have the architecture for SOC2 evidence generation. SOC2
Type II observation period not yet initiated. We produce the
evidence. You bring your auditor to verify it. SOC2 readiness
engagement: Q2 2026.

⚠️ **Third-party penetration test**
Not yet engaged. On roadmap for Q2 2026. We will share
results with design partners.

⚠️ **Named SIEM connectors**
We have a webhook that delivers the reason object to any SIEM.
We do not have pre-built Splunk apps or QRadar plugins.
We will write the Splunk HEC configuration with you during
the pilot.

---

### Why We're Telling You This

A vendor who hides these gaps will create a surprise in your
environment at the worst possible moment.

A vendor who lists them upfront is one you can plan around.

We are a pre-revenue product with a working enforcement engine,
a formal research foundation, and a clear maturity roadmap.
The gaps above are known, documented, and have closure timelines.

If any of these gaps are blockers for your organization,
tell us now. We would rather know before the pilot than after.

---

*Maturity statement current as of March 2026.
Updated quarterly or when material changes occur.*