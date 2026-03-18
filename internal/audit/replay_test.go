package audit

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/lib/pq"
)

func getTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres",
		"postgres://are:are_dev@localhost:5432/agentrepengine?sslmode=disable")
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("postgres not reachable: %v", err)
	}
	return db
}

// TestHashChainValid verifies G-TAMPER gate.
func TestHashChainValid(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	h := NewHandler(db)
	valid := h.verifyHashChain()
	if !valid {
		t.Fatal("★ G-TAMPER HARD STOP: hash chain invalid")
	}
	t.Log("✅ G-TAMPER: hash chain valid")
}

// TestReplayEndpointResponds verifies replay returns valid JSON.
func TestReplayEndpointResponds(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	h := NewHandler(db)
	req := httptest.NewRequest("GET",
		"/audit/replay?agent_did=did:jwt:test-org:finance-agent:001", nil)
	w := httptest.NewRecorder()

	h.ReplayHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var response ReplayResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	if response.AgentDID == "" {
		t.Error("agent_did missing from response")
	}
	t.Logf("✅ Replay: agent=%s events=%d",
		response.AgentDID, response.TotalEvents)
}

// TestSOC2ExportChainVerified verifies export always includes chain status.
func TestSOC2ExportChainVerified(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	h := NewHandler(db)
	req := httptest.NewRequest("GET",
		"/audit/export?from=2026-01-01&to=2026-12-31&format=soc2", nil)
	w := httptest.NewRecorder()

	h.ExportHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var export SOC2Export
	if err := json.NewDecoder(w.Body).Decode(&export); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if !export.ChainVerified {
		t.Error("★ G-TAMPER: chain_verified is false in SOC2 export")
	}

	t.Logf("✅ SOC2 export: rows=%d chain_verified=%v",
		export.TotalRows, export.ChainVerified)
}

// TestReplayRequiresAgentDID verifies missing param returns 400.
func TestReplayRequiresAgentDID(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	h := NewHandler(db)
	req := httptest.NewRequest("GET", "/audit/replay", nil)
	w := httptest.NewRecorder()

	h.ReplayHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing agent_did, got %d", w.Code)
	}
	t.Log("✅ Missing agent_did correctly returns 400")
}
