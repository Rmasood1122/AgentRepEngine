package scoring

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVRF_DeterministicSameInputs(t *testing.T) {
	did := "did:jwt:test:vrf:001"
	epoch := int64(1700000000)
	t1 := VRFThreshold(3.0, did, epoch)
	t2 := VRFThreshold(3.0, did, epoch)
	assert.Equal(t, t1, t2, "same inputs must produce same threshold")
}

func TestVRF_JitterWithinBounds(t *testing.T) {
	base := 3.0
	epoch := int64(1700000000)
	for i := 0; i < 1000; i++ {
		did := "did:jwt:test:vrf:" + string(rune('a'+i%26)) + string(rune('0'+i%10))
		threshold := VRFThreshold(base, did, epoch)
		jitter := threshold - base
		assert.True(t, math.Abs(jitter) <= VRFJitterMax,
			"jitter %f exceeds max %f for did %s", jitter, VRFJitterMax, did)
	}
}

func TestVRF_DifferentAgentsDifferentThresholds(t *testing.T) {
	base := 3.0
	epoch := int64(1700000000)
	t1 := VRFThreshold(base, "did:jwt:test:agent:001", epoch)
	t2 := VRFThreshold(base, "did:jwt:test:agent:002", epoch)
	assert.NotEqual(t, t1, t2, "different agents must get different thresholds")
}

func TestVRF_EpochRotationChangesThreshold(t *testing.T) {
	did := "did:jwt:test:vrf:rotation"
	base := 3.0
	epoch1 := int64(1700000000)
	epoch2 := epoch1 + VRFEpochSeconds
	t1 := VRFThreshold(base, did, epoch1)
	t2 := VRFThreshold(base, did, epoch2)
	assert.NotEqual(t, t1, t2, "epoch rotation must change threshold")
}

func TestVRF_EpochFunction(t *testing.T) {
	// Two timestamps in the same epoch must return the same epoch value
	ts1 := int64(1700001000)
	ts2 := int64(1700001500)
	assert.Equal(t, VRFEpoch(ts1), VRFEpoch(ts2), "same epoch window must return same epoch")

	// Timestamps in different epochs must differ
	ts3 := ts1 + VRFEpochSeconds
	assert.NotEqual(t, VRFEpoch(ts1), VRFEpoch(ts3), "different epoch windows must differ")
}

func TestVRF_BaseThresholdPreserved(t *testing.T) {
	// Jitter should never push threshold below 2.75 (base 3.0 - max 0.25)
	// or above 3.25 (base 3.0 + max 0.25)
	base := 3.0
	epoch := int64(1700000000)
	for _, did := range []string{
		"did:jwt:test:bound:001",
		"did:jwt:test:bound:002",
		"did:jwt:test:bound:003",
		"did:jwt:test:bound:004",
		"did:jwt:test:bound:005",
	} {
		threshold := VRFThreshold(base, did, epoch)
		assert.True(t, threshold >= base-VRFJitterMax && threshold <= base+VRFJitterMax,
			"threshold %f out of bounds for %s", threshold, did)
	}
}

func TestVRF_FPImpactNegligible(t *testing.T) {
	// Statistical check: mean jitter across 10000 agents should be near 0
	// (uniform distribution centered at 0)
	base := 3.0
	epoch := int64(1700000000)
	sum := 0.0
	n := 10000
	for i := 0; i < n; i++ {
		did := "did:jwt:test:stat:" + string(rune(i))
		threshold := VRFThreshold(base, did, epoch)
		sum += threshold - base
	}
	meanJitter := sum / float64(n)
	assert.True(t, math.Abs(meanJitter) < 0.01,
		"mean jitter %f should be near 0 — FP impact must be symmetric", meanJitter)
}
