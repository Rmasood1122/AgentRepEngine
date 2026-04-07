# ARE GAP CLOSURE ROADMAP
# Source: Coordination, Orchestration, Interoperability & Scalability Audit
# Date: April 7, 2026
# Authority: Supersedes all prior gap lists for sequencing decisions
# Gate rule: G-FP fires before and after every engineering commit
# Phase rule: Nothing in Sprint 1 starts before Lloyd LoU signed
# Nothing in Sprint 2 starts before first pilot go-live

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
BEFORE LLOYD (April 28) — Non-negotiable. Zero exceptions.
Total: ~2.5 hrs
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

PRE-1: ASK LLOYD ABOUT KUBERNETES                              15 min
  Action: Add one line to Lloyd confirmation message this week:
    "Quick question before we meet — do you run Kubernetes
     in your agent deployment environment, or straight Docker?"
  Why: Single answer determines whether Helm chart is a Sprint 1
       blocker or Sprint 2 optional. Changes the install story.
       Changes the 4-hour claim. Must know before April 28.
  Output: Yes/No recorded in CONTINUATION_PROMPT pipeline entry.

PRE-2: UPDATE BUILD_INTELLIGENCE FILE MAP                       1 hr
  File: docs/BUILD_INTELLIGENCE_v1.md Section 7
  Action: Add these missing packages to the file map:
    internal/audit/merkle.go       Merkle tree + daily root + proof generation
    internal/attestation/          Software TEE quote generation + verification
    internal/zkp/                  ZK-STARK composite proof (Phase1 stub)
    internal/consensus/            Raft ceiling consensus (Phase2 topology)
    internal/compliance/oscal.go   OSCAL SOC2/NIST evidence bundle generator
    cmd/dora-verify/               DORA Article 17 formatted evidence report
  Also update: "58 commits, 5,587 lines of Go" → current numbers
  Commit: "docs: BUILD_INTELLIGENCE file map updated — nuclear packages added"

PRE-3: RAFT SINGLE-NODE FALLBACK DOCUMENTATION                 30 min
  File: docs/architecture/system-architecture.md
  Action: Add one section:
    "RAFT CONSENSUS — DEPLOYMENT TOPOLOGY NOTE
     Phase 1 single-node deployment: ceiling decisions made directly
     by the scoring service. Raft consensus is not active.
     Phase 2 multi-node deployment: Raft activates when 3+ scoring
     service instances are deployed. Quorum required for ceiling changes.
     Current Phase 1 deployment is single-node. This is correct and
     documented. Do not attempt to activate Raft in Phase 1."
  Why: Lloyd's security architect will ask about this.
       Without documentation, it looks like a bug not a design choice.
  Commit: "docs: Raft single-node fallback documented — Phase1 vs Phase2 topology"

PRE-4: ZK-STARK CLAIM QUALIFICATION                            30 min
  Files: docs/enterprise/audit-trail-architecture.md
         docs/competitive/competitive-positioning.md
  Action: Find every mention of "ZK-STARK" and add qualifier:
    OLD: "ZK-STARK proof"
    NEW: "ZK-STARK API-compatible proof
          (Phase 1: SHA-256 stub with production API.
           Phase 2: production prover via starkware/cairo)"
  Why: A cryptographer reviewing the code will flag the stub immediately.
       Honest qualification is stronger than an unqualified claim.
  Commit: "docs: ZK-STARK claim qualified — Phase1 stub vs Phase2 production"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SPRINT 1 — AFTER LLOYD LOU SIGNED
Gate: Lloyd LoU must be signed. Nothing below starts before this.
Total: ~22 hrs across 3-4 sessions
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

── SESSION S1-A: CLI SURFACE (8 hrs — highest priority) ────────────────────

S1-A-1: cmd/verify-decision                                    3 hrs
  File: cmd/verify-decision/main.go
  Calls: internal/audit/merkle.go GenerateProof()
  Input: decision_id (UUID)
  Output:
    VALID — decision {id} verified against Merkle root {hash} on {date}
    INVALID — proof failed at step {n}
  Compliance claim unlocked: HIPAA §164.528 selective disclosure
  G-FP: go test ./... green before + after
  Commit: "feat: cmd/verify-decision — Merkle selective proof CLI"

S1-A-2: cmd/oscal-generate                                     3 hrs
  File: cmd/oscal-generate/main.go
  Calls: internal/compliance/oscal.go GenerateOSCALBundle()
  Input: --framework [soc2|nist|dora] --from DATE --to DATE --org ORG_ID
  Output: docs/compliance/oscal_evidence_{framework}_{date}.json
  Compliance claim unlocked: SOC2 CC1-CC9 dynamic evidence bundle
  G-FP: go test ./... green before + after
  Commit: "feat: cmd/oscal-generate — SOC2/NIST evidence bundle CLI"

S1-A-3: scripts/generate-evidence-package.sh                   2 hrs
  Wraps: cmd/verify-chain + cmd/verify-decision + cmd/dora-verify
         + cmd/oscal-generate
  Input: ORG_ID DATE_FROM DATE_TO FRAMEWORKS
  Output: ARE_AUDIT_PACKAGE_{org}_{dates}.zip containing:
    1_sequential_integrity/  (verify-chain output)
    2_selective_proofs/      (verify-decision outputs for all blocked decisions)
    3_dora_report/           (dora-verify output)
    4_oscal_bundle/          (oscal-generate output)
  This is the artifact that ends every compliance conversation.
  Commit: "feat: generate-evidence-package.sh — unified auditor ZIP"

── SESSION S1-B: PRODUCT HARDENING (6 hrs) ────────────────────────────────

S1-B-1: Kong/service schema contract test                      2 hrs
  File: tests/integration/kong_contract_test.go
  What: Verifies Kong plugin correctly parses scoring service
        response format. Catches schema drift before runtime.
  Test cases:
    - Standard ALLOW response parsed correctly
    - BLOCKED response with reason object parsed correctly
    - Score band boundaries parsed correctly (800 TRUSTED threshold)
    - Missing fields handled gracefully (fail-open)
  Why: Only protection against silent breakage on schema change.
  G-FP: go test ./... green before + after
  Commit: "test: Kong/service contract test — schema drift prevention"

S1-B-2: Redis single-point-of-failure documentation            1 hr
  File: docs/ops/redis-failover.md (update existing)
  Add: Redis unavailability behavior matrix:
    Score lookup fails    → fail-open, log, pass traffic
    Enforcement mode unknown → default to OBSERVE (safer)
    Used-token cache miss  → accept token (availability wins)
    Baseline cache miss    → fall back to PostgreSQL baseline
  Why: Lloyd's CISO will ask "what happens when Redis goes down?"
       Without a documented answer, the question kills the meeting.
  Commit: "docs: redis-failover — unavailability behavior matrix documented"

S1-B-3: Full-chain integration test                            3 hrs
  File: tests/integration/compliance_chain_test.go
  What: End-to-end test verifying:
    enforcement_decision created
    → hash chain valid
    → Merkle leaf present
    → PQC signature verifiable
    → OSCAL evidence generated
    → cmd/verify-chain passes
    → cmd/verify-decision passes
  Why: No test currently verifies the full compliance chain works
       end-to-end. One broken link breaks every compliance claim.
  G-FP: go test ./... green before + after
  Commit: "test: full compliance chain integration test"

── SESSION S1-C: INTEROPERABILITY (4 hrs) ──────────────────────────────────

S1-C-1: Helm chart (conditional — only if Lloyd confirms Kubernetes)  3 hrs
  File: helm/agentrepengine/
  What: Standard Helm chart for Kubernetes deployment
    values.yaml: scoring service config, Redis config, PG config
    templates/: deployment, service, configmap, secret, ingress
  Install command target: helm install are ./helm/agentrepengine
  4-hour install claim requires this to exist for K8s environments.
  GATE: Only build if Lloyd confirms Kubernetes in PRE-1.
  Commit: "feat: Helm chart — Kubernetes deployment for enterprise pilots"

S1-C-2: SIEM format detection                                  1 hr
  File: internal/audit/siem.go
  What: Add format flag to SIEM webhook:
    SIEM_FORMAT=splunk (default, existing)
    SIEM_FORMAT=sentinel (Microsoft Sentinel CEF format)
    SIEM_FORMAT=qradar (IBM QRadar LEEF format)
  Why: If Lloyd uses Sentinel (common in NWN Microsoft environments),
       current SIEM integration does not work.
  GATE: Ask Lloyd what SIEM they run before building.
        If Splunk → skip this item. If Sentinel → build immediately.
  Commit: "feat: SIEM multi-format — Splunk/Sentinel/QRadar support"

── SESSION S1-D: AUTOMATION (2 hrs) ────────────────────────────────────────

S1-D-1: scripts/update_continuation.sh                        2 hrs
  What: Auto-updates CONTINUATION_PROMPT on session close:
    reads git log --oneline -10
    reads go test ./... output
    updates HEAD hash, test status, recent commits
  Why: Manual CONTINUATION_PROMPT updates have error risk.
       Every session has a chance of state drift.
  Commit: "feat: update_continuation.sh — automated session state capture"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SPRINT 2 — AFTER FIRST PILOT GO-LIVE
Gate: First enterprise live in observe mode. Real traffic flowing.
Total: ~18 hrs
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

S2-1: Welford baseline scalability verification                3 hrs
  File: internal/scoring/baseline.go
  What: Benchmark test at 500, 1000, 2000 agents simultaneously.
        Measure Redis memory at each tier.
        Document max agents before Redis pressure.
  Dependency: Requires real pilot traffic patterns to tune.
  Commit: "test: Welford baseline scalability — 500/1000/2000 agent benchmark"

S2-2: PostgreSQL queue table partitioning                      3 hrs
  File: migrations/004_queue_partitioning.sql
  What: Partition agent_event_queue by week.
        Partition enforcement_decisions by month.
        Auto-archive partitions older than 90 days.
  Why: Unpartitioned tables degrade at scale. Index bloat.
  Dependency: Requires 30+ days of pilot data to size partitions.
  Commit: "feat: PostgreSQL partitioning — queue + decisions by date"

S2-3: Horizontal scaling documentation                        2 hrs
  File: docs/ops/horizontal-scaling.md
  What: Document the session-less scaling model:
    Redis as shared state → any instance serves any request
    Load balancer config (nginx / Kong upstream)
    PostgreSQL connection pooling (PgBouncer config)
    Baseline key format confirms org_id scoping works across instances
  Why: Before any second enterprise pilot, scaling story must be documented.

S2-4: Dead-letter queue                                        3 hrs
  File: internal/store/score_store.go
  What: Failed queue events → dead_letter_queue table
        Alert when dead_letter_queue depth > 0
        Manual replay command for dead-letter events
  Why: Silent event loss is acceptable for Phase 1 (≤500 agents).
       Not acceptable in production with paying enterprise.

S2-5: External query API                                       4 hrs
  File: cmd/scoring-service/main.go (new handler)
  What: GET /agents/{agent_did}/score
        → returns current score, band, last_updated, baseline_sample_count
        Authenticated via SCORING_API_KEY
        Rate-limited per org
  Why: External SOAR platforms and ticketing systems need to query
       agent scores without direct DB access.
       Enables integrations that Lloyd's security team will request.

S2-6: SPHINCS+ production binding                             4 hrs
  File: internal/audit/pqc_signer.go
  What: Replace HMAC-SHA256 stub with
        github.com/cloudflare/circl/sign/sphincsplus
        Real SPHINCS+-SHA2-256s-simple key generation and signing
  Why: Current implementation is API-compatible but not production crypto.
       Before any post-quantum security claim is made to a financial
       institution, the real library must be bound.
  Dependency: Requires circl library import and key format change.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
MASTER EXECUTION SEQUENCE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

BEFORE LLOYD (next 3 weeks):
  PRE-1  Ask Lloyd: Kubernetes or Docker?               15 min  COMMERCIAL
  PRE-2  BUILD_INTELLIGENCE file map update              1 hr   DOC
  PRE-3  Raft single-node fallback doc                  30 min  DOC
  PRE-4  ZK-STARK claim qualification                   30 min  DOC

SPRINT 1 (after LoU — ~4 weeks):
  S1-A-1 cmd/verify-decision                             3 hrs  BUILD ★
  S1-A-2 cmd/oscal-generate                              3 hrs  BUILD ★
  S1-A-3 generate-evidence-package.sh                    2 hrs  BUILD ★
  S1-B-1 Kong contract test                              2 hrs  TEST
  S1-B-2 Redis failover behavior matrix                  1 hr   DOC
  S1-B-3 Full compliance chain integration test          3 hrs  TEST
  S1-C-1 Helm chart (if K8s confirmed)                  3 hrs  BUILD
  S1-C-2 SIEM format detection (if non-Splunk)           1 hr   BUILD
  S1-D-1 update_continuation.sh automation               2 hrs  AUTO

SPRINT 2 (after pilot go-live — ~8 weeks):
  S2-1   Welford scalability benchmark                   3 hrs
  S2-2   PostgreSQL partitioning                         3 hrs
  S2-3   Horizontal scaling documentation                2 hrs
  S2-4   Dead-letter queue                               3 hrs
  S2-5   External query API                              4 hrs
  S2-6   SPHINCS+ production binding                     4 hrs

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
GATES — NON-NEGOTIABLE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

G-FP     ★  go test ./... green before AND after every commit
G-KILL   ★  Phase alignment confirmed before every build task
G-LLOYD  ★  Nothing in Sprint 1 starts before LoU signed
G-PILOT  ★  Nothing in Sprint 2 starts before pilot go-live

ANTI-SCOPE — never appear in this roadmap:
  ❌ Isolation Forest (before 90 days production data)
  ❌ Federation (before 3+ paying enterprises)
  ❌ Management UI (before pilot tells us what CISO needs)
  ❌ DID/ledger identity
  ❌ Kafka

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
THE PRODUCT AFTER ALL ITEMS COMPLETE

"AgentRepEngine is the only AI agent enforcement layer that:

  1. Enforces behavioral trust at the gateway — agents cannot see it
     or route around it. Zero code changes required.

  2. Produces a self-verifiable, quantum-resistant compliance evidence
     package covering DORA, SOC2, HIPAA, SEC, NIST AI RMF, GDPR,
     and EU AI Act from a single enforcement event stream.

  3. Scales from single-node Docker to multi-node Kubernetes with
     no architecture change — session-less design, Redis shared state.

  4. Integrates with any SIEM (Splunk, Sentinel, QRadar) and any
     agent framework (LangChain, LlamaIndex, custom) out of the box.

  5. Self-improves — every override trains the threshold engine,
     every pilot produces a published FP rate below the Visa standard.

No competitor has items 2, 3, 4, and 5 simultaneously.
Item 2 alone is worth the enterprise conversation."

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ARE GAP CLOSURE ROADMAP v1.0
Source: Coordination, orchestration, interoperability & scalability audit
April 7, 2026 | Commit to: docs/ops/ARE_GAP_CLOSURE_ROADMAP.md
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
