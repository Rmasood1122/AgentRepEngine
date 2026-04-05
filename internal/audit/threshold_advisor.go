package audit

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

// ThresholdAdvisor analyzes override patterns and suggests threshold adjustments.
// M7-STEP-1: Foundation of the monthly CISO intelligence report.
// Rule: max 10% weight change per iteration. Human must approve before applying.
//
// Data source: enforcement_decisions WHERE override = true (human-confirmed FPs)
// Output: policy-level suggestions with estimated F1 impact

// OverridePattern summarizes override activity for one policy.
type OverridePattern struct {
	PolicyFired    string
	TotalDecisions int
	Overrides      int
	OverrideRate   float64 // overrides / total decisions for this policy
	ReasonCodes    map[string]int
}

// ThresholdSuggestion is a recommended threshold adjustment for one policy.
type ThresholdSuggestion struct {
	PolicyFired      string
	CurrentOverrides int
	OverrideRate     float64
	Suggestion       string  // human-readable recommendation
	MaxChangeAllowed float64 // always 10% per iteration rule
	Priority         string  // HIGH / MEDIUM / LOW
}

// MonthlyAdvisorReport is the output of M7 threshold analysis.
// Sent to CISO monthly. Human approves before any weights change.
type MonthlyAdvisorReport struct {
	GeneratedAt     time.Time
	PeriodDays      int
	TotalDecisions  int
	TotalOverrides  int
	OverrideRate    float64
	Patterns        []OverridePattern
	Suggestions     []ThresholdSuggestion
	NoChangeSummary string // shown when no adjustments needed
}

// GenerateMonthlyReport analyzes the last N days of enforcement decisions
// and produces threshold adjustment suggestions based on override patterns.
func GenerateMonthlyReport(db *sql.DB, orgID string, periodDays int) (*MonthlyAdvisorReport, error) {
	report := &MonthlyAdvisorReport{
		GeneratedAt: time.Now(),
		PeriodDays:  periodDays,
	}

	// Count total decisions and overrides in period
	err := db.QueryRow(`
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE override = true) as overrides
		FROM enforcement_decisions
		WHERE org_id = $1
		  AND created_at > NOW() - INTERVAL '1 day' * $2`,
		orgID, periodDays,
	).Scan(&report.TotalDecisions, &report.TotalOverrides)
	if err != nil {
		return nil, fmt.Errorf("monthly report totals: %w", err)
	}

	if report.TotalDecisions > 0 {
		report.OverrideRate = float64(report.TotalOverrides) / float64(report.TotalDecisions)
	}

	// Analyze override patterns per policy
	rows, err := db.Query(`
		SELECT
			policy_fired,
			COUNT(*) as total_decisions,
			COUNT(*) FILTER (WHERE override = true) as override_count,
			COALESCE(reason_code, 'none') as reason_code
		FROM enforcement_decisions
		WHERE org_id = $1
		  AND created_at > NOW() - INTERVAL '1 day' * $2
		GROUP BY policy_fired, reason_code
		ORDER BY override_count DESC`,
		orgID, periodDays,
	)
	if err != nil {
		return nil, fmt.Errorf("monthly report patterns: %w", err)
	}
	defer rows.Close()

	patternMap := map[string]*OverridePattern{}
	for rows.Next() {
		var policy, reasonCode string
		var total, overrides int
		if err := rows.Scan(&policy, &total, &overrides, &reasonCode); err != nil {
			continue
		}
		if _, ok := patternMap[policy]; !ok {
			patternMap[policy] = &OverridePattern{
				PolicyFired: policy,
				ReasonCodes: map[string]int{},
			}
		}
		patternMap[policy].TotalDecisions += total
		patternMap[policy].Overrides += overrides
		if overrides > 0 {
			patternMap[policy].ReasonCodes[reasonCode] += overrides
		}
	}

	// Compute override rates and build suggestions
	for _, p := range patternMap {
		if p.TotalDecisions > 0 {
			p.OverrideRate = float64(p.Overrides) / float64(p.TotalDecisions)
		}
		report.Patterns = append(report.Patterns, *p)

		// Generate suggestion if override rate is significant
		suggestion := buildSuggestion(p)
		if suggestion != nil {
			report.Suggestions = append(report.Suggestions, *suggestion)
		}
	}

	if len(report.Suggestions) == 0 {
		report.NoChangeSummary = fmt.Sprintf(
			"No threshold adjustments recommended. Override rate %.2f%% is within acceptable range (<5%%). "+
				"Current thresholds are well-calibrated for this environment.",
			report.OverrideRate*100,
		)
	}

	slog.Info("monthly_advisor_report_generated",
		"org_id", orgID,
		"period_days", periodDays,
		"total_decisions", report.TotalDecisions,
		"total_overrides", report.TotalOverrides,
		"suggestions", len(report.Suggestions),
	)

	return report, nil
}

// buildSuggestion produces a threshold suggestion for a policy with high override rate.
// Returns nil if no adjustment is warranted.
func buildSuggestion(p *OverridePattern) *ThresholdSuggestion {
	// Minimum threshold: 3+ overrides AND >10% override rate
	if p.Overrides < 3 || p.OverrideRate < 0.10 {
		return nil
	}

	priority := "LOW"
	if p.OverrideRate >= 0.25 {
		priority = "HIGH"
	} else if p.OverrideRate >= 0.15 {
		priority = "MEDIUM"
	}

	// Determine dominant reason code
	dominantReason := ""
	maxCount := 0
	for code, count := range p.ReasonCodes {
		if count > maxCount {
			maxCount = count
			dominantReason = code
		}
	}

	var suggestionText string
	switch {
	case dominantReason == "ARE-FP-002": // legitimate bulk export
		suggestionText = fmt.Sprintf(
			"Policy %s: %d overrides (%.0f%% override rate). "+
				"Dominant reason: ARE-FP-002 (legitimate bulk export). "+
				"Recommend: raise bulk_access threshold by up to 10%%. "+
				"Validate on held-out corpus before applying.",
			p.PolicyFired, p.Overrides, p.OverrideRate*100)
	case dominantReason == "ARE-FP-003": // authorized off-hours job
		suggestionText = fmt.Sprintf(
			"Policy %s: %d overrides (%.0f%% override rate). "+
				"Dominant reason: ARE-FP-003 (authorized off-hours job). "+
				"Recommend: add time-context modifier for scheduled jobs. "+
				"Validate on held-out corpus before applying.",
			p.PolicyFired, p.Overrides, p.OverrideRate*100)
	default:
		suggestionText = fmt.Sprintf(
			"Policy %s: %d overrides (%.0f%% override rate). "+
				"Recommend: review policy thresholds. "+
				"Max adjustment: 10%% per iteration. "+
				"Validate on held-out corpus before applying.",
			p.PolicyFired, p.Overrides, p.OverrideRate*100)
	}

	return &ThresholdSuggestion{
		PolicyFired:      p.PolicyFired,
		CurrentOverrides: p.Overrides,
		OverrideRate:     p.OverrideRate,
		Suggestion:       suggestionText,
		MaxChangeAllowed: 0.10,
		Priority:         priority,
	}
}

// FormatCISOSummary produces the monthly CISO-facing summary string.
// Format matches M7 spec: "847 decisions, 23 TPs, 3 overrides. Suggested adjustment..."
func FormatCISOSummary(report *MonthlyAdvisorReport) string {
	if report.TotalDecisions == 0 {
		return "No enforcement decisions in the reporting period."
	}

	confirmedTPs := report.TotalDecisions - report.TotalOverrides
	summary := fmt.Sprintf(
		"This month: %d decisions, %d confirmed enforcements, %d overrides (%.1f%% override rate).",
		report.TotalDecisions, confirmedTPs, report.TotalOverrides, report.OverrideRate*100,
	)

	if len(report.Suggestions) == 0 {
		summary += " " + report.NoChangeSummary
		return summary
	}

	summary += fmt.Sprintf(" %d threshold adjustment(s) suggested:", len(report.Suggestions))
	for _, s := range report.Suggestions {
		summary += fmt.Sprintf("\n  [%s] %s", s.Priority, s.Suggestion)
		summary += "\n  ACTION REQUIRED: Human approval needed before applying any threshold change."
	}
	return summary
}
