package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/agentrepengine/are/internal/audit"
	"github.com/agentrepengine/are/internal/metrics"
	"github.com/agentrepengine/are/internal/scoring"
	"github.com/redis/go-redis/v9"
)

type ScoreResult struct {
	Score      int
	Band       string
	Source     string
	ReasonJSON string
}

type ScoreStore struct {
	db       *sql.DB
	redisURL string
	rdb      *redis.Client
	ctx      context.Context
	siem     *audit.SIEMWebhook
}

func NewScoreStore(db *sql.DB, redisURL string) *ScoreStore {
	return &ScoreStore{
		db:       db,
		redisURL: redisURL,
		ctx:      context.Background(),
		siem:     audit.NewSIEMWebhook(),
	}
}

func (s *ScoreStore) ConnectRedis() error {
	opts, err := redis.ParseURL(s.redisURL)
	if err != nil {
		return fmt.Errorf("parse redis URL: %w", err)
	}
	s.rdb = redis.NewClient(opts)
	return s.Ping()
}

func (s *ScoreStore) Ping() error {
	return s.rdb.Ping(s.ctx).Err()
}

// GetRedisClient exposes the Redis client for ModeController.
func (s *ScoreStore) GetRedisClient() *redis.Client {
	return s.rdb
}

// GetScore retrieves agent score — Redis first, PostgreSQL fallback.
// Instruments cache hit/miss metrics.
func (s *ScoreStore) GetScore(agentDID string) (*ScoreResult, error) {
	key := "score:" + agentDID

	vals, err := s.rdb.HGetAll(s.ctx, key).Result()
	if err == nil && len(vals) > 0 {
		metrics.RedisCacheHits.WithLabelValues("hit").Inc()

		score := 700
		if v, ok := vals["score"]; ok {
			fmt.Sscanf(v, "%d", &score)
		}
		band := scoring.ScoreBand(score)
		reason := vals["reason"]
		if reason == "" {
			reason = fmt.Sprintf(`{"score":%d,"band":"%s","source":"cache"}`,
				score, band)
		}
		return &ScoreResult{
			Score:      score,
			Band:       band,
			Source:     "cache",
			ReasonJSON: reason,
		}, nil
	}

	metrics.RedisCacheHits.WithLabelValues("miss").Inc()

	var currentScore int
	var status string
	err = s.db.QueryRow(`
		SELECT current_score, status
		FROM agent_identities
		WHERE did = $1`, agentDID).Scan(&currentScore, &status)

	if err == sql.ErrNoRows {
		reason := `{"score":500,"band":"MONITORED","source":"orphan","note":"unknown agent"}`
		return &ScoreResult{
			Score:      500,
			Band:       "MONITORED",
			Source:     "orphan",
			ReasonJSON: reason,
		}, nil
	}
	if err != nil {
		slog.Error("score_lookup_both_stores_failed_failing_open",
			"agent_did", agentDID,
			"error", err)
		reason := `{"score":500,"band":"MONITORED","source":"failopen","note":"both_stores_unavailable_score_may_be_stale"}`
		return &ScoreResult{
			Score:      500,
			Band:       "MONITORED",
			Source:     "failopen",
			ReasonJSON: reason,
		}, nil
	}

	band := scoring.ScoreBand(currentScore)
	reason := fmt.Sprintf(`{"score":%d,"band":"%s","source":"db","status":"%s"}`,
		currentScore, band, status)

	if band != "BLOCKED" {
		s.rdb.HSet(s.ctx, key,
			"score", currentScore,
			"band", band,
			"reason", reason,
			"updated_at", time.Now().Unix(),
		)
		s.rdb.Expire(s.ctx, key, 60*time.Second)
	}

	return &ScoreResult{
		Score:      currentScore,
		Band:       band,
		Source:     "db",
		ReasonJSON: reason,
	}, nil
}

// WriteScore writes score to PostgreSQL first, then Redis.
// Fires SIEM webhook on every BLOCKED decision.
// FP-7: logs RESTRICTED and BLOCKED decisions to fp_candidates for human review.
func (s *ScoreStore) WriteScore(agentDID string, score int, reasonObj interface{}) error {
	band := scoring.ScoreBand(score)

	reasonJSON, err := json.Marshal(reasonObj)
	if err != nil {
		reasonJSON = []byte(fmt.Sprintf(`{"score":%d,"band":"%s"}`, score, band))
	}

	policyName := "unknown"
	if rm, ok := reasonObj.(map[string]interface{}); ok {
		if pf, ok := rm["policy_fired"].(string); ok && pf != "" {
			policyName = pf
		}
	}

	if band == "BLOCKED" {
		metrics.BlockedDecisionsTotal.WithLabelValues("BLOCKED", policyName).Inc()
	}
	metrics.ScoreUpdatesTotal.WithLabelValues(band, policyName).Inc()

	_, err = s.db.Exec(`
		INSERT INTO agent_identities (did, org_id, instance_id, current_score, last_seen)
		VALUES ($1, gen_random_uuid(), gen_random_uuid(), $2, NOW())
		ON CONFLICT (did) DO UPDATE
		SET current_score = EXCLUDED.current_score,
		    last_seen = NOW()`,
		agentDID, score)
	if err != nil {
		return fmt.Errorf("postgres score write: %w", err)
	}

	if result, ok := reasonObj.(*scoring.ScoredResult); ok {
		_, err = s.db.Exec(`
			INSERT INTO scoring_explanations
				(agent_did, score_before, score_after,
				 history_score, velocity_score, z_score,
				 composite_score, band, z_score_feature,
				 reason_json, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())`,
			agentDID,
			score-result.ScoreDelta,
			score,
			result.HistoryScore,
			result.VelocityScore,
			result.WorstZScore,
			result.Score,
			result.Band,
			result.WorstFeature,
			reasonJSON,
		)
		if err != nil {
			slog.Error("scoring_explanation_write_failed",
				"agent_did", agentDID,
				"error", err)
		}
	}

	key := "score:" + agentDID

	if band == "RESTRICTED" || band == "BLOCKED" {
		_, fpErr := s.db.Exec(`
			INSERT INTO fp_candidates
				(agent_did, score, band, reason_object, flagged_at)
			VALUES ($1, $2, $3, $4, NOW())`,
			agentDID, score, band, reasonJSON,
		)
		if fpErr != nil {
			slog.Error("fp_candidate_write_failed",
				"agent_did", agentDID,
				"band", band,
				"error", fpErr)
		}
	}

	if band == "BLOCKED" {
		if err := s.rdb.Del(s.ctx, key).Err(); err != nil {
			slog.Error("redis_invalidation_failed",
				"agent_did", agentDID, "error", err)
		} else {
			slog.Info("cache_invalidated_on_block",
				"agent_did", agentDID, "score", score)
		}
		s.siem.SendBlocked(agentDID, score, 0, policyName, reasonJSON)
		slog.Info("siem_webhook_fired",
			"agent_did", agentDID,
			"score", score,
			"policy", policyName)
	} else {
		if err := s.rdb.HSet(s.ctx, key,
			"score", score,
			"band", band,
			"reason", string(reasonJSON),
			"updated_at", time.Now().Unix(),
		).Err(); err != nil {
			slog.Error("redis_write_failed",
				"agent_did", agentDID, "error", err)
		} else {
			s.rdb.Expire(s.ctx, key, 60*time.Second)
		}
	}

	return nil
}

// EnqueueEvent adds a behavioral event to the async processing queue.
// Feature vector stored atomically — T19 prevention.
func (s *ScoreStore) EnqueueEvent(agentDID, eventType string,
	vector scoring.FeatureVector, privacyTier int) error {

	vectorJSON, err := json.Marshal(vector)
	if err != nil {
		return fmt.Errorf("marshal feature vector: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO agent_events
			(event_id, agent_did, event_type, feature_vector, created_at)
		VALUES (gen_random_uuid()::text, $1, $2, $3, NOW())`,
		agentDID, eventType, vectorJSON)
	if err != nil {
		return fmt.Errorf("insert agent event: %w", err)
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"agent_did":      agentDID,
		"event_type":     eventType,
		"feature_vector": vector,
	})

	_, err = s.db.Exec(`
		INSERT INTO agent_event_queue
			(agent_did, event_type, payload, privacy_tier, created_at)
		VALUES ($1, $2, $3, $4, NOW())`,
		agentDID, eventType, payload, privacyTier)
	if err != nil {
		return fmt.Errorf("enqueue event: %w", err)
	}

	return nil
}

// GetSIRRecord retrieves the SIR lifecycle record for an agent from Redis.
// Key pattern: sir:{agent_did}
func (s *ScoreStore) GetSIRRecord(ctx context.Context, agentDID string) (*scoring.SIRRecord, error) {
	key := "sir:" + agentDID
	data, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("sir_store: not found for %s", agentDID)
	}
	var record scoring.SIRRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("sir_store: unmarshal failed: %w", err)
	}
	return &record, nil
}

// PutSIRRecord persists the SIR lifecycle record for an agent to Redis.
// TTL: 90 days (aligns with queue retention policy).
func (s *ScoreStore) PutSIRRecord(ctx context.Context, record *scoring.SIRRecord) error {
	key := "sir:" + record.AgentDID
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("sir_store: marshal failed: %w", err)
	}
	return s.rdb.Set(ctx, key, data, 90*24*time.Hour).Err()
}
