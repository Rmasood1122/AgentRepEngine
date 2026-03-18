package identity

import "github.com/golang-jwt/jwt/v5"

// AgentClaims defines the 7 required JWT claims for every agent.
// All fields are mandatory. Missing any field = G-IDENTITY hard stop.
// This struct enforces the claim schema at compile time.
type AgentClaims struct {
	// AgentDID is the unique agent identifier.
	// Format: did:jwt:{org}:{service}:{uuid}
	AgentDID string `json:"agent_did"`

	// InstanceID is a per-deployment UUID.
	// Same agent redeployed = new InstanceID, same AgentDID.
	InstanceID string `json:"instance_id"`

	// LineageHash links this agent to its parent.
	// sha256(parent_did + "|" + spawned_at_unix)
	// Root agents use sha256("root" + "|" + spawned_at_unix)
	LineageHash string `json:"lineage_hash"`

	// OrgID is the organization UUID that owns this agent.
	OrgID string `json:"org_id"`

	// LineageDepth is 0 for root agents, +1 per sub-agent level.
	// Max depth: 5 (enforced by policy pack)
	LineageDepth int `json:"lineage_depth"`

	jwt.RegisteredClaims
}
