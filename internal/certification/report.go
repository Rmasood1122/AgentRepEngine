package certification

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// CertificationReport is the signed behavioral certification for a single agent
// over a defined time window. Every field is included in the hash chain.
// DO NOT reorder fields — hash chain depends on field order.
type CertificationReport struct {
	AgentID        string              `json:"agent_id"`
	OrgID          string              `json:"org_id"`
	WindowStart    time.Time           `json:"window_start"`
	WindowEnd      time.Time           `json:"window_end"`
	BaselineParams WelfordSnapshot     `json:"baseline_params"`
	DriftEvents    []DriftEvent        `json:"drift_events"`
	EnforcementLog []EnforcementRecord `json:"enforcement_log"`
	ModelVersions  []BMVRecord         `json:"model_versions"`
	IdentityChain  []BIVRecord         `json:"identity_chain"`
	CertHash       string              `json:"cert_hash"`
	SignatureKeyID string              `json:"signature_key_id"`
	Signature      []byte              `json:"signature"`
}

// WelfordSnapshot captures the statistical baseline at certification time
type WelfordSnapshot struct {
	Mean             float64 `json:"mean"`
	Variance         float64 `json:"variance"`
	StdDev           float64 `json:"std_dev"`
	SampleCount      int64   `json:"sample_count"`
	WindowDays       int     `json:"window_days"`
	BaselineMaturity string  `json:"baseline_maturity"` // "30d_established" | "probation_Xd"
}

// DriftEvent captures a detected behavioral trajectory violation
type DriftEvent struct {
	Timestamp     time.Time `json:"timestamp"`
	ZScore        float64   `json:"z_score"`
	Category      string    `json:"category"`
	ConfidencePct float64   `json:"confidence_pct"`
	ActionTaken   string    `json:"action_taken"` // "blocked" | "flagged" | "observed"
}

// EnforcementRecord is a single enforcement decision with full reason object
type EnforcementRecord struct {
	Timestamp     time.Time              `json:"timestamp"`
	AgentID       string                 `json:"agent_id"`
	Action        string                 `json:"action"`
	Score         int                    `json:"score"`
	ReasonObject  map[string]interface{} `json:"reason_object"`
	ChainHash     string                 `json:"chain_hash"`
	PrevChainHash string                 `json:"prev_chain_hash"`
}

// BMVRecord captures behavioral continuity across model version changes
type BMVRecord struct {
	Timestamp        time.Time `json:"timestamp"`
	FromModelVersion string    `json:"from_model_version"`
	ToModelVersion   string    `json:"to_model_version"`
	BehavioralDelta  float64   `json:"behavioral_delta"`
	ContinuityPass   bool      `json:"continuity_pass"`
}

// BIVRecord captures two-factor agent authentication events
type BIVRecord struct {
	Timestamp       time.Time `json:"timestamp"`
	JWTVerified     bool      `json:"jwt_verified"`
	BehaviorMatch   bool      `json:"behavior_match"`
	FingerprintHash string    `json:"fingerprint_hash"`
}

// Generate queries PostgreSQL and builds a complete CertificationReport.
// Phase 1: windowDays relaxed for demo (agent may have minutes of history).
// Language: "behavioral attestation" — never "guarantee" or "certification".
func Generate(db *sql.DB, agentID, orgID string, windowDays int) (*CertificationReport, error) {
	now := time.Now().UTC()
	windowStart := now.AddDate(0, 0, -windowDays)

	report := &CertificationReport{
		AgentID:     agentID,
		OrgID:       orgID,
		WindowStart: windowStart,
		WindowEnd:   now,
	}

	// Enforcement log from tamper-evident audit trail
	rows, err := db.Query(`
		SELECT created_at, agent_did, decision, score, this_hash, prev_hash
		FROM enforcement_decisions
		WHERE agent_did = $1 AND created_at BETWEEN $2 AND $3
		ORDER BY chain_position ASC`, agentID, windowStart, now)
	if err != nil {
		return nil, fmt.Errorf("query enforcement log: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rec EnforcementRecord
		var thisHash, prevHash sql.NullString
		if err := rows.Scan(&rec.Timestamp, &rec.AgentID, &rec.Action,
			&rec.Score, &thisHash, &prevHash); err != nil {
			return nil, fmt.Errorf("scan enforcement record: %w", err)
		}
		if thisHash.Valid {
			rec.ChainHash = thisHash.String
		}
		if prevHash.Valid {
			rec.PrevChainHash = prevHash.String
		}
		report.EnforcementLog = append(report.EnforcementLog, rec)

		// Derive drift events from enforcement decisions
		if rec.Action == "BLOCKED" || rec.Action == "RESTRICTED" {
			report.DriftEvents = append(report.DriftEvents, DriftEvent{
				Timestamp:     rec.Timestamp,
				Category:      "enforcement_triggered",
				ActionTaken:   strings.ToLower(rec.Action),
				ConfidencePct: 100.0,
			})
		}
	}

	// Baseline snapshot from agent_baselines (best effort)
	var mean, variance, stddev float64
	var sampleCount int64
	baselineErr := db.QueryRow(`
		SELECT COALESCE(AVG(mean), 0), COALESCE(AVG(variance), 0),
		       COALESCE(AVG(std_dev), 0), COALESCE(SUM(sample_count), 0)
		FROM agent_baselines
		WHERE agent_did = $1`, agentID).Scan(&mean, &variance, &stddev, &sampleCount)
	if baselineErr == nil {
		maturity := "probation"
		if sampleCount > 100 {
			maturity = "30d_established"
		}
		report.BaselineParams = WelfordSnapshot{
			Mean:             mean,
			Variance:         variance,
			StdDev:           stddev,
			SampleCount:      sampleCount,
			WindowDays:       windowDays,
			BaselineMaturity: maturity,
		}
	}

	// Phase 1: ModelVersions and IdentityChain are empty arrays (not nil)
	if report.ModelVersions == nil {
		report.ModelVersions = []BMVRecord{}
	}
	if report.IdentityChain == nil {
		report.IdentityChain = []BIVRecord{}
	}
	if report.DriftEvents == nil {
		report.DriftEvents = []DriftEvent{}
	}
	if report.EnforcementLog == nil {
		report.EnforcementLog = []EnforcementRecord{}
	}

	return report, nil
}

// HashReport computes SHA-256 over all report fields (excluding Signature).
func HashReport(report *CertificationReport) error {
	// Zero out signature fields before hashing
	report.CertHash = ""
	report.Signature = nil
	report.SignatureKeyID = ""

	data, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("marshal for hash: %w", err)
	}
	hash := sha256.Sum256(data)
	report.CertHash = fmt.Sprintf("%x", hash[:])
	return nil
}

// ToJSON marshals the complete signed report.
func ToJSON(report *CertificationReport) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}
