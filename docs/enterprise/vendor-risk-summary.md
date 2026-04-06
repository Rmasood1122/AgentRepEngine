# AgentRepEngine — Vendor Risk Summary
**Version:** 1.0 | **Date:** April 2026 | **Classification:** Customer-Shareable

This document answers the standard vendor risk questionnaire questions that enterprise
procurement and security teams ask before approving a new vendor. It is designed to be
shared directly with your vendor risk management team.

---

## 1. COMPANY AND PRODUCT

| Field | Answer |
|---|---|
| Company name | Naseem A2A Research Lab / Call2leads Inc. (S-Corp) |
| Product name | AgentRepEngine (ARE) |
| Product version | 1.4.0 |
| Published IP | Zenodo DOI 10.5281/zenodo.19169185 (March 22, 2026) |
| GitHub | github.com/Rehanrana11/AgentRepEngine |
| Deployment model | On-premise only. No SaaS. No cloud dependency. |
| Data residency | All data remains in customer's own infrastructure. |

---

## 2. DATA HANDLING

| Question | Answer |
|---|---|
| Does ARE send data to external servers? | No. ARE runs entirely within customer infrastructure. |
| Does ARE access agent request content? | No. ARE reads metadata only: agent ID, request rate, endpoint path, response code. |
| Does ARE store PII? | No. Agent behavioral metadata does not constitute PII under GDPR Article 4. |
| What data does ARE store? | Agent DID, event type, score, timestamp, decision, hash chain. No request body content. |
| Where is data stored? | Customer's own PostgreSQL instance (port 5433). Customer controls all data. |
| What is the data retention period? | 90 days by default. Configurable via retention job. Customer controls deletion. |
| Does ARE use third-party analytics? | No. No telemetry, no analytics, no external API calls in production. |
| Does ARE use LLMs or external AI APIs? | No. Phase 1 scoring is pure statistical (z-score, velocity counting). No LLM calls. |

---

## 3. SECURITY ARCHITECTURE

| Question | Answer |
|---|---|
| Authentication model | RS256 JWT with per-org JWKS. Every agent request cryptographically verified. |
| API authentication | X-API-Key header required on all write endpoints. |
| Encryption at rest | PostgreSQL and Redis encrypted at rest per customer's infrastructure policy. ARE does not manage encryption keys. |
| Encryption in transit | TLS termination at Kong gateway. Internal service communication within customer network. |
| Audit trail | SHA-256 hash-chained enforcement log. Tamper-evident. Customer-verifiable without ARE involvement. |
| False positive rate | 0.00% on 100-scenario internal validation corpus. Target <2% in production. Auto-rollback if exceeded. |
| Fail-open behavior | If Redis or PostgreSQL unavailable: ARE fails open (agents continue operating). Enforcement suspends. SIEM alert fires. |
| Enforce mode gate | Enforce mode requires explicit CISO sign-off after 14 days of clean observe mode. |

---

## 4. COMPLIANCE

| Framework | Status |
|---|---|
| DORA (EU) | Architecture designed for DORA Article 8(4) ICT risk management. `cmd/dora-verify` produces examiner-ready audit report. |
| GDPR | Data minimization by design. No PII processed. DPA template available (see companion document). |
| SOC2 (Type II) | SOC2-ready architecture. Hash chain + audit log satisfies CC7.2 monitoring controls. Type II audit pending first enterprise deployment. |
| HIPAA | On-premise deployment satisfies data residency. No PHI processed by ARE. BAA available on request. |
| NIST AI RMF | Observe-mode baseline validation maps to NIST AI RMF GOVERN and MEASURE functions. |
| OWASP LLM Top 10 | Policy packs cover OWASP LLM01–LLM10 agent attack vectors. |

---

## 5. OPERATIONAL

| Question | Answer |
|---|---|
| Minimum infrastructure | Kong API gateway (min 2.8), Redis, PostgreSQL. Docker-deployable. |
| Deployment time | 30-day observe-mode pilot. No procurement integration required. |
| Does ARE require agent code changes? | No. Kong plugin intercepts at gateway layer. No SDK. No agent modification. |
| Support model | Direct founder support during pilot. SLA defined in pilot LoU. |
| Incident response | Auto-rollback to observe mode if FP rate exceeds 2%. SIEM alert fires within 5 minutes. |
| Uptime dependency | ARE is fail-open. Customer operations are never blocked by ARE infrastructure failure. |

---

## 6. PILOT TERMS

- **Duration:** 30 days observe mode minimum
- **Enforce mode:** Only after CISO sign-off at day 14+
- **Data:** All data stays in customer environment
- **Exit:** Customer can remove ARE at any time with zero data loss (all data in customer PostgreSQL)
- **Cost:** Pilot is free. ACV $50K–$150K post-pilot based on agent count.

---

*For procurement questions contact: [founder direct]*
*DPA template: see `docs/enterprise/data-processing-agreement-template.md`*