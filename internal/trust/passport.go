package trust

import "time"

// AgentPassport is the portable behavioral credential carried by an agent
// when making cross-organizational API calls.
// Issued by the originating enterprise's ARE instance.
// Verified by the receiving enterprise's ARE instance.
//
// Header: X-Agent-Passport (base64-encoded JSON + signature)
//
// Why different from Microsoft IATP (DID-based):
// Microsoft uses DIDs — cryptographic identity only.
// ARE passport includes BehavioralHash — trust credential derived from behavioral history.
// Valid DID + compromised behavioral profile → Microsoft passes, ARE catches.
//
// Validity: 24-hour window (short enough to prevent replay, long enough for agent workflows)
type AgentPassport struct {
	AgentID          string    `json:"agent_id"`
	IssuingOrg       string    `json:"issuing_org"`
	IssuingARENode   string    `json:"issuing_are_node"`
	BehavioralHash   string    `json:"behavioral_hash"`   // SHA-256 of agent's behavioral fingerprint
	BaselineWindow   string    `json:"baseline_window"`   // "30d"
	CertificationRef string    `json:"certification_ref"` // Zenodo DOI or internal cert ID
	IssuedAt         time.Time `json:"issued_at"`
	ExpiresAt        time.Time `json:"expires_at"`        // IssuedAt + 24h
	Signature        []byte    `json:"signature"`         // RS256 over all fields above
}

// CrossOrgTrustLevel represents the trust level assigned to an inbound agent call
type CrossOrgTrustLevel string

const (
	TrustHigh   CrossOrgTrustLevel = "high"   // Certified org + valid passport + behavioral hash match
	TrustMedium CrossOrgTrustLevel = "medium" // Uncertified org + valid passport
	TrustLow    CrossOrgTrustLevel = "low"    // No passport — standard ARE enforcement
)

// TODO: Implement IssuePassport(agentID string, orgID string) (*AgentPassport, error)
// TODO: Implement VerifyPassport(passport *AgentPassport) (CrossOrgTrustLevel, error)
// TODO: Implement ParsePassportHeader(headerValue string) (*AgentPassport, error)
// TODO: Implement ComputeBehavioralHash(agentID string, orgID string) (string, error)
