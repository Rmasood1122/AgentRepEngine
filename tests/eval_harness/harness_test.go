package eval_harness_test

import (
	"fmt"
	"math"
	"testing"

	attack "github.com/agentrepengine/are/tests/attack_corpus"
	fp "github.com/agentrepengine/are/tests/fp_scenarios"
	"github.com/agentrepengine/are/internal/scoring"
)

var clusterBaselines = map[string]struct{ mean, std float64 }{
	"tool_call_rate_per_hour":       {80, 60},
	"unique_endpoints_per_hour":     {25, 30},
	"bulk_access_count_per_session": {400, 400},
	"pii_field_access_rate":         {20, 25},
	"cross_tenant_probe_count":      {0, 0.1},
	"permission_escalation_count":   {0, 0.5},
	"sub_agent_spawn_depth":         {0, 0.3},
	"token_refresh_rate":            {1, 1},
}

func scoreVector(vector map[string]float64) int {
	H := 700.0
	worstZ := 0.0
	for feature, value := range vector {
		if b, ok := clusterBaselines[feature]; ok {
			std := b.std
			if std == 0 {
				std = 0.1
			}
			z := (value - b.mean) / std
			if z > worstZ {
				worstZ = z
			}
		}
	}
	penalty := 0.0
	if worstZ > 3.0 {
		penalty = math.Min(100*(worstZ-3.0), 300)
	}
	V := math.Max(0, 1000-penalty*3)
	return scoring.ComputeScore(H, V, scoring.DefaultWeights)
}

func TestFPGate(t *testing.T) {
	t.Log("G-FP GATE: target FP rate <= 2%")
	total := len(fp.Phase1LegitCorpus)
	fps := 0
	var fpList []string
	for _, s := range fp.Phase1LegitCorpus {
		score := scoreVector(s.Vector)
		if score < 500 {
			fps++
			fpList = append(fpList, fmt.Sprintf("FP: %s score=%d", s.ID, score))
		}
	}
	rate := float64(fps) / float64(total) * 100
	t.Logf("Total=%d FP=%d Rate=%.2f%%", total, fps, rate)
	for _, f := range fpList {
		t.Log(f)
	}
	if rate > 2.0 {
		t.Fatalf("HARD STOP: FP rate %.2f%% > 2%%", rate)
	}
	t.Logf("PASSED: FP rate %.2f%%", rate)
}

func TestAttackDetection(t *testing.T) {
	t.Log("ATTACK DETECTION: target TP rate >= 85%")
	total := len(attack.Phase1AttackCorpus)
	detected := 0
	for _, s := range attack.Phase1AttackCorpus {
		if scoreVector(s.Vector) < 500 {
			detected++
		}
	}
	rate := float64(detected) / float64(total) * 100
	t.Logf("Total=%d Detected=%d TP=%.2f%%", total, detected, rate)
	if rate < 85.0 {
		t.Errorf("TP rate %.2f%% below 85%% target", rate)
	} else {
		t.Logf("PASSED: TP rate %.2f%%", rate)
	}
}

func TestScoreBandSanity(t *testing.T) {
	bands := map[string]int{"TRUSTED": 0, "MONITORED": 0, "RESTRICTED": 0, "BLOCKED": 0}
	for _, s := range fp.Phase1LegitCorpus {
		bands[scoring.ScoreBand(scoreVector(s.Vector))]++
	}
	t.Logf("Legit: TRUSTED=%d MONITORED=%d RESTRICTED=%d BLOCKED=%d",
		bands["TRUSTED"], bands["MONITORED"], bands["RESTRICTED"], bands["BLOCKED"])
	ab := map[string]int{"TRUSTED": 0, "MONITORED": 0, "RESTRICTED": 0, "BLOCKED": 0}
	for _, s := range attack.Phase1AttackCorpus {
		ab[scoring.ScoreBand(scoreVector(s.Vector))]++
	}
	t.Logf("Attack: TRUSTED=%d MONITORED=%d RESTRICTED=%d BLOCKED=%d",
		ab["TRUSTED"], ab["MONITORED"], ab["RESTRICTED"], ab["BLOCKED"])
}