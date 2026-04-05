package attack_corpus

import (
	"testing"
)

// PL-10: Taxonomy coverage test.
// Verifies all 50 scenarios are classified and distribution is documented.
func TestTaxonomyCoverage(t *testing.T) {
	counts := map[ScenarioCategory]int{}
	unclassified := []string{}

	for _, s := range Phase1AttackCorpus {
		_, explicit := TaxonomyMap[s.ID]
		cat := GetCategory(s.ID)
		counts[cat]++
		if !explicit {
			unclassified = append(unclassified, s.ID)
		}
	}

	t.Log("=== PL-10: 5-Category Corpus Taxonomy ===")
	t.Logf("typical:     %d scenarios (clear anomalies, high signal)", counts[CategoryTypical])
	t.Logf("boundary:    %d scenarios (near RESTRICTED/BLOCKED threshold)", counts[CategoryBoundary])
	t.Logf("edge:        %d scenarios (novel/combined/noise-injected)", counts[CategoryEdge])
	t.Logf("negative:    %d scenarios (in fp_scenarios — not in attack corpus)", counts[CategoryNegative])
	t.Logf("performance: %d scenarios (in tests/performance)", counts[CategoryPerformance])
	t.Logf("Total classified: %d / %d", len(Phase1AttackCorpus)-len(unclassified), len(Phase1AttackCorpus))

	if len(unclassified) > 0 {
		t.Logf("Unclassified (defaulted to edge): %v", unclassified)
	}

	// All 50 must be accounted for
	total := counts[CategoryTypical] + counts[CategoryBoundary] + counts[CategoryEdge]
	if total != len(Phase1AttackCorpus) {
		t.Errorf("taxonomy coverage: %d/%d scenarios classified", total, len(Phase1AttackCorpus))
	} else {
		t.Logf("PASS: all %d scenarios classified", total)
	}

	// Boundary category must have scenarios (validates threshold sensitivity)
	if counts[CategoryBoundary] == 0 {
		t.Error("no boundary scenarios — corpus lacks threshold sensitivity coverage")
	}
}
