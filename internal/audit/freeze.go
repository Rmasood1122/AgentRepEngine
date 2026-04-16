package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type FreezeRequest struct {
	AgentDID    string `json:"agent_did"`
	OrgID       string `json:"org_id"`
	ReviewerID  string `json:"reviewer_id"`
	WindowStart string `json:"window_start"`
	WindowEnd   string `json:"window_end"`
	Reason      string `json:"reason"`
}

type FreezeResponse struct {
	AgentDID    string    `json:"agent_did"`
	OrgID       string    `json:"org_id"`
	ReviewerID  string    `json:"reviewer_id"`
	WindowStart time.Time `json:"window_start"`
	WindowEnd   time.Time `json:"window_end"`
	FreezeKey   string    `json:"freeze_key"`
	Message     string    `json:"message"`
}

type FreezeRecord struct {
	AgentDID    string    `json:"agent_did"`
	OrgID       string    `json:"org_id"`
	ReviewerID  string    `json:"reviewer_id"`
	WindowStart time.Time `json:"window_start"`
	WindowEnd   time.Time `json:"window_end"`
	Reason      string    `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
}

func FreezeKeyFor(orgID, agentDID string) string {
	return fmt.Sprintf("freeze:%s:%s", orgID, agentDID)
}

func NewFreezeHandler(db *sql.DB, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req FreezeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.AgentDID == "" {
			http.Error(w, "agent_did is required", http.StatusBadRequest)
			return
		}
		if req.OrgID == "" {
			http.Error(w, "org_id is required", http.StatusBadRequest)
			return
		}
		if req.ReviewerID == "" {
			http.Error(w, "reviewer_id is required", http.StatusBadRequest)
			return
		}
		if req.WindowStart == "" || req.WindowEnd == "" {
			http.Error(w, "window_start and window_end required (RFC3339)", http.StatusBadRequest)
			return
		}
		windowStart, err := time.Parse(time.RFC3339, req.WindowStart)
		if err != nil {
			http.Error(w, "window_start must be RFC3339", http.StatusBadRequest)
			return
		}
		windowEnd, err := time.Parse(time.RFC3339, req.WindowEnd)
		if err != nil {
			http.Error(w, "window_end must be RFC3339", http.StatusBadRequest)
			return
		}
		if windowEnd.Before(time.Now()) {
			http.Error(w, "window_end must be in the future", http.StatusBadRequest)
			return
		}
		if windowEnd.Before(windowStart) {
			http.Error(w, "window_end must be after window_start", http.StatusBadRequest)
			return
		}
		if windowEnd.Sub(windowStart) > 30*24*time.Hour {
			http.Error(w, "freeze window cannot exceed 30 days", http.StatusBadRequest)
			return
		}
		record := FreezeRecord{
			AgentDID:    req.AgentDID,
			OrgID:       req.OrgID,
			ReviewerID:  req.ReviewerID,
			WindowStart: windowStart,
			WindowEnd:   windowEnd,
			Reason:      req.Reason,
			CreatedAt:   time.Now(),
		}
		recordJSON, err := json.Marshal(record)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		ttl := time.Until(windowEnd)
		freezeKey := FreezeKeyFor(req.OrgID, req.AgentDID)
		ctx := context.Background()
		if err := rdb.Set(ctx, freezeKey, recordJSON, ttl).Err(); err != nil {
			slog.Error("freeze_redis_write_failed", "agent_did", req.AgentDID, "error", err)
			http.Error(w, "freeze storage failed", http.StatusInternalServerError)
			return
		}
		_, dbErr := WriteEnforcementDecision(db, EnforcementDecisionInput{
			AgentDID:    req.AgentDID,
			Decision:    "FREEZE",
			Score:       0,
			ScoreDelta:  0,
			PolicyFired: "maintenance_window",
			ReasonObj: map[string]string{
				"window_start": windowStart.Format(time.RFC3339),
				"window_end":   windowEnd.Format(time.RFC3339),
				"reason":       req.Reason,
			},
			Override:   false,
			ReviewerID: req.ReviewerID,
		})
		if dbErr != nil {
			slog.Warn("freeze_audit_log_failed", "agent_did", req.AgentDID, "error", dbErr)
		}
		slog.Info("agent_score_frozen",
			"agent_did", req.AgentDID,
			"org_id", req.OrgID,
			"reviewer_id", req.ReviewerID,
			"ttl_hours", ttl.Hours(),
		)
		resp := FreezeResponse{
			AgentDID:    req.AgentDID,
			OrgID:       req.OrgID,
			ReviewerID:  req.ReviewerID,
			WindowStart: windowStart,
			WindowEnd:   windowEnd,
			FreezeKey:   freezeKey,
			Message: fmt.Sprintf(
				"Score freeze active until %s. H-decay suspended. Freeze key: %s",
				windowEnd.Format(time.RFC3339), freezeKey),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// IsAgentFrozen checks Redis for an active freeze record.
// Fail-open: returns nil if Redis unavailable.
func IsAgentFrozen(ctx context.Context, rdb *redis.Client, orgID, agentDID string) (*FreezeRecord, error) {
	val, err := rdb.Get(ctx, FreezeKeyFor(orgID, agentDID)).Result()
	if err != nil {
		return nil, nil
	}
	var record FreezeRecord
	if err := json.Unmarshal([]byte(val), &record); err != nil {
		return nil, nil
	}
	if time.Now().After(record.WindowEnd) {
		return nil, nil
	}
	return &record, nil
}
