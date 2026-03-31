# AgentRepEngine — GDPR Position Statement
## Data Subject Rights and Agent Identity Data
Version 1.1 | For: DPO Review | Classification: Confidential

---

### Summary

AgentRepEngine stores agent behavioral metadata by default.
Agent DIDs are pseudonymous technical identifiers.
Where agent DIDs can be linked to natural persons, customers
must apply appropriate DPIA under GDPR Article 35.

ARE's enforcement decisions are made on agent behavior, not on
human subjects. The staged rollout with explicit CISO sign-off
at each phase constitutes the human oversight mechanism required
for GDPR Article 22 compliance review.

A tombstone procedure for erasure requests is documented below.

---

### What We Store (Tier 1 Default)

| Field | Description | Personal Data? |
|-------|-------------|----------------|
| agent_did | Pseudonymous software agent identifier | Potentially — see below |
| event_type | Category of action (bulk_access, tool_call) | No |
| timestamp | UTC timestamp of event | No |
| endpoint | API endpoint called | No |
| count | Request count in window | No |
| duration_ms | Request duration | No |
| decision | ALLOW/AUDIT/BLOCK | No |
| reason_object | Structured enforcement explanation | No |
| feature_vector | 8-dimension behavioral metrics | No (aggregate metrics) |

**What we never store:** prompt or completion content, user PII,
request or response payloads. Behavioral metadata only.

**Default retention:** 90 days. Configurable per-deployment.

**Regulations compatible:** DORA, SOX data governance, HIPAA (no PHI stored).

---

### GDPR Article 22 — Automated Decision-Making

ARE's enforcement decisions are automated decisions made on
**agent behavior**, not on human subjects directly.

Key distinction:
- If an agent acts autonomously (no individual user mapping):
  Article 22 does not apply to ARE's enforcement decisions
- If an agent acts on behalf of a specific human user:
  Your organization must assess Article 22 applicability
  for that specific deployment context

**ARE's human oversight documentation:**
The risk-staged deployment model (zero-impact visibility →
flagging → enforcement with human review → full enforcement)
with explicit CISO sign-off at each stage constitutes the
human oversight mechanism required for Article 22 compliance
review. Every stage transition is documented.

ARE provides the cryptographically non-repudiable enforcement
log required for any Article 22 documentation request.

---

### Are Agent DIDs Personal Data?

Under GDPR Article 4(1), personal data means information
relating to an identified or identifiable natural person.

**Agent DIDs in AgentRepEngine follow the format:**
`did:jwt:{org_id}:{service_name}:{uuid}`

Example: `did:jwt:acme-corp:data-pipeline-agent:7f3a9b2c`

**Analysis:**
- A software agent DID does not directly identify a natural person
- Where an org maps agent DIDs to specific employees or users,
  the DID may constitute personal data in that deployment context
- Where agent DIDs are assigned to software services with no
  individual user mapping, they are not personal data

**Customer responsibility:**
Customers who map agent DIDs to individual employees must:
1. Conduct a DPIA under GDPR Article 35 before deployment
2. Document the legal basis for processing under Article 6
3. Request our tombstone procedure (below) for erasure handling

---

### Tombstone Procedure for Erasure Requests

When a data subject erasure request is received and agent DIDs
are determined to be personal data in your deployment:

**Step 1 — Identify affected records:**
```sql
SELECT id, agent_did FROM enforcement_decisions
WHERE agent_did = 'did:jwt:org:service:uuid';

SELECT id FROM agent_identities
WHERE did = 'did:jwt:org:service:uuid';
```

**Step 2 — Pseudonymize enforcement records:**
```sql
-- Replace agent_did with non-reversible pseudonym
-- Cryptographically non-repudiable enforcement log integrity preserved
UPDATE enforcement_decisions
SET agent_did = encode(
  sha256(('did:jwt:org:service:uuid' ||
          'erasure-salt-' || NOW()::text)::bytea),
  'hex'
),
erased_at = NOW(),
erasure_ref = 'GDPR-REQUEST-{ticket_id}'
WHERE agent_did = 'did:jwt:org:service:uuid';
```

**Step 3 — Delete identity record:**
```sql
DELETE FROM agent_identities
WHERE did = 'did:jwt:org:service:uuid';
```

**Step 4 — Verify cryptographically non-repudiable enforcement log integrity:**
```sql
SELECT verify_hash_chain('enforcement_decisions') AS valid;
-- Must return: true
-- Hash chain continuity is preserved after pseudonymization
```

**Step 5 — Document the erasure:**
```sql
INSERT INTO erasure_log
(agent_did_hash, erasure_timestamp, erasure_ref, requestor)
VALUES (
  encode(sha256('did:jwt:org:service:uuid'::bytea), 'hex'),
  NOW(),
  'GDPR-REQUEST-{ticket_id}',
  '{requestor_email}'
);
```

---

### Hash Chain Integrity After Erasure

The tombstone procedure preserves cryptographically non-repudiable enforcement log integrity because:
- The cryptographically non-repudiable enforcement log covers: prev_hash + id + timestamp + decision
- agent_did is stored as data but is NOT part of the hash input
- Pseudonymizing agent_did does not break the chain
- Cryptographic audit integrity is maintained. Personal data is removed.

This design was intentional: the cryptographically non-repudiable
enforcement log proves the sequence and integrity of enforcement
decisions, not the identity of agents.

---

### EU Localization Note

Phase 1 reason objects are English-language only.
EU/DORA deployments requiring localized enforcement documentation
(German, French, Dutch) should plan for Phase 2 multilingual
reason objects, available on the product roadmap.

---

### Recommended Customer Guidance

**For EU-regulated deployments:**
1. Use org-level or service-level DIDs that do not map to
   individual employees where possible
2. Conduct DPIA before deployment if DIDs map to individuals
3. Document legal basis for processing (legitimate interest
   in AI governance is a valid basis under Article 6(1)(f))
4. Contact us for tombstone procedure assistance:
   rehan@naseem-a2a.com

**Data residency:**
All behavioral telemetry stays within your network perimeter.
No data transfer outside your environment without explicit
configuration. EU data residency is achievable by default.

---

### Tier 2 and Tier 3 Telemetry

Tier 2 (hashed payloads) and Tier 3 (full payload) are
OFF by default and require explicit customer opt-in plus
a signed Data Processing Agreement (DPA).

We provide a standard DPA template on request.

---

*GDPR position current as of March 2026.
Legal review recommended before EU enterprise deployment.
This document is guidance, not legal advice.*
