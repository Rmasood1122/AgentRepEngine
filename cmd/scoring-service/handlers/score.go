package handlers

import (
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/agentrepengine/are/internal/store"
)

// ScoreHandler returns the current behavioral score for an agent DID.
// Authenticated. Path: /score/{agent_did}
func ScoreHandler(s *store.ScoreStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		did := r.URL.Path[len("/score/"):]
		if did == "" {
			http.Error(w, "missing agent DID", http.StatusBadRequest)
			return
		}
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req-%d", time.Now().UnixNano())
		}
		start := time.Now()
		slog.Info("score_request", "request_id", requestID, "agent_did", did)
		result, err := s.GetScore(did)
		if err != nil {
			slog.Error("get score failed",
				"request_id", requestID, "agent_did", did, "error", err)
			http.Error(w, "score lookup failed", http.StatusInternalServerError)
			return
		}
		slog.Info("score_response",
			"request_id", requestID,
			"agent_did", did,
			"score", result.Score,
			"band", result.Band,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		w.Header().Set("Content-Type", "application/json")
		confidence := math.Min(0.99, 0.50+float64(1000-result.Score)/2000.0)
		fmt.Fprintf(w, `{
  "agent_did": "%s",
  "score": %d,
  "band": "%s",
  "confidence": %.2f,
  "source": "%s",
  "reason": %s
}`, did, result.Score, result.Band, confidence, result.Source, result.ReasonJSON)
	}
}

// CoalesceF returns val if nonzero, otherwise fallback.
// Used to substitute sensible defaults for missing feature vector fields.
func CoalesceF(val, fallback float64) float64 {
	if val != 0 {
		return val
	}
	return fallback
}
