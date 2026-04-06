package scoring

import (
	"database/sql"
	"fmt"
	"math"
	"time"
)

// ChangeClassification is the output of the two-speed baseline classifier.
type ChangeClassification string

const (
	ClassAttack           ChangeClassification = "ATTACK"
	ClassLegitimateChange ChangeClassification = "LEGITIMATE_CHANGE"
	ClassNormal           ChangeClassification = "NORMAL"
)

// TwoSpeedBaseline holds fast (3-day EWMA) and slow (30-day rolling) state.
type TwoSpeedBaseline struct {
	AgentID    string
	FastEWMA   float64
	SlowMean   float64
	SlowStdDev float64
	UpdatedAt  time.Time
}

// ClassifyChange compares a new score against the two-speed baseline.
// Fast alpha=0.3 (3-day sensitivity), slow=30-day rolling.
// Returns ATTACK | LEGITIMATE_CHANGE | NORMAL
func ClassifyChange(fast, slowMean, slowStdDev, newScore float64) ChangeClassification {
	// Update fast EWMA
	const alpha = 0.3
	updatedFast := alpha*newScore + (1-alpha)*fast

	// Divergence between fast and slow
	divergence := math.Abs(updatedFast - slowMean)

	// Threshold: 2.5 sigma for ATTACK, 1.5 sigma for LEGITIMATE_CHANGE
	if slowStdDev == 0 {
		slowStdDev = 0.01 // avoid divide by zero
	}
	sigmas := divergence / slowStdDev

	switch {
	case sigmas >= 2.5 && newScore > slowMean:
		return ClassAttack
	case sigmas >= 1.5:
		return ClassLegitimateChange
	default:
		return ClassNormal
	}
}

// LoadTwoSpeedBaseline fetches the current fast/slow state for an agent.
func LoadTwoSpeedBaseline(db *sql.DB, agentID string) (*TwoSpeedBaseline, error) {
	var b TwoSpeedBaseline
	var fastEWMA sql.NullFloat64

	err := db.QueryRow(`
		SELECT agent_id, fast_ewma, slow_mean, slow_stddev, updated_at
		FROM agent_baselines
		WHERE agent_id = $1
	`, agentID).Scan(&b.AgentID, &fastEWMA, &b.SlowMean, &b.SlowStdDev, &b.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("agent baseline not found: %s", agentID)
	}
	if err != nil {
		return nil, fmt.Errorf("load two-speed baseline: %w", err)
	}

	if fastEWMA.Valid {
		b.FastEWMA = fastEWMA.Float64
	}
	return &b, nil
}

// UpdateFastEWMA persists the updated fast EWMA after a scoring event.
func UpdateFastEWMA(db *sql.DB, agentID string, newScore float64, currentFast float64) error {
	const alpha = 0.3
	updated := alpha*newScore + (1-alpha)*currentFast

	_, err := db.Exec(`
		UPDATE agent_baselines
		SET fast_ewma = $1, updated_at = $2
		WHERE agent_id = $3
	`, updated, time.Now().UTC(), agentID)

	if err != nil {
		return fmt.Errorf("update fast ewma: %w", err)
	}
	return nil
}
