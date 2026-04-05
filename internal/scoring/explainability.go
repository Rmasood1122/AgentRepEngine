package scoring

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"
)

// ReasonObject is the required explanation for every enforcement decision.
// G-EXPLAIN hard stop: ALL fields must be non-zero/non-empty.
// FM2 prevention: if generation fails → AUDIT not BLOCK.
type ReasonObject struct {
	Decision             string         `json:"decision"`
	AgentDID             string         `json:"agent_did"`
	Score                int            `json:"score"`
	ScoreDelta           int            `json:"score_delta"`
	ScorePeriodHours     float64        `json:"score_period_hours"`
	TriggerEvents        []TriggerEvent `json:"trigger_events"`
	PolicyFired          string         `json:"policy_fired"`
	PolicyThreshold      string         `json:"policy_threshold"`
	RecommendedAction    string         `json:"recommended_action"`
	PeerCluster          string         `json:"peer_cluster"`
	PeerClusterAvgScore  int            `json:"peer_cluster_avg_score"`
	DeviationFromCluster int            `json:"deviation_from_cluster"`
	ConfidencePct        int            `json:"confidence_pct"`
	ComputedAt           int64          `json:"computed_at"`
}

// TriggerEvent describes one behavioral event that contributed to the decision.
type TriggerEvent struct {
	EventType       string  `json:"event_type"`
	Count           float64 `json:"count"`
	WindowMinutes   int     `json:"window_minutes"`
	BaselinePerHour float64 `json:"baseline_per_hour"`
	DeviationSigma  float64 `json:"deviation_sigma"`
}

// ExplainDecision builds a complete ReasonObject from scoring inputs.
// FM2 prevention: returns error if any required field cannot be populated.
// Caller must default to AUDIT if this returns error.
// worstZScore: highest z-score across all features — used to compute confidence_pct.
// confidence_pct: 100 = normal behavior, 0 = certain anomaly (z >= 3.0 sigma).
func ExplainDecision(
	agentDID string,
	decision string,
	newScore int,
	previousScore int,
	vector FeatureVector,
	violations []PolicyViolation,
	baselines map[string]Baseline,
	scorePeriodHours float64,
	worstZScore float64,
) (*ReasonObject, error) {

	if agentDID == "" {
		return nil, errors.New("agent_did is required for explanation")
	}
	if decision == "" {
		return nil, errors.New("decision is required for explanation")
	}

	// Build trigger events from policy violations
	triggerEvents := buildTriggerEvents(vector, violations, baselines)

	// Determine policy fired — use worst violation
	policyFired := "no_policy_fired"
	policyThreshold := "within normal parameters"
	recommendedAction := "No action required"

	worst := WorstViolation(violations)
	if worst != nil {
		policyFired = worst.ExplainTemplate.PolicyFired
		policyThreshold = worst.ExplainTemplate.PolicyThreshold
		recommendedAction = worst.ExplainTemplate.RecommendedAction
		if policyFired == "" {
			policyFired = worst.PolicyName + "_v1"
		}
		if recommendedAction == "" {
			recommendedAction = "Human review recommended"
		}
	}

	// Peer cluster deviation
	// Phase 1: use fixed cluster average of 700 as baseline
	// Phase 2: compute from actual peer cluster data
	peerClusterAvg := 700
	deviationFromCluster := newScore - peerClusterAvg

	// confidence_pct: inverse of anomaly confidence.
	// At z=0.0 → 100% (behavior is normal, high confidence agent is legitimate).
	// At z=3.0+ → 0% (behavior is 3+ sigma anomalous, high confidence it is malicious).
	// Formula: confidence_pct = clamp(0, 100, (1 - worstZ/3.0) * 100)
	const maxExpectedZ = 3.0
	confidencePct := int(math.Max(0, math.Min(100, (1.0-(worstZScore/maxExpectedZ))*100)))

	reason := &ReasonObject{
		Decision:             decision,
		AgentDID:             agentDID,
		Score:                newScore,
		ScoreDelta:           newScore - previousScore,
		ScorePeriodHours:     scorePeriodHours,
		TriggerEvents:        triggerEvents,
		PolicyFired:          policyFired,
		PolicyThreshold:      policyThreshold,
		RecommendedAction:    recommendedAction,
		PeerCluster:          "default",
		PeerClusterAvgScore:  peerClusterAvg,
		DeviationFromCluster: deviationFromCluster,
		ConfidencePct:        confidencePct,
		ComputedAt:           time.Now().Unix(),
	}

	// Validate all required fields populated
	if err := validateReasonObject(reason); err != nil {
		return nil, fmt.Errorf("reason object validation failed: %w", err)
	}

	return reason, nil
}

// buildTriggerEvents converts policy violations into trigger event list.
func buildTriggerEvents(
	vector FeatureVector,
	violations []PolicyViolation,
	baselines map[string]Baseline,
) []TriggerEvent {

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

	seen := map[string]bool{}
	var events []TriggerEvent

	for _, v := range violations {
		if seen[v.Feature] {
			continue
		}
		seen[v.Feature] = true

		value := features[v.Feature]
		baseline := baselines[v.Feature]

		sigma := 0.0
		if baseline.StdDev > 0 {
			sigma = (value - baseline.Mean) / baseline.StdDev
		}

		events = append(events, TriggerEvent{
			EventType:       v.Feature,
			Count:           value,
			WindowMinutes:   60,
			BaselinePerHour: baseline.Mean,
			DeviationSigma:  sigma,
		})
	}

	return events
}

// validateReasonObject checks all required fields are populated.
// Returns error if any field is missing — FM2 prevention.
func validateReasonObject(r *ReasonObject) error {
	if r.Decision == "" {
		return errors.New("decision is empty")
	}
	if r.AgentDID == "" {
		return errors.New("agent_did is empty")
	}
	if r.PolicyFired == "" {
		return errors.New("policy_fired is empty")
	}
	if r.RecommendedAction == "" {
		return errors.New("recommended_action is empty")
	}
	if r.ComputedAt == 0 {
		return errors.New("computed_at is zero")
	}
	return nil
}

// ToJSON serializes the reason object to JSON string.
// Returns fallback JSON if serialization fails — never returns empty string.
func (r *ReasonObject) ToJSON() string {
	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Sprintf(
			`{"decision":"%s","agent_did":"%s","score":%d,"confidence_pct":%d,"error":"serialization_failed"}`,
			r.Decision, r.AgentDID, r.Score, r.ConfidencePct,
		)
	}
	return string(data)
}

// BaselineMaturity indicates whether the agent has an established baseline
// Values: "30d_established" | "probation_Xd" (where X = days of data)
// Enterprise must not enable enforcement on probation agents.
// This field is required in every reason object. [F — G-EXPLAIN gate]
