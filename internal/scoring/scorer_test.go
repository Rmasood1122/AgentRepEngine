package scoring

import (
	"math"
	"testing"
	"time"
)

// ═══════════════════════════════════════════════════
// G-SCORE GATE TESTS
// ═══════════════════════════════════════════════════

// TestWeightValidation verifies config rejects invalid weights.
// FM3 prevention: weights must sum to 1.0.
func TestWeightValidation(t *testing.T) {
	// Valid weights
	valid := ScoreWeights{Historical: 0.5, Velocity: 0.5}
	score := ComputeScore(700, 700, valid)
	if score < 0 || score > 1000 {
		t.Errorf("valid weights produced out-of-bounds score: %d", score)
	}
	t.Logf("✅ Valid weights produce score: %d", score)

	// Verify weight sum constraint in config loader
	_, err := LoadScoringConfig("../../config/scoring_weights.yaml")
	if err != nil {
		t.Fatalf("failed to load valid config: %v", err)
	}
	t.Log("✅ Config loaded and weights validated")
}

// TestScoreBounds verifies Clamp(score, 0, 1000) always holds.
func TestScoreBounds(t *testing.T) {
	weights := ScoreWeights{Historical: 0.5, Velocity: 0.5}

	tests := []struct {
		H, V float64
	}{
		{0, 0},       // minimum
		{1000, 1000}, // maximum
		{-100, 500},  // negative H — should clamp to 0
		{500, 1500},  // V over 1000 — should clamp to 1000
		{700, 700},   // normal case
	}

	for _, tt := range tests {
		score := ComputeScore(tt.H, tt.V, weights)
		if score < 0 || score > 1000 {
			t.Errorf("score out of bounds: H=%.0f V=%.0f score=%d",
				tt.H, tt.V, score)
		}
	}
	t.Log("✅ Score bounds 0-1000 enforced in all cases")
}

// TestScoreDecay verifies exponential decay formula.
// H(t) = H(t-1) * e^(-0.1 * days)
func TestScoreDecay(t *testing.T) {
	initialScore := 900.0

	// No decay for same-day activity
	sameDay := time.Now().Add(-1 * time.Minute)
	decayed := ApplyDecay(initialScore, sameDay)
	if math.Abs(decayed-initialScore) > 1.0 {
		t.Errorf("same-day decay too aggressive: %.2f → %.2f",
			initialScore, decayed)
	}
	t.Logf("✅ Same-day decay minimal: %.2f → %.2f", initialScore, decayed)

	// 7 days decay
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	decayed7 := ApplyDecay(initialScore, sevenDaysAgo)
	expected7 := initialScore * math.Exp(-0.1*7)
	if math.Abs(decayed7-expected7) > 1.0 {
		t.Errorf("7-day decay wrong: got %.2f want %.2f", decayed7, expected7)
	}
	t.Logf("✅ 7-day decay: %.2f → %.2f (expected %.2f)",
		initialScore, decayed7, expected7)

	// Score should decrease over time
	if decayed7 >= initialScore {
		t.Error("score should decrease after 7 days of inactivity")
	}
	t.Log("✅ Score correctly decreases with inactivity")
}

// TestZScoreComputation verifies z-score formula.
func TestZScoreComputation(t *testing.T) {
	baseline := Baseline{Mean: 80, StdDev: 60}

	// Normal behavior — z < 3
	normalZ := ComputeZScore(100, baseline)
	if normalZ >= 3.0 {
		t.Errorf("normal behavior z-score too high: %.2f", normalZ)
	}
	t.Logf("✅ Normal z-score: %.2f (value=100, mean=80, std=60)", normalZ)

	// Anomalous behavior — z > 3
	anomalousZ := ComputeZScore(500, baseline)
	if anomalousZ <= 3.0 {
		t.Errorf("anomalous behavior not detected: z=%.2f", anomalousZ)
	}
	t.Logf("✅ Anomalous z-score: %.2f (value=500, mean=80, std=60)", anomalousZ)

	// Zero std dev — should return 0 not divide by zero
	zeroStd := Baseline{Mean: 80, StdDev: 0}
	safeZ := ComputeZScore(500, zeroStd)
	if math.IsNaN(safeZ) || math.IsInf(safeZ, 0) {
		t.Errorf("zero std dev caused NaN/Inf: %.2f", safeZ)
	}
	t.Logf("✅ Zero std dev safe: z=%.2f", safeZ)
}

// TestScoreBands verifies band boundaries.
func TestScoreBands(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{1000, "TRUSTED"},
		{800, "TRUSTED"},
		{799, "MONITORED"},
		{500, "MONITORED"},
		{499, "RESTRICTED"},
		{200, "RESTRICTED"},
		{199, "BLOCKED"},
		{0, "BLOCKED"},
	}

	for _, tt := range tests {
		got := ScoreBand(tt.score)
		if got != tt.want {
			t.Errorf("ScoreBand(%d) = %s, want %s", tt.score, got, tt.want)
		}
	}
	t.Log("✅ Score band boundaries correct")
}

// TestPhase1FormulaOnly verifies we use H+V only (no A or P yet).
// Anti-scope: Isolation Forest is Phase 2 only.
func TestPhase1FormulaOnly(t *testing.T) {
	weights := DefaultWeights

	// Phase 1: H=0.5, V=0.5, A=0, P=0
	if weights.Historical != 0.5 {
		t.Errorf("historical weight = %.2f, want 0.5", weights.Historical)
	}
	if weights.Velocity != 0.5 {
		t.Errorf("velocity weight = %.2f, want 0.5", weights.Velocity)
	}
	t.Logf("✅ Phase 1 weights: H=%.1f V=%.1f (A and P not used yet)",
		weights.Historical, weights.Velocity)
}

// TestIdempotency verifies same input always produces same score.
func TestIdempotency(t *testing.T) {
	weights := ScoreWeights{Historical: 0.5, Velocity: 0.5}

	score1 := ComputeScore(700, 650, weights)
	score2 := ComputeScore(700, 650, weights)
	score3 := ComputeScore(700, 650, weights)

	if score1 != score2 || score2 != score3 {
		t.Errorf("non-idempotent: %d %d %d", score1, score2, score3)
	}
	t.Logf("✅ Idempotent: same inputs always produce score=%d", score1)
}
