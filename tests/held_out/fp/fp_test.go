package held_out_fp

import (
	"math"
	"testing"
)

// HELD-OUT VALIDATION ONLY — do not use for threshold calibration.
// These scenarios validate final published metrics (FP rate, F1).

var baselines = map[string]struct{ mean, std float64 }{
	"tool_call_rate_per_hour":       {80, 60},
	"unique_endpoints_per_hour":     {25, 30},
	"bulk_access_count_per_session": {400, 400},
	"pii_field_access_rate":         {20, 25},
	"cross_tenant_probe_count":      {0, 0.1},
	"permission_escalation_count":   {0, 0.5},
	"sub_agent_spawn_depth":         {0, 0.3},
	"token_refresh_rate":            {1, 1},
}

func computeScore(vector map[string]float64) int {
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

func TestHeldOut_FPRate(t *testing.T) {
	falsePositives := 0
	total := len(HeldOutLegitCorpus)

	for _, scenario := range HeldOutLegitCorpus {
		score := computeScore(scenario.Vector)
		if score < 500 {
			falsePositives++
			t.Logf("FALSE POSITIVE: %s (%s) score=%d minExpected=%d",
				scenario.ID, scenario.Name, score, scenario.MinScore)
		}
	}

	fpRate := float64(falsePositives) / float64(total) * 100
	t.Logf("HELD-OUT FP Rate: %.2f%% (%d/%d scenarios scored below MONITORED band)",
		fpRate, falsePositives, total)

	if fpRate > 2.0 {
		t.Errorf("HELD-OUT VALIDATION FAILED: FP rate %.2f%% exceeds 2.0%% threshold",
			fpRate)
	}
}
