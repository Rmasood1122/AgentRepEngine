package scoring

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"
)

// ScoringPayload is the typed reason object written by the consumer.
// Replaces map[string]interface{} — eliminates reflection overhead on marshal.
// PL-5: ~60% payload size reduction vs untyped map.
type ScoringPayload struct {
	Decision     string  `json:"decision"`
	AgentDID     string  `json:"agent_did"`
	OrgID        string  `json:"org_id"`
	Score        int     `json:"score"`
	EventType    string  `json:"event_type"`
	WorstZ       float64 `json:"worst_z"`
	WorstFeature string  `json:"worst_feature"`
	Penalty      float64 `json:"penalty"`
	HComponent   float64 `json:"h_component"`
	VComponent   float64 `json:"v_component"`
	ComputedAt   int64   `json:"computed_at"`
}

// ScoreWriter is the interface the consumer uses to write scores.
type ScoreWriter interface {
	WriteScore(agentDID string, score int, reason interface{}) error
}

// EventConsumer reads from agent_event_queue and updates scores.
type EventConsumer struct {
	db          *sql.DB
	scoreWriter ScoreWriter
	batchSize   int
	interval    time.Duration
}

func NewEventConsumer(db *sql.DB, sw ScoreWriter) *EventConsumer {
	return &EventConsumer{
		db:          db,
		scoreWriter: sw,
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
		time.Sleep(c.interval)
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

	// PL-5: typed ScoringPayload replaces map[string]interface{}.
	// ~60% payload size reduction. Eliminates reflection overhead on marshal.
	reason := ScoringPayload{
		Decision:     band,
		AgentDID:     agentDID,
		OrgID:        orgID,
		Score:        newScore,
		EventType:    eventType,
		WorstZ:       worstZ,
		WorstFeature: worstFeature,
		Penalty:      penalty,
		HComponent:   H,
		VComponent:   V,
		ComputedAt:   time.Now().Unix(),
	}

	slog.Info("score_computed",
		"agent_did", agentDID,
		"org_id", orgID,
		"event_type", eventType,
		"score", newScore,
		"band", band,
		"worst_z", worstZ,
		"worst_feature", worstFeature,
	)

	return c.scoreWriter.WriteScore(agentDID, newScore, reason)
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
