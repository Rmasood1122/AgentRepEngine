package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/agentrepengine/are/internal/enforcement"
	"github.com/agentrepengine/are/internal/store"
)

// HealthHandler returns system status for Docker health checks and Kong probes.
// Unauthenticated — contains no behavioral agent data.
// Returns: redis, postgres, enforcement_mode, hash_chain_valid, queue_depth.
func HealthHandler(db *sql.DB, s *store.ScoreStore, mc *enforcement.ModeController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "ok"
		if err := db.Ping(); err != nil {
			dbStatus = "error"
		}
		redisStatus := "ok"
		if err := s.Ping(); err != nil {
			redisStatus = "error"
		}
		currentMode := mc.GetMode()
		hashChainValid := "true"
		var chainResult bool
		chainErr := db.QueryRowContext(r.Context(),
			`SELECT verify_hash_chain('enforcement_decisions') AS valid`,
		).Scan(&chainResult)
		if chainErr != nil || !chainResult {
			hashChainValid = "false"
		}
		queueDepth, _ := store.QueueDepth(r.Context(), db)
		queueStatus := "ok"
		if queueDepth > store.QueueDepthWarning {
			queueStatus = "warning"
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
  "status": "ok",
  "redis": "%s",
  "postgres": "%s",
  "enforcement_mode": "%s",
  "hash_chain_valid": %s,
  "queue_depth": %d,
  "queue_status": "%s",
  "log_format": "json"
}`, redisStatus, dbStatus, currentMode, hashChainValid, queueDepth, queueStatus)
	}
}
