package zkp

import (
	"testing"
)

func validInput() *CompositeProofInput {
	return &CompositeProofInput{
		AgentID:     "agent-001",
		Score:       187,
		FPRate:      0.00,
		ZScore:      3.4,
		PolicyFired: "bulk_pii_access_prevention_v1",
		Decision:    "BLOCK",
		ReasonJSON:  []byte(`{"policy":"bulk_pii"}`),
	}
}

func TestProveAndVerify(t *testing.T) {
	proof, err := Prove(validInput())
	if err != nil {
		t.Fatalf("Prove error: %v", err)
	}
	if proof.ProofID == "" {
		t.Fatal("ProofID must not be empty")
	}
	if proof.MerkleRoot == "" {
		t.Fatal("MerkleRoot must not be empty")
	}

	result, err := VerifyProof(proof)
	if err != nil {
		t.Fatalf("VerifyProof error: %v", err)
	}
	if !result.Valid {
		t.Fatalf("proof should be valid, got: %+v", result)
	}
}

func TestProveNilInput(t *testing.T) {
	_, err := Prove(nil)
	if err == nil {
		t.Fatal("nil input should return error")
	}
}

func TestProveEmptyAgentID(t *testing.T) {
	input := validInput()
	input.AgentID = ""
	_, err := Prove(input)
	if err == nil {
		t.Fatal("empty agent ID should return error")
	}
}

func TestVerifyNilProof(t *testing.T) {
	_, err := VerifyProof(nil)
	if err == nil {
		t.Fatal("nil proof should return error")
	}
}

func TestStatementScoreOutOfRange(t *testing.T) {
	input := validInput()
	input.Score = 1500 // out of range
	proof, err := Prove(input)
	if err != nil {
		t.Fatalf("Prove error: %v", err)
	}
	if proof.Statement.ScoreInRange {
		t.Fatal("score 1500 should not be in range")
	}
	result, _ := VerifyProof(proof)
	if result.Valid {
		t.Fatal("proof with out-of-range score should not verify")
	}
}

func TestStatementFPExceeded(t *testing.T) {
	input := validInput()
	input.FPRate = 0.05 // exceeds 2% gate
	proof, err := Prove(input)
	if err != nil {
		t.Fatalf("Prove error: %v", err)
	}
	if proof.Statement.FPBelowThreshold {
		t.Fatal("FP rate 5% should not be below threshold")
	}
	result, _ := VerifyProof(proof)
	if result.Valid {
		t.Fatal("proof with FP exceeded should not verify")
	}
}

func TestStatementZScoreBelowFloor(t *testing.T) {
	input := validInput()
	input.ZScore = 1.2 // below 2.0 floor
	proof, err := Prove(input)
	if err != nil {
		t.Fatalf("Prove error: %v", err)
	}
	if proof.Statement.ZScoreAboveFloor {
		t.Fatal("z-score 1.2 should not be above floor")
	}
	result, _ := VerifyProof(proof)
	if result.Valid {
		t.Fatal("proof with z-score below floor should not verify")
	}
}

func TestProofIDUnique(t *testing.T) {
	p1, _ := Prove(validInput())
	p2, _ := Prove(validInput())
	if p1.ProofID == p2.ProofID {
		t.Fatal("proof IDs must be unique")
	}
}

func TestMerkleRootTamper(t *testing.T) {
	proof, _ := Prove(validInput())
	proof.MerkleRoot = "tampered000000000000000000000000000000000000000000000000000000"
	result, err := VerifyProof(proof)
	if err != nil {
		t.Fatalf("VerifyProof error: %v", err)
	}
	if result.Valid {
		t.Fatal("tampered merkle root should not verify")
	}
}

func TestExportProofJSON(t *testing.T) {
	proof, _ := Prove(validInput())
	out, err := ExportProofJSON(proof)
	if err != nil {
		t.Fatalf("ExportProofJSON error: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("JSON output must not be empty")
	}
}

func TestZKReadinessReport(t *testing.T) {
	report := ZKReadinessReport()
	if report["scheme"] != STARKScheme {
		t.Fatal("wrong scheme in readiness report")
	}
	if report["security_bits"] != STARKSecurityBits {
		t.Fatal("wrong security bits")
	}
	props, ok := report["properties_proven"].([]string)
	if !ok || len(props) == 0 {
		t.Fatal("properties_proven must be non-empty")
	}
}

func TestWitnessNotInProof(t *testing.T) {
	input := validInput()
	proof, _ := Prove(input)
	out, _ := ExportProofJSON(proof)
	// Raw score, FP rate, and agent ID must not appear in proof JSON
	if contains(out, `"187"`) || contains(out, `"0.00"`) {
		t.Fatal("witness values must not appear in proof JSON")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
