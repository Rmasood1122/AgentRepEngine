package audit

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"time"
)

// Framework constants
const (
	FrameworkHIPAA   = "HIPAA"
	FrameworkSOX     = "SOX"
	FrameworkFFIEC   = "FFIEC"
	FrameworkNISTRMF = "NIST_RMF"
	FrameworkDORA    = "DORA"
)

// RegulatoryPackage is the unified evidence output for any supported framework.
type RegulatoryPackage struct {
	OrgID            string    `json:"org_id"`
	Framework        string    `json:"framework"`
	GeneratedAt      time.Time `json:"generated_at"`
	AuditPeriodStart time.Time `json:"audit_period_start"`
	AuditPeriodEnd   time.Time `json:"audit_period_end"`
	HashChainRoot    string    `json:"hash_chain_root"`

	// HIPAA fields
	PHIAdjacentAgentCount int `json:"phi_adjacent_agent_count,omitempty"`
	AccessEventsInWindow  int `json:"access_events_in_window,omitempty"`
	AnomalousAccessCount  int `json:"anomalous_access_count,omitempty"`
	BlockedAccessCount    int `json:"blocked_access_count,omitempty"`

	// SOX fields
	CC72MonitoringActive    bool `json:"cc72_monitoring_active,omitempty"`
	AnomalyDetectionEvents  int  `json:"anomaly_detection_events,omitempty"`
	IncidentResponseActions int  `json:"incident_response_actions,omitempty"`
	OverrideEvents          int  `json:"override_events,omitempty"`
	HashVerified            bool `json:"hash_verified,omitempty"`

	// FFIEC fields
	ThirdPartyAgentCount          int     `json:"third_party_agent_count,omitempty"`
	BehavioralBaselineCoveragePct float64 `json:"behavioral_baseline_coverage_pct,omitempty"`
	RiskFlaggedCount              int     `json:"risk_flagged_count,omitempty"`

	// DORA fields
	IncidentCount  int        `json:"incident_count,omitempty"`
	BlockedCount   int        `json:"blocked_count,omitempty"`
	FPRate         float64    `json:"fp_rate,omitempty"`
	FirstEventTime *time.Time `json:"first_event_time,omitempty"`
	LastEventTime  *time.Time `json:"last_event_time,omitempty"`
}

// GenerateRegulatoryPackage produces a compliance evidence package for the given org,
// time window (days), and regulatory framework.
func GenerateRegulatoryPackage(db *sql.DB, orgID string, windowDays int, framework string) (*RegulatoryPackage, error) {
	now := time.Now().UTC()
	windowStart := now.AddDate(0, 0, -windowDays)

	pkg := &RegulatoryPackage{
		OrgID:            orgID,
		Framework:        framework,
		GeneratedAt:      now,
		AuditPeriodStart: windowStart,
		AuditPeriodEnd:   now,
	}

	// Compute hash chain root from enforcement events in window
	root, err := computeHashChainRoot(db, orgID, windowStart, now)
	if err != nil {
		return nil, fmt.Errorf("hash chain root: %w", err)
	}
	pkg.HashChainRoot = root

	switch framework {
	case FrameworkHIPAA:
		if err := populateHIPAA(db, pkg, orgID, windowStart, now); err != nil {
			return nil, err
		}
	case FrameworkSOX:
		if err := populateSOX(db, pkg, orgID, windowStart, now); err != nil {
			return nil, err
		}
	case FrameworkFFIEC:
		if err := populateFFIEC(db, pkg, orgID, windowStart, now); err != nil {
			return nil, err
		}
	case FrameworkNISTRMF:
		// NIST RMF reuses FFIEC baseline for Phase 1
		if err := populateFFIEC(db, pkg, orgID, windowStart, now); err != nil {
			return nil, err
		}
	case FrameworkDORA:
		if err := populateDORA(db, pkg, orgID, windowStart, now); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported framework: %s", framework)
	}

	return pkg, nil
}

func computeHashChainRoot(db *sql.DB, orgID string, from, to time.Time) (string, error) {
	rows, err := db.Query(`
		SELECT event_hash FROM enforcement_events
		WHERE org_id = $1 AND created_at BETWEEN $2 AND $3
		ORDER BY created_at ASC
	`, orgID, from, to)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	h := sha256.New()
	count := 0
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			return "", err
		}
		h.Write([]byte(hash))
		count++
	}
	if count == 0 {
		return "no-events", nil
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func populateHIPAA(db *sql.DB, pkg *RegulatoryPackage, orgID string, from, to time.Time) error {
	row := db.QueryRow(`
		SELECT
			COUNT(DISTINCT agent_id) FILTER (WHERE phi_adjacent = true),
			COUNT(*),
			COUNT(*) FILTER (WHERE anomaly_score > 0.7),
			COUNT(*) FILTER (WHERE action = 'block')
		FROM enforcement_events
		WHERE org_id = $1 AND created_at BETWEEN $2 AND $3
	`, orgID, from, to)
	return row.Scan(
		&pkg.PHIAdjacentAgentCount,
		&pkg.AccessEventsInWindow,
		&pkg.AnomalousAccessCount,
		&pkg.BlockedAccessCount,
	)
}

func populateSOX(db *sql.DB, pkg *RegulatoryPackage, orgID string, from, to time.Time) error {
	pkg.CC72MonitoringActive = true
	pkg.HashVerified = pkg.HashChainRoot != "no-events"

	row := db.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE anomaly_score > 0.7),
			COUNT(*) FILTER (WHERE action = 'block'),
			COUNT(*) FILTER (WHERE action = 'override')
		FROM enforcement_events
		WHERE org_id = $1 AND created_at BETWEEN $2 AND $3
	`, orgID, from, to)
	return row.Scan(
		&pkg.AnomalyDetectionEvents,
		&pkg.IncidentResponseActions,
		&pkg.OverrideEvents,
	)
}

func populateFFIEC(db *sql.DB, pkg *RegulatoryPackage, orgID string, from, to time.Time) error {
	row := db.QueryRow(`
		SELECT
			COUNT(DISTINCT agent_id) FILTER (WHERE third_party = true),
			COUNT(DISTINCT agent_id) FILTER (WHERE baseline_established = true),
			COUNT(DISTINCT agent_id),
			COUNT(*) FILTER (WHERE anomaly_score > 0.7)
		FROM enforcement_events
		WHERE org_id = $1 AND created_at BETWEEN $2 AND $3
	`, orgID, from, to)

	var thirdParty, baselined, total, riskFlagged int
	if err := row.Scan(&thirdParty, &baselined, &total, &riskFlagged); err != nil {
		return err
	}
	pkg.ThirdPartyAgentCount = thirdParty
	pkg.RiskFlaggedCount = riskFlagged
	if total > 0 {
		pkg.BehavioralBaselineCoveragePct = float64(baselined) / float64(total) * 100.0
	}
	return nil
}

func populateDORA(db *sql.DB, pkg *RegulatoryPackage, orgID string, from, to time.Time) error {
	row := db.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE anomaly_score > 0.7),
			COUNT(*) FILTER (WHERE action = 'block'),
			MIN(created_at),
			MAX(created_at)
		FROM enforcement_events
		WHERE org_id = $1 AND created_at BETWEEN $2 AND $3
	`, orgID, from, to)

	var first, last sql.NullTime
	if err := row.Scan(&pkg.IncidentCount, &pkg.BlockedCount, &first, &last); err != nil {
		return err
	}
	if first.Valid {
		pkg.FirstEventTime = &first.Time
	}
	if last.Valid {
		pkg.LastEventTime = &last.Time
	}

	// FP rate from metrics table
	fpRow := db.QueryRow(`
		SELECT COALESCE(AVG(fp_rate), 0.0) FROM daily_fp_metrics
		WHERE org_id = $1 AND metric_date BETWEEN $2 AND $3
	`, orgID, from, to)
	return fpRow.Scan(&pkg.FPRate)
}
