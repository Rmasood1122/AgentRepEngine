package scoring

import (
	"encoding/json"
	"testing"
)

// ═══════════════════════════════════════════════════
// G-EXPLAIN GATE TESTS — HARD STOP IF ANY FAIL
// ═══════════════════════════════════════════════════

// TestExplainDecisionComplete verifies full reason object is generated.
func TestExplainDecisionComplete(t *testing.T) {
	vector := FeatureVector{
		PIIFieldAccessRate:        600,
		BulkAccessCountPerSession: 8000,
		ToolCallRatePerHour:       200,
	}

	engine, err := NewPolicyEngine("../../config/policy_packs")
	if err != nil {
		t.Fatalf("load policies: %v", err)
	}

	violations := engine.Evaluate(vector)
	baselines := map[string]Baseline{
		"pii_field_access_rate":         {Mean: 20, StdDev: 25},
		"bulk_access_count_per_session": {Mean: 400, StdDev: 400},
		"tool_call_rate_per_hour":       {Mean: 80, StdDev: 60},
	}

	reason, err := ExplainDecision(
		"did:jwt:test-org:finance-agent:001",
		"BLOCKED",
		187, 743,
		vector, violations, baselines,
		6.0,
	)
	if err != nil {
		t.Fatalf("★ G-EXPLAIN HARD STOP: ExplainDecision failed: %v", err)
	}

	// Verify all required fields
	if reason.Decision == "" {
		t.Error("decision is empty")
	}
	if reason.AgentDID == "" {
		t.Error("agent_did is empty")
	}
	if reason.Score == 0 {
		t.Error("score is zero")
	}
	if reason.PolicyFired == "" {
		t.Error("policy_fired is empty")
	}
	if reason.RecommendedAction == "" {
		t.Error("recommended_action is empty")
	}
	if len(reason.TriggerEvents) == 0 {
		t.Error("trigger_events is empty")
	}
	if reason.ComputedAt == 0 {
		t.Error("computed_at is zero")
	}

	t.Logf("✅ Decision: %s", reason.Decision)
	t.Logf("✅ Agent: %s", reason.AgentDID)
	t.Logf("✅ Score: %d (delta: %d)", reason.Score, reason.ScoreDelta)
	t.Logf("✅ Policy fired: %s", reason.PolicyFired)
	t.Logf("✅ Trigger events: %d", len(reason.TriggerEvents))
	t.Logf("✅ Recommended action: %s", reason.RecommendedAction)
}

// TestExplainRejectsEmptyAgentDID verifies FM2 prevention.
// Missing agent_did must return error — not produce incomplete object.
func TestExplainRejectsInvalidInput(t *testing.T) {
	vector := FeatureVector{}
	baselines := map[string]Baseline{}

	_, err := ExplainDecision("", "BLOCKED", 100, 700,
		vector, nil, baselines, 1.0)
	if err == nil {
		t.Error("expected error for empty agent_did, got nil")
	} else {
		t.Logf("✅ Correctly rejected empty agent_did: %v", err)
	}

	_, err = ExplainDecision("did:jwt:org:agent:001", "", 100, 700,
		vector, nil, baselines, 1.0)
	if err == nil {
		t.Error("expected error for empty decision, got nil")
	} else {
		t.Logf("✅ Correctly rejected empty decision: %v", err)
	}
}

// TestExplainJSONSerializes verifies reason object produces valid JSON.
// G-EXPLAIN requires reason object in SIEM webhook and audit log.
func TestExplainJSONSerializes(t *testing.T) {
	engine, _ := NewPolicyEngine("../../config/policy_packs")
	vector := FeatureVector{PIIFieldAccessRate: 600}
	violations := engine.Evaluate(vector)
	baselines := map[string]Baseline{
		"pii_field_access_rate": {Mean: 20, StdDev: 25},
	}

	reason, err := ExplainDecision(
		"did:jwt:org:agent:001", "BLOCKED",
		150, 700, vector, violations, baselines, 2.0,
	)
	if err != nil {
		t.Fatalf("ExplainDecision failed: %v", err)
	}

	jsonStr := reason.ToJSON()
	if jsonStr == "" {
		t.Error("ToJSON returned empty string")
	}

	// Verify it parses as valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Errorf("reason object is not valid JSON: %v", err)
	}

	// Verify required top-level keys
	required := []string{
		"decision", "agent_did", "score", "score_delta",
		"policy_fired", "recommended_action", "computed_at",
	}
	for _, key := range required {
		if _, ok := parsed[key]; !ok {
			t.Errorf("missing required key in JSON: %s", key)
		}
	}

	t.Logf("✅ Valid JSON produced: %d bytes", len(jsonStr))
	t.Logf("✅ All required keys present")
}

// TestExplainCleanDecision verifies reason object on ALLOW decision.
// ALLOW decisions also need reason objects — not just blocks.
func TestExplainCleanDecision(t *testing.T) {
	vector := FeatureVector{
		ToolCallRatePerHour: 60,
		PIIFieldAccessRate:  8,
	}
	baselines := map[string]Baseline{
		"tool_call_rate_per_hour": {Mean: 80, StdDev: 60},
		"pii_field_access_rate":   {Mean: 20, StdDev: 25},
	}

	reason, err := ExplainDecision(
		"did:jwt:org:clean-agent:001", "ALLOW",
		750, 700, vector, nil, baselines, 1.0,
	)
	if err != nil {
		t.Fatalf("ExplainDecision failed for ALLOW: %v", err)
	}

	if reason.Decision != "ALLOW" {
		t.Errorf("expected ALLOW, got %s", reason.Decision)
	}
	if len(reason.TriggerEvents) != 0 {
		t.Errorf("clean agent should have 0 trigger events, got %d",
			len(reason.TriggerEvents))
	}

	t.Logf("✅ ALLOW decision explained: score=%d policy=%s",
		reason.Score, reason.PolicyFired)
}
