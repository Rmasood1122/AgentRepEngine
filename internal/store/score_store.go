package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/agentrepengine/are/internal/scoring"
	"github.com/redis/go-redis/v9"
)

// ScoreResult is returned by GetScore.
type ScoreResult struct {
	Score      int
	Band       string
	Source     string // "cache" or "db"
	ReasonJSON string
}

// ScoreStore handles reading and writing scores to Redis and PostgreSQL.
// PostgreSQL is source of truth. Redis is cache only.
type ScoreStore struct {
	db       *sql.DB
	redisURL string
	rdb      *redis.Client
	ctx      context.Context
}

// NewScoreStore creates a new ScoreStore.
func NewScoreStore(db *sql.DB, redisURL string) *ScoreStore {
	return &ScoreStore{
		db:       db,
		redisURL: redisURL,
		ctx:      context.Background(),
	}
}

// ConnectRedis initializes the Redis client.
func (s *ScoreStore) ConnectRedis() error {
	opts, err := redis.ParseURL(s.redisURL)
	if err != nil {
		return fmt.Errorf("parse redis URL: %w", err)
	}
	s.rdb = redis.NewClient(opts)
	return s.Ping()
}

// Ping checks Redis connectivity.
func (s *ScoreStore) Ping() error {
	return s.rdb.Ping(s.ctx).Err()
}

// GetScore retrieves agent score — Redis first, PostgreSQL fallback.
// FM2 prevention: PostgreSQL is always source of truth.
func (s *ScoreStore) GetScore(agentDID string) (*ScoreResult, error) {
	key := "score:" + agentDID

	// Try Redis cache first
	vals, err := s.rdb.HGetAll(s.ctx, key).Result()
	if err == nil && len(vals) > 0 {
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

	// Cache miss — fetch from PostgreSQL
	var currentScore int
	var status string
	err = s.db.QueryRow(`
		SELECT current_score, status
		FROM agent_identities
		WHERE did = $1`, agentDID).Scan(&currentScore, &status)

	if err == sql.ErrNoRows {
		// Unknown agent — orphan score
		reason := fmt.Sprintf(`{"score":500,"band":"MONITORED","source":"orphan","note":"unknown agent"}`)
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

	// Backfill Redis cache
	s.rdb.HSet(s.ctx, key,
		"score", currentScore,
		"band", band,
		"reason", reason,
		"updated_at", time.Now().Unix(),
	)
	s.rdb.Expire(s.ctx, key, 60*time.Second)

	return &ScoreResult{
		Score:      currentScore,
		Band:       band,
		Source:     "db",
		ReasonJSON: reason,
	}, nil
}

// WriteScore writes a score update to both PostgreSQL and Redis.
// PostgreSQL first — source of truth.
// FM2 prevention: never write Redis only.
func (s *ScoreStore) WriteScore(agentDID string, score int, reasonObj interface{}) error {
	band := scoring.ScoreBand(score)

	reasonJSON, err := json.Marshal(reasonObj)
	if err != nil {
		reasonJSON = []byte(fmt.Sprintf(`{"score":%d,"band":"%s"}`, score, band))
	}

	// Write PostgreSQL first
	_, err = s.db.Exec(`
		UPDATE agent_identities
		SET current_score = $1, last_seen = NOW()
		WHERE did = $2`,
		score, agentDID)
	if err != nil {
		return fmt.Errorf("postgres score write: %w", err)
	}

	// Write Redis cache — GAP 3 fix: log failures, never silently ignore
	// Redis failure is non-fatal — Postgres is source of truth
	key := "score:" + agentDID
	if err := s.rdb.HSet(s.ctx, key,
		"score", score,
		"band", band,
		"reason", string(reasonJSON),
		"updated_at", time.Now().Unix(),
	).Err(); err != nil {
		slog.Error("redis_write_failed",
			"agent_did", agentDID,
			"error", err,
		)
		// Do not return error — Postgres write succeeded
		// Kong will miss cache on next request — acceptable
	} else {
		s.rdb.Expire(s.ctx, key, 60*time.Second)
	}

	return nil
}

// EnqueueEvent adds a behavioral event to the async processing queue.
// Feature vector stored atomically with event — T19 prevention.
func (s *ScoreStore) EnqueueEvent(agentDID, eventType string,
	vector scoring.FeatureVector, privacyTier int) error {

	vectorJSON, err := json.Marshal(vector)
	if err != nil {
		return fmt.Errorf("marshal feature vector: %w", err)
	}

	// Store event with feature vector in one atomic INSERT
	// T19 prevention: feature_vector is in same row as event
	_, err = s.db.Exec(`
		INSERT INTO agent_events
			(event_id, agent_did, event_type, feature_vector, created_at)
		VALUES (gen_random_uuid()::text, $1, $2, $3, NOW())`,
		agentDID, eventType, vectorJSON)
	if err != nil {
		return fmt.Errorf("insert agent event: %w", err)
	}

	// Also queue for async score update
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
