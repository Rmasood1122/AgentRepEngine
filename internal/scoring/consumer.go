package scoring

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"
)

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

	// FIX 1: Correct baselines — pii_field_access_rate is a ratio (0.0-1.0)
	// not a count. Normal analyst: ~5% PII rate. Anomalous: >30%.
	baselines := map[string]struct{ mean, std float64 }{
		"tool_call_rate_per_hour":       {80, 60},
		"unique_endpoints_per_hour":     {25, 30},
		"bulk_access_count_per_session": {400, 400},
		"pii_field_access_rate":         {0.05, 0.08}, // FIX: was {20, 25}
		"cross_tenant_probe_count":      {0, 0.1},
		"permission_escalation_count":   {0, 0.5},
		"sub_agent_spawn_depth":         {0, 0.3},
		"token_refresh_rate":            {1, 1},
	}

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
		if b, ok := baselines[feature]; ok {
			std := b.std
			if std == 0 {
				std = 0.1
			}
			z := (value - b.mean) / std
			if z > worstZ {
				worstZ = z
				worstFeature = feature
			}
		}
	}

	penalty := 0.0
	if worstZ > 3.0 {
		penalty = min(100*(worstZ-3.0), 300)
	}
	V := max(0, 1000-penalty*3)

	// FIX 2: Use actual agent score from DB as H component.
	// Preserves behavioral history across events.
	// Falls back to 700 bootstrap only for unknown agents.
	H := c.getAgentScore(agentDID)

	newScore := ComputeScore(H, V, DefaultWeights)
	band := ScoreBand(newScore)

	reason := map[string]interface{}{
		"decision":      band,
		"agent_did":     agentDID,
		"score":         newScore,
		"event_type":    eventType,
		"worst_z":       worstZ,
		"worst_feature": worstFeature,
		"penalty":       penalty,
		"h_component":   H,
		"v_component":   V,
		"computed_at":   time.Now().Unix(),
	}

	slog.Info("score_computed",
		"agent_did", agentDID,
		"event_type", eventType,
		"score", newScore,
		"band", band,
		"worst_z", worstZ,
		"worst_feature", worstFeature,
	)

	return c.scoreWriter.WriteScore(agentDID, newScore, reason)
}

// getAgentScore fetches current agent score from DB for H component.
// Returns 700 (bootstrap) if agent not found or error.
func (c *EventConsumer) getAgentScore(agentDID string) float64 {
	var score int
	err := c.db.QueryRow(`
		SELECT current_score FROM agent_identities
		WHERE did = $1`, agentDID).Scan(&score)
	if err != nil {
		return 700.0
	}
	return float64(score)
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
