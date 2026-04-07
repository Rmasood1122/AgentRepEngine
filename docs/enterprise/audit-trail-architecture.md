# ARE Audit Trail Architecture
# Version: 1.0 — April 7, 2026
# Purpose: Defines the two complementary audit mechanisms and their
#           relationship. One source of truth. Two verification tools.
#           Five compliance frameworks covered.

## THE ONE SOURCE OF TRUTH

enforcement_decisions (PostgreSQL)
  - INSERT-only enforced at DB permission level (REVOKE DELETE/UPDATE)
  - Every enforcement decision written here exactly once
  - Nothing is ever modified or deleted
  - Both audit mechanisms read from this table

## TWO COMPLEMENTARY MECHANISMS

### LAYER 1 — Sequential Integrity (SHA-256 Hash Chain)

Question answered:
  "Has the entire audit trail been tampered with since deployment?"

How it works:
  Every row contains SHA-256 hash of (row_data + previous_hash).
  Sequential. Every decision linked to every decision before it.
  A single deleted or modified row breaks the entire chain.
  The break location identifies exactly which decision was touched.

Verification command:
  cmd/verify-chain
  Expected output: VALID — N decisions verified, chain intact
  If INVALID: reports exact break location (agent_did + timestamp)

Regulator mapping:
  DORA Article 17    — ICT incident sequential audit trail
  SOC 2 CC7.2        — System monitoring and anomaly detection
  SEC Rule 17a-4     — Electronic records preservation
  NIST AI RMF MG-2.2 — AI risk monitoring records

CISO claim:
  "Run cmd/verify-chain. If it returns VALID, every enforcement
   decision since deployment day one is cryptographically sealed.
   No one — including us — can delete or modify a past decision.
   You run the verification. We watch."

### LAYER 2 — Selective Proof (Merkle Tree)

Question answered:
  "Can you prove this specific decision existed without revealing
   any other agent's data?"

How it works:
  Daily Merkle tree built over enforcement_decisions rows.
  Root hash stored in merkle_roots table.
  Individual proofs generated per decision_id.
  Proof path = 32 bytes. Reveals nothing about other decisions.

Verification command:
  cmd/verify-decision [decision_id]
  Expected output: VALID — decision {id} verified against root {hash}
  If INVALID: proof failed, specific decision may be compromised

Regulator mapping:
  HIPAA §164.528     — Access reporting for specific PHI incidents
  GDPR Article 22    — Selective disclosure without exposing other subjects
  EU AI Act Art.86   — Incident-specific evidence without full log exposure

CISO claim:
  "When a regulator asks for proof of one specific enforcement
   decision, we produce a 32-byte proof. They verify it against
   the published daily root hash. They see nothing else.
   No other agent's data is exposed."

## THE INTEGRATION

SPHINCS+ PQC and Merkle are architecturally integrated:
  pqc_signer.go SignAuditEvent() takes merkleLeaf as parameter.
  One enforcement event = Merkle leaf + PQC signature over that leaf.
  Hash chain + Merkle + PQC = one unified audit layer, not three systems.

This means:
  Every enforcement decision is simultaneously:
  1. Linked into the sequential hash chain (DORA/SOC2)
  2. A leaf in the daily Merkle tree (HIPAA/GDPR)
  3. Signed with a post-quantum signature (NIST SP 800-208)

## VERIFICATION SEQUENCE FOR REGULATORS

Step 1: cmd/verify-chain
  → proves nothing was deleted or modified since deployment
  → DORA Article 17, SOC 2 CC7.2

Step 2: cmd/verify-decision [decision_id]
  → proves specific incident existed without exposing others
  → HIPAA §164.528, GDPR Article 22

Step 3: Reason object (JSON from enforcement_decisions)
  → proves decision was explainable at time of enforcement
  → SEC Rule 15c3-5, GDPR Article 22, EU AI Act Article 13

Step 4: cmd/dora-verify [date_range]
  → produces Article 17 formatted evidence report
  → DORA RTS audit trail requirements

Step 5: cmd/oscal-generate [framework] (Sprint 1)
  → produces SOC2/NIST OSCAL evidence bundle
  → SOC 2 Type II CC1-CC9

These five steps = complete compliance evidence package.
No vendor trust required. Every step runs independently.

## COMPLIANCE COVERAGE MAP

Component           DORA    SOC2    HIPAA   SEC     NIST    GDPR    EU AI
────────────────────────────────────────────────────────────────────────────
Hash Chain ✅        Art.17  CC7.2   §164.3  17a-4   MG-2.2  Art.5   Art.9
Merkle Proofs ✅     RTS     CC7.3   §164.5  —       —       Art.22  Art.86
SPHINCS+ PQC ✅      RTS'26  —       HTI-1   —       800-208 —       —
ZK-STARK ✅          —       —       §164.5  —       —       Art.22  Art.86
OSCAL Bundle ✅      —       CC1-9   —       —       All     —       —
Reason Object ✅     Art.28  CC6.1   —       15c3-5  EX-1.1  Art.22  Art.13
Auto-rollback ✅     Art.11  A1.3    §164.3  —       MG-4.1  —       —

✅ = wired and callable today

## CLI ENTRY POINTS STATUS

cmd/verify-chain          ✅ exists — hash chain sequential integrity
cmd/dora-verify           ✅ exists — DORA Article 17 evidence report
cmd/verify-decision       ⏳ Sprint 1 — calls internal/audit/merkle.go GenerateProof()
cmd/oscal-generate        ⏳ Sprint 1 — calls internal/compliance/oscal.go GenerateOSCALBundle()
scripts/generate-evidence-package.sh  ⏳ Sprint 1 — unified auditor ZIP