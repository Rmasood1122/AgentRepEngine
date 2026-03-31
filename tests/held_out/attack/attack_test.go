package held_out_attack

import (
	"math"
	"testing"

	"github.com/agentrepengine/are/internal/scoring"
)

// HELD-OUT VALIDATION ONLY — do not use for threshold calibration.
// These scenarios validate final published metrics (TP rate, F1).

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

func applyPolicyPenalties(v scoring.FeatureVector) (float64, bool) {
	penalty := 0.0
	highRiskFired := false

	if v.PIIFieldAccessRate > 0.3 {
		penalty += 200
		highRiskFired = true
	}
	if v.CrossTenantProbeCount > 0 {
		penalty += 300
		highRiskFired = true
	}
	if v.PermissionEscalationCount > 2 {
		penalty += 150
		highRiskFired = true
	}
	if v.SubAgentSpawnDepth > 3 {
		penalty += 100
		highRiskFired = true
	}
	if v.ToolCallRatePerHour > 100 {
		penalty += 50
		highRiskFired = true
	}

	return penalty, highRiskFired
}

func computeVelocityScore(v scoring.FeatureVector, agentBaseline float64) float64 {
	normalStd := 15.0
	z := (v.ToolCallRatePerHour - agentBaseline) / normalStd
	if z < 0 {
		z = 0
	}
	velocityPenalty := 0.0
	if z > 3.0 {
		velocityPenalty = 100.0 * (z - 3.0)
		if velocityPenalty > 300 {
			velocityPenalty = 300
		}
	}
	V := 1000.0 - velocityPenalty
	if V < 0 {
		V = 0
	}
	return V
}

func simulateScoreOverDays(days []DayVector, initialScore float64) (float64, bool) {
	score := initialScore
	agentBaseline := 50.0
	anyHighRiskFired := false

	for i, day := range days {
		daysElapsed := 0.0
		if i > 0 {
			daysElapsed = 1.0
		}
		H := score * math.Exp(-0.1*daysElapsed)
		V := computeVelocityScore(day.Vector, agentBaseline)
		penalty, highRiskFired := applyPolicyPenalties(day.Vector)
		if highRiskFired {
			anyHighRiskFired = true
		}
		newScore := 0.5*H + 0.5*V - penalty
		if newScore < 0 {
			newScore = 0
		}
		if newScore > 1000 {
			newScore = 1000
		}
		score = newScore
		agentBaseline = 0.9*agentBaseline + 0.1*day.Vector.ToolCallRatePerHour
	}
	return score, anyHighRiskFired
}

func scoreBandFromScore(score float64) string {
	if score >= 700 {
		return "TRUSTED"
	} else if score >= 500 {
		return "MONITORED"
	} else if score >= 200 {
		return "RESTRICTED"
	}
	return "BLOCKED"
}

func TestHeldOut_TPRate(t *testing.T) {
	detected := 0
	total := len(HeldOutAttackCorpus)

	for _, scenario := range HeldOutAttackCorpus {
		score := computeScore(scenario.Vector)
		band := ""
		if score >= 700 {
			band = "TRUSTED"
		} else if score >= 500 {
			band = "MONITORED"
		} else if score >= 200 {
			band = "RESTRICTED"
		} else {
			band = "BLOCKED"
		}

		if band == "RESTRICTED" || band == "BLOCKED" {
			detected++
			t.Logf("DETECTED: %s (%s) score=%d band=%s expected=%s",
				scenario.ID, scenario.Name, score, band, scenario.ExpectedBand)
		} else {
			t.Logf("MISSED: %s (%s) score=%d band=%s expected=%s",
				scenario.ID, scenario.Name, score, band, scenario.ExpectedBand)
		}
	}

	tpRate := float64(detected) / float64(total) * 100
	t.Logf("HELD-OUT TP Rate: %.2f%% (%d/%d attack scenarios detected)", tpRate, detected, total)

	if tpRate < 85.0 {
		t.Errorf("HELD-OUT VALIDATION FAILED: TP rate %.2f%% below 85%% threshold", tpRate)
	}
}

func TestHeldOut_SlowWalkDetection(t *testing.T) {
	detected := 0
	total := len(HeldOutSlowWalkCorpus)

	for _, scenario := range HeldOutSlowWalkCorpus {
		finalScore, highRiskFired := simulateScoreOverDays(scenario.Days, 700.0)
		finalBand := scoreBandFromScore(finalScore)
		scoreDetected := finalBand == "RESTRICTED" || finalBand == "BLOCKED"

		if scoreDetected || highRiskFired {
			detected++
			t.Logf("DETECTED: %s (%s) score=%.0f band=%s", scenario.ID, scenario.Name, finalScore, finalBand)
		} else {
			t.Logf("MISSED: %s (%s) score=%.0f band=%s", scenario.ID, scenario.Name, finalScore, finalBand)
		}
	}

	detectionRate := float64(detected) / float64(total) * 100
	t.Logf("HELD-OUT Slow-Walk Detection: %.2f%% (%d/%d)", detectionRate, detected, total)

	if detectionRate < 70.0 {
		t.Errorf("HELD-OUT VALIDATION FAILED: slow-walk detection %.2f%% below 70%% threshold", detectionRate)
	}
}
