package handlers

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/agentrepengine/are/internal/enforcement"
)

// OverrideHandler records a human override of an enforcement decision.
// Authenticated. Path: POST /enforcement/override
// Required fields: agent_did, reason_code, operator_id
func OverrideHandler(db *sql.DB, mc *enforcement.ModeController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			AgentDID   string `json:"agent_did"`
			ReasonCode string `json:"reason_code"`
			OperatorID string `json:"operator_id"`
		}
		if err := decodeJSON(r, &body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if body.AgentDID == "" || body.ReasonCode == "" || body.OperatorID == "" {
			http.Error(w, "agent_did, reason_code, operator_id required", http.StatusBadRequest)
			return
		}
		_, err := db.Exec(
			`INSERT INTO enforcement_decisions
				(agent_did, decision, reason_code, operator_id, override, created_at)
			 VALUES ($1, 'OVERRIDE', $2, $3, true, NOW())`,
			body.AgentDID, body.ReasonCode, body.OperatorID,
		)
		if err != nil {
			slog.Error("override_insert_failed", "agent_did", body.AgentDID, "error", err)
			http.Error(w, "override failed", http.StatusInternalServerError)
			return
		}
		slog.Info("override_recorded",
			"agent_did", body.AgentDID,
			"reason_code", body.ReasonCode,
			"operator_id", body.OperatorID,
		)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"override_recorded","agent_did":%q}`, body.AgentDID)
	}
}
