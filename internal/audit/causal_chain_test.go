package audit

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestBuildCausalChain_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT agent_id, action, anomaly_score, created_at").
		WithArgs("missing-event").
		WillReturnRows(sqlmock.NewRows([]string{}))

	_, err = BuildCausalChain(db, "missing-event")
	if err == nil {
		t.Error("expected error for missing event")
	}
}

func TestBuildCausalChain_SingleAgent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Now().UTC()

	mock.ExpectQuery("SELECT agent_id, action, anomaly_score, created_at").
		WithArgs("event-001").
		WillReturnRows(sqlmock.NewRows([]string{"agent_id", "action", "anomaly_score", "created_at"}).
			AddRow("agent-root", "block", 0.92, now))

	mock.ExpectQuery("WITH RECURSIVE chain").
		WithArgs("agent-root").
		WillReturnRows(sqlmock.NewRows([]string{"callee_agent_id", "event_id", "action", "anomaly_score", "created_at", "caller_agent_id"}))

	chain, err := BuildCausalChain(db, "event-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.RootAgentID != "agent-root" {
		t.Errorf("expected agent-root, got %s", chain.RootAgentID)
	}
	if chain.ChainLength != 1 {
		t.Errorf("expected chain length 1, got %d", chain.ChainLength)
	}
	if chain.EarliestEnforceableAgentID != "agent-root" {
		t.Errorf("expected agent-root as earliest enforceable, got %s", chain.EarliestEnforceableAgentID)
	}
}

func TestBuildCausalChain_MultiAgent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Now().UTC()

	mock.ExpectQuery("SELECT agent_id, action, anomaly_score, created_at").
		WithArgs("event-002").
		WillReturnRows(sqlmock.NewRows([]string{"agent_id", "action", "anomaly_score", "created_at"}).
			AddRow("agent-A", "block", 0.95, now))

	mock.ExpectQuery("WITH RECURSIVE chain").
		WithArgs("agent-A").
		WillReturnRows(sqlmock.NewRows([]string{"callee_agent_id", "event_id", "action", "anomaly_score", "created_at", "caller_agent_id"}).
			AddRow("agent-B", "event-003", "allow", 0.75, now, "agent-A").
			AddRow("agent-C", "event-004", "allow", 0.30, now, "agent-B"))

	chain, err := BuildCausalChain(db, "event-002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.ChainLength != 3 {
		t.Errorf("expected chain length 3, got %d", chain.ChainLength)
	}
	if chain.RootAgentID != "agent-A" {
		t.Errorf("expected agent-A as root, got %s", chain.RootAgentID)
	}
	// agent-B has score 0.75 > 0.5, so it becomes earliest enforceable last
	if chain.EarliestEnforceableAgentID != "agent-B" {
		t.Errorf("expected agent-B as earliest enforceable, got %s", chain.EarliestEnforceableAgentID)
	}
}

func TestBuildCausalChain_EnforcementEventID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Now().UTC()

	mock.ExpectQuery("SELECT agent_id, action, anomaly_score, created_at").
		WithArgs("event-005").
		WillReturnRows(sqlmock.NewRows([]string{"agent_id", "action", "anomaly_score", "created_at"}).
			AddRow("agent-X", "block", 0.88, now))

	mock.ExpectQuery("WITH RECURSIVE chain").
		WithArgs("agent-X").
		WillReturnRows(sqlmock.NewRows([]string{"callee_agent_id", "event_id", "action", "anomaly_score", "created_at", "caller_agent_id"}))

	chain, err := BuildCausalChain(db, "event-005")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.EnforcementEventID != "event-005" {
		t.Errorf("expected event-005, got %s", chain.EnforcementEventID)
	}
	if chain.Chain[0].Action != "block" {
		t.Errorf("expected block action, got %s", chain.Chain[0].Action)
	}
}
