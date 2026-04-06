package scoring

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"math"
)

const (
	// VRFJitterMax is the maximum threshold adjustment in z-score units.
	// ±0.25 sigma — keeps FP impact negligible while defeating threshold probing.
	VRFJitterMax = 0.25

	// VRFEpochSeconds defines how often the jitter rotates per agent.
	// 3600 = hourly rotation — attacker cannot hold a probing window open.
	VRFEpochSeconds = 3600
)

// VRFThreshold computes a deterministic per-agent threshold jitter.
// Inputs: agentDID (stable identity) + epochTS (rotates hourly).
// Output: base threshold adjusted by [-VRFJitterMax, +VRFJitterMax].
//
// Properties:
//   - Deterministic: same inputs → same output (audit reproducible)
//   - Unpredictable: HMAC-SHA256 — attacker cannot predict without secret
//   - Zero external deps: pure stdlib crypto
//   - Cheap: single HMAC-SHA256 per scored request
func VRFThreshold(baseThreshold float64, agentDID string, epochTS int64) float64 {
	jitter := vrfJitter(agentDID, epochTS)
	return baseThreshold + jitter
}

// VRFEpoch returns the current epoch timestamp for a given unix timestamp.
// Epochs rotate every VRFEpochSeconds — callers pass time.Now().Unix().
func VRFEpoch(unixTS int64) int64 {
	return (unixTS / VRFEpochSeconds) * VRFEpochSeconds
}

// vrfJitter derives a deterministic jitter value in [-VRFJitterMax, +VRFJitterMax].
// Uses HMAC-SHA256(agentDID, epoch) — deterministic but unpredictable per agent.
func vrfJitter(agentDID string, epochTS int64) float64 {
	// Encode epoch as 8-byte big-endian key material
	epochBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(epochBytes, uint64(epochTS))

	mac := hmac.New(sha256.New, epochBytes)
	mac.Write([]byte(agentDID))
	digest := mac.Sum(nil)

	// Take first 8 bytes → uint64 → normalize to [0, 1]
	raw := binary.BigEndian.Uint64(digest[:8])
	normalized := float64(raw) / float64(math.MaxUint64)

	// Map [0, 1] → [-VRFJitterMax, +VRFJitterMax]
	return (normalized*2 - 1) * VRFJitterMax
}
