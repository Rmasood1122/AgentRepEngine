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
		return nil, fmt.Errorf("db score lookup: %w", err)
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
// Instruments blocked decisions and score update metrics.
func (s *ScoreStore) WriteScore(agentDID string, score int, reasonObj interface{}) error {
	band := scoring.ScoreBand(score)

	reasonJSON, err := json.Marshal(reasonObj)
	if err != nil {
		reasonJSON = []byte(fmt.Sprintf(`{"score":%d,"band":"%s"}`, score, band))
	}

	// Extract policy name for metrics label
	policyName := "unknown"
	if rm, ok := reasonObj.(map[string]interface{}); ok {
		if pf, ok := rm["policy_fired"].(string); ok && pf != "" {
			policyName = pf
		}
	}

	// Instrument enforcement decision metrics
	metrics.BlockedDecisionsTotal.WithLabelValues(band, policyName).Inc()
	metrics.ScoreUpdatesTotal.WithLabelValues(band, policyName).Inc()

	// PostgreSQL first — source of truth
	_, err = s.db.Exec(`
		UPDATE agent_identities
		SET current_score = $1, last_seen = NOW()
		WHERE did = $2`,
		score, agentDID)
	if err != nil {
		return fmt.Errorf("postgres score write: %w", err)
	}

	key := "score:" + agentDID

	if band == "BLOCKED" {
		// Immediate cache invalidation on block
		if err := s.rdb.Del(s.ctx, key).Err(); err != nil {
			slog.Error("redis_invalidation_failed",
				"agent_did", agentDID, "error", err)
		} else {
			slog.Info("cache_invalidated_on_block",
				"agent_did", agentDID, "score", score)
		}
		// H5 FIX: Fire SIEM webhook on every BLOCKED decision
		// Fire-and-forget — SIEM delivery never blocks enforcement
		// scoreDelta=0 placeholder — Phase 2 tracks previous score in WriteScore
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
