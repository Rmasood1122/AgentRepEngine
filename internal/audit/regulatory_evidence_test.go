package audit

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGenerateRegulatoryPackage_UnsupportedFramework(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = GenerateRegulatoryPackage(db, "org1", 30, "UNKNOWN")
	if err == nil {
		t.Fatal("expected error for unsupported framework")
	}
}

func TestGenerateRegulatoryPackage_HIPAA(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// hash chain query — now uses enforcement_decisions, no org_id
	mock.ExpectQuery("SELECT this_hash FROM enforcement_decisions").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"this_hash"}).AddRow("abc123"))

	// HIPAA query — no org_id
	mock.ExpectQuery("SELECT").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"phi", "access", "anomalous", "blocked"}).
			AddRow(3, 100, 5, 2))

	pkg, err := GenerateRegulatoryPackage(db, "org-hipaa", 30, FrameworkHIPAA)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.Framework != FrameworkHIPAA {
		t.Errorf("expected HIPAA, got %s", pkg.Framework)
	}
	if pkg.PHIAdjacentAgentCount != 3 {
		t.Errorf("expected 3 PHI agents, got %d", pkg.PHIAdjacentAgentCount)
	}
	if pkg.AccessEventsInWindow != 100 {
		t.Errorf("expected 100 access events, got %d", pkg.AccessEventsInWindow)
	}
	if pkg.BlockedAccessCount != 2 {
		t.Errorf("expected 2 blocked, got %d", pkg.BlockedAccessCount)
	}
	if pkg.HashChainRoot == "" || pkg.HashChainRoot == "no-events" {
		t.Error("expected non-empty hash chain root")
	}
}

func TestGenerateRegulatoryPackage_SOX(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT this_hash FROM enforcement_decisions").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"this_hash"}).AddRow("def456"))

	mock.ExpectQuery("SELECT").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"anomaly", "incident", "override"}).
			AddRow(10, 3, 1))

	pkg, err := GenerateRegulatoryPackage(db, "org-sox", 30, FrameworkSOX)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !pkg.CC72MonitoringActive {
		t.Error("expected CC72 monitoring active")
	}
	if !pkg.HashVerified {
		t.Error("expected hash verified")
	}
	if pkg.OverrideEvents != 1 {
		t.Errorf("expected 1 override, got %d", pkg.OverrideEvents)
	}
}

func TestGenerateRegulatoryPackage_FFIEC(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// hash chain — empty
	mock.ExpectQuery("SELECT this_hash FROM enforcement_decisions").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"this_hash"}))

	// FFIEC enforcement_decisions query
	mock.ExpectQuery("SELECT").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"total_agents", "blocked_agents"}).
			AddRow(10, 3))

	// agent_baselines count
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(8))

	// agent_identities count
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	pkg, err := GenerateRegulatoryPackage(db, "org-ffiec", 30, FrameworkFFIEC)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.RiskFlaggedCount != 3 {
		t.Errorf("expected 3 risk-flagged agents, got %d", pkg.RiskFlaggedCount)
	}
	if pkg.BehavioralBaselineCoveragePct != 80.0 {
		t.Errorf("expected 80.0%% coverage, got %f", pkg.BehavioralBaselineCoveragePct)
	}
}

func TestGenerateRegulatoryPackage_DORA(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Now().UTC()
	first := now.AddDate(0, 0, -20)
	last := now.AddDate(0, 0, -1)

	mock.ExpectQuery("SELECT this_hash FROM enforcement_decisions").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"this_hash"}).AddRow("ghi789"))

	mock.ExpectQuery("SELECT").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"incident", "blocked", "first", "last"}).
			AddRow(7, 4, sql.NullTime{Time: first, Valid: true}, sql.NullTime{Time: last, Valid: true}))

	mock.ExpectQuery("SELECT COALESCE").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"fp_rate"}).AddRow(0.0))

	pkg, err := GenerateRegulatoryPackage(db, "org-dora", 30, FrameworkDORA)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.IncidentCount != 7 {
		t.Errorf("expected 7 incidents, got %d", pkg.IncidentCount)
	}
	if pkg.BlockedCount != 4 {
		t.Errorf("expected 4 blocked, got %d", pkg.BlockedCount)
	}
	if pkg.FPRate != 0.0 {
		t.Errorf("expected 0.0 FP rate, got %f", pkg.FPRate)
	}
	if pkg.FirstEventTime == nil {
		t.Error("expected first event time")
	}
}
