package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Gateway enforcement duration — G-LATENCY gate metric
	GatewayEnforcementDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "are_gateway_enforcement_duration_ms",
			Help:    "Gateway enforcement overhead in milliseconds",
			Buckets: []float64{1, 2, 5, 10, 25, 50, 100},
		},
		[]string{"decision", "band"},
	)

	// Score updates per second
	ScoreUpdatesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "are_score_updates_total",
			Help: "Total score updates processed",
		},
		[]string{"band", "policy"},
	)

	// False positive rate gauge — G-FP gate metric
	FPRateGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "are_fp_rate_7d",
		Help: "False positive rate over last 7 days",
	})

	// Active agent count
	ActiveAgentCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "are_active_agents_total",
		Help: "Number of active agent identities",
	})

	// Blocked decisions counter
	BlockedDecisionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "are_blocked_decisions_total",
			Help: "Total enforcement decisions by type",
		},
		[]string{"decision", "policy"},
	)

	// Event queue depth — pipeline health
	EventQueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "are_event_queue_depth",
		Help: "Unprocessed events in the async queue",
	})

	// Redis cache hit rate
	RedisCacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "are_redis_cache_total",
			Help: "Redis cache hits and misses",
		},
		[]string{"result"}, // "hit" or "miss"
	)

	// SIEM delivery status
	SIEMDeliveryTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "are_siem_delivery_total",
			Help: "SIEM webhook delivery attempts",
		},
		[]string{"status"}, // "success" or "failed"
	)

	// JWT verify fallback counter — incremented when Kong falls back to
	// unverified claim extraction because /verify service was unavailable.
	// Nonzero value in production means RS256 verification is not running.
	// Monitor: are_jwt_verify_fallback_total should be 0 in healthy deployment.
	JWTVerifyFallbackTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "are_jwt_verify_fallback_total",
		Help: "JWT verifications that fell back to unverified extraction (verify service unavailable)",
	})

	// A3 Hardening Sprint: Redis memory utilization gauge.
	// Alert threshold: 80%. Prevents silent score staleness from Redis OOM.
	RedisMemoryPct = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "are_redis_memory_pct",
		Help: "Redis memory usage as percentage of maxmemory (0-100)",
	})

	// OTel Feature: Per-agent reputation score — Kong dashboard visibility.
	// Background updater sets latest score per agent every 60s.
	// High cardinality acceptable for Phase 1 pilot agent counts.
	AgentScoreGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "are_agent_score",
		Help: "Current reputation score per agent",
	}, []string{"agent_did", "band"})

	// OTel Feature: Per-agent anomaly events — behavioral drift counter.
	// Incremented at scoring path when z-score exceeds threshold.
	AgentAnomaliesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "are_agent_anomalies_total",
		Help: "Total behavioral anomalies detected per agent",
	}, []string{"agent_did"})

	// OTel Feature: Enforcement decisions by regulatory framework.
	// Incremented when compliance export tags a decision to a framework.
	EnforcementByFramework = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "are_enforcement_by_framework_total",
		Help: "Enforcement decisions tagged by regulatory framework",
	}, []string{"framework", "decision"})
)
