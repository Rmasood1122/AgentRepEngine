package scoring

import (
	"fmt"
	"log/slog"
	"math"
	"time"
)

// ScoredResult is the output of a full scoring computation.
type ScoredResult struct {
	AgentDID     string
	Score        int
	Band         string
	ScoreDelta   int
	WorstFeature string
	WorstZScore  float64
	Penalty      float64
	Explanation  map[string]interface{}
}

// Scorer computes behavioral scores using live baselines.
type Scorer struct {
	baselines *BaselineStore
	config    *ScoringConfig
}

// NewScorer creates a scorer with live baselines and config.
func NewScorer(baselines *BaselineStore, config *ScoringConfig) *Scorer {
	return &Scorer{
		baselines: baselines,
		config:    config,
	}
}

// Score computes a new score for an agent given a feature vector.
// Uses live baselines from DB — updates agent baselines after scoring.
func (s *Scorer) Score(orgID, agentDID string, vector FeatureVector,
	previousScore int, lastSeen time.Time) (*ScoredResult, error) {

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

	// Apply decay to historical score
	// FM2 prevention: decay applied on every computation
	H := ApplyDecay(float64(previousScore), lastSeen)

	// Compute worst z-score across all features
	worstZ := 0.0
	worstFeature := ""
	featureZScores := map[string]float64{}

	for feature, value := range features {
		baseline := s.baselines.GetBaseline(orgID, agentDID, feature)
		z := ComputeZScore(value, baseline)
		featureZScores[feature] = z

		if z > worstZ {
			worstZ = z
			worstFeature = feature
		}

		// Update agent baseline with this observation
		if err := s.baselines.UpdateAgentBaseline(orgID, agentDID, feature, value); err != nil {
			slog.Warn("baseline update failed",
				"agent_did", agentDID,
				"feature", feature,
				"error", err,
			)
		}

		// Update cluster baseline
		s.baselines.UpdateClusterBaseline("default", feature, value)
	}

	// Compute velocity penalty
	penalty := 0.0
	if worstZ > s.config.ZScoreThreshold {
		penalty = math.Min(
			s.config.PenaltyPerSigma*(worstZ-s.config.ZScoreThreshold),
			s.config.MaxPenalty,
		)
	}
	V := math.Max(0, 1000-penalty*3)

	// Final score
	weights := s.config.ToScoreWeights()
	newScore := ComputeScore(H, V, weights)
	band := ScoreBand(newScore)
	delta := newScore - previousScore

	explanation := map[string]interface{}{
		"decision":         band,
		"agent_did":        agentDID,
		"score":            newScore,
		"score_delta":      delta,
		"historical_input": H,
		"velocity_input":   V,
		"worst_feature":    worstFeature,
		"worst_z_score":    fmt.Sprintf("%.2f", worstZ),
		"penalty":          penalty,
		"feature_z_scores": featureZScores,
		"weights":          map[string]float64{"H": weights.Historical, "V": weights.Velocity},
		"computed_at":      time.Now().Unix(),
	}

	slog.Info("score_computed",
		"agent_did", agentDID,
		"previous_score", previousScore,
		"new_score", newScore,
		"band", band,
		"worst_feature", worstFeature,
		"worst_z", worstZ,
		"delta", delta,
	)

	return &ScoredResult{
		AgentDID:     agentDID,
		Score:        newScore,
		Band:         band,
		ScoreDelta:   delta,
		WorstFeature: worstFeature,
		WorstZScore:  worstZ,
		Penalty:      penalty,
		Explanation:  explanation,
	}, nil
}
