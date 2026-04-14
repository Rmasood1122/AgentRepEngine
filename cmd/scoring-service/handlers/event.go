package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/agentrepengine/are/internal/metrics"
	"github.com/agentrepengine/are/internal/scoring"
	"github.com/agentrepengine/are/internal/store"
)

// EventHandler receives behavioral events from Kong plugin and enqueues
// them for async processing by the event consumer.
// Authenticated. Path: POST /event
func EventHandler(db *sql.DB, s *store.ScoreStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			AgentDID    string `json:"agent_did"`
			OrgID       string `json:"org_id"`
			EventType   string `json:"event_type"`
			PrivacyTier int    `json:"privacy_tier"`
			FeatureVector struct {
				ToolCallRatePerHour       float64 `json:"tool_call_rate_per_hour"`
				UniqueEndpointsPerHour    float64 `json:"unique_endpoints_per_hour"`
				BulkAccessCountPerSession float64 `json:"bulk_access_count_per_session"`
				PIIFieldAccessRate        float64 `json:"pii_field_access_rate"`
				CrossTenantProbeCount     float64 `json:"cross_tenant_probe_count"`
				PermissionEscalationCount float64 `json:"permission_escalation_count"`
				SubAgentSpawnDepth        float64 `json:"sub_agent_spawn_depth"`
				TokenRefreshRate          float64 `json:"token_refresh_rate"`
			} `json:"feature_vector"`
			Payload struct {
				Method         string  `json:"method"`
				Path           string  `json:"path"`
				StatusCode     int     `json:"status_code"`
				ScoreAtRequest float64 `json:"score_at_request"`
				BandAtRequest  string  `json:"band_at_request"`
			} `json:"payload"`
		}

		if err := decodeJSON(r, &body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if body.AgentDID == "" || body.EventType == "" {
			http.Error(w, "agent_did and event_type required", http.StatusBadRequest)
			return
		}

		// Track JWT verify fallback — Kong sets this header when /verify
		// service was unavailable and unverified extraction was used.
		// Nonzero in production = RS256 verification not running.
		if r.Header.Get("X-Verify-Fallback") == "true" {
			metrics.JWTVerifyFallbackTotal.Inc()
		}

		privacyTier := body.PrivacyTier
		if privacyTier == 0 {
			privacyTier = 1
		}

		fv := body.FeatureVector
		vector := scoring.FeatureVector{
			ToolCallRatePerHour:       CoalesceF(fv.ToolCallRatePerHour, 1.0),
			UniqueEndpointsPerHour:    CoalesceF(fv.UniqueEndpointsPerHour, 1.0),
			BulkAccessCountPerSession: fv.BulkAccessCountPerSession,
			PIIFieldAccessRate:        fv.PIIFieldAccessRate,
			CrossTenantProbeCount:     fv.CrossTenantProbeCount,
			PermissionEscalationCount: fv.PermissionEscalationCount,
			SubAgentSpawnDepth:        fv.SubAgentSpawnDepth,
			TokenRefreshRate:          fv.TokenRefreshRate,
		}

		if err := s.EnqueueEvent(body.AgentDID, body.EventType,
			vector, privacyTier); err != nil {
			log.Printf("ERROR enqueue event agent=%s: %v", body.AgentDID, err)
			http.Error(w, "enqueue failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, `{"status":"queued"}`)
	}
}
