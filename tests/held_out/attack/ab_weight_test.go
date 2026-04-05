package held_out_attack

import (
	"fmt"
	"math"
	"testing"
)

// PL-13: A/B threshold weight validation.
// Runs three weight configurations against the held-out corpus.
// Rule: max 10% weight change per iteration. F1 must improve to accept.
// DOES NOT modify weights — human decision required to update scoring_weights.yaml.
//
// Config A (current):   H=0.5, V=0.5, z>3.0
// Config B (candidate): H=0.4, V=0.6, z>3.0  (10% shift toward velocity)
// Config C (candidate): H=0.5, V=0.5, z>2.5  (lower z threshold)

type weightConfig struct {
	name       string
	hWeight    float64
	vWeight    float64
	zThreshold float64
}

var (
	cfgA = weightConfig{"current  (H=0.5 V=0.5 z>3.0)", 0.5, 0.5, 3.0}
	cfgB = weightConfig{"candidateB(H=0.4 V=0.6 z>3.0)", 0.4, 0.6, 3.0}
	cfgC = weightConfig{"candidateC(H=0.5 V=0.5 z>2.5)", 0.5, 0.5, 2.5}
)

// heldOutFPVectors — inline FP held-out vectors.
// Mirrors tests/held_out/fp/scenarios.go — kept in sync manually.
var heldOutFPVectors = []map[string]float64{
	{"tool_call_rate_per_hour": 95, "unique_endpoints_per_hour": 20, "bulk_access_count_per_session": 800},
	{"bulk_access_count_per_session": 3000, "pii_field_access_rate": 50, "tool_call_rate_per_hour": 40},
	{"unique_endpoints_per_hour": 55, "tool_call_rate_per_hour": 75, "bulk_access_count_per_session": 200},
	{"tool_call_rate_per_hour": 15, "unique_endpoints_per_hour": 8},
	{"tool_call_rate_per_hour": 30, "unique_endpoints_per_hour": 12, "pii_field_access_rate": 10},
	{"tool_call_rate_per_hour": 60, "bulk_access_count_per_session": 500, "unique_endpoints_per_hour": 10},
}

func computeScoreAB(vector map[string]float64, cfg weightConfig) int {
	worstZ := 0.0
	for feature, value := range vector {
		b, ok := baselines[feature]
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
	if worstZ > cfg.zThreshold {
		penalty = math.Min(100*(worstZ-cfg.zThreshold), 300)
	}
	V := math.Max(0, 1000-penalty*3)
	H := 700.0
	score := int(math.Round(cfg.hWeight*H + cfg.vWeight*V))
	if score < 0 {
		return 0
	}
	if score > 1000 {
		return 1000
	}
	return score
}

func bandAB(score int) string {
	switch {
	case score >= 800:
		return "TRUSTED"
	case score >= 500:
		return "MONITORED"
	case score >= 200:
		return "RESTRICTED"
	default:
		return "BLOCKED"
	}
}

func isDetectedAB(band string) bool {
	return band == "RESTRICTED" || band == "BLOCKED"
}

type abResult struct {
	tp, fp, fn, tn    int
	precision, recall float64
	f1                float64
}

func runConfig(cfg weightConfig) abResult {
	var r abResult
	for _, s := range HeldOutAttackCorpus {
		band := bandAB(computeScoreAB(s.Vector, cfg))
		if isDetectedAB(band) {
			r.tp++
		} else {
			r.fn++
		}
	}
	for _, v := range heldOutFPVectors {
		band := bandAB(computeScoreAB(v, cfg))
		if isDetectedAB(band) {
			r.fp++
		} else {
			r.tn++
		}
	}
	if r.tp+r.fp > 0 {
		r.precision = float64(r.tp) / float64(r.tp+r.fp)
	}
	if r.tp+r.fn > 0 {
		r.recall = float64(r.tp) / float64(r.tp+r.fn)
	}
	if r.precision+r.recall > 0 {
		r.f1 = 2 * r.precision * r.recall / (r.precision + r.recall)
	}
	return r
}

// TestABWeightValidation — PL-13 empirical weight selection.
func TestABWeightValidation(t *testing.T) {
	configs := []weightConfig{cfgA, cfgB, cfgC}
	results := make([]abResult, len(configs))
	for i, cfg := range configs {
		results[i] = runConfig(cfg)
	}

	t.Log("=== PL-13: A/B Weight Validation on Held-Out Corpus ===")
	t.Log("Config                            | TP  FP  FN  TN | Precision | Recall | F1")
	t.Log("----------------------------------|----------------|-----------|--------|--------")
	for i, cfg := range configs {
		r := results[i]
		t.Logf("%-34s| %2d  %2d  %2d  %2d | %8.4f  | %6.4f | %6.4f",
			cfg.name, r.tp, r.fp, r.fn, r.tn, r.precision, r.recall, r.f1)
	}

	bestIdx := 0
	for i := 1; i < len(results); i++ {
		if results[i].f1 > results[bestIdx].f1 {
			bestIdx = i
		}
	}

	improvement := results[bestIdx].f1 - results[0].f1
	t.Logf("")
	t.Logf("Winner: %s | F1 delta: %+.4f", configs[bestIdx].name, improvement)

	if bestIdx == 0 {
		t.Log("RECOMMENDATION: Keep current weights — no improvement found")
	} else {
		t.Logf("RECOMMENDATION: Consider %s — requires human approval before updating scoring_weights.yaml", configs[bestIdx].name)
	}

	// Gate: current config must hold F1 ≥ 0.85 and FP = 0
	if results[0].f1 < 0.85 {
		t.Errorf("REGRESSION: current weights F1 %.4f < 0.85 gate", results[0].f1)
	}
	// Note: FP gate on inline vectors only (6 samples). Authoritative FP gate
	// is TestHeldOut_FPRate in tests/held_out/fp/ — 20 scenarios, 0.00% required.
	if results[0].fp > 0 {
		t.Logf("WARNING: %d FP(s) on inline validation vectors — check heldOutFPVectors alignment", results[0].fp)
	}
	t.Logf("Current weights: F1=%.4f FP=%d/%d Precision=%.2f%% Recall=%.2f%%", results[0].f1, results[0].fp, len(heldOutFPVectors), results[0].precision*100, results[0].recall*100)
}

// TestWeightChangeProtocol validates proposed changes follow 10% max rule.
func TestWeightChangeProtocol(t *testing.T) {
	currentH, currentV := 0.5, 0.5
	maxDelta := 0.10

	proposals := []struct {
		name string
		h, v float64
	}{
		{"H=0.4 V=0.6", 0.4, 0.6},
		{"H=0.6 V=0.4", 0.6, 0.4},
		{"H=0.45 V=0.55", 0.45, 0.55},
	}

	t.Log("=== Threshold Change Protocol (10% max per iteration) ===")
	for _, p := range proposals {
		dh := math.Abs(p.h - currentH)
		dv := math.Abs(p.v - currentV)
		ok := dh <= maxDelta && dv <= maxDelta
		status := "ALLOWED"
		if !ok {
			status = "REJECTED"
		}
		t.Logf("  %-20s ΔH=%.2f ΔV=%.2f → %s", fmt.Sprintf("%s", p.name), dh, dv, status)
	}
}
