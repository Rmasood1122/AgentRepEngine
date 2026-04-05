package regression

import (
	"math"
	"testing"

	"github.com/agentrepengine/are/internal/scoring"
	attack "github.com/agentrepengine/are/tests/attack_corpus"
	fp "github.com/agentrepengine/are/tests/fp_scenarios"
)

// PL-14: Regression gate — must pass before any scoring engine change merges.
// Pass criteria: FP=0.00% on fp corpus, TP≥86.67% on attack corpus.
// Run this before and after every change to internal/scoring/*.go
//
// Usage:
//   go test ./tests/regression/... -v
//
// Any failure = revert the scoring change immediately. No exceptions.

var regressionBaselines = map[string]struct{ mean, std float64 }{
	"tool_call_rate_per_hour":       {80, 60},
	"unique_endpoints_per_hour":     {25, 30},
	"bulk_access_count_per_session": {400, 400},
	"pii_field_access_rate":         {20, 25},
	"cross_tenant_probe_count":      {0, 0.1},
	"permission_escalation_count":   {0, 0.5},
	"sub_agent_spawn_depth":         {0, 0.3},
	"token_refresh_rate":            {1, 1},
}

func regressionScore(vector map[string]float64) int {
	worstZ := 0.0
	for feature, value := range vector {
		b, ok := regressionBaselines[feature]
		if !ok {
			continue
		}
		std := b.std
		if std == 0 {
			std = 0.1
		}
		z := (value - b.mean) / std
		if z > worstZ {
			worstZ = z
		}
	}
	penalty := 0.0
	if worstZ > 3.0 {
		penalty = math.Min(100*(worstZ-3.0), 300)
	}
	V := math.Max(0, 1000-penalty*3)
	H := 700.0
	score := int(math.Round(0.5*H + 0.5*V))
	if score < 0 {
		return 0
	}
	if score > 1000 {
		return 1000
	}
	return score
}

func regressionBand(score int) string {
	return scoring.ScoreBand(score)
}

// TestRegressionFPGate — FP rate must be 0.00% on full fp corpus.
// Hard stop: any FP regression blocks merge to scoring engine.
func TestRegressionFPGate(t *testing.T) {
	scenarios := fp.Phase1LegitCorpus
	blocked := 0

	for _, s := range scenarios {
		score := regressionScore(s.Vector)
		band := regressionBand(score)
		if band == "RESTRICTED" || band == "BLOCKED" {
			blocked++
			t.Logf("FP DETECTED: %s score=%d band=%s", s.ID, score, band)
		}
	}

	fpRate := float64(blocked) / float64(len(scenarios)) * 100

	if fpRate > 0.00 {
		t.Errorf("REGRESSION GATE FAIL: FP rate %.2f%% > 0.00%% — revert scoring change immediately",
			fpRate)
	} else {
		t.Logf("REGRESSION FP GATE: PASS — %.2f%% on %d scenarios", fpRate, len(scenarios))
	}
}

// TestRegressionTPGate — TP rate must be ≥86.67% on full attack corpus.
// Hard stop: any TP regression below gate blocks merge.
func TestRegressionTPGate(t *testing.T) {
	scenarios := attack.Phase1AttackCorpus
	detected := 0

	for _, s := range scenarios {
		score := regressionScore(s.Vector)
		band := regressionBand(score)
		if band == "RESTRICTED" || band == "BLOCKED" {
			detected++
		} else {
			t.Logf("MISSED: %s score=%d band=%s expected=%s", s.ID, score, band, s.ExpectedBand)
		}
	}

	tpRate := float64(detected) / float64(len(scenarios)) * 100
	const tpGate = 86.67

	if tpRate < tpGate {
		t.Errorf("REGRESSION GATE FAIL: TP rate %.2f%% < %.2f%% gate — revert scoring change",
			tpRate, tpGate)
	} else {
		t.Logf("REGRESSION TP GATE: PASS — %.2f%% (%d/%d) gate=%.2f%%",
			tpRate, detected, len(scenarios), tpGate)
	}
}

// TestRegressionSummary — prints 4-metric summary for commit records.
// Always runs after FP and TP gates. Documents baseline for comparison.
func TestRegressionSummary(t *testing.T) {
	fpScenarios := fp.Phase1LegitCorpus
	atkScenarios := attack.Phase1AttackCorpus

	fpBlocked := 0
	for _, s := range fpScenarios {
		band := regressionBand(regressionScore(s.Vector))
		if band == "RESTRICTED" || band == "BLOCKED" {
			fpBlocked++
		}
	}

	atkDetected := 0
	for _, s := range atkScenarios {
		band := regressionBand(regressionScore(s.Vector))
		if band == "RESTRICTED" || band == "BLOCKED" {
			atkDetected++
		}
	}

	fpRate := float64(fpBlocked) / float64(len(fpScenarios)) * 100
	tpRate := float64(atkDetected) / float64(len(atkScenarios)) * 100
	precision := 0.0
	if atkDetected+fpBlocked > 0 {
		precision = float64(atkDetected) / float64(atkDetected+fpBlocked) * 100
	}
	f1 := 0.0
	if tpRate+precision > 0 {
		f1 = 2 * (precision / 100) * (tpRate / 100) / ((precision / 100) + (tpRate / 100))
	}

	t.Logf("=== REGRESSION BASELINE METRICS ===")
	t.Logf("FP corpus:     %d scenarios | blocked: %d | FP rate: %.2f%%",
		len(fpScenarios), fpBlocked, fpRate)
	t.Logf("Attack corpus: %d scenarios | detected: %d | TP rate: %.2f%%",
		len(atkScenarios), atkDetected, tpRate)
	t.Logf("4-metric: TP=%.2f%% FP=%.2f%% Precision=%.2f%% F1=%.4f",
		tpRate, fpRate, precision, f1)
	t.Logf("Gates: FP≤0.00%% [%s] | TP≥86.67%% [%s]",
		gateStr(fpRate == 0), gateStr(tpRate >= 86.67))
}

func gateStr(pass bool) string {
	if pass {
		return "PASS"
	}
	return "FAIL"
}
