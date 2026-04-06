package scoring

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestClassifyChange_Attack(t *testing.T) {
	// High new score, well above slow mean — should be ATTACK
	result := ClassifyChange(0.5, 0.3, 0.08, 0.95)
	if result != ClassAttack {
		t.Errorf("expected ATTACK, got %s", result)
	}
}

func TestClassifyChange_LegitimateChange(t *testing.T) {
	// Moderate divergence — should be LEGITIMATE_CHANGE
	result := ClassifyChange(0.4, 0.3, 0.08, 0.55)
	if result != ClassLegitimateChange {
		t.Errorf("expected LEGITIMATE_CHANGE, got %s", result)
	}
}

func TestClassifyChange_Normal(t *testing.T) {
	// Small deviation — should be NORMAL
	result := ClassifyChange(0.3, 0.3, 0.08, 0.32)
	if result != ClassNormal {
		t.Errorf("expected NORMAL, got %s", result)
	}
}

func TestClassifyChange_ZeroStdDev(t *testing.T) {
	// StdDev=0 should not panic — uses floor of 0.01
	result := ClassifyChange(0.3, 0.3, 0.0, 0.95)
	if result == "" {
		t.Error("expected valid classification, got empty")
	}
}

func TestClassifyChange_BelowMean(t *testing.T) {
	// High divergence but score below mean — LEGITIMATE_CHANGE not ATTACK
	result := ClassifyChange(0.5, 0.8, 0.08, 0.2)
	if result == ClassAttack {
		t.Errorf("expected LEGITIMATE_CHANGE or NORMAL for below-mean score, got ATTACK")
	}
}

func TestLoadTwoSpeedBaseline_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Now().UTC()

	mock.ExpectQuery("SELECT agent_id, fast_ewma, slow_mean, slow_stddev, updated_at").
		WithArgs("agent-001").
		WillReturnRows(sqlmock.NewRows([]string{"agent_id", "fast_ewma", "slow_mean", "slow_stddev", "updated_at"}).
			AddRow("agent-001", 0.42, 0.35, 0.08, now))

	b, err := LoadTwoSpeedBaseline(db, "agent-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.AgentID != "agent-001" {
		t.Errorf("expected agent-001, got %s", b.AgentID)
	}
	if b.FastEWMA != 0.42 {
		t.Errorf("expected 0.42, got %f", b.FastEWMA)
	}
	if b.SlowMean != 0.35 {
		t.Errorf("expected 0.35, got %f", b.SlowMean)
	}
}

func TestLoadTwoSpeedBaseline_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT agent_id, fast_ewma, slow_mean, slow_stddev, updated_at").
		WithArgs("ghost").
		WillReturnRows(sqlmock.NewRows([]string{}))

	_, err = LoadTwoSpeedBaseline(db, "ghost")
	if err == nil {
		t.Error("expected error for missing agent")
	}
}

func TestUpdateFastEWMA(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec("UPDATE agent_baselines").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "agent-001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = UpdateFastEWMA(db, "agent-001", 0.9, 0.4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
