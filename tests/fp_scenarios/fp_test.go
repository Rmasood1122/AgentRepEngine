package fp_scenarios

import (
	"math"
	"testing"
)

// Phase 1 scoring constants — mirrors consumer.go bootstrap baselines
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

func TestFPRate_LegitimateCorpus(t *testing.T) {
	falsePositives := 0
	total := len(Phase1LegitCorpus)

	for _, scenario := range Phase1LegitCorpus {
		score := computeScore(scenario.Vector)
		if score < 500 {
			falsePositives++
			t.Logf("FALSE POSITIVE: %s (%s) score=%d minExpected=%d",
				scenario.ID, scenario.Name, score, scenario.MinScore)
		}
	}

	fpRate := float64(falsePositives) / float64(total) * 100
	t.Logf("FP Rate: %.2f%% (%d/%d scenarios scored below MONITORED band)",
		fpRate, falsePositives, total)

	if fpRate > 2.0 {
		t.Errorf("GATE 3 FAILED: FP rate %.2f%% exceeds 2.0%% threshold (%d false positives)",
			fpRate, falsePositives)
	}
}

func TestFPRate_MinScoreRequirements(t *testing.T) {
	failures := 0

	for _, scenario := range Phase1LegitCorpus {
		score := computeScore(scenario.Vector)
		if score < scenario.MinScore {
			failures++
			t.Logf("SCORE TOO LOW: %s (%s) got=%d expected>=%d",
				scenario.ID, scenario.Name, score, scenario.MinScore)
		}
	}

	if failures > 0 {
		t.Errorf("%d scenarios scored below their minimum expected score", failures)
	}
}

func TestScoreBands_Sanity(t *testing.T) {
	tests := []struct {
		name     string
		vector   map[string]float64
		minScore int
		maxScore int
	}{
		{
			name:     "idle agent",
			vector:   map[string]float64{"tool_call_rate_per_hour": 1},
			minScore: 700, maxScore: 1000,
		},
		{
			name:     "normal agent",
			vector:   map[string]float64{"tool_call_rate_per_hour": 80, "unique_endpoints_per_hour": 25},
			minScore: 600, maxScore: 1000,
		},
		{
			name:     "elevated but legitimate",
			vector:   map[string]float64{"tool_call_rate_per_hour": 140, "bulk_access_count_per_session": 800},
			minScore: 500, maxScore: 1000,
		},
		{
			name:     "clear attacker",
			vector:   map[string]float64{"cross_tenant_probe_count": 10, "permission_escalation_count": 5},
			minScore: 0, maxScore: 499,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := computeScore(tt.vector)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("%s: score=%d want [%d, %d]",
					tt.name, score, tt.minScore, tt.maxScore)
			}
		})
	}
}
