package certification

import "time"

// CertificationReport is the signed behavioral certification for a single agent
// over a defined time window. Every field is included in the hash chain.
// DO NOT reorder fields — hash chain depends on field order.
type CertificationReport struct {
	AgentID        string           `json:"agent_id"`
	OrgID          string           `json:"org_id"`
	WindowStart    time.Time        `json:"window_start"`
	WindowEnd      time.Time        `json:"window_end"`
	BaselineParams WelfordSnapshot  `json:"baseline_params"`
	DriftEvents    []DriftEvent     `json:"drift_events"`
	EnforcementLog []EnforcementRecord `json:"enforcement_log"`
	ModelVersions  []BMVRecord      `json:"model_versions"`
	IdentityChain  []BIVRecord      `json:"identity_chain"`
	CertHash       string           `json:"cert_hash"`
	SignatureKeyID string           `json:"signature_key_id"`
	Signature      []byte           `json:"signature"`
}

// WelfordSnapshot captures the statistical baseline at certification time
type WelfordSnapshot struct {
	Mean            float64 `json:"mean"`
	Variance        float64 `json:"variance"`
	StdDev          float64 `json:"std_dev"`
	SampleCount     int64   `json:"sample_count"`
	WindowDays      int     `json:"window_days"`
	BaselineMaturity string `json:"baseline_maturity"` // "30d_established" | "probation_Xd"
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
	Timestamp      time.Time              `json:"timestamp"`
	AgentID        string                 `json:"agent_id"`
	Action         string                 `json:"action"`
	Score          int                    `json:"score"`
	ReasonObject   map[string]interface{} `json:"reason_object"`
	ChainHash      string                 `json:"chain_hash"`
	PrevChainHash  string                 `json:"prev_chain_hash"`
}

// BMVRecord captures behavioral continuity across model version changes
type BMVRecord struct {
	Timestamp       time.Time `json:"timestamp"`
	FromModelVersion string   `json:"from_model_version"`
	ToModelVersion  string    `json:"to_model_version"`
	BehavioralDelta float64   `json:"behavioral_delta"`
	ContinuityPass  bool      `json:"continuity_pass"`
}

// BIVRecord captures two-factor agent authentication events
type BIVRecord struct {
	Timestamp      time.Time `json:"timestamp"`
	JWTVerified    bool      `json:"jwt_verified"`
	BehaviorMatch  bool      `json:"behavior_match"`
	FingerprintHash string   `json:"fingerprint_hash"`
}

// TODO: Implement Generate() method
// TODO: Implement HashAndSign() method
// TODO: Implement ToPDF() method
// TODO: Implement ToJSON() method
