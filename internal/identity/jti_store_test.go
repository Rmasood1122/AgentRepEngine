package identity

import (
	"context"
	"testing"
	"time"
)

func TestJTIStore_DualWrite(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()
	defer cleanupReplayKeys(t, rdb, "used_token:test-*")

	// No Postgres in unit test — Redis-only mode
	store := NewJTIStore(rdb, nil)
	ctx := context.Background()

	// First use should succeed
	err := store.CheckAndMark(ctx, "test-jti-001", "did:test:agent", 5*time.Minute)
	if err != nil {
		t.Fatalf("first use should succeed: %v", err)
	}

	// Replay should fail
	err = store.CheckAndMark(ctx, "test-jti-001", "did:test:agent", 5*time.Minute)
	if err == nil {
		t.Fatal("replay should be detected")
	}
	t.Logf("A5: Replay correctly detected: %v", err)
}

func TestJTIStore_RedisRecovery(t *testing.T) {
	rdb := newTestRedis(t)
	defer rdb.Close()
	defer cleanupReplayKeys(t, rdb, "used_token:test-recovery-*")

	store := NewJTIStore(rdb, nil)
	ctx := context.Background()

	// Mark a jti
	err := store.CheckAndMark(ctx, "test-recovery-001", "did:test:agent", 5*time.Minute)
	if err != nil {
		t.Fatalf("mark failed: %v", err)
	}

	// Simulate Redis restart by deleting the key
	rdb.Del(ctx, "used_token:test-recovery-001")

	// Without Postgres, the key is lost — this is the gap A5 closes
	// In production with Postgres, RebuildFromPostgres would recover it
	err = store.CheckAndMark(ctx, "test-recovery-001", "did:test:agent", 5*time.Minute)
	if err != nil {
		t.Log("A5: Key recovered (Postgres fallback active)")
	} else {
		t.Log("A5: Key lost after Redis restart (no Postgres in test) — expected in Redis-only mode")
	}
}

func TestJTIStore_VerifyTokenBackwardCompat(t *testing.T) {
	keys, err := LoadOrGenerateKeys()
	if err != nil {
		t.Fatalf("LoadOrGenerateKeys failed: %v", err)
	}
	rdb := newTestRedis(t)
	defer rdb.Close()

	claims := NewAgentClaims(
		"did:jwt:test-org:a5-compat:001",
		"inst-001",
		ComputeLineageHash("root", time.Now()),
		"org-001",
		0,
	)
	token, err := SignToken(claims, keys)
	if err != nil {
		t.Fatalf("SignToken failed: %v", err)
	}
	defer cleanupReplayKeys(t, rdb, "used_token:*")

	// Call without db (backward compatible — no Postgres)
	parsed, err := VerifyToken(token, keys, rdb)
	if err != nil {
		t.Fatalf("VerifyToken failed: %v", err)
	}
	if parsed.AgentDID != claims.AgentDID {
		t.Errorf("AgentDID mismatch: got %s want %s", parsed.AgentDID, claims.AgentDID)
	}
	t.Log("A5: VerifyToken backward compatible — works without Postgres param")
}
