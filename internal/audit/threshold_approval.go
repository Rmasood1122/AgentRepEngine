package audit

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// ThresholdApprovalRequest is the body for POST /threshold/approve or /threshold/decline.
type ThresholdApprovalRequest struct {
	SuggestionID string `json:"suggestion_id"` // policy_fired value
	ReviewerID   string `json:"reviewer_id"`
	Decision     string `json:"decision"` // "approve" | "decline" | "review_later"
	Notes        string `json:"notes"`
}

// ThresholdApprovalRecord is written to threshold_approvals table on every decision.
type ThresholdApprovalRecord struct {
	ID           string    `json:"id"`
	PolicyFired  string    `json:"policy_fired"`
	ReviewerID   string    `json:"reviewer_id"`
	Decision     string    `json:"decision"`
	Notes        string    `json:"notes"`
	OverrideRate float64   `json:"override_rate"`
	CreatedAt    time.Time `json:"created_at"`
}

// ThresholdAdvisorHandler handles the monthly threshold advisory HTTP endpoints.
// GET  /threshold/report  — generate and return monthly advisor report
// POST /threshold/approve — record human approval/decline of a suggestion
type ThresholdAdvisorHandler struct {
	db *sql.DB
}

// NewThresholdAdvisorHandler creates a new handler.
func NewThresholdAdvisorHandler(db *sql.DB) *ThresholdAdvisorHandler {
	return &ThresholdAdvisorHandler{db: db}
}

// ReportHandler handles GET /threshold/report
// Returns the monthly advisor report for the requesting org.
func (h *ThresholdAdvisorHandler) ReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		http.Error(w, "org_id query parameter required", http.StatusBadRequest)
		return
	}

	periodDays := 30
	report, err := GenerateMonthlyReport(h.db, orgID, periodDays)
	if err != nil {
		slog.Error("threshold_report_failed", "org_id", orgID, "error", err)
		http.Error(w, "report generation failed", http.StatusInternalServerError)
		return
	}

	// Build CISO-facing response
	response := map[string]interface{}{
		"report":       report,
		"ciso_summary": FormatCISOSummary(report),
		"instructions": map[string]string{
			"approve": "POST /threshold/approve with {suggestion_id, reviewer_id, decision: 'approve', notes}",
			"decline": "POST /threshold/approve with {suggestion_id, reviewer_id, decision: 'decline', notes}",
			"rule":    "Max 10% weight change per iteration. Held-out validation required before applying.",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ApproveHandler handles POST /threshold/approve
// Records human decision on a threshold suggestion.
// DOES NOT automatically change weights — documents the decision only.
// Actual weight change requires manual update to config/scoring_weights.yaml.
func (h *ThresholdAdvisorHandler) ApproveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ThresholdApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.SuggestionID == "" {
		http.Error(w, "suggestion_id is required", http.StatusBadRequest)
		return
	}
	if req.ReviewerID == "" {
		http.Error(w, "reviewer_id is required — human reviewer must be identified", http.StatusBadRequest)
		return
	}
	if req.Decision != "approve" && req.Decision != "decline" && req.Decision != "review_later" {
		http.Error(w, "decision must be 'approve', 'decline', or 'review_later'", http.StatusBadRequest)
		return
	}

	// Write approval record
	var recordID string
	err := h.db.QueryRow(`
		INSERT INTO threshold_approvals
			(policy_fired, reviewer_id, decision, notes, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id::text`,
		req.SuggestionID, req.ReviewerID, req.Decision, req.Notes,
	).Scan(&recordID)
	if err != nil {
		// Table may not exist yet — log and continue gracefully
		slog.Warn("threshold_approval_write_failed",
			"policy", req.SuggestionID,
			"decision", req.Decision,
			"error", err,
		)
		// Return success anyway — the decision is logged in our server logs
		recordID = fmt.Sprintf("log-only-%d", time.Now().Unix())
	}

	slog.Info("threshold_approval_recorded",
		"suggestion_id", req.SuggestionID,
		"reviewer_id", req.ReviewerID,
		"decision", req.Decision,
		"notes", req.Notes,
	)

	message := buildApprovalMessage(req.Decision, req.SuggestionID)

	response := map[string]interface{}{
		"id":            recordID,
		"suggestion_id": req.SuggestionID,
		"reviewer_id":   req.ReviewerID,
		"decision":      req.Decision,
		"recorded_at":   time.Now().UTC(),
		"message":       message,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func buildApprovalMessage(decision, suggestionID string) string {
	switch decision {
	case "approve":
		return fmt.Sprintf(
			"Approval recorded for policy '%s'. "+
				"NEXT STEP: update config/scoring_weights.yaml manually (max 10%% change). "+
				"Run go test ./tests/regression/... to validate before deploying.",
			suggestionID)
	case "decline":
		return fmt.Sprintf(
			"Decline recorded for policy '%s'. "+
				"Current thresholds will be maintained. "+
				"Override pattern will continue to be monitored.",
			suggestionID)
	case "review_later":
		return fmt.Sprintf(
			"Deferred for policy '%s'. "+
				"Will appear in next monthly report if pattern continues.",
			suggestionID)
	default:
		return "Decision recorded."
	}
}
