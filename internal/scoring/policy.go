package scoring

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
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
