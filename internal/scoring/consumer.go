package scoring

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"
)

// ScoreWriter is the interface the consumer uses to write scores.
// Defined here to avoid import cycles.
type ScoreWriter interface {
	WriteScore(agentDID string, score int, reason interface{}) error
}

// EventConsumer reads from agent_event_queue and updates scores.
// Runs as a background goroutine.
// Batch size: 100 events every 5 seconds per APEX v5.2 spec.
type EventConsumer struct {
	db          *sql.DB
	scoreWriter ScoreWriter
	batchSize   int
	interval    time.Duration
}

// NewEventConsumer creates a new consumer.
func NewEventConsumer(db *sql.DB, sw ScoreWriter) *EventConsumer {
	return &EventConsumer{
		db:          db,
		scoreWriter: sw,
		batchSize:   100,
		interval:    5 * time.Second,
	}
}

// Start runs the consumer loop. Call as goroutine.
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

// processBatch reads up to batchSize unprocessed events and scores them.
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
			// Move to DLQ
			c.db.Exec(`
				INSERT INTO agent_event_dlq (payload, error, created_at)
				VALUES ($1, $2, NOW())`,
				row.payload, err.Error())
		} else {
			processed++
		}

		// Mark processed regardless of outcome
		c.db.Exec(`
			UPDATE agent_event_queue
			SET processed = true, processed_at = NOW()
			WHERE id = $1`, row.id)
	}

	return processed, nil
}

// processEvent computes a score update for one event.
func (c *EventConsumer) processEvent(id int64, agentDID,
	eventType string, payload []byte) error {

	// Parse feature vector from payload
	var data struct {
		FeatureVector FeatureVector `json:"feature_vector"`
	}
	if err := json.Unmarshal(payload, &data); err != nil {
		return err
	}

	// Phase 1 scoring: H+V formula
	// H: use current score from DB as historical component
	// V: compute z-score from feature vector
	// Simplified bootstrap: use fixed cluster baselines
	baselines := map[string]struct{ mean, std float64 }{
		"tool_call_rate_per_hour":       {80, 60},
		"unique_endpoints_per_hour":     {25, 30},
		"bulk_access_count_per_session": {400, 400},
		"pii_field_access_rate":         {20, 25},
		"cross_tenant_probe_count":      {0, 0.1},
		"permission_escalation_count":   {0, 0.5},
		"sub_agent_spawn_depth":         {0, 0.3},
		"token_refresh_rate":            {1, 1},
	}

	// Compute worst z-score across all features
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
	for feature, value := range features {
		if b, ok := baselines[feature]; ok {
			std := b.std
			if std == 0 {
				std = 0.1
			}
			z := (value - b.mean) / std
			if z > worstZ {
				worstZ = z
			}
		}
	}

	penalty := 0.0
	if worstZ > 3.0 {
		penalty = min(100*(worstZ-3.0), 300)
	}
	V := max(0, 1000-penalty*3)
	H := 700.0 // bootstrap historical score

	newScore := ComputeScore(H, V, DefaultWeights)
	band := ScoreBand(newScore)

	reason := map[string]interface{}{
		"decision":    band,
		"agent_did":   agentDID,
		"score":       newScore,
		"event_type":  eventType,
		"worst_z":     worstZ,
		"penalty":     penalty,
		"computed_at": time.Now().Unix(),
	}

	slog.Info("score_computed",
		"agent_did", agentDID,
		"event_type", eventType,
		"score", newScore,
		"band", band,
		"worst_z", worstZ,
	)

	return c.scoreWriter.WriteScore(agentDID, newScore, reason)
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
