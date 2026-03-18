package scoring

import "math"

type ScoreWeights struct {
	Historical float64
	Velocity   float64
}

var DefaultWeights = ScoreWeights{Historical: 0.5, Velocity: 0.5}

func HistoricalDecay(prev, days float64) float64 {
	return prev * math.Exp(-0.1*days)
}

func VelocityScore(observed, mean, std float64) float64 {
	if std == 0 {
		return 1000
	}
	z := (observed - mean) / std
	z = math.Max(0, math.Min(z, 3.0))
	return 1000 * (1 - z/3.0)
}

func ComputeScore(historical, velocity float64, w ScoreWeights) int {
	raw := w.Historical*historical + w.Velocity*velocity
	return int(math.Max(0, math.Min(raw, 1000)))
}

func ScoreBand(score int) string {
	switch {
	case score >= 800:
		return "TRUSTED"
	case score >= 500:
		return "MONITORED"
	case score >= 200:
		return "RESTRICTED"
	default:
		return "BLOCKED"
	}
}
