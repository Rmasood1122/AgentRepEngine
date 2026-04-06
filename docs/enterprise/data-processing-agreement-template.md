# Data Processing Agreement — Template
**AgentRepEngine (ARE) | Version 1.0 | April 2026**

This Data Processing Agreement ("DPA") is entered into between:

**Controller:** [CUSTOMER LEGAL NAME] ("Customer")
**Processor:** Call2leads Inc. / Naseem A2A Research Lab ("ARE Provider")

---

## ARTICLE 1 — SUBJECT MATTER

1.1 This DPA governs the processing of personal data by ARE Provider on behalf of Customer
in connection with the deployment of AgentRepEngine ("ARE") within Customer's infrastructure.

1.2 ARE processes AI agent behavioral metadata only. This data does not constitute personal
data under GDPR Article 4 in standard deployments because it describes machine behavior,
not natural persons. This DPA applies as a precautionary measure where Customer's legal
counsel determines that any processed data may fall within GDPR scope.

---

## ARTICLE 2 — NATURE AND PURPOSE OF PROCESSING

| Field | Detail |
|---|---|
| Nature | Automated scoring of AI agent behavioral patterns |
| Purpose | ICT risk management and AI agent trust enforcement |
| Data types | Agent identifier (DID), event type, request metadata, behavioral score, enforcement decision |
| Data subjects | None in standard deployment (machine-generated data only) |
| Duration | Duration of pilot agreement plus 90-day retention period |

---

## ARTICLE 3 — CUSTOMER OBLIGATIONS (CONTROLLER)

3.1 Customer shall ensure a lawful basis exists for any processing under this DPA.

3.2 Customer shall notify ARE Provider of any data subject requests within 5 business days.

3.3 Customer is responsible for configuring ARE's data retention period in accordance with
Customer's data governance policy.

---

## ARTICLE 4 — ARE PROVIDER OBLIGATIONS (PROCESSOR)

4.1 **On-premise deployment only.** ARE Provider does not operate any infrastructure that
processes Customer data. All ARE components run within Customer's own environment.

4.2 **No data transfer.** ARE Provider does not transfer, copy, or access Customer data
except during on-site support sessions explicitly authorized by Customer.

4.3 **No sub-processors.** ARE has no sub-processors in Phase 1. No Customer data is
shared with third parties.

4.4 **Security measures.** ARE implements: RS256 JWT authentication, SHA-256 hash-chained
audit trail, Redis ACL access controls, PostgreSQL row-level data isolation.

4.5 **Deletion.** Upon termination, Customer controls all data deletion. ARE Provider
holds no copies of Customer data.

---

## ARTICLE 5 — DATA SUBJECT RIGHTS

5.1 As ARE processes machine-generated behavioral metadata rather than personal data in
standard deployments, data subject rights under GDPR Articles 15–22 are not triggered.

5.2 Where Customer's legal counsel determines otherwise, Customer shall handle all data
subject requests directly, as Customer controls all data storage.

---

## ARTICLE 6 — SECURITY INCIDENT NOTIFICATION

6.1 ARE's auto-rollback mechanism detects anomalous enforcement patterns within 5 minutes
and fires a SIEM alert to Customer's security team.

6.2 ARE Provider shall notify Customer of any security vulnerability in ARE software
within 24 hours of discovery.

6.3 As Customer controls all data infrastructure, Customer is responsible for breach
notification under GDPR Article 33.

---

## ARTICLE 7 — AUDIT RIGHTS

7.1 Customer may audit ARE's compliance with this DPA at any time by running
`cmd/dora-verify` against their own PostgreSQL instance. No ARE Provider involvement required.

7.2 Customer may request a copy of ARE's security documentation at any time.

---

## ARTICLE 8 — TERM AND TERMINATION

8.1 This DPA is effective from the pilot start date and terminates 90 days after
the end of the commercial agreement.

8.2 Upon termination, Customer deletes ARE components from their infrastructure.
ARE Provider retains no Customer data.

---

## SIGNATURES

| Role | Name | Title | Date | Signature |
|---|---|---|---|---|
| Controller | | | | |
| Processor | | ARE Provider | | |

---

*This template is provided for Customer review. Customer's legal counsel should
review before execution. ARE Provider can execute this DPA as-is or negotiate
modifications on request.*