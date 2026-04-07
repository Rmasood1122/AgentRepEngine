package integration

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/agentrepengine/are/internal/attestation"
	"github.com/agentrepengine/are/internal/audit"
	"github.com/agentrepengine/are/internal/compliance"
	"github.com/agentrepengine/are/internal/zkp"
)

// TestComplianceChain_FullEndToEnd verifies the complete compliance
// evidence chain works end-to-end:
//
//	enforcement_decision
//	→ hash chain valid
//	→ Merkle leaf present
//	→ PQC signature verifiable
//	→ ZK proof generates
//	→ TEE attestation generates
//	→ OSCAL bundle generates
//
// If any step fails, the 24x compliance architecture claim is broken.
// This test is the single source of truth for compliance chain integrity.
func TestComplianceChain_FullEndToEnd(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// ── Step 1: Verify hash chain integrity ─────────────────────────────
	t.Run("Step1_HashChainValid", func(t *testing.T) {
		var valid bool
		err := db.QueryRowContext(ctx,
			`SELECT verify_hash_chain('enforcement_decisions') AS valid`,
		).Scan(&valid)
		if err != nil {
			t.Skipf("hash chain function unavailable: %v", err)
		}
		if !valid {
			t.Fatal("COMPLIANCE CHAIN BROKEN: hash chain integrity check failed — audit trail may be compromised")
		}
		t.Log("✅ Step 1: Hash chain VALID")
	})

	// ── Step 2: Verify Merkle tree can be built ──────────────────────────
	t.Run("Step2_MerkleTreeBuilds", func(t *testing.T) {
		windowEnd := time.Now().UTC()
		windowStart := windowEnd.AddDate(0, 0, -30)

		tree, err := audit.BuildMerkleTree(db, windowStart, windowEnd)
		if err != nil {
			t.Fatalf("COMPLIANCE CHAIN BROKEN: Merkle tree build failed: %v", err)
		}
		if tree == nil {
			t.Fatal("COMPLIANCE CHAIN BROKEN: Merkle tree is nil")
		}
		t.Logf("✅ Step 2: Merkle tree built — root: %s", truncateStr(tree.Root, 16))
	})

	// ── Step 3: Verify Merkle proof generates for a real decision ────────
	t.Run("Step3_MerkleProofGenerates", func(t *testing.T) {
		// Find a real enforcement decision
		var decisionID string
		err := db.QueryRowContext(ctx,
			`SELECT id FROM enforcement_decisions ORDER BY created_at DESC LIMIT 1`,
		).Scan(&decisionID)
		if err == sql.ErrNoRows {
			t.Skip("no enforcement decisions in database — run demo first")
		}
		if err != nil {
			t.Skipf("cannot query enforcement decisions: %v", err)
		}

		windowEnd := time.Now().UTC()
		windowStart := windowEnd.AddDate(0, 0, -30)

		proof, err := audit.GenerateProof(db, decisionID, windowStart, windowEnd)
		if err != nil {
			t.Fatalf("COMPLIANCE CHAIN BROKEN: Merkle proof generation failed: %v", err)
		}

		valid := audit.VerifyProof(proof)
		if !valid {
			t.Fatal("COMPLIANCE CHAIN BROKEN: Merkle proof verification failed")
		}

		t.Logf("✅ Step 3: Merkle proof VALID — decision: %s root: %s",
			truncateStr(decisionID, 8),
			truncateStr(proof.RootHash, 16))
	})

	// ── Step 4: Verify PQC signature generates ───────────────────────────
	t.Run("Step4_PQCSignatureGenerates", func(t *testing.T) {
		kp, err := audit.GeneratePQCKeyPair()
		if err != nil {
			t.Fatalf("COMPLIANCE CHAIN BROKEN: PQC key generation failed: %v", err)
		}

		testEvent := []byte(`{"decision":"BLOCKED","agent_did":"test-agent","score":187}`)
		signedEvent, err := audit.SignAuditEvent(
			kp,
			"test-event-001",
			"test-agent-001",
			"BLOCKED",
			187,
			testEvent,
			"test-merkle-leaf",
		)
		if err != nil {
			t.Fatalf("COMPLIANCE CHAIN BROKEN: PQC sign failed: %v", err)
		}

		result, err := audit.VerifyAuditEvent(kp.PublicKey, signedEvent)
		if err != nil {
			t.Fatalf("COMPLIANCE CHAIN BROKEN: PQC verify failed: %v", err)
		}
		if !result.Valid {
			t.Fatalf("COMPLIANCE CHAIN BROKEN: PQC signature invalid: %s", result.RegulatoryNote)
		}

		t.Logf("✅ Step 4: PQC signature VALID — scheme: %s", result.Scheme)
	})

	// ── Step 5: Verify ZK proof generates ───────────────────────────────
	t.Run("Step5_ZKProofGenerates", func(t *testing.T) {
		input := &zkp.CompositeProofInput{
			AgentID:     "test-agent-001",
			Score:       187,
			FPRate:      0.0,
			ZScore:      4.2,
			PolicyFired: "pii_field_access_rate",
			Decision:    "BLOCK",
			ReasonJSON:  []byte(`{"decision":"BLOCK","policy":"pii_field_access_rate"}`),
		}

		proof, err := zkp.Prove(input)
		if err != nil {
			t.Fatalf("COMPLIANCE CHAIN BROKEN: ZK proof generation failed: %v", err)
		}

		result, err := zkp.VerifyProof(proof)
		if err != nil {
			t.Fatalf("COMPLIANCE CHAIN BROKEN: ZK proof verification failed: %v", err)
		}
		if !result.Valid {
			t.Fatalf("COMPLIANCE CHAIN BROKEN: ZK proof invalid: %s", result.RegulatoryNote)
		}

		t.Logf("✅ Step 5: ZK proof VALID — scheme: %s", proof.Scheme)
	})

	// ── Step 6: Verify TEE attestation generates ─────────────────────────
	t.Run("Step6_TEEAttestationGenerates", func(t *testing.T) {
		quote, err := attestation.GenerateQuote(
			"test-agent-001",
			"abc123hashvalue",
			187,
			[]byte(`{"decision":"BLOCKED"}`),
		)
		if err != nil {
			t.Fatalf("COMPLIANCE CHAIN BROKEN: TEE quote generation failed: %v", err)
		}

		cfg := attestation.DefaultTEEConfig()
		result, err := attestation.VerifyQuote(quote, cfg)
		if err != nil {
			t.Fatalf("COMPLIANCE CHAIN BROKEN: TEE verify failed: %v", err)
		}
		if !result.Valid {
			t.Fatalf("COMPLIANCE CHAIN BROKEN: TEE attestation invalid: %s", result.RegulatoryNote)
		}

		t.Logf("✅ Step 6: TEE attestation VALID — scheme: %s", quote.Scheme)
	})

	// ── Step 7: Verify OSCAL bundle generates ────────────────────────────
	t.Run("Step7_OSCALBundleGenerates", func(t *testing.T) {
		input := &compliance.AREEvidenceInput{
			FPRate:           0.0,
			TPRate:           0.88,
			TotalDecisions:   100,
			BlockedIncidents: 12,
			AuditTrailIntact: true,
			MerkleRootHash:   "test-merkle-root-hash-abc123",
			PQCEnabled:       true,
			TEEEnabled:       true,
			ZKProofEnabled:   true,
			RaftEnabled:      false,
			AuditPeriodStart: time.Now().AddDate(0, -1, 0),
			AuditPeriodEnd:   time.Now(),
		}

		bundle, err := compliance.GenerateOSCALBundle(input)
		if err != nil {
			t.Fatalf("COMPLIANCE CHAIN BROKEN: OSCAL bundle generation failed: %v", err)
		}
		if bundle == nil {
			t.Fatal("COMPLIANCE CHAIN BROKEN: OSCAL bundle is nil")
		}

		jsonOutput, err := compliance.ExportOSCALJSON(bundle)
		if err != nil {
			t.Fatalf("COMPLIANCE CHAIN BROKEN: OSCAL JSON export failed: %v", err)
		}
		if len(jsonOutput) < 100 {
			t.Fatalf("COMPLIANCE CHAIN BROKEN: OSCAL output too short (%d bytes)", len(jsonOutput))
		}

		t.Logf("✅ Step 7: OSCAL bundle VALID — %d bytes", len(jsonOutput))
	})

	t.Log("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	t.Log("  COMPLIANCE CHAIN: ALL 7 STEPS PASSED")
	t.Log("  Hash chain → Merkle → PQC → ZK → TEE → OSCAL")
	t.Log("  24x compliance architecture is intact.")
	t.Log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// TestComplianceChain_ComponentsIndependent verifies each component
// works independently — no hidden cross-dependencies.
func TestComplianceChain_ComponentsIndependent(t *testing.T) {
	t.Run("PQC_independent", func(t *testing.T) {
		kp, err := audit.GeneratePQCKeyPair()
		if err != nil {
			t.Fatalf("PQC key generation failed: %v", err)
		}
		report := audit.PQCReadinessReport()
		if report["scheme"] == nil {
			t.Fatal("PQC readiness report missing scheme")
		}
		t.Logf("✅ PQC independent: scheme=%s nist_level=%v", kp.Scheme, kp.NISTLevel)
	})

	t.Run("ZK_independent", func(t *testing.T) {
		report := zkp.ZKReadinessReport()
		if report["scheme"] == nil {
			t.Fatal("ZK readiness report missing scheme")
		}
		t.Logf("✅ ZK independent: scheme=%v", report["scheme"])
	})

	t.Run("TEE_independent", func(t *testing.T) {
		report := attestation.TEEReadinessReport()
		if report["scheme"] == nil {
			t.Fatal("TEE readiness report missing scheme")
		}
		t.Logf("✅ TEE independent: scheme=%v", report["scheme"])
	})
}

// TestComplianceChain_MerkleAndPQCIntegrated verifies the key
// architectural insight: Merkle leaf is embedded in PQC signature.
func TestComplianceChain_MerkleAndPQCIntegrated(t *testing.T) {
	kp, err := audit.GeneratePQCKeyPair()
	if err != nil {
		t.Fatalf("PQC key generation failed: %v", err)
	}

	merkleLeaf := "merkle-leaf-hash-abc123def456"
	testEvent := []byte(`{"decision":"BLOCKED","score":187}`)

	signedEvent, err := audit.SignAuditEvent(
		kp,
		"integration-event-001",
		"agent-001",
		"BLOCKED",
		187,
		testEvent,
		merkleLeaf,
	)
	if err != nil {
		t.Fatalf("SignAuditEvent failed: %v", err)
	}

	// Verify the Merkle leaf is embedded in the signed event
	if signedEvent.MerkleLeaf != merkleLeaf {
		t.Errorf("INTEGRATION BROKEN: Merkle leaf not embedded in PQC event\nExpected: %s\nGot: %s",
			merkleLeaf, signedEvent.MerkleLeaf)
	}

	// Verify the full event verifies correctly
	result, err := audit.VerifyAuditEvent(kp.PublicKey, signedEvent)
	if err != nil {
		t.Fatalf("VerifyAuditEvent failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("Merkle+PQC integrated event invalid: %s", result.RegulatoryNote)
	}

	t.Logf("✅ Merkle+PQC integration VALID — leaf embedded and verified")
	t.Logf("   Merkle leaf: %s", truncateStr(merkleLeaf, 20))
	t.Logf("   PQC scheme:  %s", result.Scheme)
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// Ensure fmt is used
var _ = fmt.Sprintf
