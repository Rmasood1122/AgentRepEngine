// internal/scoring/context.go
// #4 Temporal context multiplier
// Raises the anomaly threshold when context predicts legitimate high activity.
// Eliminates schedule-driven FPs — batch jobs bursting at their scheduled time
// should not trigger z-score enforcement.
// Config-driven via scoring_weights.yaml scheduled_jobs section.
// Zero infrastructure changes — additive only.

package scoring

import (
	"math"
	"time"
)

// ScheduledJob declares a known burst window for a specific agent type.
// Configured in config/scoring_weights.yaml under scheduled_jobs.
type ScheduledJob struct {
	AgentType      string  `yaml:"agent_type"`
	CronHour       int     `yaml:"cron_hour"`       // 0-23 UTC
	CronMinute     int     `yaml:"cron_minute"`     // 0-59
	WindowMinutes  int     `yaml:"window_minutes"`  // burst window size
	RateMultiplier float64 `yaml:"rate_multiplier"` // max threshold raise
}

// ContextualThresholdMultiplier returns a multiplier > 1.0 when the current
// time context predicts legitimate high-activity behavior for this agent type.
// The multiplier raises the effective z-score threshold, preventing FPs during
// known burst windows.
//
// Example: batch_processor scheduled at 02:00 UTC with window_minutes=60
// and rate_multiplier=5.0 will have its threshold raised 5x between 01:30-03:00.
// A burst that would normally fire at z=3.0 now requires z=15.0 to fire.
//
// Returns 1.0 (no change) when no scheduled job context applies.
// Hard ceiling: never returns more than MaxContextMultiplier.
func ContextualThresholdMultiplier(agentType string, t time.Time, jobs []ScheduledJob) float64 {
	const MaxContextMultiplier = 8.0

	multiplier := 1.0

	for _, job := range jobs {
		if job.AgentType != agentType {
			continue
		}

		// Compute minutes since/until scheduled job time today
		jobTime := time.Date(t.Year(), t.Month(), t.Day(),
			job.CronHour, job.CronMinute, 0, 0, time.UTC)

		// Also check yesterday's job (for jobs near midnight)
		jobTimeYesterday := jobTime.Add(-24 * time.Hour)
		jobTimeTomorrow := jobTime.Add(24 * time.Hour)

		for _, jt := range []time.Time{jobTimeYesterday, jobTime, jobTimeTomorrow} {
			diff := t.Sub(jt).Minutes()
			windowHalf := float64(job.WindowMinutes) / 2.0

			if math.Abs(diff) <= windowHalf {
				// Inside the burst window — compute proximity-weighted multiplier
				// Peaks at job time, tapers to 1.0 at window edges
				proximity := 1.0 - (math.Abs(diff) / windowHalf)
				candidateMultiplier := 1.0 + (job.RateMultiplier-1.0)*proximity
				if candidateMultiplier > multiplier {
					multiplier = candidateMultiplier
				}
			}
		}
	}

	return math.Min(multiplier, MaxContextMultiplier)
}

// FleetActivityMultiplier returns a multiplier > 1.0 when a large fraction
// of the agent fleet is simultaneously elevated.
// If >50% of agents of same type are elevated, it is likely a fleet event
// (deployment, market open, scheduled maintenance) — not an attack.
// Prevents the entire class of correlated legitimate burst FPs.
//
// elevatedFraction: fraction of same-type agents currently above normal rate (0.0-1.0)
// Returns 1.0 when <25% of fleet is elevated (no adjustment).
// Returns up to MaxFleetMultiplier when >75% of fleet is elevated.
func FleetActivityMultiplier(elevatedFraction float64) float64 {
	const (
		FleetThresholdLow  = 0.25 // below this: no adjustment
		FleetThresholdHigh = 0.75 // above this: maximum adjustment
		MaxFleetMultiplier = 4.0
	)

	if elevatedFraction < FleetThresholdLow {
		return 1.0
	}

	// Linear scale from 1.0 at 25% to MaxFleetMultiplier at 75%
	fraction := (elevatedFraction - FleetThresholdLow) /
		(FleetThresholdHigh - FleetThresholdLow)
	fraction = math.Min(fraction, 1.0)

	return 1.0 + (MaxFleetMultiplier-1.0)*fraction
}

// EffectiveZScoreThreshold returns the adjusted z-score threshold after
// applying all context multipliers. The base threshold is multiplied by
// the highest applicable context multiplier.
//
// Usage in scorer.go:
//
//	threshold := EffectiveZScoreThreshold(
//	    s.config.ZScoreThreshold,
//	    agentType,
//	    time.Now().UTC(),
//	    s.config.ScheduledJobs,
//	    elevatedFraction,
//	)
//	if blendedZ > threshold { ... apply penalty ... }
func EffectiveZScoreThreshold(
	baseThreshold float64,
	agentType string,
	t time.Time,
	scheduledJobs []ScheduledJob,
	elevatedFraction float64,
) float64 {
	temporalMult := ContextualThresholdMultiplier(agentType, t, scheduledJobs)
	fleetMult := FleetActivityMultiplier(elevatedFraction)

	// Use the highest applicable multiplier — not additive
	// Rationale: context signals are independent; the strongest one applies
	multiplier := math.Max(temporalMult, fleetMult)

	return baseThreshold * multiplier
}
