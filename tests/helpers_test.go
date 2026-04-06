// tests/helpers_test.go
// Shared helpers for the ARE 11x test suite
// This file resolves import and fmt dependencies across layers
package tests

import "fmt"

// fmt shim — ensures fmt is used across all test files in package
var _ = fmt.Sprintf

// ─────────────────────────────────────────────────────────────────────────────
// ARE 11x TEST SUITE — MASTER EXECUTION PLAN
// ─────────────────────────────────────────────────────────────────────────────
//
// PASTE INTO CLAUDE CODE (git bash on windows):
//
// STEP 1: Copy test files into your repo
//   cp layer1_unit_test.go          /path/to/AgentRepEngine/tests/layer1_unit_test.go
//   cp layer2_integration_test.go   /path/to/AgentRepEngine/tests/layer2_integration_test.go
//   cp layer3_adversarial_test.go   /path/to/AgentRepEngine/tests/layer3_adversarial_test.go
//   cp layer4_fp_corpus_test.go     /path/to/AgentRepEngine/tests/layer4_fp_corpus_test.go
//   cp layer5_to_9_test.go          /path/to/AgentRepEngine/tests/layer5_to_9_test.go
//   cp helpers_test.go              /path/to/AgentRepEngine/tests/helpers_test.go
//
// STEP 2: Update package declarations to match your existing test package
//   grep -r "^package" tests/ | head -5
//   # Update all new files to match (likely: package tests or package are_test)
//
// STEP 3: Fix import paths for your actual packages
//   # These tests use inline implementations for portability
//   # For integration tests, replace inline scoring with actual imports:
//   # import "github.com/Rehanrana11/AgentRepEngine/internal/scoring"
//
// STEP 4: Run all layers
//   # Layer 1 only (no docker needed)
//   go test ./tests/... -run TestLayer1 -v
//
//   # Layer 3 only (adversarial — no docker needed)
//   go test ./tests/... -run TestLayer3 -v
//
//   # Layer 4 only (FP corpus — no docker needed)
//   go test ./tests/... -run TestLayer4 -v
//
//   # Layer 2 + 5 (requires docker stack)
//   docker compose up -d && sleep 20
//   go test ./tests/... -run "TestLayer2|TestLayer5" -v -tags integration
//
//   # Full 9-layer suite (unit only, no docker)
//   go test ./tests/... -run "TestLayer1|TestLayer3|TestLayer4|TestLayer7|TestLayer8|TestLayer9|TestCompetitor" -v
//
//   # Full suite with integration
//   go test ./tests/... -v
//
//   # With coverage report
//   go test ./tests/... -v -coverprofile=coverage.out
//   go tool cover -html=coverage.out -o coverage.html
//
// STEP 5: Generate test report for Lloyd
//   go test ./tests/... -v -json 2>&1 | python3 scripts/test_report.py
//
// ─────────────────────────────────────────────────────────────────────────────
// EXPECTED OUTPUT WHEN ALL TESTS PASS:
// ─────────────────────────────────────────────────────────────────────────────
//
// Layer 1 — UNIT (35 tests):
//   ✅ Score formula: nominal, clamp floor/ceiling, weights, new agent default
//   ✅ Band assignment: all boundary conditions (700, 699.9, 500, 499, 200, 199, 0)
//   ✅ H decay: 7-day half-life, zero days, never below zero
//   ✅ Policy: cross-tenant -300, PII boundary, escalation threshold, stacking
//   ✅ Z-score: below threshold, exact threshold, high anomaly cap, zero stddev
//   ✅ Hash chain: 50-event valid, tampered score, tampered band, severed link, deletion
//
// Layer 2 — INTEGRATION (14 tests, requires docker):
//   ✅ Health: service, redis, postgres, queue_depth, hash_chain
//   ✅ JWT: valid token scores, JWKS endpoint, /verify endpoint, missing key rejected
//   ✅ Score: structured response, band consistency, reason object fields
//   ✅ Events: submission accepted, score updates after events (async)
//   ✅ Fail-open: health fast (<200ms), score fast (<100ms)
//   ✅ SIEM: blocked metric correct (BUG 2 fix verified)
//
// Layer 3 — ADVERSARIAL (14 scenarios):
//   ✅ Slow-walk: baseline poisoning detected by HIGH_RISK VERIFY
//   ✅ Slow-walk: gradual permission escalation caught
//   ✅ Burst: 10x normal → BLOCKED, 3x normal → not TRUSTED
//   ✅ Burst: high-variance agent still anomalous at 3x
//   ✅ Cross-tenant: single probe forces trusted agent out of TRUSTED band
//   ✅ Multi-agent: Phase 1 scope documented, peer cluster analysis verified
//   ✅ Sub-agent spawn: depth > 3 triggers HIGH_RISK
//   ✅ Baseline poisoning: absolute policy threshold defeats evasion
//   ✅ Stacked attack: all vectors → BLOCKED + HIGH_RISK + near-zero score
//   ✅ Recovery: clean agent can return to TRUSTED (no permanent penalty)
//
// Layer 4 — FP CORPUS (8 tests):
//   ✅ All 7 archetypes: 0.00% FP rate
//   ✅ All archetypes: zero false HIGH_RISK alerts
//   ✅ Batch agent month-end spike: not blocked
//   ✅ New agent cold-start: TRUSTED default
//   ✅ PII rate at exact threshold: not flagged
//   ✅ Bias audit: 0.00% FP across all archetypes
//
// Layer 5 — ENFORCEMENT (6 tests):
//   ✅ Default observe mode
//   ✅ Auto-rollback at FP > 2%
//   ✅ No rollback at FP = 2.0% exactly
//   ✅ No rollback in observe mode
//   ✅ Enforce gate activation sequence
//   ✅ SIEM callback on rollback
//
// Layer 6 — COMPLIANCE (6 tests):
//   ✅ DORA Article 22: immutable audit trail
//   ✅ DORA: timestamped incident records
//   ✅ HIPAA: PII breach attribution
//   ✅ SOX Section 404: permission violations logged
//   ✅ 200-event audit chain: auditor-ready
//   ✅ Tamper detection: auditor scenario
//
// Layer 7 — PERFORMANCE (3 tests):
//   ✅ Score calculation: sub-microsecond
//   ✅ 100-event chain: under 1 second
//   ✅ 1000 concurrent agents: under 10ms
//
// Layer 8 — SDK (4 tests):
//   ✅ DID format: valid formats accepted
//   ✅ DID format: invalid formats rejected
//   ✅ LangChain payload structure
//   ✅ Framework-agnostic scoring (LangChain = LlamaIndex)
//
// Layer 9 — DEMO INTEGRITY (10 tests):
//   ✅ C1: gateway enforcement
//   ✅ C4: confidence score populated
//   ✅ C6: fail-open behavior
//   ✅ C8: observe mode default
//   ✅ C11: human-readable reason object
//   ✅ All claims pass simultaneously (Lloyd demo gate)
//   ✅ Regression: score formula unchanged post-dashboard commits
//   ✅ Regression: FP rate 0.00% anchored
//   ✅ Regression: TP rate >= 80% anchored
//   ✅ Competitor gaps: Gen Digital, Lakera, Protect AI, Validia
//
// TOTAL: ~100 tests across 9 layers
// TARGET: 0 failures, 0 skips (unit tests) before Lloyd meeting
//
// ─────────────────────────────────────────────────────────────────────────────
// KNOWN INTEGRATION TEST SKIP CONDITIONS (not failures):
// ─────────────────────────────────────────────────────────────────────────────
//   TestLayer2_JWT_ValidTokenScores: Set ARE_TEST_JWT env var
//   TestLayer2_SIEM_BlockedDecisionTriggersAlert: Set ARE_SIEM_TEST_URL
//   TestLayer2_InsertOnly_*: Run manually or set PG env vars
//
// ─────────────────────────────────────────────────────────────────────────────
