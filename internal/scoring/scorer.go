package scoring

import (
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/agentrepengine/are/internal/metrics"
)

// ScoredResult is the output of a full scoring computation.
type ScoredResult struct {
	AgentDID      string
	Score         int
	Band          string
	ScoreDelta    int
	WorstFeature  string
	WorstZScore   float64
	Penalty       float64
	Explanation   map[string]interface{}
	HistoryScore  int
	VelocityScore int
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
	clusterWorstZ := 0.0 // M5-STEP-1: worst z-score vs cluster baseline
	featureZScores := map[string]float64{}

	for feature, value := range features {
		baseline := s.baselines.GetBaseline(orgID, agentDID, feature)
		z := ComputeZScore(value, baseline, s.config.MinStdDev)
		featureZScores[feature] = z

		if z > worstZ {
			worstZ = z
			worstFeature = feature
		}

		// M5-STEP-1: compute cluster z-score for peer deviation signal
		clusterBaseline := s.baselines.GetClusterBaseline("default", feature)
		clusterZ := ComputeZScore(value, clusterBaseline, s.config.MinStdDev)
		if clusterZ > clusterWorstZ {
			clusterWorstZ = clusterZ
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

	// Compute velocity penalty — M5-STEP-1: blend own baseline (70%) + cluster (30%)
	// Phase 1: own baseline only used when cluster has <100 samples (bootstrap period)
	// Phase 2: full blend once cluster has 100+ samples per feature
	blendedZ := worstZ
	clusterBaseline := s.baselines.GetClusterBaseline("default", "tool_call_rate_per_hour")
	if clusterBaseline.SampleCount >= 100 {
		// Cluster is established — apply peer deviation blend
		blendedZ = worstZ*0.7 + clusterWorstZ*0.3
	}

	penalty := 0.0
	if blendedZ > s.config.ZScoreThreshold {
		penalty = math.Min(
			s.config.PenaltyPerSigma*(blendedZ-s.config.ZScoreThreshold),
			s.config.MaxPenalty,
		)
		metrics.AgentAnomaliesTotal.WithLabelValues(agentDID).Inc()
	}
	V := math.Max(0, 1000-penalty*3)

	// Final score
	weights := s.config.ToScoreWeights()
	newScore := ComputeScore(H, V, weights)

	// #2 max_score_after_violation cap
	// Prevents trust-shield attacks where high historical trust (H=900)
	// masks a confirmed z-score violation. A trusted agent that just
	// triggered a genuine anomaly must drop to RESTRICTED minimum.
	// Without this cap: trusted agent at H=900 + V penalty stays at ~650 MONITORED.
	// With this cap: drops to ≤499 RESTRICTED — forces human review.
	const MaxScoreAfterViolation = 499
	if blendedZ > s.config.ZScoreThreshold && newScore > MaxScoreAfterViolation {
		slog.Warn("trust_shield_cap_applied",
			"agent_did", agentDID,
			"pre_cap_score", newScore,
			"capped_to", MaxScoreAfterViolation,
			"blended_z", blendedZ,
		)
		newScore = MaxScoreAfterViolation
	}

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
		"blended_z":        fmt.Sprintf("%.2f", blendedZ),
		"cluster_worst_z":  fmt.Sprintf("%.2f", clusterWorstZ),
		"cluster_samples":  clusterBaseline.SampleCount,
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
		AgentDID:      agentDID,
		Score:         newScore,
		Band:          band,
		ScoreDelta:    delta,
		WorstFeature:  worstFeature,
		WorstZScore:   worstZ,
		Penalty:       penalty,
		Explanation:   explanation,
		HistoryScore:  int(H),
		VelocityScore: int(V),
	}, nil
}
