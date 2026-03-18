package scoring

import (
	"fmt"
	"math"
	"os"

	"gopkg.in/yaml.v3"
)

// ScoringConfig holds all tunable scoring parameters.
// Loaded from config/scoring_weights.yaml at startup.
type ScoringConfig struct {
	Weights struct {
		Historical float64 `yaml:"historical"`
		Velocity   float64 `yaml:"velocity"`
	} `yaml:"weights"`
	DecayRate                float64 `yaml:"decay_rate"`
	ZScoreThreshold          float64 `yaml:"zscore_threshold"`
	PenaltyPerSigma          float64 `yaml:"penalty_per_sigma"`
	MaxPenalty               float64 `yaml:"max_penalty"`
	MinStdDev                float64 `yaml:"min_std_dev"`
	BootstrapSampleThreshold int     `yaml:"bootstrap_sample_threshold"`
}

// LoadScoringConfig reads and validates the scoring config.
// FM3 prevention: refuses to load if weights do not sum to 1.0.
func LoadScoringConfig(path string) (*ScoringConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read scoring config: %w", err)
	}

	var cfg ScoringConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse scoring config: %w", err)
	}

	// Validate weights sum to 1.0
	sum := cfg.Weights.Historical + cfg.Weights.Velocity
	if math.Abs(sum-1.0) > 0.001 {
		return nil, fmt.Errorf(
			"scoring weights must sum to 1.0, got %.3f (H=%.3f V=%.3f)",
			sum, cfg.Weights.Historical, cfg.Weights.Velocity,
		)
	}

	return &cfg, nil
}

// ToScoreWeights converts config to the ScoreWeights struct.
func (c *ScoringConfig) ToScoreWeights() ScoreWeights {
	return ScoreWeights{
		Historical: c.Weights.Historical,
		Velocity:   c.Weights.Velocity,
	}
}
