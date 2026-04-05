package attack_corpus

// PL-10: 5-category corpus taxonomy.
// Labels all scenarios for enterprise reviewers and FAANG security auditors.
// ADDITIVE — no existing files modified.
//
// Categories:
//   typical     — standard attack pattern, clearly anomalous
//   boundary    — score near band threshold (±50 points from boundary)
//   edge        — novel/rare/combined attack vector
//   negative    — should NOT block (FP test — these are in fp_scenarios, not here)
//   performance — high-concurrency burst (in tests/performance/)

type ScenarioCategory string

const (
	CategoryTypical     ScenarioCategory = "typical"
	CategoryBoundary    ScenarioCategory = "boundary"
	CategoryEdge        ScenarioCategory = "edge"
	CategoryNegative    ScenarioCategory = "negative"
	CategoryPerformance ScenarioCategory = "performance"
)

// TaxonomyMap classifies each attack scenario by category.
var TaxonomyMap = map[string]ScenarioCategory{
	// typical — clear anomalies, high signal, expected to block
	"ATK-001": CategoryTypical,
	"ATK-003": CategoryTypical,
	"ATK-004": CategoryTypical,
	"ATK-006": CategoryTypical,
	"ATK-007": CategoryTypical,
	"ATK-011": CategoryTypical,
	"ATK-014": CategoryTypical,
	"ATK-017": CategoryTypical,
	"ATK-018": CategoryTypical,
	"ATK-022": CategoryTypical,
	"ATK-024": CategoryTypical,
	"ATK-027": CategoryTypical,
	"ATK-029": CategoryTypical,
	"ATK-030": CategoryTypical,
	"ATK-037": CategoryTypical,
	"ATK-038": CategoryTypical,
	"ATK-039": CategoryTypical,
	"ATK-040": CategoryTypical,
	"ATK-041": CategoryTypical,
	"ATK-042": CategoryTypical,
	"ATK-043": CategoryTypical,
	"ATK-044": CategoryTypical,
	"ATK-045": CategoryTypical,

	// boundary — score near RESTRICTED/BLOCKED boundary (expected: RESTRICTED)
	"ATK-002": CategoryBoundary,
	"ATK-005": CategoryBoundary,
	"ATK-009": CategoryBoundary,
	"ATK-012": CategoryBoundary,
	"ATK-013": CategoryBoundary,
	"ATK-016": CategoryBoundary,
	"ATK-021": CategoryBoundary,
	"ATK-026": CategoryBoundary,
	"ATK-032": CategoryBoundary,
	"ATK-034": CategoryBoundary,
	"ATK-036": CategoryBoundary,
	"ATK-046": CategoryBoundary,

	// edge — novel, combined, or noise-injected vectors
	"ATK-008": CategoryEdge,
	"ATK-010": CategoryEdge,
	"ATK-015": CategoryEdge,
	"ATK-019": CategoryEdge,
	"ATK-020": CategoryEdge,
	"ATK-023": CategoryEdge,
	"ATK-025": CategoryEdge,
	"ATK-028": CategoryEdge,
	"ATK-031": CategoryEdge,
	"ATK-033": CategoryEdge,
	"ATK-035": CategoryEdge,
	"ATK-047": CategoryEdge,
	"ATK-048": CategoryEdge,
	"ATK-049": CategoryEdge,
	"ATK-050": CategoryEdge,
}

// GetCategory returns the taxonomy category for a scenario ID.
// Returns CategoryEdge as default for unclassified scenarios.
func GetCategory(scenarioID string) ScenarioCategory {
	if cat, ok := TaxonomyMap[scenarioID]; ok {
		return cat
	}
	return CategoryEdge
}

// ByCategory returns all scenarios in Phase1AttackCorpus matching a category.
func ByCategory(cat ScenarioCategory) []AttackScenario {
	var result []AttackScenario
	for _, s := range Phase1AttackCorpus {
		if GetCategory(s.ID) == cat {
			result = append(result, s)
		}
	}
	return result
}
