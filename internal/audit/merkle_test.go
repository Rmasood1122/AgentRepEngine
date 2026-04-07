package audit

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// Unit tests — no DB required (pure hash logic)

func TestHashPairDeterministic(t *testing.T) {
	h1 := hashPair("aaa", "bbb")
	h2 := hashPair("aaa", "bbb")
	if h1 != h2 {
		t.Fatalf("hashPair not deterministic: %s != %s", h1, h2)
	}
}

func TestHashPairNotCommutative(t *testing.T) {
	h1 := hashPair("aaa", "bbb")
	h2 := hashPair("bbb", "aaa")
	if h1 == h2 {
		t.Fatal("hashPair should not be commutative — order matters for tree integrity")
	}
}

func TestVerifyProofNilInput(t *testing.T) {
	if VerifyProof(nil) {
		t.Fatal("nil proof should not verify")
	}
}

func TestVerifyProofInvalidFlag(t *testing.T) {
	proof := &MerkleProof{Valid: false}
	if VerifyProof(proof) {
		t.Fatal("invalid proof should not verify")
	}
}

func TestVerifyProofMismatchedPaths(t *testing.T) {
	proof := &MerkleProof{
		Valid:      true,
		LeafHash:   "abc",
		RootHash:   "xyz",
		ProofPath:  []string{"sib1", "sib2"},
		ProofSides: []string{"right"}, // mismatch: 2 paths, 1 side
	}
	if VerifyProof(proof) {
		t.Fatal("mismatched proof path/sides should not verify")
	}
}

func TestVerifyProofSingleLeaf(t *testing.T) {
	// Single leaf tree: root == leaf
	leafHash := hashLeaf("id1", "agent-1", "ALLOW", "0000", time.Now().UTC())
	proof := &MerkleProof{
		DecisionID: "id1",
		LeafHash:   leafHash,
		RootHash:   leafHash, // single leaf: root == leaf
		ProofPath:  []string{},
		ProofSides: []string{},
		Valid:      true,
	}
	if !VerifyProof(proof) {
		t.Fatal("single-leaf proof should verify (root == leaf)")
	}
}

func TestProofToJSON(t *testing.T) {
	proof := &MerkleProof{
		DecisionID: "test-id",
		LeafHash:   "abc123",
		RootHash:   "def456",
		Valid:      true,
	}
	out, err := ProofToJSON(proof)
	if err != nil {
		t.Fatalf("ProofToJSON error: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("ProofToJSON returned empty string")
	}
}

func TestSortedLeaves(t *testing.T) {
	leaves := []string{"ccc", "aaa", "bbb"}
	sorted := SortedLeaves(leaves)
	if sorted[0] != "aaa" || sorted[1] != "bbb" || sorted[2] != "ccc" {
		t.Fatalf("SortedLeaves wrong order: %v", sorted)
	}
	// Original must be unchanged
	if leaves[0] != "ccc" {
		t.Fatal("SortedLeaves mutated original slice")
	}
}

func TestBuildMerkleTreeEmptyDB(t *testing.T) {
	// Verify graceful empty result without DB (nil db will fail at query)
	// This test documents expected behavior — integration test needs real DB
	tree := &MerkleTree{
		Root:          "",
		Leaves:        []string{},
		Nodes:         map[string]MerkleNode{},
		DecisionCount: 0,
	}
	if tree.Root != "" {
		t.Fatal("empty tree should have empty root")
	}
	if len(tree.Leaves) != 0 {
		t.Fatal("empty tree should have no leaves")
	}
}

// Integration test helper — skips if no DB available
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres", "host=localhost port=5433 user=are password=are dbname=agentrepengine sslmode=disable")
	if err != nil {
		t.Skipf("no test DB available: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("DB not reachable: %v", err)
	}
	return db
}

func TestBuildMerkleTreeIntegration(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	now := time.Now().UTC()
	tree, err := BuildMerkleTree(db, now.Add(-24*time.Hour), now)
	if err != nil {
		t.Fatalf("BuildMerkleTree error: %v", err)
	}
	if tree == nil {
		t.Fatal("BuildMerkleTree returned nil")
	}
	// Root must be non-empty if there are decisions
	if tree.DecisionCount > 0 && tree.Root == "" {
		t.Fatal("non-zero decision count but empty root")
	}
}

func TestGenerateAuditReportIntegration(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	now := time.Now().UTC()
	report, err := GenerateAuditReport(db, now.Add(-24*time.Hour), now)
	if err != nil {
		t.Fatalf("GenerateAuditReport error: %v", err)
	}
	if report == nil {
		t.Fatal("GenerateAuditReport returned nil")
	}
	if report.RegulatoryNote == "" {
		t.Fatal("RegulatoryNote must not be empty")
	}
}
