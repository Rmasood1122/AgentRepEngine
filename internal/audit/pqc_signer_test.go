package audit

import (
	"testing"
)

func TestGeneratePQCKeyPair(t *testing.T) {
	kp, err := GeneratePQCKeyPair()
	if err != nil {
		t.Fatalf("GeneratePQCKeyPair error: %v", err)
	}
	if kp.KeyID == "" {
		t.Fatal("KeyID must not be empty")
	}
	if kp.PublicKey == "" {
		t.Fatal("PublicKey must not be empty")
	}
	if kp.Scheme != SPHINCSScheme {
		t.Fatalf("expected scheme %s, got %s", SPHINCSScheme, kp.Scheme)
	}
	if kp.NISTLevel != 1 {
		t.Fatalf("expected NIST level 1, got %d", kp.NISTLevel)
	}
}

func TestSignAndVerify(t *testing.T) {
	kp, err := GeneratePQCKeyPair()
	if err != nil {
		t.Fatalf("key gen error: %v", err)
	}

	msg := []byte("test-event|agent-001|BLOCK|187|abc123|merkle-leaf-xyz")
	sig, err := kp.Sign(msg)
	if err != nil {
		t.Fatalf("Sign error: %v", err)
	}
	if sig.Signature == "" {
		t.Fatal("signature must not be empty")
	}
	if sig.MessageHash == "" {
		t.Fatal("message hash must not be empty")
	}

	result, err := Verify(kp.PublicKey, msg, sig)
	if err != nil {
		t.Fatalf("Verify error: %v", err)
	}
	if !result.Valid {
		t.Fatal("signature should be valid")
	}
	if !result.QuantumSafe {
		t.Fatal("result should be marked quantum safe")
	}
}

func TestVerifyTamperedMessage(t *testing.T) {
	kp, _ := GeneratePQCKeyPair()
	msg := []byte("original message")
	sig, _ := kp.Sign(msg)

	tampered := []byte("tampered message")
	result, err := Verify(kp.PublicKey, tampered, sig)
	if err != nil {
		t.Fatalf("Verify error: %v", err)
	}
	if result.Valid {
		t.Fatal("tampered message should not verify")
	}
}

func TestVerifyNilSignature(t *testing.T) {
	_, err := Verify("pubkey", []byte("msg"), nil)
	if err == nil {
		t.Fatal("nil signature should return error")
	}
}

func TestVerifyEmptyMessage(t *testing.T) {
	kp, _ := GeneratePQCKeyPair()
	sig, _ := kp.Sign([]byte("valid"))
	_, err := Verify(kp.PublicKey, []byte{}, sig)
	if err == nil {
		t.Fatal("empty message should return error")
	}
}

func TestSignEmptyMessage(t *testing.T) {
	kp, _ := GeneratePQCKeyPair()
	_, err := kp.Sign([]byte{})
	if err == nil {
		t.Fatal("signing empty message should return error")
	}
}

func TestSignNilKeyPair(t *testing.T) {
	var kp *PQCKeyPair
	_, err := kp.Sign([]byte("msg"))
	if err == nil {
		t.Fatal("nil keypair should return error")
	}
}

func TestSignAuditEvent(t *testing.T) {
	kp, _ := GeneratePQCKeyPair()
	event, err := SignAuditEvent(kp,
		"evt-001", "agent-001", "BLOCK", 187,
		[]byte(`{"policy":"bulk_pii"}`), "merkle-leaf-abc",
	)
	if err != nil {
		t.Fatalf("SignAuditEvent error: %v", err)
	}
	if event.EventID != "evt-001" {
		t.Fatal("wrong event ID")
	}
	if event.Signature.Signature == "" {
		t.Fatal("signature must not be empty")
	}
}

func TestVerifyAuditEvent(t *testing.T) {
	kp, _ := GeneratePQCKeyPair()
	event, _ := SignAuditEvent(kp,
		"evt-002", "agent-002", "ALLOW", 850,
		[]byte(`{"policy":"none"}`), "",
	)

	result, err := VerifyAuditEvent(kp.PublicKey, event)
	if err != nil {
		t.Fatalf("VerifyAuditEvent error: %v", err)
	}
	if !result.Valid {
		t.Fatal("audit event signature should be valid")
	}
}

func TestVerifyAuditEventNil(t *testing.T) {
	_, err := VerifyAuditEvent("pubkey", nil)
	if err == nil {
		t.Fatal("nil event should return error")
	}
}

func TestExportSignedEventJSON(t *testing.T) {
	kp, _ := GeneratePQCKeyPair()
	event, _ := SignAuditEvent(kp, "evt-003", "agent-003", "THROTTLE", 450,
		[]byte(`{}`), "leaf-xyz",
	)
	out, err := ExportSignedEventJSON(event)
	if err != nil {
		t.Fatalf("ExportSignedEventJSON error: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("JSON output must not be empty")
	}
}

func TestPQCReadinessReport(t *testing.T) {
	report := PQCReadinessReport()
	if report["scheme"] != SPHINCSScheme {
		t.Fatal("wrong scheme in readiness report")
	}
	if report["quantum_safe"] != true {
		t.Fatal("quantum_safe must be true")
	}
	if report["nist_level"] != SPHINCSNISTLevel {
		t.Fatal("wrong NIST level in readiness report")
	}
}

func TestKeyIDUnique(t *testing.T) {
	kp1, _ := GeneratePQCKeyPair()
	kp2, _ := GeneratePQCKeyPair()
	if kp1.KeyID == kp2.KeyID {
		t.Fatal("key IDs must be unique across keypairs")
	}
}
