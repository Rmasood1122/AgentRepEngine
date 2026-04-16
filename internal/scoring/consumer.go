package scoring

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"math"
	"time"

	"github.com/agentrepengine/are/internal/audit"
	are_metrics "github.com/agentrepengine/are/internal/metrics"
)

// ScoringPayload is the typed reason object written by the consumer.
// Replaces map[string]interface{} — eliminates reflection overhead on marshal.
// PL-5: ~60% payload size reduction vs untyped map.
type ScoringPayload struct {
	Decision      string  `json:"decision"`
	AgentDID      string  `json:"agent_did"`
	OrgID         string  `json:"org_id"`
	Score         int     `json:"score"`
	EventType     string  `json:"event_type"`
	WorstZ        float64 `json:"worst_z"`
	WorstFeature  string  `json:"worst_feature"`
	Penalty       float64 `json:"penalty"`
	HComponent    float64 `json:"h_component"`
	VComponent    float64 `json:"v_component"`
	ComputedAt    int64   `json:"computed_at"`
	PolicyFired   string  `json:"policy_fired"`
	ConfidencePct int     `json:"confidence_pct"`
}

// ScoreWriter is the interface the consumer uses to write scores.
type ScoreWriter interface {
	WriteScore(agentDID string, score int, reason interface{}) error
}

// EventConsumer reads from agent_event_queue and updates scores.
type EventConsumer struct {
	db          *sql.DB
	scoreWriter ScoreWriter
	policy      *PolicyEngine
	batchSize   int
	interval    time.Duration
	siem        *audit.SIEMWebhook
}

// QueueBacklogThreshold is the unprocessed event count that triggers a SIEM
// alert. At 1000 unprocessed events the consumer is materially behind —
// agent scores may not reflect recent behavioral events.
// Addresses FMEA RPN-210: async pipeline backlog → stale agent scores.
const QueueBacklogThreshold = 1000

func NewEventConsumer(db *sql.DB, sw ScoreWriter) *EventConsumer {
	var pe *PolicyEngine
	if loaded, err := NewPolicyEngine("config/policy_packs"); err != nil {
		slog.Warn("policy engine load failed — running without policy evaluation",
			"error", err,
		)
	} else {
		pe = loaded
	}
	return &EventConsumer{
		db:          db,
		scoreWriter: sw,
		policy:      pe,
		siem:        audit.NewSIEMWebhook(),
		batchSize:   100,
		interval:    5 * time.Second,
	}
}

func (c *EventConsumer) Start() {
	slog.Info("event consumer running",
		"batch_size", c.batchSize,
		"interval_seconds", c.interval.Seconds(),
	)
	for {
		processed, err := c.processBatch()
		if err != nil {
			slog.Error("consumer batch error", "error", err)
		} else if processed > 0 {
			slog.Info("consumer batch processed", "count", processed)
		}
		c.updateQueueDepth()
		time.Sleep(c.interval)
	}
}

// updateQueueDepth queries the unprocessed event count, updates the
// Prometheus gauge, and fires a SIEM alert if the backlog exceeds
// QueueBacklogThreshold. Called after every consumer batch cycle.
// Addresses FMEA RPN-210: async pipeline backlog → stale agent scores.
func (c *EventConsumer) updateQueueDepth() {
	var depth int
	err := c.db.QueryRow(`
                SELECT COUNT(*)
                FROM agent_event_queue
                WHERE processed = false`).Scan(&depth)
	if err != nil {
		slog.Warn("queue_depth_check_failed", "error", err)
		return
	}
	are_metrics.EventQueueDepth.Set(float64(depth))
	slog.Debug("queue_depth_updated", "depth", depth)
	if depth > QueueBacklogThreshold {
		slog.Warn("event_queue_backlog_alert",
			"depth", depth,
			"threshold", QueueBacklogThreshold,
			"action", "investigate_consumer_lag",
		)
		c.siem.SendQueueBacklog(depth)
	}
}

func (c *EventConsumer) processBatch() (int, error) {
	rows, err := c.db.Query(`
		SELECT id, agent_did, event_type, payload
		FROM agent_event_queue
		WHERE processed = false
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED`,
		c.batchSize)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type queueRow struct {
		id        int64
		agentDID  string
		eventType string
		payload   []byte
	}

	var batch []queueRow
	for rows.Next() {
		var row queueRow
		if err := rows.Scan(&row.id, &row.agentDID,
			&row.eventType, &row.payload); err != nil {
			continue
		}
		batch = append(batch, row)
	}
	rows.Close()

	if len(batch) == 0 {
		return 0, nil
	}

	processed := 0
	for _, row := range batch {
		if err := c.processEvent(row.id, row.agentDID,
			row.eventType, row.payload); err != nil {
			slog.Error("process event failed",
				"event_id", row.id,
				"agent_did", row.agentDID,
				"error", err,
			)
			c.db.Exec(`
				INSERT INTO agent_event_dlq (payload, error, created_at)
				VALUES ($1, $2, NOW())`,
				row.payload, err.Error())
		} else {
			processed++
		}

		c.db.Exec(`
			UPDATE agent_event_queue
			SET processed = true, processed_at = NOW()
			WHERE id = $1`, row.id)
	}

	return processed, nil
}

func (c *EventConsumer) processEvent(id int64, agentDID,
	eventType string, payload []byte) error {

	var data struct {
		FeatureVector FeatureVector `json:"feature_vector"`
	}
	if err := json.Unmarshal(payload, &data); err != nil {
		return err
	}

	// Fetch org_id + current score together — single DB round trip.
	// org_id required for org-scoped baseline isolation (TW-PRE-2).
	H, orgID := c.getAgentContext(agentDID)

	// Live org-scoped baselines via BaselineStore.
	// Falls back to cluster → hardcoded if agent has <100 samples.
	bs := NewBaselineStore(c.db)

	v := data.FeatureVector
	features := map[string]float64{
		"tool_call_rate_per_hour":       v.ToolCallRatePerHour,
		"unique_endpoints_per_hour":     v.UniqueEndpointsPerHour,
		"bulk_access_count_per_session": v.BulkAccessCountPerSession,
		"pii_field_access_rate":         v.PIIFieldAccessRate,
		"cross_tenant_probe_count":      v.CrossTenantProbeCount,
		"permission_escalation_count":   v.PermissionEscalationCount,
		"sub_agent_spawn_depth":         v.SubAgentSpawnDepth,
		"token_refresh_rate":            v.TokenRefreshRate,
	}

	worstZ := 0.0
	worstFeature := ""
	for feature, value := range features {
		// Org-scoped baseline lookup — prevents cross-tenant contamination.
		baseline := bs.GetBaseline(orgID, agentDID, feature)
		std := baseline.StdDev
		if std < 0.01 {
			std = 0.1
		}
		z := (value - baseline.Mean) / std
		if z > worstZ {
			worstZ = z
			worstFeature = feature
		}
		// Update agent baseline with this observation (Welford's online).
		if err := bs.UpdateAgentBaseline(orgID, agentDID, feature, value); err != nil {
			slog.Warn("baseline update failed",
				"org_id", orgID,
				"agent_did", agentDID,
				"feature", feature,
				"error", err,
			)
		}
		// Update cluster baseline.
		bs.UpdateClusterBaseline("default", feature, value)
	}

	penalty := 0.0
	if worstZ > 3.0 {
		penalty = min(100*(worstZ-3.0), 300)
	}
	V := max(0, 1000-penalty*3)

	newScore := ComputeScore(H, V, DefaultWeights)
	band := ScoreBand(newScore)

	// confidence_pct: inverse of anomaly confidence.
	// At z=0 → 100%, at z≥3.0 → 0%.
	confidencePct := int(math.Max(0, math.Min(100, (1.0-(worstZ/3.0))*100)))

	// TW-6: Variance growth rate check — early warning for slow-walk attacks.
	// Runs after z-score scoring. If variance doubled in last 7 days → HIGH_RISK.
	// This catches slow-walk attacks that evade z-score by gradually training
	// the baseline over multiple days. Z-score stays low; variance growth catches it.
	policyFired := "no_policy_fired"
	varianceHighRisk := false
	vgrResults, vgrErr := CheckVarianceGrowthRate(c.db, orgID, agentDID)
	if vgrErr != nil {
		slog.Warn("variance_growth_check_failed",
			"org_id", orgID,
			"agent_did", agentDID,
			"error", vgrErr,
		)
	} else {
		for _, vgr := range vgrResults {
			if vgr.HighRisk {
				varianceHighRisk = true
				policyFired = "variance_growth_rate_exceeded"
				slog.Warn("slow_walk_early_warning",
					"org_id", orgID,
					"agent_did", agentDID,
					"feature", vgr.Feature,
					"growth_rate", vgr.GrowthRate,
					"threshold", VarianceGrowthThreshold,
					"window_days", VarianceWindowDays,
				)
				break
			}
		}
	}

	// Apply variance growth penalty — additive on top of z-score penalty.
	// HIGH_RISK variance growth: -150 points, flags for human review.
	// Rationale: slow-walk attack must not silently execute high-value actions
	// even when z-score looks normal.
	if varianceHighRisk {
		newScore = int(max(0, float64(newScore)-150))
		band = ScoreBand(newScore)
	}

	// Policy evaluation — ceiling override model.
	// Threshold breaches apply penalty regardless of behavioral score.
	// A TRUSTED agent cannot silently execute HIGH_RISK actions.
	// Second detection layer: catches what z-score misses.
	if c.policy != nil {
		violations := c.policy.Evaluate(v)
		if worst := WorstViolation(violations); worst != nil {
			penalty := worst.ScorePenalty
			if penalty < 0 {
				penalty = -penalty
			}
			newScore = int(max(0, float64(newScore)-float64(penalty)))
			band = ScoreBand(newScore)
			if policyFired == "no_policy_fired" {
				policyFired = worst.ExplainTemplate.PolicyFired
			}
			slog.Warn("policy_violation_detected",
				"agent_did", agentDID,
				"policy", worst.PolicyName,
				"feature", worst.Feature,
				"value", worst.Value,
				"threshold", worst.Threshold,
				"level", worst.Level,
				"penalty", penalty,
			)
		}
	}

	reason := ScoringPayload{
		Decision:      band,
		AgentDID:      agentDID,
		OrgID:         orgID,
		Score:         int(newScore),
		EventType:     eventType,
		WorstZ:        worstZ,
		WorstFeature:  worstFeature,
		Penalty:       penalty,
		HComponent:    H,
		VComponent:    V,
		ComputedAt:    time.Now().Unix(),
		PolicyFired:   policyFired,
		ConfidencePct: confidencePct,
	}

	slog.Info("score_computed",
		"agent_did", agentDID,
		"org_id", orgID,
		"event_type", eventType,
		"score", int(newScore),
		"band", band,
		"worst_z", worstZ,
		"worst_feature", worstFeature,
		"policy_fired", policyFired,
	)

	return c.scoreWriter.WriteScore(agentDID, int(newScore), reason)
}

// getAgentContext fetches current score and org_id from agent_identities.
// Returns (700, "default") bootstrap values if agent not found.
// org_id is required for org-scoped baseline isolation.
func (c *EventConsumer) getAgentContext(agentDID string) (float64, string) {
	var score int
	var orgID string
	err := c.db.QueryRow(`
		SELECT current_score, org_id::text
		FROM agent_identities
		WHERE did = $1`, agentDID).Scan(&score, &orgID)
	if err != nil {
		return 700.0, "default"
	}
	return float64(score), orgID
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
