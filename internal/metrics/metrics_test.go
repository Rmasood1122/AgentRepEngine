package metrics

import (
	"testing"
)

func TestRedisMemoryPctGaugeExists(t *testing.T) {
	// Verify the gauge is registered and can be set without panic
	RedisMemoryPct.Set(42.5)
	t.Log("A3: RedisMemoryPct gauge registered and settable")
}
