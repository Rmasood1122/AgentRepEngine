package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/agentrepengine/are/internal/audit"
	_ "github.com/lib/pq"
)

const usage = `
cmd/verify-decision — ARE Merkle Selective Proof Verifier
HIPAA §164.528 | GDPR Article 22 | EU AI Act Article 86

Usage:
  verify-decision <decision_id>
  verify-decision <decision_id> --window-days <N>
  verify-decision --help

Arguments:
  decision_id     UUID of the enforcement decision to verify
  --window-days   Days window to search (default: 1)
  --json          Output raw JSON proof

Examples:
  verify-decision 550e8400-e29b-41d4-a716-446655440000
  verify-decision 550e8400-e29b-41d4-a716-446655440000 --window-days 7
  verify-decision 550e8400-e29b-41d4-a716-446655440000 --json

Exit codes:
  0 = VALID proof
  1 = INVALID proof or decision not found
  2 = usage error
  3 = database connection error
`

func main() {
	if len(os.Args) < 2 || os.Args[1] == "--help" {
		fmt.Print(usage)
		os.Exit(2)
	}

	decisionID := os.Args[1]
	windowDays := 1
	outputJSON := false

	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--window-days":
			if i+1 < len(os.Args) {
				fmt.Sscanf(os.Args[i+1], "%d", &windowDays)
				i++
			}
		case "--json":
			outputJSON = true
		}
	}

	// Connect to PostgreSQL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://are:are_dev_password@localhost:5433/agentrepengine?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: database connection failed: %v\n", err)
		os.Exit(3)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: database unreachable: %v\n", err)
		os.Exit(3)
	}

	// Define time window around now
	windowEnd := time.Now().UTC()
	windowStart := windowEnd.AddDate(0, 0, -windowDays)

	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("  ARE Merkle Selective Proof Verifier\n")
	fmt.Printf("  Decision ID: %s\n", decisionID)
	fmt.Printf("  Window: %s → %s\n",
		windowStart.Format("2006-01-02"),
		windowEnd.Format("2006-01-02"))
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	fmt.Printf("  [1/3] Building Merkle tree over window...\n")

	// Generate Merkle proof
	proof, err := audit.GenerateProof(db, decisionID, windowStart, windowEnd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n  RESULT: INVALID\n")
		fmt.Fprintf(os.Stderr, "  Reason: %v\n\n", err)
		os.Exit(1)
	}

	fmt.Printf("  [2/3] Verifying proof path...\n")

	// Verify the proof
	valid := audit.VerifyProof(proof)

	fmt.Printf("  [3/3] Checking Merkle root...\n\n")

	if outputJSON {
		proofJSON, err := audit.ProofToJSON(proof)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: proof serialization failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(proofJSON)
		if !valid {
			os.Exit(1)
		}
		return
	}

	// Human-readable output
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	if valid {
		fmt.Printf("  RESULT: ✅ VALID\n\n")
		fmt.Printf("  Decision %s EXISTS in the audit trail.\n", decisionID)
		fmt.Printf("  Merkle root: %s\n", truncate(proof.RootHash, 32))
		fmt.Printf("  Proof steps: %d\n", len(proof.ProofPath))
		fmt.Printf("  Window: %s → %s\n",
			proof.WindowStart.Format("2006-01-02 15:04:05 UTC"),
			proof.WindowEnd.Format("2006-01-02 15:04:05 UTC"))
		fmt.Printf("\n  This proof confirms the enforcement decision existed\n")
		fmt.Printf("  in the audit trail without revealing any other decisions.\n")
		fmt.Printf("\n  Compliance use:\n")
		fmt.Printf("    HIPAA §164.528  — specific incident disclosure\n")
		fmt.Printf("    GDPR Art.22     — selective enforcement evidence\n")
		fmt.Printf("    EU AI Act 86    — individual decision accountability\n")
	} else {
		fmt.Printf("  RESULT: ❌ INVALID\n\n")
		fmt.Printf("  Decision %s FAILED Merkle proof verification.\n", decisionID)
		fmt.Printf("  The audit trail may have been tampered with.\n")
		fmt.Printf("  Run cmd/verify-chain for full chain integrity check.\n")
	}
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	if !valid {
		os.Exit(1)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// outputSummary prints a JSON summary for programmatic use
func outputSummary(decisionID string, valid bool, proof *audit.MerkleProof) {
	summary := map[string]interface{}{
		"decision_id": decisionID,
		"valid":       valid,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}
	if proof != nil {
		fmt.Printf("  Merkle root: %s\n", truncate(proof.RootHash, 32))
		fmt.Printf("  Proof steps: %d\n", len(proof.ProofPath))
		summary["window_start"] = proof.WindowStart.Format(time.RFC3339)
		summary["window_end"] = proof.WindowEnd.Format(time.RFC3339)
	}
	data, _ := json.MarshalIndent(summary, "", "  ")
	fmt.Println(string(data))
}
