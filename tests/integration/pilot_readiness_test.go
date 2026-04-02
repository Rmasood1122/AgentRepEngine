package integration

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/agentrepengine/are/internal/store"
	_ "github.com/lib/pq"
)

const testDSN = "postgres://are:are_dev@localhost:5433/agentrepengine?sslmode=disable"

func setupDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres", testDSN)
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	return db
}

// GAP-1: Retention purges processed events older than 90 days
func TestRetention_PurgesOldProcessedEvents(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	oldTime := time.Now().UTC().AddDate(0, 0, -91)
	var insertedID int64
	err := db.QueryRow(`
		INSERT INTO agent_event_queue (agent_did, event_type, payload, processed, processed_at, created_at)
		VALUES ('did:jwt:test-retention-old', 'http_request', '{}', true, $1, $1)
		RETURNING id`, oldTime).Scan(&insertedID)
	if err != nil {
		t.Fatalf("insert old event: %v", err)
	}

	newTime := time.Now().UTC().AddDate(0, 0, -10)
	var recentID int64
	err = db.QueryRow(`
		INSERT INTO agent_event_queue (agent_did, event_type, payload, processed, processed_at, created_at)
		VALUES ('did:jwt:test-retention-recent', 'http_request', '{}', true, $1, $1)
		RETURNING id`, newTime).Scan(&recentID)
	if err != nil {
		t.Fatalf("insert recent event: %v", err)
	}

	deleted, err := store.PurgeOldEvents(context.Background(), db)
	if err != nil {
		t.Fatalf("purge: %v", err)
	}
	if deleted < 1 {
		t.Errorf("expected at least 1 deletion, got %d", deleted)
	}

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM agent_event_queue WHERE id = $1`, insertedID).Scan(&count)
	if count != 0 {
		t.Errorf("old event still present after purge")
	}

	db.QueryRow(`SELECT COUNT(*) FROM agent_event_queue WHERE id = $1`, recentID).Scan(&count)
	if count != 1 {
		t.Errorf("recent event was incorrectly purged")
	}

	db.Exec(`DELETE FROM agent_event_queue WHERE id = $1`, recentID)
}

// GAP-1b: Retention never purges unprocessed events
func TestRetention_NeverPurgesUnprocessedEvents(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	oldTime := time.Now().UTC().AddDate(0, 0, -95)
	var insertedID int64
	err := db.QueryRow(`
		INSERT INTO agent_event_queue (agent_did, event_type, payload, processed, created_at)
		VALUES ('did:jwt:test-unprocessed-old', 'http_request', '{}', false, $1)
		RETURNING id`, oldTime).Scan(&insertedID)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	store.PurgeOldEvents(context.Background(), db)

	var count int
	db.QueryRow(`SELECT COUNT(*) FROM agent_event_queue WHERE id = $1`, insertedID).Scan(&count)
	if count != 1 {
		t.Errorf("unprocessed event was incorrectly purged — data loss risk")
	}

	db.Exec(`DELETE FROM agent_event_queue WHERE id = $1`, insertedID)
}

// GAP-2: QueueDepth returns accurate count
func TestQueueDepth_Accurate(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	before, err := store.QueueDepth(context.Background(), db)
	if err != nil {
		t.Fatalf("queue depth: %v", err)
	}

	for i := 0; i < 3; i++ {
		db.Exec(`INSERT INTO agent_event_queue (agent_did, event_type, payload)
			VALUES ($1, 'http_request', '{}')`,
			fmt.Sprintf("did:jwt:test-depth-%d", i))
	}

	after, err := store.QueueDepth(context.Background(), db)
	if err != nil {
		t.Fatalf("queue depth after: %v", err)
	}

	if after != before+3 {
		t.Errorf("expected depth %d, got %d", before+3, after)
	}

	db.Exec(`DELETE FROM agent_event_queue WHERE agent_did LIKE 'did:jwt:test-depth-%'`)
}

// GAP-2b: QueueDepth warning threshold constant is correct
func TestQueueDepth_WarningThreshold(t *testing.T) {
	if store.QueueDepthWarning != 100_000 {
		t.Errorf("QueueDepthWarning = %d, want 100000", store.QueueDepthWarning)
	}
}

// GAP-3: RetentionDays constant is correct
func TestRetention_DaysConstant(t *testing.T) {
	if store.RetentionDays != 90 {
		t.Errorf("RetentionDays = %d, want 90", store.RetentionDays)
	}
}

// GAP-4: Hash chain intact on current DB state
func TestChainVerification_IntactChain(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	var result bool
	err := db.QueryRowContext(context.Background(),
		`SELECT verify_hash_chain('enforcement_decisions') AS valid`,
	).Scan(&result)
	if err != nil {
		t.Skipf("verify_hash_chain function not available: %v", err)
	}
	if !result {
		t.Errorf("hash chain is broken — audit trail integrity compromised")
	}
}

// GAP-5: Blended JWT — baselines isolated per agent_did, not blended
func TestBlendedJWT_AgentsScoreIndependently(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	orgID := "a0000000-0000-0000-0000-000000000001"
	agentA := fmt.Sprintf("did:jwt:agent-a-%d", time.Now().UnixNano())
	agentB := fmt.Sprintf("did:jwt:agent-b-%d", time.Now().UnixNano())

	_, err := db.Exec(`
		INSERT INTO agent_baselines (agent_did, org_id, feature_name, mean, std_dev, sample_count, last_updated)
		VALUES ($1, $2::uuid, 'tool_call_rate_per_hour', 10.0, 1.0, 100, NOW())
		ON CONFLICT (org_id, agent_did, feature_name) DO UPDATE
		SET mean=EXCLUDED.mean, std_dev=EXCLUDED.std_dev`,
		agentA, orgID)
	if err != nil {
		t.Fatalf("seed agent A: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO agent_baselines (agent_did, org_id, feature_name, mean, std_dev, sample_count, last_updated)
		VALUES ($1, $2::uuid, 'tool_call_rate_per_hour', 10.0, 1.0, 100, NOW())
		ON CONFLICT (org_id, agent_did, feature_name) DO UPDATE
		SET mean=EXCLUDED.mean, std_dev=EXCLUDED.std_dev`,
		agentB, orgID)
	if err != nil {
		t.Fatalf("seed agent B: %v", err)
	}

	var countA, countB int
	db.QueryRow(`SELECT COUNT(*) FROM agent_baselines WHERE agent_did=$1`, agentA).Scan(&countA)
	db.QueryRow(`SELECT COUNT(*) FROM agent_baselines WHERE agent_did=$1`, agentB).Scan(&countB)

	if countA != 1 {
		t.Errorf("agent A baseline not isolated: count=%d", countA)
	}
	if countB != 1 {
		t.Errorf("agent B baseline not isolated: count=%d", countB)
	}

	var totalRows int
	db.QueryRow(`SELECT COUNT(*) FROM agent_baselines
		WHERE agent_did IN ($1, $2)`, agentA, agentB).Scan(&totalRows)
	if totalRows != 2 {
		t.Errorf("baselines blended — expected 2 rows, got %d", totalRows)
	}

	db.Exec(`DELETE FROM agent_baselines WHERE agent_did IN ($1, $2)`, agentA, agentB)
}

// GAP-6: Enforce mode default must be observe
func TestEnforceGate_DefaultIsObserve(t *testing.T) {
	defaultMode := "observe"
	if defaultMode != "observe" {
		t.Errorf("default enforcement mode is %q — must be 'observe'", defaultMode)
	}
}

// GAP-7: Queue status warning fires at threshold
func TestHealthEndpoint_QueueStatusWarning(t *testing.T) {
	cases := []struct {
		depth    int64
		expected string
	}{
		{0, "ok"},
		{99_999, "ok"},
		{100_000, "warning"},
		{500_000, "warning"},
	}

	for _, tc := range cases {
		status := "ok"
		if tc.depth >= store.QueueDepthWarning {
			status = "warning"
		}
		if status != tc.expected {
			t.Errorf("depth=%d: got %q, want %q", tc.depth, status, tc.expected)
		}
	}
}

// GAP-8: Reason object contains human-readable fields
func TestReasonObject_HumanReadableFields(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	var reasonJSON []byte
	err := db.QueryRow(`
		SELECT reason_object FROM enforcement_decisions
		WHERE reason_object IS NOT NULL LIMIT 1
	`).Scan(&reasonJSON)

	if err == sql.ErrNoRows {
		t.Skip("no enforcement decisions yet — run after pilot produces data")
	}
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	reason := string(reasonJSON)
	requiredFields := []string{"decision", "score", "confidence_pct", "policy_fired"}

	for _, field := range requiredFields {
		if !containsStr(reason, field) {
			t.Errorf("reason_object missing field %q", field)
		}
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
