package scoring

import (
	"testing"
	"time"
)

func TestContextualThresholdMultiplier_NoJobs(t *testing.T) {
	mult := ContextualThresholdMultiplier("batch_processor", time.Now().UTC(), nil)
	if mult != 1.0 {
		t.Errorf("expected 1.0 with no jobs, got %.2f", mult)
	}
}

func TestContextualThresholdMultiplier_InsideWindow(t *testing.T) {
	jobs := []ScheduledJob{
		{
			AgentType:      "batch_processor",
			CronHour:       2,
			CronMinute:     0,
			WindowMinutes:  60,
			RateMultiplier: 5.0,
		},
	}
	// Exactly at job time — should be near max multiplier
	jobTime := time.Date(2026, 1, 1, 2, 0, 0, 0, time.UTC)
	mult := ContextualThresholdMultiplier("batch_processor", jobTime, jobs)
	if mult < 4.0 {
		t.Errorf("expected high multiplier at job time, got %.2f", mult)
	}
}

func TestContextualThresholdMultiplier_OutsideWindow(t *testing.T) {
	jobs := []ScheduledJob{
		{
			AgentType:      "batch_processor",
			CronHour:       2,
			CronMinute:     0,
			WindowMinutes:  60,
			RateMultiplier: 5.0,
		},
	}
	// 3 hours after job time — outside window
	outsideTime := time.Date(2026, 1, 1, 5, 0, 0, 0, time.UTC)
	mult := ContextualThresholdMultiplier("batch_processor", outsideTime, jobs)
	if mult != 1.0 {
		t.Errorf("expected 1.0 outside window, got %.2f", mult)
	}
}

func TestContextualThresholdMultiplier_WrongAgentType(t *testing.T) {
	jobs := []ScheduledJob{
		{
			AgentType:      "batch_processor",
			CronHour:       2,
			CronMinute:     0,
			WindowMinutes:  60,
			RateMultiplier: 5.0,
		},
	}
	jobTime := time.Date(2026, 1, 1, 2, 0, 0, 0, time.UTC)
	mult := ContextualThresholdMultiplier("api_agent", jobTime, jobs)
	if mult != 1.0 {
		t.Errorf("expected 1.0 for wrong agent type, got %.2f", mult)
	}
}

func TestContextualThresholdMultiplier_MaxCeiling(t *testing.T) {
	jobs := []ScheduledJob{
		{
			AgentType:      "batch_processor",
			CronHour:       2,
			CronMinute:     0,
			WindowMinutes:  60,
			RateMultiplier: 100.0, // extreme value
		},
	}
	jobTime := time.Date(2026, 1, 1, 2, 0, 0, 0, time.UTC)
	mult := ContextualThresholdMultiplier("batch_processor", jobTime, jobs)
	if mult > 8.0 {
		t.Errorf("expected max ceiling 8.0, got %.2f", mult)
	}
}

func TestFleetActivityMultiplier_LowElevation(t *testing.T) {
	mult := FleetActivityMultiplier(0.10)
	if mult != 1.0 {
		t.Errorf("expected 1.0 at 10%% elevation, got %.2f", mult)
	}
}

func TestFleetActivityMultiplier_HighElevation(t *testing.T) {
	mult := FleetActivityMultiplier(0.80)
	if mult < 3.0 {
		t.Errorf("expected high multiplier at 80%% elevation, got %.2f", mult)
	}
}

func TestEffectiveZScoreThreshold_NoContext(t *testing.T) {
	threshold := EffectiveZScoreThreshold(3.0, "api_agent", time.Now().UTC(), nil, 0.0)
	if threshold != 3.0 {
		t.Errorf("expected 3.0 with no context, got %.2f", threshold)
	}
}

func TestEffectiveZScoreThreshold_ScheduledJob(t *testing.T) {
	jobs := []ScheduledJob{
		{
			AgentType:      "batch_processor",
			CronHour:       2,
			CronMinute:     0,
			WindowMinutes:  60,
			RateMultiplier: 5.0,
		},
	}
	jobTime := time.Date(2026, 1, 1, 2, 0, 0, 0, time.UTC)
	threshold := EffectiveZScoreThreshold(3.0, "batch_processor", jobTime, jobs, 0.0)
	if threshold < 12.0 {
		t.Errorf("expected threshold > 12.0 during scheduled window, got %.2f", threshold)
	}
}

func TestEffectiveZScoreThreshold_FleetEvent(t *testing.T) {
	threshold := EffectiveZScoreThreshold(3.0, "api_agent", time.Now().UTC(), nil, 0.80)
	if threshold < 9.0 {
		t.Errorf("expected threshold > 9.0 during fleet event, got %.2f", threshold)
	}
}
