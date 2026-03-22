# AgentRepEngine — APEX Laws to ATP/ATG Invariants Mapping
## Formal Connection Between Implementation and Research Foundation
Version 1.0 | For: Technical Due Diligence | IP Validation

Research corpus: Agentic Trust Architecture (Z0–Z6 + ID-RTP)
Author: Rehan Masood, Naseem A2A Research Lab, December 2025
DOI: 10.5281/zenodo.17917085

---

### Purpose

This document formally maps every APEX v5.2 Immutable Law to
its corresponding research invariant in the Agentic Trust
Architecture corpus. It demonstrates that AgentRepEngine is
not a product inspired by research — it is the direct
commercial implementation of formally published primitives.

Every architectural decision in AgentRepEngine traces to a
citable, timestamped, DOI-registered research publication.
No competitor can replicate this chain retroactively.

---

### The Seven APEX Laws — Research Grounding

---

#### L1 — One blocked incident in production is worth more than ten architectural improvements on paper. Ship the wedge first.

**ATP Invariant I₁ — Action Primacy**
Observable enforcement outcomes take precedence over
theoretical correctness. A system that cannot demonstrate
a blocked incident has not proven its trust model.

**ATG Connection:**
The Agentic Trust Graph accumulates behavioral evidence
over time. Without real deployment data, the graph has
no signal. L1 enforces that the graph must be populated
with real enforcement events before any architectural
enhancement is justified.

**Implementation:**
- Prime Directive: one enterprise, one blocked incident
- Task 8 (T8) is the only gate that matters in Phase 1
- All other tasks exist to make T8 possible

---

#### L2 — Fail closed on enforcement. An unexplained block is an untrustworthy block.

**ATP Invariant I₅ — Deferral Precedence**
When authority preconditions are unmet (reason object
cannot be generated), action must be suspended rather
than executed without authorization.

**Z5 Connection — Deferral Principle:**
"A system that acts without being able to explain why
it acted has exceeded its authority bounds. Deferral
is not failure — it is constitutional constraint."

**Implementation:**
- If reason object generation fails → AUDIT, never BLOCK
- Enforcement failure path: fail-closed (no unexplained blocks)
- Infrastructure failure path: fail-open (availability preserved)
- These are two distinct failure classes with two correct behaviors

**Note:** The fail-open/fail-closed distinction maps directly
to Z5's constitutional separation between infrastructure
availability and enforcement authority.

---

#### L3 — False positive rate is measured before any enforcement goes live. Target ≤2%.

**Z5 Connection — Non-Terminal Deferral:**
Every deferral (enforcement decision) must be non-terminal
— the system must be able to recover from incorrect decisions.
A false positive that cannot be measured cannot be corrected.

**ATP Invariant I₄ — Score Idempotency:**
Identical behavioral input must produce identical scoring
output. This invariant enables FP rate measurement — if
scoring is non-deterministic, FP rate is unmeasurable.

**Implementation:**
- G-FP gate: HARD STOP before any enforcement ships
- 100-scenario legitimate corpus: 0.00% FP rate
- Production monitoring: daily_fp_metrics view
- Auto-rollback: ModeController rolls back at >2% FP

---

#### L4 — The identity model must be adoptable in one day. JWT/JWKS is a 4-hour integration.

**Z3 Connection — Entity Engineering:**
Entity Cards must be practically adoptable, not
theoretically elegant. An identity model that requires
6 months of procurement conversation violates the
Entity Engineering prerequisite — no trust graph can
be built without identity primitives in place.

**ATG Structural Invariant I₁ — Non-Mergeability:**
Identity nodes in ATG are formally non-mergeable by default.
JWT/RS256 with org-managed JWKS enforces this at the
cryptographic layer — no two agents can share an identity.

**Implementation:**
- JWT/RS256 identity — 4-hour enterprise install
- DID/ledger rejected from Phase 1 (procurement blocker)
- JWKS endpoint: GET /jwks — org-managed key rotation
- Identity cycling detection: >5 new DIDs/org/hour → alert

---

#### L5 — Every score change must produce a structured reason object. Score alone is not enough.

**ATG Invariant I₃ — Lineage Recoverability:**
Every state change in the Agentic Trust Graph must be
recoverable and attributable. A score change without
a reason object breaks the lineage chain — the graph
state cannot be forensically reconstructed.

**Z4 Connection — Authority Invariant:**
Every enforcement action must be attributable to a
specific authority source. "The score was 187" is not
an authority source. "Policy bulk_pii_access_prevention_v1
fired at 4.2σ above baseline" is.

**Implementation:**
```json
{
  "decision": "BLOCKED",
  "agent_did": "did:jwt:org:agent:001",
  "score": 187,
  "score_delta": -556,
  "policy_fired": "bulk_pii_access_prevention_v1",
  "policy_threshold": "PII access >500 records in 2h window",
  "trigger_events": [...],
  "peer_cluster_avg_score": 700,
  "deviation_from_cluster": -513
}
```
- G-EXPLAIN gate: HARD STOP — no enforcement without reason object
- INSERT-only enforcement_decisions table
- SHA-256 hash chain: tamper-evident lineage

---

#### L6 — The behavioral dataset is the moat. Instrument to collect it from day one.

**ATG Invariant I₂ — Conflict Visibility:**
Every behavioral conflict must be observable in the
Agentic Trust Graph. An ATG with no behavioral data
has no conflict visibility — the graph is structurally
present but functionally empty.

**ATG Connection — Trust Accumulation:**
ATG edges accumulate trust signal over time. The moat
is not the graph structure — it is the behavioral data
that populates the edges. Day 1 of deployment = day 1
of moat accumulation.

**Implementation:**
- Feature vector stored on EVERY scored event (8 dimensions)
- L74 verification: COUNT(feature_vector) = COUNT(*)
- Data moat defensible at 90 days per enterprise
- Network effect at Phase 3 (5+ enterprises federated)

**8 behavioral feature dimensions:**
```
tool_call_rate_per_hour
unique_endpoints_per_hour
bulk_access_count_per_session
pii_field_access_rate
cross_tenant_probe_count
permission_escalation_count
sub_agent_spawn_depth
token_refresh_rate
```

---

#### L7 — Federation is a Phase 3 moat, not a Phase 1 requirement.

**ATG Invariant I₄ — Trust Monotonicity:**
Trust in ATG is monotonically earned — it cannot be
artificially inflated by federation before standalone
value is proven. Cross-tenant trust sharing requires
that each tenant's individual trust graph is validated
independently first.

**Z1 Connection — ATP Protocol Sequencing:**
ATP's state machine (RESOLVE → VERIFY → DEFER) operates
on a single tenant's behavioral history. Federation adds
cross-tenant signal but cannot substitute for individual
behavioral evidence. Phase 1 validates the single-tenant
state machine. Phase 3 adds federation when N≥5 tenants
provide independent validation.

**Implementation:**
- Phase 1: single-tenant enforcement (validated)
- Phase 2: intelligence layer (ML scoring)
- Phase 3: federation (cross-tenant signal sharing)
- Anti-scope: federation triggers G-KILL if requested in Phase 1

---

### Score Band to ATP State Mapping

This table formally connects AgentRepEngine numerical
score bands to ATP protocol states from the research corpus.

| Score Range | ATP State | Enforcement Action | Conflict Severity Γ |
|-------------|-----------|-------------------|---------------------|
| 700–1000 | RESOLVE | ALLOW + passive log | Γ < 0.3 |
| 400–699 | VERIFY | ALLOW + active audit + human review on HIGH_RISK | 0.3 ≤ Γ < 0.7 |
| 200–399 | RESTRICT | THROTTLE + human-in-loop | 0.5 ≤ Γ < 0.7 |
| 0–199 | DEFER | Synthetic response (not 403) | Γ ≥ 0.7 |

**ATP State Definitions (from Z1):**
- RESOLVE: Authority preconditions met. Action authorized.
- VERIFY: Authority uncertain. Human verification required.
- DEFER: Authority preconditions unmet. Action suspended.

**Γ(C_o) — Conflict Severity Function:**
Γ is monotonic — adding conflicts never decreases severity.
Defined in config/scoring_weights.yaml:
```yaml
conflict_severity:
  bulk_pii_access:          0.40
  permission_escalation:    0.35
  cross_tenant_probe:       0.50
  recursive_spawn:          0.30
  credential_access:        0.45
  high_frequency_tool_abuse: 0.25
```

---

### Research-to-Product Mapping Summary

| Research Primitive | Source | Product Implementation |
|-------------------|--------|----------------------|
| ATP State Machine | Z1 | Kong gateway enforcement bands |
| ATG Behavioral Graph | Z2 | agent_events + feature vectors |
| Entity Cards | Z3 | JWT claims schema (AgentClaims) |
| Authority Invariant | Z4 | YAML policy packs + override workflow |
| Deferral Principle | Z5 | Fail-closed enforcement + AUDIT fallback |
| Trust Representation | Z6 | Reason object schema |
| Truth Kernel™ | Z0 | Kong gateway enforcement boundary |
| ConflictSet | Z2 | trigger_events in reason object |
| Lineage L(o) | Z3 | lineage_hash + lineage_depth in JWT |
| ID-RTP Protocol | ID-RTP | HIGH_RISK VERIFY escalation workflow |

---

### Acquisition Value Statement

This mapping demonstrates that acquiring AgentRepEngine
means acquiring:

1. **The product** — working runtime enforcement engine
2. **The research** — formally published primitives that
   define the category standard (DOI: 10.5281/zenodo.17917085)
3. **The vocabulary** — "Agent Trust Infrastructure" named
   and defined before any competitor
4. **The priority** — December 2025 Zenodo timestamps
   establish prior art that cannot be challenged retroactively

No competitor can build this chain backwards. The research
was published before the product. The product implements
the research. The chain is complete, citable, and permanent.

---

*Mapping document current as of March 2026.
Updated when new research publications are added to corpus.*