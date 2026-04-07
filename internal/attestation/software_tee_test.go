package attestation

import (
	"testing"
	"time"
)

func TestGenerateQuoteBasic(t *testing.T) {
	quote, err := GenerateQuote("agent-001", "decision-hash-abc", 187, []byte(`{}`))
	if err != nil {
		t.Fatalf("GenerateQuote error: %v", err)
	}
	if quote.QuoteID == "" {
		t.Fatal("QuoteID must not be empty")
	}
	if quote.QuoteBody == "" {
		t.Fatal("QuoteBody must not be empty")
	}
	if quote.Measurement.MREnclave == "" {
		t.Fatal("MREnclave must not be empty")
	}
	if quote.ReportData.AgentID != "agent-001" {
		t.Fatal("wrong agent ID in report data")
	}
}

func TestGenerateQuoteEmptyAgentID(t *testing.T) {
	_, err := GenerateQuote("", "hash", 100, []byte(`{}`))
	if err == nil {
		t.Fatal("empty agent ID should return error")
	}
}

func TestGenerateQuoteEmptyDecisionHash(t *testing.T) {
	_, err := GenerateQuote("agent-001", "", 100, []byte(`{}`))
	if err == nil {
		t.Fatal("empty decision hash should return error")
	}
}

func TestVerifyQuoteValid(t *testing.T) {
	quote, err := GenerateQuote("agent-001", "decision-hash-abc", 187, []byte(`{}`))
	if err != nil {
		t.Fatalf("GenerateQuote error: %v", err)
	}

	result, err := VerifyQuote(quote, DefaultTEEConfig())
	if err != nil {
		t.Fatalf("VerifyQuote error: %v", err)
	}
	if !result.Valid {
		t.Fatalf("quote should be valid: %+v", result)
	}
	if !result.MeasurementMatch {
		t.Fatal("measurement should match")
	}
	if !result.TrustedEnclave {
		t.Fatal("enclave should be trusted")
	}
}

func TestVerifyQuoteNil(t *testing.T) {
	_, err := VerifyQuote(nil, DefaultTEEConfig())
	if err == nil {
		t.Fatal("nil quote should return error")
	}
}

func TestVerifyQuoteTamperedMREnclave(t *testing.T) {
	quote, _ := GenerateQuote("agent-001", "hash", 187, []byte(`{}`))
	quote.Measurement.MREnclave = "tampered000000000000000000000000000000000000000000000000000000"

	result, err := VerifyQuote(quote, DefaultTEEConfig())
	if err != nil {
		t.Fatalf("VerifyQuote error: %v", err)
	}
	if result.MeasurementMatch {
		t.Fatal("tampered MREnclave should not match")
	}
	if result.Valid {
		t.Fatal("tampered quote should not be valid")
	}
}

func TestVerifyQuoteTamperedBody(t *testing.T) {
	quote, _ := GenerateQuote("agent-001", "hash", 187, []byte(`{}`))
	quote.QuoteBody = "tampered000000000000000000000000000000000000000000000000000000"

	result, err := VerifyQuote(quote, DefaultTEEConfig())
	if err != nil {
		t.Fatalf("VerifyQuote error: %v", err)
	}
	if result.Valid {
		t.Fatal("tampered quote body should not be valid")
	}
}

func TestVerifyQuoteStaleTimestamp(t *testing.T) {
	quote, _ := GenerateQuote("agent-001", "hash", 187, []byte(`{}`))
	// Set timestamp 10 minutes in the past
	quote.ReportData.Timestamp = time.Now().UTC().Add(-10 * time.Minute).UnixNano()

	result, err := VerifyQuote(quote, DefaultTEEConfig())
	if err != nil {
		t.Fatalf("VerifyQuote error: %v", err)
	}
	if result.FreshnessValid {
		t.Fatal("stale timestamp should fail freshness check")
	}
}

func TestVerifyQuoteNilConfig(t *testing.T) {
	quote, _ := GenerateQuote("agent-001", "hash", 187, []byte(`{}`))
	// nil config should use DefaultTEEConfig
	result, err := VerifyQuote(quote, nil)
	if err != nil {
		t.Fatalf("VerifyQuote with nil config error: %v", err)
	}
	if !result.Valid {
		t.Fatal("quote with nil config should use default and be valid")
	}
}

func TestMREnclaveDeterministic(t *testing.T) {
	m1 := deriveMREnclave()
	m2 := deriveMREnclave()
	if m1 != m2 {
		t.Fatal("MREnclave must be deterministic")
	}
}

func TestQuoteIDUnique(t *testing.T) {
	q1, _ := GenerateQuote("agent-001", "hash-1", 187, []byte(`{}`))
	q2, _ := GenerateQuote("agent-001", "hash-2", 187, []byte(`{}`))
	if q1.QuoteID == q2.QuoteID {
		t.Fatal("quote IDs must be unique")
	}
}

func TestExportQuoteJSON(t *testing.T) {
	quote, _ := GenerateQuote("agent-001", "hash", 187, []byte(`{}`))
	out, err := ExportQuoteJSON(quote)
	if err != nil {
		t.Fatalf("ExportQuoteJSON error: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("JSON output must not be empty")
	}
}

func TestTEEReadinessReport(t *testing.T) {
	report := TEEReadinessReport()
	if report["scheme"] != TEEScheme {
		t.Fatal("wrong scheme in readiness report")
	}
	if report["mr_enclave"] == "" {
		t.Fatal("mr_enclave must not be empty in report")
	}
}

func TestRegulatoryNotePresent(t *testing.T) {
	quote, _ := GenerateQuote("agent-001", "hash", 187, []byte(`{}`))
	result, _ := VerifyQuote(quote, DefaultTEEConfig())
	if result.RegulatoryNote == "" {
		t.Fatal("RegulatoryNote must not be empty")
	}
}
