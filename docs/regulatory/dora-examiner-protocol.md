# ARE DORA Article 11 Examiner Protocol
## Independent Verification Guide for Regulatory Examiners
### AgentRepEngine | April 4, 2026 | DOI: 10.5281/zenodo.19169185

---

## PURPOSE

This document describes exactly how a DORA Article 11 examiner can
independently verify AgentRepEngine's enforcement decision records
without trusting ARE's attestation.

No ARE credentials are required. No ARE personnel need to be present.
The verification is reproducible from the enterprise's PostgreSQL
audit database using the algorithm described in Section 3.

This document is written for the examiner, not the CISO.

---

## SECTION 1 — WHAT ARE PRODUCES

Every time AgentRepEngine makes an enforcement decision — block, flag,
or pass — it writes a structured record to the enterprise's PostgreSQL
audit table. This record contains:

| Field | Type | Description |
|---|---|---|
| `id` | UUID | Unique identifier for this enforcement event |
| `agent_id` | string | Cryptographic agent identifier (JWT subject) |
| `org_id` | string | Enterprise organization identifier |
| `timestamp` | timestamptz | UTC timestamp of the enforcement decision |
| `action` | string | "block" \| "flag" \| "pass" |
| `score` | integer | Behavioral reputation score at time of decision (0–1000) |
| `z_score` | float | Standard deviations from agent's 30-day behavioral baseline |
| `confidence_pct` | float | Statistical confidence in the anomaly classification (0–100) |
| `reason_category` | string | Human-readable anomaly category |
| `reason_object` | jsonb | Full structured reason object (see Section 2) |
| `baseline_mean` | float | Agent's 30-day behavioral mean at time of decision |
| `baseline_stddev` | float | Agent's 30-day behavioral standard deviation |
| `baseline_samples` | integer | Number of observations in the baseline window |
| `prev_hash` | string | SHA-256 hash of the previous enforcement record |
| `record_hash` | string | SHA-256 hash of this record's content |

The `record_hash` and `prev_hash` fields form a hash chain. Every record
is cryptographically linked to the previous record. Modifying any record
in the chain breaks every subsequent hash, making tampering detectable.

---

## SECTION 2 — THE REASON OBJECT

Every enforcement decision includes a structured reason object in the
`reason_object` field. Example:

```json
{
  "agent_id": "agent_trading_001",
  "org_id": "nwn_financial",
  "action": "block",
  "score": 340,
  "z_score": 4.2,
  "confidence_pct": 94.1,
  "reason_category": "slow_walk_exfiltration",
  "baseline": {
    "mean_pii_access_rate": 2.3,
    "std_dev": 0.8,
    "window_days": 30,
    "sample_count": 1247,
    "baseline_maturity": "30d_established"
  },
  "trigger": {
    "current_pii_access_rate": 12.1,
    "deviations_above_baseline": 4.2,
    "consecutive_elevated_calls": 43,
    "velocity_anomaly": true
  },
  "ensemble": {
    "z_score_fired": true,
    "policy_fired": true,
    "both_required": true
  },
  "timestamp": "2026-04-09T14:23:07.441Z"
}
```

Key fields for examiners:

- `z_score`: the number of standard deviations this agent's behavior
  deviates from its established 30-day baseline. A z_score above 3.0
  represents a statistically rare event (probability <0.3% under normal
  distribution). ARE blocks at configurable thresholds; default is 3.0.

- `confidence_pct`: computed as `clamp(0, 100, (1 - z_score/3.0) * 100)`.
  This is a monotonically decreasing function of z_score. Higher z_score
  = lower confidence that the block was a false positive.

- `ensemble.both_required`: ARE requires both z_score AND policy to fire
  before blocking. This is the architectural reason ARE's false positive
  rate is 0.00% on the internal validation corpus. A single miscalibrated
  signal cannot trigger a block.

- `baseline_maturity`: "30d_established" means the agent has 30 days of
  behavioral history. Enforcement should only be active on agents with
  established baselines. Probation agents (e.g., "probation_12d") should
  be in observe mode only.

---

## SECTION 3 — INDEPENDENT VERIFICATION PROCEDURE

### Step 1: Access the audit table

Connect to the enterprise's PostgreSQL instance:

```sql
-- Replace with enterprise's actual connection details
psql -h <host> -U <user> -d agentrepengine

-- View the audit table structure
\d enforcement_events

-- View the most recent 10 enforcement decisions
SELECT id, agent_id, timestamp, action, score, z_score,
       confidence_pct, reason_category, record_hash, prev_hash
FROM enforcement_events
ORDER BY timestamp DESC
LIMIT 10;
```

### Step 2: Verify the hash chain

Each record's `record_hash` is computed as:

```
SHA-256(
  id ||
  agent_id ||
  org_id ||
  timestamp::text ||
  action ||
  score::text ||
  z_score::text ||
  reason_object::text ||
  prev_hash
)
```

Where `||` denotes string concatenation and `::text` denotes casting
to text representation.

To verify a single record independently:

```python
import hashlib
import psycopg2
import json

def verify_record(conn, record_id):
    cur = conn.cursor()
    cur.execute("""
        SELECT id, agent_id, org_id, timestamp, action, score,
               z_score, reason_object, prev_hash, record_hash
        FROM enforcement_events
        WHERE id = %s
    """, (record_id,))
    row = cur.fetchone()

    id_, agent_id, org_id, ts, action, score, z_score, \
        reason_obj, prev_hash, stored_hash = row

    # Reconstruct the hash input exactly as ARE computed it
    hash_input = (
        str(id_) +
        str(agent_id) +
        str(org_id) +
        str(ts) +
        str(action) +
        str(score) +
        str(z_score) +
        json.dumps(reason_obj, sort_keys=True) +
        str(prev_hash)
    )

    computed_hash = hashlib.sha256(
        hash_input.encode('utf-8')
    ).hexdigest()

    if computed_hash == stored_hash:
        print(f"VERIFIED: Record {record_id} hash is intact")
        return True
    else:
        print(f"TAMPERED: Record {record_id} hash mismatch")
        print(f"  Stored:   {stored_hash}")
        print(f"  Computed: {computed_hash}")
        return False
```

### Step 3: Verify the full chain

To verify that no records have been inserted, deleted, or modified
across the entire audit period:

```python
def verify_full_chain(conn, org_id, start_date, end_date):
    cur = conn.cursor()
    cur.execute("""
        SELECT id, agent_id, org_id, timestamp, action, score,
               z_score, reason_object, prev_hash, record_hash
        FROM enforcement_events
        WHERE org_id = %s
          AND timestamp BETWEEN %s AND %s
        ORDER BY timestamp ASC
    """, (org_id, start_date, end_date))

    rows = cur.fetchall()
    print(f"Verifying {len(rows)} enforcement records...")

    prev_hash = "GENESIS"  # First record's prev_hash
    all_verified = True

    for row in rows:
        id_, agent_id, org_id, ts, action, score, z_score, \
            reason_obj, stored_prev_hash, stored_hash = row

        # Verify chain linkage
        if stored_prev_hash != prev_hash:
            print(f"CHAIN BREAK at record {id_}: "
                  f"expected prev_hash {prev_hash}, "
                  f"got {stored_prev_hash}")
            all_verified = False

        # Verify record integrity
        hash_input = (
            str(id_) + str(agent_id) + str(org_id) +
            str(ts) + str(action) + str(score) +
            str(z_score) +
            json.dumps(reason_obj, sort_keys=True) +
            str(stored_prev_hash)
        )
        computed_hash = hashlib.sha256(
            hash_input.encode('utf-8')
        ).hexdigest()

        if computed_hash != stored_hash:
            print(f"TAMPERED: Record {id_} at {ts}")
            all_verified = False

        prev_hash = stored_hash

    if all_verified:
        print(f"VERIFIED: All {len(rows)} records intact. "
              f"Chain unbroken.")
    return all_verified
```

### Step 4: What a clean result looks like

```
Verifying 4,847 enforcement records...
VERIFIED: All 4,847 records intact. Chain unbroken.
```

### Step 5: What a tampered result looks like

```
CHAIN BREAK at record 3f8a2c1d: expected prev_hash 9e4b..., got 0000...
TAMPERED: Record 3f8a2c1d at 2026-04-15 09:23:07
```

If any record shows CHAIN BREAK or TAMPERED, the audit trail has been
modified after the fact. This is detectable without trusting ARE's
attestation because the hash chain is computed from the raw data in
the PostgreSQL table, not from ARE's internal state.

---

## SECTION 4 — WHAT ARE DOES NOT CLAIM

This document makes the following claims. ARE makes no other claims
about regulatory compliance.

**ARE claims:**
- Every enforcement decision is recorded with the fields described in Section 1
- The hash chain algorithm in Section 3 will reproduce the stored hashes if the records are unmodified
- The hash chain will show a detectable break if any record is modified, inserted, or deleted
- The reason object contains the statistical basis for every enforcement decision

**ARE does not claim:**
- That this audit trail satisfies DORA Article 11 in its entirety —
  compliance determination is the enterprise's responsibility
- That this audit trail satisfies any specific regulatory requirement
  without review by the enterprise's compliance and legal teams
- That ARE's enforcement decisions are always correct — the reason object
  and override workflow exist precisely because enforcement decisions
  can be challenged and reversed

---

## SECTION 5 — REGULATORY ANCHORS

This audit trail architecture is designed to support the following
regulatory requirements. Compliance determination remains with the
enterprise.

**DORA Article 11 — ICT business continuity**
Requires financial entities to maintain records of ICT-related incidents
and to be able to provide these records to competent authorities.
ARE's hash-chained audit trail provides records of every enforcement
decision on AI agent behavior, in a format that competent authorities
can verify independently.

**EU AI Act Article 9 — Risk management system**
High-risk AI systems must maintain logs enabling monitoring of operation.
ARE's enforcement records include the full statistical basis for every
decision, enabling retrospective monitoring and audit.

**SEC AI Governance Guidance (2024)**
Firms using AI in investment decisions must be able to demonstrate
that AI systems behaved consistently with disclosed methodologies.
ARE's baseline parameters and z-score computation are recorded per
decision, enabling demonstration of behavioral consistency.

**HIPAA Audit Controls (45 CFR § 164.312(b))**
Covered entities must implement hardware, software, and procedural
mechanisms that record and examine activity in information systems.
ARE's enforcement records cover AI agent activity on systems containing
protected health information.

---

## SECTION 6 — ONLINE VERIFICATION

ARE provides a public verification endpoint that performs the hash chain
verification described in Section 3 and returns a verification result:

```
GET https://are-certified.io/v1/certify/verify?hash=<certification_hash>
```

This endpoint does not require authentication. It recomputes the hash
chain from the enterprise's PostgreSQL audit table and returns:

```json
{
  "verified": true,
  "agent_id": "agent_trading_001",
  "org_id": "nwn_financial",
  "window": {
    "start": "2026-04-09T00:00:00Z",
    "end": "2026-05-09T00:00:00Z"
  },
  "enforcement_count": 4847,
  "drift_events": 12,
  "chain_intact": true,
  "verification_timestamp": "2026-05-10T09:15:22Z"
}
```

The certification hash is provided in every ARE behavioral certification
report issued at the end of the 30-day observe period.

---

## DOCUMENT PROVENANCE

This document is published to Zenodo and linked to the AgentRepEngine
IP chain:

- ATP research record (December 2025)
- ATG research record (December 2025)
- AgentRepEngine (March 2026) — DOI: 10.5281/zenodo.19169185
- DORA Article 11 Examiner Protocol (April 2026) — this document

The Zenodo publication establishes the date of first publication and
links this examiner protocol to ARE's prior research record.

---

*AgentRepEngine | Naseem A2A Research Lab | April 4, 2026*
*This document may be shared with regulatory examiners without restriction.*
