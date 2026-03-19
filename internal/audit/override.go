package audit

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// OverrideRequest is the body for POST /enforcement/override.
type OverrideRequest struct {
	DecisionID string `json:"decision_id"`
	ReviewerID string `json:"reviewer_id"`
	ReasonCode string `json:"reason_code"`
	Notes      string `json:"notes"`
}

// OverrideResponse is returned after a successful override.
type OverrideResponse struct {
	DecisionID string    `json:"decision_id"`
	ReviewerID string    `json:"reviewer_id"`
	ReasonCode string    `json:"reason_code"`
	OverrideAt time.Time `json:"override_at"`
	AgentDID   string    `json:"agent_did"`
	PrevScore  int       `json:"previous_score"`
	Message    string    `json:"message"`
}

// ValidReasonCodes — standardized override reason vocabulary.
// ARE-xxx codes from the open header spec.
var ValidReasonCodes = map[string]bool{
	"ARE-FP-001": true, // confirmed false positive — threshold too tight
	"ARE-FP-002": true, // confirmed false positive — legitimate bulk export
	"ARE-FP-003": true, // confirmed false positive — authorized off-hours job
	"ARE-FP-004": true, // confirmed false positive — new agent legitimate
	"ARE-TP-001": true, // confirmed true positive — reviewed and accepted
	"ARE-EX-001": true, // exception granted — business justification on file
	"ARE-EX-002": true, // exception granted — temporary elevated access
}

// OverrideHandler handles POST /enforcement/override.
// Requires: decision_id, reviewer_id, reason_code.
// Logs override to tamper-evident audit trail.
// Does NOT automatically restore score — manual reset only per APEX spec.
func (h *Handler) OverrideHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req OverrideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.DecisionID == "" {
		http.Error(w, "decision_id is required", http.StatusBadRequest)
		return
	}
	if req.ReviewerID == "" {
		http.Error(w, "reviewer_id is required — human reviewer must be identified",
			http.StatusBadRequest)
		return
	}
	if req.ReasonCode == "" {
		http.Error(w, "reason_code is required", http.StatusBadRequest)
		return
	}
	if !ValidReasonCodes[req.ReasonCode] {
		http.Error(w,
			fmt.Sprintf("invalid reason_code: %s. Valid codes: ARE-FP-001 through ARE-EX-002",
				req.ReasonCode),
			http.StatusBadRequest)
		return
	}

	// Fetch the original decision
	var agentDID string
	var prevScore int
	var currentOverride bool
	err := h.db.QueryRow(`
		SELECT agent_did, score, override
		FROM enforcement_decisions
		WHERE id = $1`,
		req.DecisionID).Scan(&agentDID, &prevScore, &currentOverride)

	if err == sql.ErrNoRows {
		http.Error(w, "decision not found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("override fetch failed",
			"decision_id", req.DecisionID, "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if currentOverride {
		http.Error(w, "decision already overridden", http.StatusConflict)
		return
	}

	// Write override — logged in tamper-evident audit trail
	_, err = h.db.Exec(`
		UPDATE enforcement_decisions
		SET override = true,
		    reason_code = $1,
		    reviewer_id = $2
		WHERE id = $3`,
		req.ReasonCode, req.ReviewerID, req.DecisionID)
	if err != nil {
		slog.Error("override write failed",
			"decision_id", req.DecisionID, "error", err)
		http.Error(w, "override failed", http.StatusInternalServerError)
		return
	}

	slog.Info("enforcement_override",
		"decision_id", req.DecisionID,
		"agent_did", agentDID,
		"reviewer_id", req.ReviewerID,
		"reason_code", req.ReasonCode,
		"notes", req.Notes,
	)

	resp := OverrideResponse{
		DecisionID: req.DecisionID,
		ReviewerID: req.ReviewerID,
		ReasonCode: req.ReasonCode,
		OverrideAt: time.Now(),
		AgentDID:   agentDID,
		PrevScore:  prevScore,
		Message: fmt.Sprintf(
			"Override recorded. Score NOT automatically restored — "+
				"use POST /score/reset to manually restore agent %s",
			agentDID),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
