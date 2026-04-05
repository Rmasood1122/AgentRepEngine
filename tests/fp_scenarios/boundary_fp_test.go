package fp_scenarios

import (
	"testing"
)

// TestFPRate_BoundaryCorpus validates that agents operating near the
// score=500 enforcement boundary are not incorrectly blocked.
//
// These 50 scenarios represent legitimate enterprise agents with
// elevated behavioral metrics (z-score 4.0–5.5 on one or more features).
// They are the highest-risk FP scenarios in the corpus — if the model
// over-fires, it fires here first.
//
// Independence statement: all scenarios in boundary_scenarios.go were
// constructed from behavioral specifications written before running
// against the scoring model. No scenario was modified after observing
// model output. Spec commit hash is embedded in boundary_scenarios.go.
func TestFPRate_BoundaryCorpus(t *testing.T) {
	falsePositives := 0
	total := len(Phase1BoundaryCorpus)

	for _, scenario := range Phase1BoundaryCorpus {
		score := computeScore(scenario.Vector)
		if score < 500 {
			falsePositives++
			t.Logf("FALSE POSITIVE [BOUNDARY]: %s (%s) score=%d minExpected=%d",
				scenario.ID, scenario.Name, score, scenario.MinScore)
		}
	}

	fpRate := float64(falsePositives) / float64(total) * 100
	t.Logf("Boundary FP Rate: %.2f%% (%d/%d boundary scenarios scored below MONITORED band)",
		fpRate, falsePositives, total)

	if falsePositives > 0 {
		t.Errorf("BOUNDARY GATE FAILED: %d boundary scenarios produced false positives — "+
			"model is over-firing near the enforcement threshold. "+
			"These are legitimate enterprise agents. Fix the model before pilot.",
			falsePositives)
	}
}

// TestFPRate_BoundaryMinScores verifies each boundary scenario meets
// its documented minimum expected score.
func TestFPRate_BoundaryMinScores(t *testing.T) {
	failures := 0

	for _, scenario := range Phase1BoundaryCorpus {
		score := computeScore(scenario.Vector)
		if score < scenario.MinScore {
			failures++
			t.Logf("SCORE TOO LOW [BOUNDARY]: %s (%s) got=%d expected>=%d",
				scenario.ID, scenario.Name, score, scenario.MinScore)
		}
	}

	if failures > 0 {
		t.Errorf("%d boundary scenarios scored below their minimum expected score", failures)
	}
}

// TestFPRate_CombinedCorpus reports the combined FP rate across
// all 150 scenarios (100 base + 50 boundary).
// This is the number used in investor and CISO communications
// after boundary corpus is added.
func TestFPRate_CombinedCorpus(t *testing.T) {
	allScenarios := append(Phase1LegitCorpus, Phase1BoundaryCorpus...)
	falsePositives := 0
	total := len(allScenarios)

	for _, scenario := range allScenarios {
		score := computeScore(scenario.Vector)
		if score < 500 {
			falsePositives++
			t.Logf("FALSE POSITIVE [COMBINED]: %s (%s) score=%d",
				scenario.ID, scenario.Name, score)
		}
	}

	fpRate := float64(falsePositives) / float64(total) * 100
	t.Logf("=== COMBINED CORPUS FP RATE: %.2f%% (%d/%d total scenarios) ===",
		fpRate, falsePositives, total)
	t.Logf("95%% CI upper bound (Clopper-Pearson): if 0 FPs on %d scenarios → <%.1f%%",
		total, 300.0/float64(total))

	if fpRate > 2.0 {
		t.Errorf("COMBINED GATE FAILED: FP rate %.2f%% exceeds 2.0%% threshold", fpRate)
	}
}
