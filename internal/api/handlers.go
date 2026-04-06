package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/agentrepengine/are/internal/scoring"
)

// Handler holds dependencies for API endpoints.
type Handler struct {
	sirMachine *scoring.SIRMachine
}

// NewHandler constructs a Handler with required dependencies.
func NewHandler(sirMachine *scoring.SIRMachine) *Handler {
	return &Handler{sirMachine: sirMachine}
}

// HandleAgentClear processes a human operator clearing an INCIDENT agent.
// POST /agent/{id}/clear
// Body: {"cleared_by": "soc_analyst:jsmith"}
// 200: agent moved to REMEDIATED, probation window started
// 409: agent is not in INCIDENT state
func (h *Handler) HandleAgentClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract agent ID from path: /agent/{id}/clear
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	agentID := parts[1]

	var req struct {
		ClearedBy string `json:"cleared_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ClearedBy == "" {
		http.Error(w, `{"error":"cleared_by required"}`, http.StatusBadRequest)
		return
	}

	record, err := h.sirMachine.HumanClear(r.Context(), agentID, req.ClearedBy)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"error":         err.Error(),
			"current_state": "not_incident",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agent_did":      record.AgentDID,
		"state":          record.CurrentState,
		"probation_ends": record.ProbationEndsAt,
		"cleared_by":     record.ClearRequestedBy,
		"incident_count": record.IncidentCount,
	})
}
