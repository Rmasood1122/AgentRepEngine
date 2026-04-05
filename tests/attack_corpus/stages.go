package attack_corpus

// PL-8: Staged curriculum taxonomy for attack corpus.
// Maps Phase1AttackCorpus into 4 progressive stages for enterprise reviewers.
// ADDITIVE — Phase1AttackCorpus in scenarios.go is unchanged.
//
// Stage 1: Basic rate anomaly (LLM01, LLM04) — single-vector, high signal
// Stage 2: PII access patterns (LLM06) — data exfiltration variants
// Stage 3: Slow-walk evasion (LLM07, LLM08, LLM09) — privilege + spawn
// Stage 4: Multi-vector combined (combined OWASP refs + noise-injected)

// StageMap classifies each scenario ID into its curriculum stage.
var StageMap = map[string]int{
	// Stage 1 — Basic rate anomaly
	"ATK-001": 1, "ATK-002": 1, "ATK-003": 1, "ATK-004": 1, "ATK-005": 1,
	"ATK-006": 1, "ATK-007": 1, "ATK-008": 1, "ATK-009": 1, "ATK-010": 1,
	// Stage 2 — PII access patterns
	"ATK-011": 2, "ATK-012": 2, "ATK-013": 2, "ATK-014": 2, "ATK-015": 2,
	"ATK-016": 2,
	// Stage 3 — Slow-walk evasion + privilege escalation + spawn
	"ATK-017": 3, "ATK-018": 3, "ATK-019": 3, "ATK-020": 3, "ATK-021": 3,
	"ATK-022": 3, "ATK-023": 3, "ATK-024": 3, "ATK-025": 3, "ATK-026": 3,
	// Stage 4 — Multi-vector combined + noise-injected
	"ATK-027": 4, "ATK-028": 4, "ATK-029": 4, "ATK-030": 4,
	"ATK-031": 4, "ATK-032": 4, "ATK-033": 4, "ATK-034": 4, "ATK-035": 4,
	"ATK-036": 4, "ATK-037": 4, "ATK-038": 4, "ATK-039": 4, "ATK-040": 4,
	"ATK-041": 4, "ATK-042": 4, "ATK-043": 4, "ATK-044": 4, "ATK-045": 4,
	"ATK-046": 4, "ATK-047": 4, "ATK-048": 4, "ATK-049": 4, "ATK-050": 4,
}

// Stage1Corpus — basic rate anomaly scenarios (LLM01 + LLM04)
var Stage1Corpus []AttackScenario

// Stage2Corpus — PII access pattern scenarios (LLM06)
var Stage2Corpus []AttackScenario

// Stage3Corpus — slow-walk evasion + privilege + spawn (LLM07/08/09)
var Stage3Corpus []AttackScenario

// Stage4Corpus — multi-vector combined + noise-injected
var Stage4Corpus []AttackScenario

func init() {
	for _, s := range Phase1AttackCorpus {
		stage, ok := StageMap[s.ID]
		if !ok {
			stage = 4 // default unknown to stage 4
		}
		switch stage {
		case 1:
			Stage1Corpus = append(Stage1Corpus, s)
		case 2:
			Stage2Corpus = append(Stage2Corpus, s)
		case 3:
			Stage3Corpus = append(Stage3Corpus, s)
		case 4:
			Stage4Corpus = append(Stage4Corpus, s)
		}
	}
}
