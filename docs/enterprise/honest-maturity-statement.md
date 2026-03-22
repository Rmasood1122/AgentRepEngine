# AgentRepEngine — Product Maturity Statement
## What We Have, What We Don't, and What's Coming

This document exists because security vendors who oversell maturity
create surprises. We would rather you know exactly what you're
getting before we install anything.

---

### What Is Production-Ready Today

✅ **Runtime enforcement engine**
Gateway-level behavioral scoring and enforcement via Kong plugin.
Tested. Gate-verified. 0.00% false positive rate on 100 enterprise
agent scenarios. Latency under 10ms p99.

✅ **Agent identity model**
JWT/RS256 signed identity with org-managed JWKS. Sub-agent lineage
tracking. Identity cycling detection. Probation mode for new agents.
No DID or ledger dependency — designed for 4-hour enterprise install.

✅ **JWT replay detection**
Implemented. jti claim on every token. Redis used-token cache with
TTL matching token expiry. Replay rejected on second use. Fail-open
if Redis unavailable — infrastructure failure never blocks agents.

✅ **Redis ACL hardening**
Authentication required on all connections. Default user disabled.
Scoring service restricted to score and replay detection key
namespaces only. Rate limiting at 500 RPS per org enforced at
gateway. Unauthenticated access rejected.

✅ **Explainability on every decision**
Every enforcement decision produces a structured reason object:
what triggered, what threshold, what baseline deviation,
what policy fired. Nothing blocks without explanation on record.

✅ **Tamper-evident audit log**
SHA-256 hash chain on enforcement_decisions. INSERT-only at DB
permission level. Verifiable: SELECT verify_hash_chain() = true.
SOC2-compatible export endpoint available.

✅ **5 OWASP-aligned policy packs**
Pre-built enforcement rules covering prompt injection, bulk PII
access, scope escalation, high-frequency tool abuse, and recursive
agent spawning. All configurable via YAML — no code changes required
for threshold adjustments.

✅ **Observe mode + override workflow**
48-hour observe window standard. Human auth + reason code required
for every override. Full audit trail on all override decisions.

---

### What Is Not Yet Production-Ready

⚠️ **Management UI**
Phase 1 monitoring is through Grafana dashboards. A dedicated
management console is Phase 2, built based on what the first
enterprise actually needs. Security engineers are comfortable
with Grafana. We are honest that it is not a polished product UI.

⚠️ **SOC2 Type II certification**
We have the architecture for SOC2 evidence generation. We do not
have a SOC2 auditor's report. This is a pre-revenue company.
We produce the evidence. You bring your auditor to verify it.

⚠️ **Penetration test of AgentRepEngine itself**
We have not yet engaged a third-party pen tester on our own
infrastructure. This is on the roadmap for Q2 2026. We will share
results with design partners.

⚠️ **Named SIEM connectors**
We have a webhook that delivers the reason object to any SIEM.
We do not have pre-built Splunk apps or QRadar plugins.
We will write the Splunk HEC configuration with you during the pilot.

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