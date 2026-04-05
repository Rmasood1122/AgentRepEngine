package scoring

import (
	"database/sql"
	"fmt"
	"log/slog"
	"math"
	"time"
)

// Baseline holds mean and standard deviation for one feature.
type Baseline struct {
	Mean        float64
	StdDev      float64
	SampleCount int
	LastUpdated time.Time
}

// BaselineStore reads and writes behavioral baselines.
// Per-agent baselines used when sample_count >= 100.
// Cluster baselines used as fallback (bootstrap period).
type BaselineStore struct {
	db *sql.DB
}

// NewBaselineStore creates a new baseline store.
func NewBaselineStore(db *sql.DB) *BaselineStore {
	return &BaselineStore{db: db}
}

// GetBaseline returns the best available baseline for an agent+feature.
// Keyed on (org_id, agent_did, feature_name) — org-scoped isolation.
// Priority: agent-level (if >=100 samples) → cluster-level → hardcoded.
// FM1 prevention: never use agent baseline with <100 samples.
func (s *BaselineStore) GetBaseline(orgID, agentDID, featureName string) Baseline {
	// Try agent-level baseline first — scoped to org_id + agent_did
	var mean, stdDev float64
	var sampleCount int
	err := s.db.QueryRow(`
		SELECT mean, std_dev, sample_count
		FROM agent_baselines
		WHERE org_id = $1 AND agent_did = $2 AND feature_name = $3`,
		orgID, agentDID, featureName,
	).Scan(&mean, &stdDev, &sampleCount)

	if err == nil && sampleCount >= 100 {
		return Baseline{Mean: mean, StdDev: stdDev, SampleCount: sampleCount}
	}

	// Fall back to cluster baseline
	err = s.db.QueryRow(`
		SELECT mean, std_dev, sample_count
		FROM cluster_baselines
		WHERE cluster_id = 'default' AND feature_name = $1`,
		featureName,
	).Scan(&mean, &stdDev, &sampleCount)

	if err == nil {
		return Baseline{Mean: mean, StdDev: stdDev, SampleCount: sampleCount}
	}

	// Final fallback — hardcoded defaults from APEX v5.2
	return hardcodedBaseline(featureName)
}

// GetClusterBaseline returns the cluster-level baseline for a feature directly.
// Used by M5-STEP-1 peer cluster deviation scoring — bypasses agent-level lookup.
// Falls back to hardcoded defaults if cluster_baselines table has no data yet.
func (s *BaselineStore) GetClusterBaseline(clusterID, featureName string) Baseline {
	var mean, stdDev float64
	var sampleCount int
	err := s.db.QueryRow(`
		SELECT mean, std_dev, sample_count
		FROM cluster_baselines
		WHERE cluster_id = $1 AND feature_name = $2`,
		clusterID, featureName,
	).Scan(&mean, &stdDev, &sampleCount)
	if err == nil {
		return Baseline{Mean: mean, StdDev: stdDev, SampleCount: sampleCount}
	}
	return hardcodedBaseline(featureName)
}

// UpdateAgentBaseline updates the running mean and std dev for an agent.
// Keyed on (org_id, agent_did, feature_name) — org-scoped isolation.
// Uses Welford's online algorithm — no need to store all historical values.
func (s *BaselineStore) UpdateAgentBaseline(orgID, agentDID, featureName string,
	newValue float64) error {

	_, err := s.db.Exec(`
		INSERT INTO agent_baselines
			(org_id, agent_did, feature_name, mean, std_dev, sample_count, last_updated)
		VALUES ($1, $2, $3, $4, 1.0, 1, NOW())
		ON CONFLICT (org_id, agent_did, feature_name) DO UPDATE SET
			sample_count = agent_baselines.sample_count + 1,
			mean = agent_baselines.mean +
				($4 - agent_baselines.mean) /
				(agent_baselines.sample_count + 1),
			std_dev = CASE
				WHEN agent_baselines.sample_count < 2 THEN 1.0
				ELSE GREATEST(
					SQRT(
						(agent_baselines.sample_count - 1) *
						POWER(agent_baselines.std_dev, 2) /
						agent_baselines.sample_count +
						POWER($4 - agent_baselines.mean, 2) /
						agent_baselines.sample_count
					), 0.1
				)
			END,
			last_updated = NOW()`,
		orgID, agentDID, featureName, newValue,
	)
	if err != nil {
		return fmt.Errorf("update agent baseline: %w", err)
	}
	return nil
}

// UpdateClusterBaseline updates the cluster-level baseline.
// Called by the hourly baseline updater job.
func (s *BaselineStore) UpdateClusterBaseline(clusterID, featureName string,
	newValue float64) error {

	_, err := s.db.Exec(`
		INSERT INTO cluster_baselines
			(cluster_id, feature_name, mean, std_dev, sample_count, last_updated)
		VALUES ($1, $2, $3, 1.0, 1, NOW())
		ON CONFLICT (cluster_id, feature_name) DO UPDATE SET
			sample_count = cluster_baselines.sample_count + 1,
			mean = cluster_baselines.mean +
				($3 - cluster_baselines.mean) /
				(cluster_baselines.sample_count + 1),
			std_dev = CASE
				WHEN cluster_baselines.sample_count < 2 THEN 1.0
				ELSE GREATEST(
					SQRT(
						(cluster_baselines.sample_count - 1) *
						POWER(cluster_baselines.std_dev, 2) /
						cluster_baselines.sample_count +
						POWER($3 - cluster_baselines.mean, 2) /
						cluster_baselines.sample_count
					), 0.1
				)
			END,
			last_updated = NOW()`,
		clusterID, featureName, newValue,
	)
	if err != nil {
		return fmt.Errorf("update cluster baseline: %w", err)
	}
	return nil
}

// ComputeZScore returns (value - mean) / std_dev.
// Returns 0 if std_dev is effectively zero — no signal yet.
func ComputeZScore(value float64, b Baseline) float64 {
	if b.StdDev < 0.01 {
		return 0
	}
	return (value - b.Mean) / b.StdDev
}

// ApplyDecay applies exponential decay to historical score.
// H(t) = H(t-1) * e^(-0.1 * days_since_last_event)
// FM2 prevention: decay applied on every score computation.
func ApplyDecay(previousScore float64, lastSeen time.Time) float64 {
	days := time.Since(lastSeen).Hours() / 24
	if days < 0.01 {
		return previousScore // same day — no decay
	}
	decayed := previousScore * math.Exp(-0.1*days)
	slog.Info("score_decay_applied",
		"previous_score", previousScore,
		"days_since_seen", days,
		"decayed_score", decayed,
	)
	return decayed
}

// hardcodedBaseline returns APEX v5.2 spec defaults.
// Last resort — only used if DB is unavailable.
func hardcodedBaseline(featureName string) Baseline {
	defaults := map[string]Baseline{
		"tool_call_rate_per_hour":       {Mean: 80, StdDev: 60},
		"unique_endpoints_per_hour":     {Mean: 25, StdDev: 30},
		"bulk_access_count_per_session": {Mean: 400, StdDev: 400},
		"pii_field_access_rate":         {Mean: 20, StdDev: 25},
		"cross_tenant_probe_count":      {Mean: 0, StdDev: 0.1},
		"permission_escalation_count":   {Mean: 0, StdDev: 0.5},
		"sub_agent_spawn_depth":         {Mean: 0, StdDev: 0.3},
		"token_refresh_rate":            {Mean: 1, StdDev: 1},
	}
	if b, ok := defaults[featureName]; ok {
		return b
	}
	return Baseline{Mean: 0, StdDev: 1}
}
