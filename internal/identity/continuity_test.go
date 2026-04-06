package identity

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSealAgentIdentity(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE agent_baselines").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "agent-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	seal, err := SealAgentIdentity(db, "agent-001", "baselinehash123", "jti-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if seal.AgentID != "agent-001" {
		t.Errorf("expected agent-001, got %s", seal.AgentID)
	}
	if seal.Seal == "" {
		t.Error("expected non-empty seal")
	}
	if seal.SealedAt.IsZero() {
		t.Error("expected non-zero sealed_at")
	}
}

func TestVerifyContinuity_Match(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	expectedSeal := computeSeal("baselinehash123", "jti-abc")

	mock.ExpectQuery("SELECT identity_seal, sealed_at").
		WithArgs("agent-001").
		WillReturnRows(sqlmock.NewRows([]string{"identity_seal", "sealed_at"}).
			AddRow(expectedSeal, time.Now()))

	ok, err := VerifyContinuity(db, "agent-001", "baselinehash123", "jti-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected continuity verified")
	}
}

func TestVerifyContinuity_Mismatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT identity_seal, sealed_at").
		WithArgs("agent-001").
		WillReturnRows(sqlmock.NewRows([]string{"identity_seal", "sealed_at"}).
			AddRow("oldseal999", time.Now()))

	ok, err := VerifyContinuity(db, "agent-001", "newbaselinehash", "jti-xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected continuity failure — seal should not match")
	}
}

func TestVerifyContinuity_NoSeal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT identity_seal, sealed_at").
		WithArgs("agent-002").
		WillReturnRows(sqlmock.NewRows([]string{"identity_seal", "sealed_at"}).
			AddRow("", nil))

	ok, err := VerifyContinuity(db, "agent-002", "anyhash", "anyjti")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected false — no seal stored")
	}
}

func TestVerifyContinuity_AgentNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT identity_seal, sealed_at").
		WithArgs("ghost-agent").
		WillReturnError(sqlmock.ErrCancelled)

	_, err = VerifyContinuity(db, "ghost-agent", "hash", "jti")
	if err == nil {
		t.Error("expected error for missing agent")
	}
}
