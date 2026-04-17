package scoring

import (
	"os"
	"path/filepath"
	"testing"
)

// ═══════════════════════════════════════════════════════════════
// A1 — YAML POLICY SCHEMA VALIDATION ON STARTUP
// Hardening Sprint Phase A — Defensive Infrastructure
// Risk: ZERO — startup only, no runtime change
// ═══════════════════════════════════════════════════════════════

func TestPolicyValidation_MissingName(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, dir, "bad.yaml", `
version: "1.0.0"
owasp_ref: LLM01
thresholds:
  tool_call_rate_per_hour:
    warning: 5.0
    throttle: 10.0
    block: 15.0
score_penalties:
  warning: -25
  throttle: -75
  block: -200
enforcement_action: RATE_LIMIT
escalation_action: BLOCK
`)
	_, err := NewPolicyEngine(dir)
	if err == nil {
		t.Fatal("expected error for policy missing name, got nil")
	}
	t.Logf("✅ A1: Missing name correctly rejected: %v", err)
}

func TestPolicyValidation_MissingVersion(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, dir, "bad.yaml", `
name: test_policy
owasp_ref: LLM01
thresholds:
  tool_call_rate_per_hour:
    warning: 5.0
score_penalties:
  warning: -25
  throttle: -75
  block: -200
enforcement_action: RATE_LIMIT
escalation_action: BLOCK
`)
	_, err := NewPolicyEngine(dir)
	if err == nil {
		t.Fatal("expected error for policy missing version, got nil")
	}
	t.Logf("✅ A1: Missing version correctly rejected: %v", err)
}

func TestPolicyValidation_MissingEnforcementAction(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, dir, "bad.yaml", `
name: test_policy
version: "1.0.0"
thresholds:
  tool_call_rate_per_hour:
    warning: 5.0
score_penalties:
  warning: -25
  throttle: -75
  block: -200
escalation_action: BLOCK
`)
	_, err := NewPolicyEngine(dir)
	if err == nil {
		t.Fatal("expected error for policy missing enforcement_action, got nil")
	}
	t.Logf("✅ A1: Missing enforcement_action correctly rejected: %v", err)
}

func TestPolicyValidation_MissingEscalationAction(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, dir, "bad.yaml", `
name: test_policy
version: "1.0.0"
thresholds:
  tool_call_rate_per_hour:
    warning: 5.0
score_penalties:
  warning: -25
  throttle: -75
  block: -200
enforcement_action: RATE_LIMIT
`)
	_, err := NewPolicyEngine(dir)
	if err == nil {
		t.Fatal("expected error for policy missing escalation_action, got nil")
	}
	t.Logf("✅ A1: Missing escalation_action correctly rejected: %v", err)
}

func TestPolicyValidation_NoThresholds(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, dir, "bad.yaml", `
name: test_policy
version: "1.0.0"
score_penalties:
  warning: -25
  throttle: -75
  block: -200
enforcement_action: RATE_LIMIT
escalation_action: BLOCK
`)
	_, err := NewPolicyEngine(dir)
	if err == nil {
		t.Fatal("expected error for policy with no thresholds, got nil")
	}
	t.Logf("✅ A1: No thresholds correctly rejected: %v", err)
}

func TestPolicyValidation_NegativeThreshold(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, dir, "bad.yaml", `
name: test_policy
version: "1.0.0"
thresholds:
  tool_call_rate_per_hour:
    warning: -5.0
    throttle: 10.0
    block: 15.0
score_penalties:
  warning: -25
  throttle: -75
  block: -200
enforcement_action: RATE_LIMIT
escalation_action: BLOCK
`)
	_, err := NewPolicyEngine(dir)
	if err == nil {
		t.Fatal("expected error for negative threshold, got nil")
	}
	t.Logf("✅ A1: Negative threshold correctly rejected: %v", err)
}

func TestPolicyValidation_ThresholdOrderInverted(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, dir, "bad.yaml", `
name: test_policy
version: "1.0.0"
thresholds:
  tool_call_rate_per_hour:
    warning: 15.0
    throttle: 10.0
    block: 5.0
score_penalties:
  warning: -25
  throttle: -75
  block: -200
enforcement_action: RATE_LIMIT
escalation_action: BLOCK
`)
	_, err := NewPolicyEngine(dir)
	if err == nil {
		t.Fatal("expected error for inverted thresholds (warning > throttle > block), got nil")
	}
	t.Logf("✅ A1: Inverted thresholds correctly rejected: %v", err)
}

func TestPolicyValidation_InvalidYAMLSyntax(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, dir, "bad.yaml", `
name: test_policy
version: "1.0.0"
thresholds:
  tool_call_rate_per_hour:
    warning: [[[invalid yaml
`)
	_, err := NewPolicyEngine(dir)
	if err == nil {
		t.Fatal("expected error for invalid YAML syntax, got nil")
	}
	t.Logf("✅ A1: Invalid YAML syntax correctly rejected: %v", err)
}

func TestPolicyValidation_ValidPackPasses(t *testing.T) {
	dir := t.TempDir()
	writeTestYAML(t, dir, "good.yaml", `
name: test_valid_policy
version: "1.0.0"
owasp_ref: LLM04
description: "Test policy for validation"
thresholds:
  tool_call_rate_per_hour:
    warning: 5.0
    throttle: 10.0
    block: 15.0
score_penalties:
  warning: -25
  throttle: -75
  block: -200
enforcement_action: RATE_LIMIT
escalation_action: BLOCK
`)
	engine, err := NewPolicyEngine(dir)
	if err != nil {
		t.Fatalf("valid policy pack should load: %v", err)
	}
	if len(engine.packs) != 1 {
		t.Errorf("expected 1 pack, got %d", len(engine.packs))
	}
	t.Log("✅ A1: Valid policy pack loads successfully")
}

func TestPolicyValidation_ExistingPacksAllValid(t *testing.T) {
	engine, err := NewPolicyEngine("../../config/policy_packs")
	if err != nil {
		t.Fatalf("existing policy packs must pass validation: %v", err)
	}
	t.Logf("✅ A1: All %d existing policy packs pass schema validation", len(engine.packs))
}

func TestScoringConfigValidation_NegativeDecayRate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad_config.yaml")
	writeTestYAML(t, dir, "bad_config.yaml", `
weights:
  historical: 0.5
  velocity: 0.5
decay_rate: -0.1
zscore_threshold: 3.0
penalty_per_sigma: 100.0
max_penalty: 300.0
min_std_dev: 0.1
bootstrap_sample_threshold: 100
`)
	_, err := LoadScoringConfig(path)
	if err == nil {
		t.Fatal("expected error for negative decay_rate, got nil")
	}
	t.Logf("✅ A1: Negative decay_rate correctly rejected: %v", err)
}

func TestScoringConfigValidation_ZeroZScoreThreshold(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad_config.yaml")
	writeTestYAML(t, dir, "bad_config.yaml", `
weights:
  historical: 0.5
  velocity: 0.5
decay_rate: 0.1
zscore_threshold: 0.0
penalty_per_sigma: 100.0
max_penalty: 300.0
min_std_dev: 0.1
bootstrap_sample_threshold: 100
`)
	_, err := LoadScoringConfig(path)
	if err == nil {
		t.Fatal("expected error for zero zscore_threshold, got nil")
	}
	t.Logf("✅ A1: Zero zscore_threshold correctly rejected: %v", err)
}

func TestScoringConfigValidation_ExistingConfigValid(t *testing.T) {
	cfg, err := LoadScoringConfig("../../config/scoring_weights.yaml")
	if err != nil {
		t.Fatalf("existing scoring config must pass validation: %v", err)
	}
	t.Logf("✅ A1: Existing config valid — H=%.1f V=%.1f decay=%.1f z=%.1f",
		cfg.Weights.Historical, cfg.Weights.Velocity, cfg.DecayRate, cfg.ZScoreThreshold)
}

// writeTestYAML is a helper that writes YAML content to a temp file.
func writeTestYAML(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write test YAML: %v", err)
	}
}
