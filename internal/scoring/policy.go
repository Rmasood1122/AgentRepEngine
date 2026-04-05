package scoring

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Variance growth rate monitoring constants.
// Slow-walk attacks require multiple days to shift an agent's baseline.
// A 2x variance increase in one window is statistically anomalous —
// it indicates the baseline is being actively trained by an attacker.
// This fires as an early warning before the attack score recovers to
// a normal range, because the mean may look stable while variance grows.
const (
	// VarianceWindowDays is the lookback window for variance growth monitoring.
	VarianceWindowDays = 7

	// VarianceGrowthThreshold is the multiplier above which variance growth
	// is flagged as HIGH_RISK. A value of 2.0 means "variance doubled in
	// one window period" — statistically anomalous for established agents.
	VarianceGrowthThreshold = 2.0
)

// PolicyThreshold defines warning/throttle/block levels for one feature.
type PolicyThreshold struct {
	Warning  float64 `yaml:"warning"`
	Throttle float64 `yaml:"throttle"`
	Block    float64 `yaml:"block"`
}

// PolicyPenalties defines score penalties per enforcement level.
type PolicyPenalties struct {
	Warning  int `yaml:"warning"`
	Throttle int `yaml:"throttle"`
	Block    int `yaml:"block"`
}

// PolicyPack is one YAML policy file loaded into memory.
type PolicyPack struct {
	Name                string                     `yaml:"name"`
	Version             string                     `yaml:"version"`
	OWASPRef            string                     `yaml:"owasp_ref"`
	Description         string                     `yaml:"description"`
	Thresholds          map[string]PolicyThreshold `yaml:"thresholds"`
	ScorePenalties      PolicyPenalties            `yaml:"score_penalties"`
	EnforcementAction   string                     `yaml:"enforcement_action"`
	EscalationAction    string                     `yaml:"escalation_action"`
	HumanReviewRequired bool                       `yaml:"human_review_required"`
	ExplainTemplate     struct {
		PolicyFired       string `yaml:"policy_fired"`
		PolicyThreshold   string `yaml:"policy_threshold"`
		RecommendedAction string `yaml:"recommended_action"`
	} `yaml:"explainability_template"`
}

// PolicyViolation is returned when a policy fires.
type PolicyViolation struct {
	PolicyName          string
	OWASPRef            string
	Feature             string
	Value               float64
	Threshold           float64
	Level               string // "warning", "throttle", "block"
	ScorePenalty        int
	EnforcementAction   string
	HumanReviewRequired bool
	ExplainTemplate     struct {
		PolicyFired       string
		PolicyThreshold   string
		RecommendedAction string
	}
}

// PolicyEngine loads and evaluates all policy packs.
type PolicyEngine struct {
	packs []*PolicyPack
}

// NewPolicyEngine loads all YAML policy packs from a directory.
// FM3 prevention: validates each pack on load, fails fast on bad YAML.
func NewPolicyEngine(policyDir string) (*PolicyEngine, error) {
	files, err := filepath.Glob(filepath.Join(policyDir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("glob policy packs: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no policy packs found in %s", policyDir)
	}

	engine := &PolicyEngine{}
	for _, f := range files {
		pack, err := loadPolicyPack(f)
		if err != nil {
			return nil, fmt.Errorf("load policy %s: %w", f, err)
		}
		engine.packs = append(engine.packs, pack)
	}

	return engine, nil
}

func loadPolicyPack(path string) (*PolicyPack, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read policy file: %w", err)
	}

	var pack PolicyPack
	if err := yaml.Unmarshal(data, &pack); err != nil {
		return nil, fmt.Errorf("parse policy YAML: %w", err)
	}

	if pack.Name == "" {
		return nil, fmt.Errorf("policy pack missing name in %s", path)
	}

	return &pack, nil
}

// Evaluate checks a feature vector against all loaded policy packs.
// Returns all violations found — caller applies worst penalty.
func (e *PolicyEngine) Evaluate(vector FeatureVector) []PolicyViolation {
	features := map[string]float64{
		"tool_call_rate_per_hour":       vector.ToolCallRatePerHour,
		"unique_endpoints_per_hour":     vector.UniqueEndpointsPerHour,
		"bulk_access_count_per_session": vector.BulkAccessCountPerSession,
		"pii_field_access_rate":         vector.PIIFieldAccessRate,
		"cross_tenant_probe_count":      vector.CrossTenantProbeCount,
		"permission_escalation_count":   vector.PermissionEscalationCount,
		"sub_agent_spawn_depth":         vector.SubAgentSpawnDepth,
		"token_refresh_rate":            vector.TokenRefreshRate,
	}

	var violations []PolicyViolation

	for _, pack := range e.packs {
		for feature, threshold := range pack.Thresholds {
			value, ok := features[feature]
			if !ok {
				continue
			}

			var level string
			var thresholdValue float64
			var penalty int

			switch {
			case threshold.Block > 0 && value >= threshold.Block:
				level = "block"
				thresholdValue = threshold.Block
				penalty = pack.ScorePenalties.Block
			case threshold.Throttle > 0 && value >= threshold.Throttle:
				level = "throttle"
				thresholdValue = threshold.Throttle
				penalty = pack.ScorePenalties.Throttle
			case threshold.Warning > 0 && value >= threshold.Warning:
				level = "warning"
				thresholdValue = threshold.Warning
				penalty = pack.ScorePenalties.Warning
			default:
				continue // no violation
			}

			action := pack.EnforcementAction
			if level == "block" {
				action = pack.EscalationAction
			}

			v := PolicyViolation{
				PolicyName:          pack.Name,
				OWASPRef:            pack.OWASPRef,
				Feature:             feature,
				Value:               value,
				Threshold:           thresholdValue,
				Level:               level,
				ScorePenalty:        penalty,
				EnforcementAction:   action,
				HumanReviewRequired: pack.HumanReviewRequired,
			}
			v.ExplainTemplate.PolicyFired = pack.ExplainTemplate.PolicyFired
			v.ExplainTemplate.PolicyThreshold = pack.ExplainTemplate.PolicyThreshold
			v.ExplainTemplate.RecommendedAction = pack.ExplainTemplate.RecommendedAction

			violations = append(violations, v)
		}
	}

	return violations
}

// WorstViolation returns the highest-severity violation.
// Used to determine the primary enforcement action.
func WorstViolation(violations []PolicyViolation) *PolicyViolation {
	if len(violations) == 0 {
		return nil
	}
	worst := &violations[0]
	for i := range violations {
		if violations[i].ScorePenalty < worst.ScorePenalty {
			worst = &violations[i]
		}
	}
	return worst
}

// VarianceGrowthResult holds the outcome of a variance growth rate check.
type VarianceGrowthResult struct {
	Feature      string
	GrowthRate   float64 // ratio of current variance to previous variance
	HighRisk     bool    // true if growth rate exceeds VarianceGrowthThreshold
	PrevVariance float64
	CurrVariance float64
}

// CheckVarianceGrowthRate queries the agent_baselines table for the current
// std_dev and compares it against a snapshot from VarianceWindowDays ago.
// If any feature's variance has grown by more than VarianceGrowthThreshold,
// it returns HIGH_RISK results for those features.
//
// This is the early-warning layer for slow-walk baseline poisoning:
// even as an attacker shifts the baseline mean toward attack behavior,
// the variance of the baseline grows because the new observations are
// further from the original distribution. A 2x variance increase in one
// 7-day window is statistically anomalous for an established agent.
func CheckVarianceGrowthRate(db *sql.DB, orgID, agentDID string) ([]VarianceGrowthResult, error) {
	rows, err := db.Query(`
		SELECT
			ab.feature_name,
			ab.std_dev AS current_std_dev,
			COALESCE(snap.std_dev, ab.std_dev) AS previous_std_dev
		FROM agent_baselines ab
		LEFT JOIN agent_baseline_snapshots snap
			ON snap.org_id = ab.org_id
			AND snap.agent_did = ab.agent_did
			AND snap.feature_name = ab.feature_name
			AND snap.snapshot_date >= NOW() - INTERVAL '1 day' * $3
		WHERE ab.org_id = $1
			AND ab.agent_did = $2
			AND ab.sample_count >= 100
		ORDER BY ab.feature_name`,
		orgID, agentDID, VarianceWindowDays,
	)
	if err != nil {
		return nil, fmt.Errorf("query variance growth: %w", err)
	}
	defer rows.Close()

	var results []VarianceGrowthResult
	for rows.Next() {
		var feature string
		var currStdDev, prevStdDev float64
		if err := rows.Scan(&feature, &currStdDev, &prevStdDev); err != nil {
			return nil, fmt.Errorf("scan variance row: %w", err)
		}

		// Variance = std_dev^2
		currVariance := currStdDev * currStdDev
		prevVariance := prevStdDev * prevStdDev

		// Avoid division by zero for features with no prior variance
		if prevVariance < 0.001 {
			continue
		}

		growthRate := currVariance / prevVariance
		highRisk := growthRate >= VarianceGrowthThreshold

		if highRisk {
			slog.Warn("variance_growth_high_risk",
				"agent_did", agentDID,
				"feature", feature,
				"growth_rate", fmt.Sprintf("%.2f", growthRate),
				"current_variance", fmt.Sprintf("%.4f", currVariance),
				"previous_variance", fmt.Sprintf("%.4f", prevVariance),
				"window_days", VarianceWindowDays,
				"threshold", VarianceGrowthThreshold,
			)
		}

		results = append(results, VarianceGrowthResult{
			Feature:      feature,
			GrowthRate:   growthRate,
			HighRisk:     highRisk,
			PrevVariance: prevVariance,
			CurrVariance: currVariance,
		})
	}

	return results, rows.Err()
}

// CoordinatedAttackResult holds the outcome of a coordinated attack check.
// A coordinated attack is detected when multiple agents in the same org
// show simultaneous anomalous behavior — each looks normal individually
// but the fleet-wide deviation pattern is statistically anomalous.
type CoordinatedAttackResult struct {
	OrgID             string
	DeviatingAgents   int     // number of agents currently in RESTRICTED or BLOCKED band
	TotalAgents       int     // total active agents in org (seen in last 24h)
	DeviationRate     float64 // deviatingAgents / totalAgents
	IsCoordinated     bool    // true if deviation rate exceeds threshold
	ThresholdExceeded float64 // the threshold that was exceeded
}

const (
	// CoordinatedAttackThreshold is the fraction of org agents that must be
	// simultaneously anomalous to trigger a coordinated attack alert.
	// 0.25 = 25% of fleet deviating simultaneously = statistically anomalous.
	CoordinatedAttackThreshold = 0.25

	// CoordinatedAttackMinAgents is the minimum fleet size before the check fires.
	// Prevents false positives on small orgs (1-2 agents, both happen to spike).
	CoordinatedAttackMinAgents = 4
)

// CheckCoordinatedAttack detects simultaneous multi-agent deviation in an org.
// Queries agent_identities for agents with score < 500 (RESTRICTED/BLOCKED band).
// If ≥25% of the fleet is simultaneously anomalous → coordinated attack signal.
// M5-STEP-2: new detection claim — "ARE detects both individual and coordinated attacks."
func CheckCoordinatedAttack(db *sql.DB, orgID string) (*CoordinatedAttackResult, error) {
	result := &CoordinatedAttackResult{
		OrgID:             orgID,
		ThresholdExceeded: CoordinatedAttackThreshold,
	}

	// Count total active agents (seen in last 24 hours)
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM agent_identities
		WHERE org_id = $1
		  AND last_seen > NOW() - INTERVAL '24 hours'`,
		orgID,
	).Scan(&result.TotalAgents)
	if err != nil {
		return nil, fmt.Errorf("coordinated attack total count: %w", err)
	}

	// Not enough agents to detect coordinated attack
	if result.TotalAgents < CoordinatedAttackMinAgents {
		return result, nil
	}

	// Count agents currently in RESTRICTED or BLOCKED band (score < 500)
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM agent_identities
		WHERE org_id = $1
		  AND last_seen > NOW() - INTERVAL '24 hours'
		  AND current_score < 500`,
		orgID,
	).Scan(&result.DeviatingAgents)
	if err != nil {
		return nil, fmt.Errorf("coordinated attack deviation count: %w", err)
	}

	result.DeviationRate = float64(result.DeviatingAgents) / float64(result.TotalAgents)
	result.IsCoordinated = result.DeviationRate >= CoordinatedAttackThreshold

	if result.IsCoordinated {
		slog.Warn("coordinated_attack_detected",
			"org_id", orgID,
			"deviating_agents", result.DeviatingAgents,
			"total_agents", result.TotalAgents,
			"deviation_rate", fmt.Sprintf("%.2f", result.DeviationRate),
			"threshold", CoordinatedAttackThreshold,
		)
	}

	return result, nil
}
