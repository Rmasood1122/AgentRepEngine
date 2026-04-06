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

	orgID := "org-hipaa"
	now := time.Now().UTC()
	windowStart := now.AddDate(0, 0, -30)

	// hash chain query
	mock.ExpectQuery("SELECT event_hash FROM enforcement_events").
		WithArgs(orgID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"event_hash"}).AddRow("abc123"))

	// HIPAA query
	mock.ExpectQuery("SELECT").
		WithArgs(orgID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"phi", "access", "anomalous", "blocked"}).
			AddRow(3, 100, 5, 2))

	pkg, err := GenerateRegulatoryPackage(db, orgID, 30, FrameworkHIPAA)
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
	_ = windowStart
}

func TestGenerateRegulatoryPackage_SOX(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	orgID := "org-sox"

	mock.ExpectQuery("SELECT event_hash FROM enforcement_events").
		WithArgs(orgID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"event_hash"}).AddRow("def456"))

	mock.ExpectQuery("SELECT").
		WithArgs(orgID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"anomaly", "incident", "override"}).
			AddRow(10, 3, 1))

	pkg, err := GenerateRegulatoryPackage(db, orgID, 30, FrameworkSOX)
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

	orgID := "org-ffiec"

	mock.ExpectQuery("SELECT event_hash FROM enforcement_events").
		WithArgs(orgID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"event_hash"}))

	mock.ExpectQuery("SELECT").
		WithArgs(orgID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"third_party", "baselined", "total", "risk_flagged"}).
			AddRow(5, 8, 10, 3))

	pkg, err := GenerateRegulatoryPackage(db, orgID, 30, FrameworkFFIEC)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.ThirdPartyAgentCount != 5 {
		t.Errorf("expected 5 third-party agents, got %d", pkg.ThirdPartyAgentCount)
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

	orgID := "org-dora"
	now := time.Now().UTC()
	first := now.AddDate(0, 0, -20)
	last := now.AddDate(0, 0, -1)

	mock.ExpectQuery("SELECT event_hash FROM enforcement_events").
		WithArgs(orgID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"event_hash"}).AddRow("ghi789"))

	mock.ExpectQuery("SELECT").
		WithArgs(orgID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"incident", "blocked", "first", "last"}).
			AddRow(7, 4, sql.NullTime{Time: first, Valid: true}, sql.NullTime{Time: last, Valid: true}))

	mock.ExpectQuery("SELECT COALESCE").
		WithArgs(orgID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"fp_rate"}).AddRow(0.0))

	pkg, err := GenerateRegulatoryPackage(db, orgID, 30, FrameworkDORA)
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
