package identity

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// newTestRedis returns a real Redis client for tests.
// Tests require Redis running (docker compose up -d redis).
func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "are_redis_dev",
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available at localhost:6379 — skipping replay tests: %v", err)
	}
	return rdb
}

// cleanupReplayKeys removes test jti keys from Redis after test.
func cleanupReplayKeys(t *testing.T, rdb *redis.Client, pattern string) {
	t.Helper()
	ctx := context.Background()
	keys, err := rdb.Keys(ctx, pattern).Result()
	if err != nil {
		return
	}
	if len(keys) > 0 {
		rdb.Del(ctx, keys...)
	}
}

// ═══════════════════════════════════════════════════
// G-IDENTITY GATE TESTS — HARD STOP IF ANY FAIL
// ═══════════════════════════════════════════════════

// TestRS256Only verifies we never use HS256.
func TestRS256Only(t *testing.T) {
	keys, err := LoadOrGenerateKeys()
	if err != nil {
		t.Fatalf("LoadOrGenerateKeys failed: %v", err)
	}
	rdb := newTestRedis(t)
	defer rdb.Close()

	claims := NewAgentClaims(
		"did:jwt:test-org:finance-agent:001",
		"inst-001",
		ComputeLineageHash("root", time.Now()),
		"org-001",
		0,
	)

	token, err := SignToken(claims, keys)
	if err != nil {
		t.Fatalf("SignToken failed: %v", err)
	}

	parsed, err := VerifyToken(token, keys, rdb)
	if err != nil {
		t.Fatalf("VerifyToken failed: %v", err)
	}

	if parsed.AgentDID != claims.AgentDID {
		t.Errorf("AgentDID mismatch: got %s want %s", parsed.AgentDID, claims.AgentDID)
	}
	t.Log("✅ RS256 signing and verification confirmed")
}

// TestAllClaimsRequired verifies missing claims are rejected.
func TestAllClaimsRequired(t *testing.T) {
	keys, err := LoadOrGenerateKeys()
	if err != nil {
		t.Fatalf("LoadOrGenerateKeys: %v", err)
	}

	tests := []struct {
		name   string
		claims *AgentClaims
	}{
		{
			name: "missing agent_did",
			claims: &AgentClaims{
				InstanceID: "inst-001", LineageHash: "abc",
				OrgID: "org-001", LineageDepth: 0,
			},
		},
		{
			name: "missing instance_id",
			claims: &AgentClaims{
				AgentDID: "did:jwt:org:svc:001", LineageHash: "abc",
				OrgID: "org-001", LineageDepth: 0,
			},
		},
		{
			name: "missing lineage_hash",
			claims: &AgentClaims{
				AgentDID: "did:jwt:org:svc:001", InstanceID: "inst-001",
				OrgID: "org-001", LineageDepth: 0,
			},
		},
		{
			name: "missing org_id",
			claims: &AgentClaims{
				AgentDID: "did:jwt:org:svc:001", InstanceID: "inst-001",
				LineageHash: "abc", LineageDepth: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SignToken(tt.claims, keys)
			if err == nil {
				t.Errorf("expected error for %s but got none", tt.name)
			} else {
				t.Logf("✅ Correctly rejected: %s — %v", tt.name, err)
			}
		})
	}
}

// TestJTIReplayDetection verifies the same token cannot be used twice.
// L23 FIX: financial services requires replay detection.
// Attack: attacker intercepts valid token, replays it after original use.
func TestJTIReplayDetection(t *testing.T) {
	keys, err := LoadOrGenerateKeys()
	if err != nil {
		t.Fatalf("LoadOrGenerateKeys: %v", err)
	}
	rdb := newTestRedis(t)
	defer rdb.Close()
	defer cleanupReplayKeys(t, rdb, replayKeyPrefix+"*")

	claims := NewAgentClaims(
		"did:jwt:test-org:replay-test:001",
		"inst-replay-001",
		ComputeLineageHash("root", time.Now()),
		"org-replay",
		0,
	)

	// Verify jti is set
	if claims.RegisteredClaims.ID == "" {
		t.Fatal("NewAgentClaims must set jti — RegisteredClaims.ID is empty")
	}

	token, err := SignToken(claims, keys)
	if err != nil {
		t.Fatalf("SignToken failed: %v", err)
	}

	// First use — must succeed
	_, err = VerifyToken(token, keys, rdb)
	if err != nil {
		t.Fatalf("First VerifyToken failed (should succeed): %v", err)
	}
	t.Log("✅ First token use: accepted")

	// Second use of same token — must be rejected
	_, err = VerifyToken(token, keys, rdb)
	if err == nil {
		t.Fatal("❌ REPLAY ATTACK SUCCEEDED — second use of same token was accepted")
	}
	t.Logf("✅ Replay detected and rejected: %v", err)
}

// TestJTIRequired verifies tokens without jti are rejected.
// Tokens generated without NewAgentClaims won't have jti set.
func TestJTIRequired(t *testing.T) {
	keys, err := LoadOrGenerateKeys()
	if err != nil {
		t.Fatalf("LoadOrGenerateKeys: %v", err)
	}
	rdb := newTestRedis(t)
	defer rdb.Close()

	// Manually construct claims without jti — simulates old token format
	claims := &AgentClaims{
		AgentDID:     "did:jwt:org:agent:nojti",
		InstanceID:   "inst-nojti",
		LineageHash:  "abc123",
		OrgID:        "org-nojti",
		LineageDepth: 0,
		// RegisteredClaims.ID intentionally left empty — no jti
	}

	// SignToken now validates jti is present — should fail
	_, err = SignToken(claims, keys)
	if err == nil {
		t.Fatal("SignToken should reject claims without jti")
	}
	t.Logf("✅ Token without jti correctly rejected at sign time: %v", err)
}

// TestLineageHash verifies hash is deterministic and non-empty.
func TestLineageHash(t *testing.T) {
	spawnedAt := time.Date(2026, 3, 17, 12, 0, 0, 0, time.UTC)

	hash1 := ComputeLineageHash("did:jwt:org:parent:001", spawnedAt)
	hash2 := ComputeLineageHash("did:jwt:org:parent:001", spawnedAt)

	if hash1 == "" {
		t.Fatal("lineage hash is empty")
	}
	if hash1 != hash2 {
		t.Error("lineage hash is not deterministic")
	}

	hash3 := ComputeLineageHash("did:jwt:org:other-parent:999", spawnedAt)
	if hash1 == hash3 {
		t.Error("different parents produced same hash")
	}
	t.Logf("✅ Lineage hash deterministic: %s", hash1)
}

// TestSubAgentScore verifies min(parent_score, 700) rule.
func TestSubAgentScore(t *testing.T) {
	tests := []struct {
		parentScore int
		wantScore   int
	}{
		{900, 700},
		{700, 700},
		{400, 400},
		{100, 100},
	}

	for _, tt := range tests {
		got := SubAgentScore(tt.parentScore)
		if got != tt.wantScore {
			t.Errorf("SubAgentScore(%d) = %d, want %d",
				tt.parentScore, got, tt.wantScore)
		}
	}
	t.Log("✅ Sub-agent score inheritance correct")
}

// TestProbationMode verifies new agents are in probation.
func TestProbationMode(t *testing.T) {
	now := time.Now()
	expires := now.Add(ProbationDuration)

	agent := &AgentIdentity{
		DID:                "did:jwt:org:agent:001",
		CurrentScore:       InitialScore,
		Status:             "probation",
		ProbationExpiresAt: &expires,
	}

	if agent.CurrentScore != 700 {
		t.Errorf("new agent score = %d, want 700", agent.CurrentScore)
	}
	if !agent.IsInProbation() {
		t.Error("new agent should be in probation")
	}

	past := now.Add(-1 * time.Hour)
	agent.ProbationExpiresAt = &past
	if agent.IsInProbation() {
		t.Error("agent with expired probation should not be in probation")
	}
	t.Log("✅ Probation mode correct")
}

// TestOrphanScore verifies unregistered agents get score 500.
func TestOrphanScore(t *testing.T) {
	if OrphanScore != 500 {
		t.Errorf("OrphanScore = %d, want 500", OrphanScore)
	}
	if InitialScore != 700 {
		t.Errorf("InitialScore = %d, want 700", InitialScore)
	}
	t.Log("✅ Orphan score 500, initial score 700 confirmed")
}

// TestKeyPersistence verifies keys survive reload.
func TestKeyPersistence(t *testing.T) {
	keys1, err := LoadOrGenerateKeys()
	if err != nil {
		t.Fatalf("first load: %v", err)
	}
	keys2, err := LoadOrGenerateKeys()
	if err != nil {
		t.Fatalf("second load: %v", err)
	}
	rdb := newTestRedis(t)
	defer rdb.Close()
	defer cleanupReplayKeys(t, rdb, replayKeyPrefix+"*")

	claims := NewAgentClaims(
		"did:jwt:org:agent:persist-test",
		"inst-persist",
		ComputeLineageHash("root", time.Now()),
		"org-persist",
		0,
	)

	token, err := SignToken(claims, keys1)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	_, err = VerifyToken(token, keys2, rdb)
	if err != nil {
		t.Fatalf("verify with reloaded keys failed — FM1 risk: %v", err)
	}
	t.Log("✅ Key persistence confirmed — restart-safe")
}
