package trust

// CrossOrgScorer evaluates inbound agent calls from external organizations.
// Integrates with the Kong plugin to apply cross-org trust levels
// to enforcement decisions.
//
// Trust level determination:
// Certified org + valid passport + behavioral hash match → TrustHigh
//   → reduced enforcement friction (higher score threshold before blocking)
// Uncertified org + valid passport                      → TrustMedium
//   → observe mode enforcement on cross-org calls
// No passport                                           → TrustLow
//   → standard ARE enforcement (same as today, zero regression)
//
// All cross-org trust decisions logged to audit trail with full reason object.
// The audit trail entry includes: originating org, trust level, passport hash,
// behavioral hash match result, and enforcement action taken.

// TODO: Implement ScoreCrossOrgCall(passport *AgentPassport, inboundOrgID string) (CrossOrgTrustLevel, error)
// TODO: Implement LogCrossOrgDecision(call CrossOrgCall, trustLevel CrossOrgTrustLevel) error
// TODO: Integrate with Kong plugin: add X-Agent-Passport header extraction
