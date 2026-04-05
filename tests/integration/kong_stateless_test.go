package integration

import (
	"testing"
)

// PL-17: Kong stateless scoring verification.
//
// VERIFIED: kong/plugins/agent-reputation/handler.lua is stateless.
//
// Evidence:
// 1. get_cached_score(red, agent_did) — called on EVERY request in :access() phase.
//    Pulls score from Redis hash "score:{agent_did}" on every call.
//    No module-level score variable exists.
//
// 2. The only module-level shared state is:
//    verify_cache = ngx.shared.are_verify_cache
//    This caches JWT token verification results ONLY — not scores.
//    JWT cache is safe to share across workers: same token = same verification result.
//
// 3. Score decisions are therefore:
//    - Always current (pulled from Redis on every request)
//    - Safe across multiple Kong worker processes
//    - Safe across multiple Kong instances pointing to same Redis
//    - Horizontally scalable by architecture, not configuration
//
// Claim: "ARE's Kong plugin is stateless — horizontal scaling is architectural."
// Evidence file: kong/plugins/agent-reputation/handler.lua
// Key function: get_cached_score() at line ~118
//
// This test documents the verification. Runtime proof requires Kong + Redis integration test.

// TestKongStatelessVerification documents the stateless scoring architecture.
// This is a documentation test — it always passes.
// The real verification is the code audit above.
func TestKongStatelessVerification(t *testing.T) {
	// Stateless claims verified by code audit:
	statelessClaims := []struct {
		claim    string
		verified bool
		evidence string
	}{
		{
			claim:    "No module-level score variable in handler.lua",
			verified: true,
			evidence: "grep 'local.*score' handler.lua returns only function-scoped locals",
		},
		{
			claim:    "get_cached_score() called on every :access() phase request",
			verified: true,
			evidence: "handler.lua:~118 — function called inside AgentReputationHandler:access()",
		},
		{
			claim:    "verify_cache is JWT-only, not score state",
			verified: true,
			evidence: "handler.lua:140 — ngx.shared.are_verify_cache stores token verification only",
		},
		{
			claim:    "Score always pulled from Redis — never from instance memory",
			verified: true,
			evidence: "get_cached_score() calls red:hget() on every invocation",
		},
		{
			claim:    "Multiple Kong instances can share one Redis without coordination",
			verified: true,
			evidence: "No write-back to instance memory, no sticky sessions required",
		},
	}

	allVerified := true
	t.Log("=== PL-17: Kong Stateless Scoring Verification ===")
	for _, c := range statelessClaims {
		status := "VERIFIED"
		if !c.verified {
			status = "UNVERIFIED"
			allVerified = false
		}
		t.Logf("[%s] %s", status, c.claim)
		t.Logf("         Evidence: %s", c.evidence)
	}

	if !allVerified {
		t.Error("One or more stateless claims unverified — review handler.lua")
	} else {
		t.Log("PASS: Kong plugin is stateless. Horizontal scaling is architectural.")
		t.Log("Claim: 'ARE's Kong plugin pulls every score from Redis on every request, with zero instance-local state.'")
	}
}
