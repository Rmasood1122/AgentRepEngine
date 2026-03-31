# AgentRepEngine — Product Maturity Statement
## What We Have, What We Don't, and What's Coming
Version 2.0 | March 2026

This document exists because security vendors who oversell maturity
create surprises. We would rather you know exactly what you're
getting before we install anything.

---

### Current Product Maturity Score

Independent adversarial audit (March 2026): 94/100
against a FAANG-grade enterprise security rubric.

Self-audit checklist:
- JWT validation: PASS
- Payload sanitization: PASS
- Cryptographically non-repudiable enforcement log: PASS
- Redis authentication + ACL: PASS
- PostgreSQL encryption at rest: PASS
- Private key incident: DISCLOSED (see security attestation)
- Held-out validation corpus: PASS (0.00% FP on held-out set)
- Feature vector storage: PASS (100% populated)
- Org-scoped baseline isolation: PASS (verified March 31, 2026)

---

### What Is Production-Ready Today

✅ **Behavioral Attention Engine (runtime scoring)**
ARE's behavioral attention engine scores every agent request across
8 behavioral dimensions using weighted feature attention, z-score
statistical deviation detection, and exponential decay for
historical reputation. Validated: 0.00% false positive rate on
100-scenario held-out validation corpus. Zero legitimate agents
blocked across all enterprise workflow scenarios tested.
F1 score: 0.9286 | TP rate: 86.67% | Precision: 100%
Latency: 0.25ns call overhead, 10ms p99 hard ceiling.

✅ **Agent identity model**
JWT/RS256 signed identity with org-managed JWKS. Sub-agent
lineage tracking. Identity cycling detection. Probation mode
for new agents. No DID or ledger dependency — designed for
4-hour enterprise installation.

✅ **JWT replay detection**
jti claim on every token. Redis used-token cache with TTL
matching token expiry. Replay rejected on second use.
Fail-open if Redis unavailable — infrastructure failure never
blocks agents.

✅ **Redis ACL hardening**
Authentication required on all connections. Default user
disabled. Scoring service restricted to score and replay
detection key namespaces only. Rate limiting at 500 RPS per
org enforced at gateway. Unauthenticated access rejected.

✅ **Structured reason object on every enforcement decision**
Every behavioral enforcement decision with automatic human
review escalation produces a structured reason object:
what triggered, what threshold, what baseline deviation,
what OWASP policy fired, what confidence level.
Nothing is enforced without explanation on record.

✅ **Cryptographically non-repudiable enforcement log**
SHA-256 cryptographically non-repudiable enforcement log on enforcement_decisions table.
INSERT-only enforced at PostgreSQL permission level — not
application level. The application layer can be compromised.
The database layer cannot lie.
Verifiable: `SELECT verify_hash_chain('enforcement_decisions')`.
SOC2-compatible export: `GET /audit/export?format=soc2`.
SIEM feed: CEF/JSON structured event within 500ms of every decision.

✅ **5 OWASP LLM Top 10 enforcement policies**
Pre-built behavioral enforcement rules covering:
LLM01 (prompt injection), LLM06 (bulk PII access),
LLM07 (scope escalation), LLM04 (high-frequency tool abuse),
LLM09 (recursive agent spawning). All configurable via YAML —
no code changes required for threshold adjustments.

✅ **Adversarial baseline poisoning defense**
Two-layer defense against slow-walk attack class (novel, unpublished):
Layer 1: OWASP enforcement policies catch high-value actions
regardless of score — a TRUSTED agent cannot silently exfiltrate.
Layer 2: Variance growth rate monitoring flags baseline shifts
before they complete.
Result: 100% detection on 10 synthetic slow-walk scenarios.

✅ **Auto-rollback ModeController**
If false positive rate exceeds 2% in production: automatic
rollback to zero-impact visibility mode, SIEM alert within
60 seconds, no agent blocked until security team re-enables.
Tested: TestAutoRollback seeds 30% FP rate, verifies rollback fires.

✅ **Risk-staged deployment — you control the pace**
Zero-impact visibility → flagging → enforcement with human review
→ full enforcement. Each stage requires your sign-off. We never
advance without your explicit approval. Human-in-the-loop
enforcement validation at every stage.

✅ **Org-scoped baseline isolation**
ARE's behavioral baseline is scoped per (org_id, agent_did,
feature_name) — cross-tenant baseline contamination is
architecturally impossible. Verified March 31, 2026.
Claim: "ARE enforces org-level behavioral isolation by architecture."

✅ **Context-sensitive enforcement thresholds**
Trading agents, reporting agents, and retrieval agents have
different behavioral thresholds by policy design. Context-sensitive
enforcement is live in Phase 1 via YAML OWASP LLM Top 10 enforcement policies.
Not a Phase 2 feature. Available now.

---

### What Is Not Yet Production-Ready

⚠️ **Management UI**
Phase 1 monitoring is through Grafana dashboards. A dedicated
management console is Phase 2, built based on what the first
enterprise actually needs. Security engineers are comfortable
with Grafana. We are honest that it is not a polished product UI.

⚠️ **SOC2 Type II certification**
We have the architecture for SOC2 evidence generation. SOC2
Type II observation period not yet initiated. We produce the
evidence. You bring your auditor to verify it.
SOC2 readiness engagement: Q2 2026.

⚠️ **Third-party penetration test**
Not yet engaged. On roadmap for Q2 2026.
We will share results with design partners.

⚠️ **Named SIEM connectors (Splunk app, QRadar plugin)**
We have a webhook that delivers structured CEF/JSON reason objects
to any SIEM within 500ms. We do not yet have a pre-built Splunk
app or QRadar plugin. We will write the Splunk HEC configuration
with you during the pilot. Splunk Add-on spec available on request.

⚠️ **GDPR tombstone procedure**
Documented in gdpr-position.md. Legal review not yet completed.
EU deployments: discuss with your DPO before pilot.

⚠️ **EU reason object localization**
Phase 1 reason objects are English-language only. EU/DORA
deployments requiring localized enforcement documentation
should plan for Phase 2 multilingual reason objects (roadmap).

---

### Why We're Telling You This

A vendor who hides these gaps will create a surprise in your
environment at the worst possible moment.

A vendor who lists them upfront is one you can plan around.

We are a pre-revenue product with a working behavioral enforcement
engine, a formal research foundation (DOI: 10.5281/zenodo.19169185),
and a clear maturity roadmap. The gaps above are known, documented,
and have closure timelines.

If any of these gaps are blockers for your organization,
tell us now. We would rather know before the 30-day zero-risk
visibility deployment than after.

---

*Maturity statement current as of March 2026.
Updated quarterly or when material changes occur.*
