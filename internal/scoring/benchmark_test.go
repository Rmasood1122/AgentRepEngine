package scoring

// Linux production benchmark results (golang:1.24-alpine, AMD Ryzen 7 5800HS):
// BenchmarkCallLevelOverhead:   0.25ns — gateway decision path
// BenchmarkScoreBand:           0.25ns — band assignment
// BenchmarkComputeZScore:       0.25ns — z-score computation
// BenchmarkComputeScore:        5.2ns  — full formula
// BenchmarkAgentLevelScoring:  22.8ns  — complete pipeline
// Zero memory allocations on all hot paths
//
// Windows dev numbers ~30-40% slower (expected — not production)
// Run Linux benchmark:
//   MSYS_NO_PATHCONV=1 docker run --rm \
//     -v "C:/Users/rmaso/AgentRepEngine:/app" -w /app \
//     golang:1.24-alpine go test ./internal/scoring/... -bench=. -benchmem

import (
	"testing"
	"time"
)

var benchVector = FeatureVector{
	ToolCallRatePerHour:       75,
	UniqueEndpointsPerHour:    12,
	BulkAccessCountPerSession: 50,
	PIIFieldAccessRate:        0.10,
	CrossTenantProbeCount:     0,
	PermissionEscalationCount: 0,
	SubAgentSpawnDepth:        1,
	TokenRefreshRate:          1,
}

// BenchmarkComputeScore — pure formula, no DB, no Redis.
// Minimum latency floor for any scoring decision.
// Target: sub-100ns (this is just math).
func BenchmarkComputeScore(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ComputeScore(700.0, 850.0, DefaultWeights)
	}
}

// BenchmarkScoreBand — band assignment overhead.
// Target: sub-10ns.
func BenchmarkScoreBand(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ScoreBand(750)
	}
}

// BenchmarkVelocityScore — z-score computation per feature.
// Target: sub-100ns.
func BenchmarkVelocityScore(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VelocityScore(75.0, 80.0, 60.0)
	}
}

// BenchmarkHistoricalDecay — decay computation.
// Target: sub-100ns.
func BenchmarkHistoricalDecay(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HistoricalDecay(700.0, 1.0)
	}
}

// BenchmarkComputeZScore — z-score with baseline lookup.
func BenchmarkComputeZScore(b *testing.B) {
	baseline := Baseline{Mean: 80.0, StdDev: 60.0, SampleCount: 1000}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeZScore(75.0, baseline)
	}
}

// BenchmarkCallLevelOverhead — gateway decision path simulation.
// Score lookup (cache hit assumed) + band assignment + enforcement decision.
// This is what runs on EVERY agent request at the gateway.
// Target: sub-1ms (well within p99 10ms budget).
func BenchmarkCallLevelOverhead(b *testing.B) {
	score := 750
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		band := ScoreBand(score)
		_ = band == "BLOCKED"
		_ = band == "RESTRICTED"
	}
}

// BenchmarkAgentLevelScoring — full scoring pipeline (no DB).
// Simulates async score update path (off critical path).
// Target: sub-500ms p95 (async, not on request path).
func BenchmarkAgentLevelScoring(b *testing.B) {
	lastSeen := time.Now().Add(-1 * time.Hour)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		H := HistoricalDecay(700.0, 1.0)
		V := VelocityScore(
			benchVector.ToolCallRatePerHour,
			80.0,
			60.0,
		)
		score := ComputeScore(H, V, DefaultWeights)
		_ = ScoreBand(score)
		_ = lastSeen
	}
}
