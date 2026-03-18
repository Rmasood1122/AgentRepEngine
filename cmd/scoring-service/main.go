package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/agentrepengine/are/internal/audit"
	"github.com/agentrepengine/are/internal/scoring"
	"github.com/agentrepengine/are/internal/store"
	_ "github.com/lib/pq"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	dbURL := getEnv("DATABASE_URL",
		"postgres://are:are_dev@localhost:5432/agentrepengine?sslmode=disable")
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	port := getEnv("PORT", "8080")
	enforcementMode := getEnv("ENFORCEMENT_MODE", "observe")

	slog.Info("starting scoring service",
		"port", port,
		"enforcement_mode", enforcementMode,
	)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Error("postgres connect failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	for i := 0; i < 10; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		slog.Info("waiting for postgres...", "attempt", i+1)
		time.Sleep(2 * time.Second)
	}
	if err := db.Ping(); err != nil {
		slog.Error("postgres not ready", "error", err)
		os.Exit(1)
	}
	slog.Info("postgres connected")

	scoreStore := store.NewScoreStore(db, redisURL)
	if err := scoreStore.ConnectRedis(); err != nil {
		slog.Error("redis connect failed", "error", err)
		os.Exit(1)
	}
	slog.Info("redis connected")

	consumer := scoring.NewEventConsumer(db, scoreStore)
	go consumer.Start()
	slog.Info("event consumer started")

	mux := http.NewServeMux()

	// Health endpoint — no auth required
	mux.HandleFunc("/health", healthHandler(db, scoreStore, enforcementMode))

	// Protected endpoints — GAP 1 + GAP 6 fix
	mux.HandleFunc("/score/", requireAPIKey(scoreHandler(scoreStore)))
	mux.HandleFunc("/event", requireAPIKey(eventHandler(db, scoreStore)))

	auditHandler := audit.NewHandler(db)
	mux.HandleFunc("/audit/replay", requireAPIKey(auditHandler.ReplayHandler))
	mux.HandleFunc("/audit/export", requireAPIKey(auditHandler.ExportHandler))

	slog.Info("scoring service ready", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

// requireAPIKey middleware — GAP 1 + GAP 6 fix.
// Blocks direct agent access to scoring endpoints.
// Kong gateway sets X-Gateway-Verified after JWT validation.
// External tools use X-API-Key header.
func requireAPIKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Kong gateway — already verified JWT, trust it
		if r.Header.Get("X-Gateway-Verified") == "true" {
			next(w, r)
			return
		}

		// API key auth for direct access (ops, CI, dashboards)
		apiKey := os.Getenv("SCORING_API_KEY")
		if apiKey == "" {
			slog.Warn("SCORING_API_KEY not set — endpoints unprotected in dev mode")
			next(w, r)
			return
		}

		provided := r.Header.Get("X-API-Key")
		if provided != apiKey {
			slog.Warn("unauthorized access attempt",
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
			)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func healthHandler(db *sql.DB, s *store.ScoreStore, mode string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "ok"
		if err := db.Ping(); err != nil {
			dbStatus = "error"
		}
		redisStatus := "ok"
		if err := s.Ping(); err != nil {
			redisStatus = "error"
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
  "status": "ok",
  "redis": "%s",
  "postgres": "%s",
  "enforcement_mode": "%s",
  "log_format": "json"
}`, redisStatus, dbStatus, mode)
	}
}

func scoreHandler(s *store.ScoreStore) http.HandlerFunc {
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
		slog.Info("score_request",
			"request_id", requestID,
			"agent_did", did,
		)

		result, err := s.GetScore(did)
		if err != nil {
			slog.Error("get score failed",
				"request_id", requestID,
				"agent_did", did,
				"error", err,
			)
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
		fmt.Fprintf(w, `{
  "agent_did": "%s",
  "score": %d,
  "band": "%s",
  "source": "%s",
  "reason": %s
}`, did, result.Score, result.Band, result.Source, result.ReasonJSON)
	}
}

func eventHandler(db *sql.DB, s *store.ScoreStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"queued"}`)
	}
}
