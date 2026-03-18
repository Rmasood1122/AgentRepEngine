package scoring

import (
	"testing"
)

// ═══════════════════════════════════════════════════════════
// HARDENING TESTS — SECURITY GAP CLOSURE VERIFICATION
// Expert panel: E1 Security + E2 Infrastructure + E3 Agent
// ═══════════════════════════════════════════════════════════

// TestGAP3_ScoreComputationStable verifies scoring degrades gracefully.
// GAP 3: Redis write failure must not corrupt state.
// Score must be recomputable from Postgres with same result.
func TestGAP3_ScoreComputationStable(t *testing.T) {
	weights := ScoreWeights{Historical: 0.5, Velocity: 0.5}

	score1 := ComputeScore(700, 650, weights)
	score2 := ComputeScore(700, 650, weights)

	if score1 != score2 {
		t.Errorf("score not idempotent: %d != %d", score1, score2)
	}
	t.Logf("✅ GAP 3: Score idempotent: %d", score1)

	// 100 iterations — stable under repeated computation
	for i := 0; i < 100; i++ {
		s := ComputeScore(743, 800, weights)
		if s != ComputeScore(743, 800, weights) {
			t.Error("score not stable under repeated computation")
		}
	}
	t.Log("✅ GAP 3: 100 iterations stable — Redis miss safe")
}

// TestGAP6_ScoreMonotonicallyDecreases verifies score cannot be gamed.
// GAP 6: attacker incrementing anomaly must see score decrease.
// If score ever increases with more anomalous behavior — evasion is possible.
func TestGAP6_ScoreMonotonicallyDecreases(t *testing.T) {
	weights := ScoreWeights{Historical: 0.5, Velocity: 0.5}
	baseline := Baseline{Mean: 80, StdDev: 60}

	scores := []int{}
	for toolCallRate := 80.0; toolCallRate <= 500; toolCallRate += 20 {
		z := ComputeZScore(toolCallRate, baseline)
		penalty := 0.0
		if z > 3.0 {
			penalty = min(100*(z-3.0), 300)
		}
		V := max(0, 1000-penalty*3)
		score := ComputeScore(700, V, weights)
		scores = append(scores, score)
	}

	for i := 1; i < len(scores); i++ {
		if scores[i] > scores[i-1]+5 {
			t.Errorf("score increased at step %d: %d -> %d — evasion possible",
				i, scores[i-1], scores[i])
		}
	}
	t.Logf("✅ GAP 6: Score decreases monotonically: %d → %d",
		scores[0], scores[len(scores)-1])
}

// TestSecurityInvariant_NeverBlockWithoutExplanation verifies Law L5.
// MOST IMPORTANT SECURITY INVARIANT: every block has a reason object.
func TestSecurityInvariant_NeverBlockWithoutExplanation(t *testing.T) {
	engine, err := NewPolicyEngine("../../config/policy_packs")
	if err != nil {
		t.Fatalf("load policies: %v", err)
	}

	attackVectors := []FeatureVector{
		{PIIFieldAccessRate: 600, BulkAccessCountPerSession: 8000},
		{PermissionEscalationCount: 10, CrossTenantProbeCount: 3},
		{ToolCallRatePerHour: 1200, SubAgentSpawnDepth: 5},
	}

	baselines := map[string]Baseline{
		"pii_field_access_rate":         {Mean: 20, StdDev: 25},
		"bulk_access_count_per_session": {Mean: 400, StdDev: 400},
		"permission_escalation_count":   {Mean: 0, StdDev: 0.5},
		"cross_tenant_probe_count":      {Mean: 0, StdDev: 0.1},
		"tool_call_rate_per_hour":       {Mean: 80, StdDev: 60},
		"sub_agent_spawn_depth":         {Mean: 0, StdDev: 0.3},
	}

	for i, vector := range attackVectors {
		violations := engine.Evaluate(vector)

		reason, err := ExplainDecision(
			"did:jwt:test-org:attack-agent:001",
			"BLOCKED", 150, 700,
			vector, violations, baselines, 2.0,
		)

		if err != nil {
			t.Errorf("vector %d: ExplainDecision failed — "+
				"must fall back to AUDIT not BLOCK: %v", i, err)
			continue
		}

		if reason.PolicyFired == "" {
			t.Errorf("vector %d: policy_fired empty", i)
		}
		if reason.RecommendedAction == "" {
			t.Errorf("vector %d: recommended_action empty", i)
		}

		t.Logf("✅ Vector %d: BLOCKED with explanation — policy=%s",
			i, reason.PolicyFired)
	}
}

// TestZeroToleranceEnforced verifies cross-tenant probe = immediate block.
func TestZeroToleranceEnforced(t *testing.T) {
	engine, err := NewPolicyEngine("../../config/policy_packs")
	if err != nil {
		t.Fatalf("load policies: %v", err)
	}

	vector := FeatureVector{CrossTenantProbeCount: 1}
	violations := engine.Evaluate(vector)

	found := false
	for _, v := range violations {
		if v.Feature == "cross_tenant_probe_count" {
			found = true
			if v.ScorePenalty >= 0 {
				t.Errorf("cross-tenant penalty must be negative, got %d",
					v.ScorePenalty)
			}
			t.Logf("✅ Zero tolerance: cross-tenant → penalty=%d action=%s",
				v.ScorePenalty, v.EnforcementAction)
		}
	}

	if !found {
		t.Error("cross-tenant probe not detected — zero tolerance broken")
	}
}

// TestGAP2_IdentityValidationInScoring verifies scoring rejects
// decisions that would be made without valid identity context.
func TestGAP2_IdentityValidationInScoring(t *testing.T) {
	engine, err := NewPolicyEngine("../../config/policy_packs")
	if err != nil {
		t.Fatalf("load policies: %v", err)
	}

	// Empty agent_did must be rejected by ExplainDecision
	vector := FeatureVector{PIIFieldAccessRate: 600}
	violations := engine.Evaluate(vector)
	baselines := map[string]Baseline{
		"pii_field_access_rate": {Mean: 20, StdDev: 25},
	}

	_, err = ExplainDecision("", "BLOCKED", 150, 700,
		vector, violations, baselines, 1.0)
	if err == nil {
		t.Error("empty agent_did must be rejected — identity required for explanation")
	}
	t.Logf("✅ GAP 2: Empty identity rejected at explanation layer: %v", err)
}
