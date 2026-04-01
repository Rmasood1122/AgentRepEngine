package enforcement

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// newTestRedis returns Redis client for tests.
func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Username: "are_admin",
		Password: "are_redis_dev",
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
	}
	return rdb
}

// newTestDB returns a PostgreSQL connection for tests.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres",
		"postgres://are:are_dev@localhost:5433/agentrepengine?sslmode=disable")
	if err != nil {
		t.Skipf("PostgreSQL not available: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("PostgreSQL not reachable: %v", err)
	}
	return db
}

// TestAutoRollback verifies auto-rollback fires when FP rate exceeds 2%.
// V6 FIX: this is the most important pilot safety test.
func TestAutoRollback(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()
	db := newTestDB(t)
	defer db.Close()

	mc := NewModeController(rdb, db, ModeEnforce)

	// Verify starts in enforce mode
	if mc.GetMode() != ModeEnforce {
		t.Fatalf("expected enforce mode, got %s", mc.GetMode())
	}
	t.Log("✅ Started in enforce mode")

	// Seed enforcement_decisions with FP rate > 2%
	// Insert 10 BLOCKED decisions, 3 overridden = 30% FP rate
	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		DELETE FROM enforcement_decisions
		WHERE agent_did = 'did:jwt:test:rollback-test:001'
	`)
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	// Get last hash for chain continuity
	var prevHash string
	err = db.QueryRowContext(ctx, `
		SELECT COALESCE(
			(SELECT this_hash FROM enforcement_decisions
			 ORDER BY created_at DESC LIMIT 1),
			'genesis'
		)
	`).Scan(&prevHash)
	if err != nil {
		t.Fatalf("get prev hash: %v", err)
	}

	// Insert 10 BLOCKED decisions
	for i := 0; i < 10; i++ {
		override := i < 3 // first 3 are FPs (overridden)
		thisHash := fmt.Sprintf("test-hash-%d", i)
		_, err = db.ExecContext(ctx, `
			INSERT INTO enforcement_decisions
				(agent_did, decision, score, reason_object,
				 override, prev_hash, this_hash, created_at)
			VALUES ($1, 'BLOCKED', 187, '{"policy_fired":"test"}',
				$2, $3, $4, NOW())`,
			"did:jwt:test:rollback-test:001",
			override, prevHash, thisHash)
		if err != nil {
			t.Fatalf("insert decision %d failed: %v", i, err)
		}
		prevHash = thisHash
	}
	t.Log("✅ Seeded 10 BLOCKED decisions (3 overridden = 30% FP)")

	// Measure FP rate directly
	fpRate, err := mc.measureFPRate()
	if err != nil {
		t.Fatalf("measureFPRate failed: %v", err)
	}
	if fpRate < 0 {
		t.Fatal("no FP data found — seed failed")
	}
	t.Logf("✅ FP rate measured: %.2f%%", fpRate)

	if fpRate <= FPRateThreshold {
		t.Fatalf("expected FP rate > %.2f%%, got %.2f%%",
			FPRateThreshold, fpRate)
	}

	// Trigger rollback check
	mc.checkAndRollbackIfNeeded()

	// Verify mode rolled back to observe
	if mc.GetMode() != ModeObserve {
		t.Fatalf("❌ AUTO-ROLLBACK FAILED — mode is still %s", mc.GetMode())
	}
	t.Logf("✅ Auto-rollback triggered at %.2f%% FP rate", fpRate)
	t.Log("✅ Mode confirmed: OBSERVE")

	// Verify enforce cannot be re-enabled automatically
	// (requires explicit SetEnforce call)
	mc.checkAndRollbackIfNeeded() // second check — already in observe
	if mc.GetMode() != ModeObserve {
		t.Fatal("❌ Mode changed without manual action")
	}
	t.Log("✅ Mode stays OBSERVE — manual re-enable required")

	// Cleanup
	db.ExecContext(ctx, `
		DELETE FROM enforcement_decisions
		WHERE agent_did = 'did:jwt:test:rollback-test:001'
	`)
}

// TestModeDefault verifies new ModeController defaults to observe.
func TestModeDefault(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()
	db := newTestDB(t)
	defer db.Close()

	mc := NewModeController(rdb, db, ModeObserve)
	if mc.GetMode() != ModeObserve {
		t.Fatalf("expected observe, got %s", mc.GetMode())
	}
	t.Log("✅ Default mode is observe")
}

// TestManualEnforce verifies SetEnforce works and requires manual action.
func TestManualEnforce(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()
	db := newTestDB(t)
	defer db.Close()

	mc := NewModeController(rdb, db, ModeObserve)

	if err := mc.SetEnforce(); err != nil {
		t.Fatalf("SetEnforce failed: %v", err)
	}
	if mc.GetMode() != ModeEnforce {
		t.Fatalf("expected enforce, got %s", mc.GetMode())
	}
	t.Log("✅ Manual enforce mode set correctly")
}

// TestFPMonitorNoRollbackBelowThreshold verifies no rollback when FP rate is healthy.
func TestFPMonitorNoRollbackBelowThreshold(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()
	db := newTestDB(t)
	defer db.Close()

	mc := NewModeController(rdb, db, ModeEnforce)

	// Seed 100 BLOCKED decisions, 1 overridden = 1% FP (below 2% threshold)
	ctx := context.Background()
	db.ExecContext(ctx, `
		DELETE FROM enforcement_decisions
		WHERE agent_did = 'did:jwt:test:no-rollback-test:001'
	`)

	var prevHash string
	db.QueryRowContext(ctx, `
		SELECT COALESCE(
			(SELECT this_hash FROM enforcement_decisions
			 ORDER BY created_at DESC LIMIT 1),
			'genesis'
		)
	`).Scan(&prevHash)

	for i := 0; i < 100; i++ {
		override := i == 0 // only first is FP = 1%
		thisHash := fmt.Sprintf("no-rollback-hash-%d-%d", i, time.Now().UnixNano())
		db.ExecContext(ctx, `
			INSERT INTO enforcement_decisions
				(agent_did, decision, score, reason_object,
				 override, prev_hash, this_hash, created_at)
			VALUES ($1, 'BLOCKED', 187, '{"policy_fired":"test"}',
				$2, $3, $4, NOW())`,
			"did:jwt:test:no-rollback-test:001",
			override, prevHash, thisHash)
		prevHash = thisHash
	}

	mc.checkAndRollbackIfNeeded()

	if mc.GetMode() != ModeEnforce {
		t.Fatal("❌ Rolled back at 1% FP — should not have triggered")
	}
	t.Log("✅ No rollback at 1% FP rate — threshold respected")

	db.ExecContext(ctx, `
		DELETE FROM enforcement_decisions
		WHERE agent_did = 'did:jwt:test:no-rollback-test:001'
	`)
}
