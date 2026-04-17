package identity

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// JTIStore provides dual-write jti tracking: Redis (fast) + Postgres (durable).
// A5 Hardening Sprint: closes replay window during Redis restarts.
type JTIStore struct {
	rdb *redis.Client
	db  *sql.DB
}

// NewJTIStore creates a JTI store with Redis primary and Postgres backup.
// If db is nil, operates in Redis-only mode (backward compatible).
func NewJTIStore(rdb *redis.Client, db *sql.DB) *JTIStore {
	return &JTIStore{rdb: rdb, db: db}
}

// CheckAndMark checks if a jti has been used, and marks it as used if not.
// Returns error if jti was already used (replay detected).
// Fail-open on infrastructure failure per APEX Law L2.
func (s *JTIStore) CheckAndMark(ctx context.Context, jti, agentDID string, ttl time.Duration) error {
	replayKey := replayKeyPrefix + jti

	// Check Redis first (fast path)
	used, err := s.rdb.Exists(ctx, replayKey).Result()
	if err != nil {
		slog.Warn("jti_redis_check_failed", "jti", jti, "error", err)
		// Redis down — fall back to Postgres
		return s.checkPostgres(ctx, jti, agentDID, ttl)
	}

	if used > 0 {
		return fmt.Errorf("token replay detected: jti %s already used", jti)
	}

	// Not in Redis — check Postgres (covers Redis restart scenario)
	if s.db != nil {
		var exists bool
		pgErr := s.db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM jti_cache WHERE jti = $1 AND expires_at > NOW())`,
			jti,
		).Scan(&exists)
		if pgErr == nil && exists {
			// Found in Postgres but not Redis — Redis was restarted
			// Re-populate Redis and reject
			s.rdb.SetEx(ctx, replayKey, "1", ttl)
			return fmt.Errorf("token replay detected: jti %s already used (recovered from Postgres)", jti)
		}
	}

	// Mark as used — dual write
	if setErr := s.rdb.SetEx(ctx, replayKey, "1", ttl).Err(); setErr != nil {
		slog.Warn("jti_redis_write_failed", "jti", jti, "error", setErr)
	}

	if s.db != nil {
		expiresAt := time.Now().Add(ttl)
		if _, pgErr := s.db.ExecContext(ctx,
			`INSERT INTO jti_cache (jti, agent_did, expires_at) VALUES ($1, $2, $3)
			 ON CONFLICT (jti) DO NOTHING`,
			jti, agentDID, expiresAt,
		); pgErr != nil {
			slog.Warn("jti_postgres_write_failed", "jti", jti, "error", pgErr)
		}
	}

	return nil
}

// checkPostgres is the fallback path when Redis is unavailable.
func (s *JTIStore) checkPostgres(ctx context.Context, jti, agentDID string, ttl time.Duration) error {
	if s.db == nil {
		// No Postgres — fail-open per APEX Law L2
		slog.Warn("jti_both_stores_unavailable_failing_open", "jti", jti)
		return nil
	}

	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM jti_cache WHERE jti = $1 AND expires_at > NOW())`,
		jti,
	).Scan(&exists)
	if err != nil {
		slog.Warn("jti_postgres_check_failed_failing_open", "jti", jti, "error", err)
		return nil // fail-open
	}
	if exists {
		return fmt.Errorf("token replay detected: jti %s already used (Postgres fallback)", jti)
	}

	// Mark in Postgres only (Redis is down)
	expiresAt := time.Now().Add(ttl)
	if _, pgErr := s.db.ExecContext(ctx,
		`INSERT INTO jti_cache (jti, agent_did, expires_at) VALUES ($1, $2, $3)
		 ON CONFLICT (jti) DO NOTHING`,
		jti, agentDID, expiresAt,
	); pgErr != nil {
		slog.Warn("jti_postgres_write_failed_failing_open", "jti", jti, "error", pgErr)
	}

	return nil
}

// RebuildFromPostgres loads all non-expired JTIs from Postgres into Redis.
// Called on startup to recover from Redis restart.
func (s *JTIStore) RebuildFromPostgres(ctx context.Context) (int, error) {
	if s.db == nil {
		return 0, nil
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT jti, expires_at FROM jti_cache WHERE expires_at > NOW()`)
	if err != nil {
		return 0, fmt.Errorf("query jti_cache: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var jti string
		var expiresAt time.Time
		if err := rows.Scan(&jti, &expiresAt); err != nil {
			continue
		}
		ttl := time.Until(expiresAt)
		if ttl <= 0 {
			continue
		}
		replayKey := replayKeyPrefix + jti
		s.rdb.SetEx(ctx, replayKey, "1", ttl)
		count++
	}

	slog.Info("jti_cache_rebuilt_from_postgres", "count", count)
	return count, rows.Err()
}
