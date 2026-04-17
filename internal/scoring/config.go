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

	// A1 Hardening Sprint: validate all field ranges, not just weight sum
	if err := validateScoringConfig(&cfg); err != nil {
		return nil, fmt.Errorf("scoring config validation failed: %w", err)
	}

	return &cfg, nil
}

// validateScoringConfig checks all scoring parameters are within valid ranges.
// Prevents silent misconfiguration that causes incorrect enforcement decisions.
func validateScoringConfig(cfg *ScoringConfig) error {
	if cfg.DecayRate < 0 {
		return fmt.Errorf("decay_rate must be >= 0, got %.4f", cfg.DecayRate)
	}
	if cfg.ZScoreThreshold <= 0 {
		return fmt.Errorf("zscore_threshold must be > 0, got %.4f", cfg.ZScoreThreshold)
	}
	if cfg.PenaltyPerSigma < 0 {
		return fmt.Errorf("penalty_per_sigma must be >= 0, got %.4f", cfg.PenaltyPerSigma)
	}
	if cfg.MaxPenalty < 0 {
		return fmt.Errorf("max_penalty must be >= 0, got %.4f", cfg.MaxPenalty)
	}
	if cfg.MinStdDev <= 0 {
		return fmt.Errorf("min_std_dev must be > 0, got %.4f", cfg.MinStdDev)
	}
	if cfg.BootstrapSampleThreshold < 0 {
		return fmt.Errorf("bootstrap_sample_threshold must be >= 0, got %d", cfg.BootstrapSampleThreshold)
	}
	return nil
}

// ToScoreWeights converts config to the ScoreWeights struct.
func (c *ScoringConfig) ToScoreWeights() ScoreWeights {
	return ScoreWeights{
		Historical: c.Weights.Historical,
		Velocity:   c.Weights.Velocity,
	}
}
