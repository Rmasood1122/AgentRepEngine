package scoring

import (
	"testing"
)

func TestPolicyEngineLoads(t *testing.T) {
	engine, err := NewPolicyEngine("../../config/policy_packs")
	if err != nil {
		t.Fatalf("failed to load policy packs: %v", err)
	}
	if len(engine.packs) != 7 {
		t.Errorf("expected 7 policy packs, got %d", len(engine.packs))
	}
	for _, p := range engine.packs {
		t.Logf("✅ Loaded: %s (%s)", p.Name, p.OWASPRef)
	}
}

func TestPolicyFiresOnAttack(t *testing.T) {
	engine, err := NewPolicyEngine("../../config/policy_packs")
	if err != nil {
		t.Fatalf("load policies: %v", err)
	}

	// Bulk PII attack vector
	attackVector := FeatureVector{
		PIIFieldAccessRate:        600,
		BulkAccessCountPerSession: 8000,
		ToolCallRatePerHour:       200,
	}

	violations := engine.Evaluate(attackVector)
	if len(violations) == 0 {
		t.Fatal("expected policy violations on attack vector, got none")
	}

	worst := WorstViolation(violations)
	if worst == nil {
		t.Fatal("WorstViolation returned nil")
	}

	t.Logf("✅ %d violations detected", len(violations))
	t.Logf("✅ Worst: policy=%s feature=%s level=%s penalty=%d",
		worst.PolicyName, worst.Feature, worst.Level, worst.ScorePenalty)
}

func TestPolicyCleanOnLegitimate(t *testing.T) {
	engine, err := NewPolicyEngine("../../config/policy_packs")
	if err != nil {
		t.Fatalf("load policies: %v", err)
	}

	// Normal data analyst vector
	legitVector := FeatureVector{
		ToolCallRatePerHour:       60,
		UniqueEndpointsPerHour:    12,
		BulkAccessCountPerSession: 200,
		PIIFieldAccessRate:        8,
	}

	violations := engine.Evaluate(legitVector)
	blockViolations := 0
	for _, v := range violations {
		if v.Level == "block" {
			blockViolations++
		}
	}

	if blockViolations > 0 {
		t.Errorf("legitimate agent triggered %d block violations", blockViolations)
	}
	t.Logf("✅ Legitimate agent: %d violations (0 blocks)", len(violations))
}

func TestVarianceGrowthConstants(t *testing.T) {
	if VarianceWindowDays != 7 {
		t.Errorf("VarianceWindowDays = %d, want 7", VarianceWindowDays)
	}
	if VarianceGrowthThreshold != 2.0 {
		t.Errorf("VarianceGrowthThreshold = %f, want 2.0", VarianceGrowthThreshold)
	}
	t.Logf("✅ Variance growth constants: window=%d days, threshold=%.1fx",
		VarianceWindowDays, VarianceGrowthThreshold)
}

func TestVarianceGrowthResultHighRisk(t *testing.T) {
	// Simulate a feature where variance doubled (slow-walk indicator)
	result := VarianceGrowthResult{
		Feature:      "tool_call_rate_per_hour",
		PrevVariance: 100.0,
		CurrVariance: 250.0,
		GrowthRate:   2.5,
		HighRisk:     2.5 >= VarianceGrowthThreshold,
	}
	if !result.HighRisk {
		t.Error("expected HIGH_RISK for 2.5x variance growth")
	}
	t.Logf("✅ Variance growth 2.5x → HIGH_RISK=true for %s", result.Feature)

	// Simulate normal variance — should not trigger
	normal := VarianceGrowthResult{
		Feature:      "unique_endpoints_per_hour",
		PrevVariance: 100.0,
		CurrVariance: 130.0,
		GrowthRate:   1.3,
		HighRisk:     1.3 >= VarianceGrowthThreshold,
	}
	if normal.HighRisk {
		t.Error("expected no HIGH_RISK for 1.3x variance growth")
	}
	t.Logf("✅ Variance growth 1.3x → HIGH_RISK=false for %s", normal.Feature)
}

func TestZeroTolerancePolicies(t *testing.T) {
	engine, err := NewPolicyEngine("../../config/policy_packs")
	if err != nil {
		t.Fatalf("load policies: %v", err)
	}

	// Any cross-tenant probe = immediate block
	crossTenantVector := FeatureVector{
		CrossTenantProbeCount: 1,
	}

	violations := engine.Evaluate(crossTenantVector)
	found := false
	for _, v := range violations {
		if v.Feature == "cross_tenant_probe_count" {
			found = true
			t.Logf("✅ Cross-tenant probe blocked: level=%s penalty=%d",
				v.Level, v.ScorePenalty)
		}
	}
	if !found {
		t.Error("cross-tenant probe not detected — zero tolerance policy not firing")
	}
}
