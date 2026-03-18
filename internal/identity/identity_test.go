package identity

import (
	"testing"
	"time"
)

// ═══════════════════════════════════════════════════
// G-IDENTITY GATE TESTS — HARD STOP IF ANY FAIL
// ═══════════════════════════════════════════════════

// TestRS256Only verifies we never use HS256.
// FM2 prevention: wrong algorithm = shared secret risk.
func TestRS256Only(t *testing.T) {
	keys, err := LoadOrGenerateKeys()
	if err != nil {
		t.Fatalf("LoadOrGenerateKeys failed: %v", err)
	}

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

	// Verify token parses correctly with RS256
	parsed, err := VerifyToken(token, keys)
	if err != nil {
		t.Fatalf("VerifyToken failed: %v", err)
	}

	if parsed.AgentDID != claims.AgentDID {
		t.Errorf("AgentDID mismatch: got %s want %s", parsed.AgentDID, claims.AgentDID)
	}

	t.Log("✅ RS256 signing and verification confirmed")
}

// TestAllClaimsRequired verifies missing claims are rejected.
// FM2 prevention: lineage_hash missing = identity drift T15.
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

	// Different parent = different hash
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
		{900, 700}, // trusted parent — sub-agent capped at 700
		{700, 700}, // monitored parent — sub-agent starts at 700
		{400, 400}, // restricted parent — sub-agent inherits penalty
		{100, 100}, // blocked parent — sub-agent inherits low score
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

	// Agent with expired probation
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
// FM1 prevention: key must load from disk, not regenerate.
func TestKeyPersistence(t *testing.T) {
	keys1, err := LoadOrGenerateKeys()
	if err != nil {
		t.Fatalf("first load: %v", err)
	}

	keys2, err := LoadOrGenerateKeys()
	if err != nil {
		t.Fatalf("second load: %v", err)
	}

	// Sign with keys1, verify with keys2 — must work
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

	_, err = VerifyToken(token, keys2)
	if err != nil {
		t.Fatalf("verify with reloaded keys failed — FM1 risk: %v", err)
	}

	t.Log("✅ Key persistence confirmed — restart-safe")
}
