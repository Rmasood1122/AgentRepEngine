package fp_scenarios

import (
	"strings"
	"testing"
)

// PL-16: Per-agent-type FP bias audit.
// Confirms 0.00% FP rate holds across ALL agent archetypes, not just overall.
// FAANG security reviewer will ask this question.
// Classification by name prefix — additive, zero modifications to existing scenarios.
//
// Agent archetypes derived from scenario names:
//   data_analyst   — "Data analyst *"
//   authorized     — "Authorized *"
//   research       — "Research *" / "Multi-step *"
//   monitoring     — "Monitoring *" / "Watchdog *" / "Health *"
//   batch          — "Batch *" / "Scheduled *" / "Nightly *"
//   cold_start     — "New agent *" / "Cold start *" / "Bootstrap *"
//   recovery       — "Recovery *" / "Recovering *" / "Rebuilding *"
//   other          — everything else

func classifyAgent(name string) string {
	n := strings.ToLower(name)
	switch {
	case strings.HasPrefix(n, "data analyst"):
		return "data_analyst"
	case strings.HasPrefix(n, "authorized"):
		return "authorized"
	case strings.HasPrefix(n, "research") || strings.HasPrefix(n, "multi-step"):
		return "research"
	case strings.HasPrefix(n, "monitoring") || strings.HasPrefix(n, "watchdog") || strings.HasPrefix(n, "health"):
		return "monitoring"
	case strings.HasPrefix(n, "batch") || strings.HasPrefix(n, "scheduled") || strings.HasPrefix(n, "nightly"):
		return "batch"
	case strings.HasPrefix(n, "new agent") || strings.HasPrefix(n, "cold start") || strings.HasPrefix(n, "bootstrap"):
		return "cold_start"
	case strings.HasPrefix(n, "recover") || strings.HasPrefix(n, "rebuilding"):
		return "recovery"
	default:
		return "other"
	}
}

func isFPBiasAudit(vector map[string]float64) bool {
	return computeScore(vector) < 500
}

// TestFPBiasAudit — per-agent-type FP rate.
// All archetypes must show 0.00% FP. Any archetype with FP > 0 = bias detected.
func TestFPBiasAudit(t *testing.T) {
	type archetypeResult struct {
		total   int
		blocked int
	}

	results := map[string]*archetypeResult{}

	for _, s := range Phase1LegitCorpus {
		archetype := classifyAgent(s.Name)
		if _, ok := results[archetype]; !ok {
			results[archetype] = &archetypeResult{}
		}
		results[archetype].total++

		if isFPBiasAudit(s.Vector) {
			results[archetype].blocked++
			t.Logf("FP DETECTED [%s] %s: score=%d", archetype, s.ID, computeScore(s.Vector))
		}
	}

	t.Log("=== PL-16: Per-Agent-Type FP Bias Audit ===")
	t.Log("Archetype        | Total | FP | FP Rate")
	t.Log("-----------------|-------|----|--------")

	totalFPs := 0
	archetypesFailing := 0

	for _, archetype := range []string{"data_analyst", "authorized", "research", "monitoring", "batch", "cold_start", "recovery", "other"} {
		r, ok := results[archetype]
		if !ok {
			t.Logf("%-17s| %5d | %2d | %6.2f%% (no scenarios)", archetype, 0, 0, 0.0)
			continue
		}
		fpRate := float64(r.blocked) / float64(r.total) * 100
		t.Logf("%-17s| %5d | %2d | %6.2f%%", archetype, r.total, r.blocked, fpRate)
		totalFPs += r.blocked
		if r.blocked > 0 {
			archetypesFailing++
		}
	}

	t.Logf("---")
	t.Logf("Total: %d scenarios | FPs: %d | Archetypes with FP: %d",
		len(Phase1LegitCorpus), totalFPs, archetypesFailing)

	if archetypesFailing > 0 {
		t.Errorf("FP BIAS DETECTED: %d archetype(s) show FP > 0.00%% — scoring is biased against specific agent types",
			archetypesFailing)
	} else {
		t.Log("BIAS AUDIT PASS: 0.00% FP across all agent archetypes")
	}
}
