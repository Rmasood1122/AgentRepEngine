package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agentrepengine/are/internal/audit"
	"github.com/agentrepengine/are/internal/metrics"
	"github.com/agentrepengine/are/internal/scoring"
	"github.com/agentrepengine/are/internal/store"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	// FAANG standard: explicit connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

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
	// Background metrics collector — updates gauges every 60 seconds
	go func() {
		for {
			var agentCount float64
			db.QueryRow(`SELECT COUNT(*) FROM agent_identities
				WHERE last_seen > NOW() - INTERVAL '24 hours'`).Scan(&agentCount)
			metrics.ActiveAgentCount.Set(agentCount)

			var queueDepth float64
			db.QueryRow(`SELECT COUNT(*) FROM agent_event_queue
				WHERE processed = false`).Scan(&queueDepth)
			metrics.EventQueueDepth.Set(queueDepth)

			var fpRate float64
			db.QueryRow(`SELECT COALESCE(fp_rate, 0) FROM daily_fp_metrics
				ORDER BY date DESC LIMIT 1`).Scan(&fpRate)
			metrics.FPRateGauge.Set(fpRate)

			time.Sleep(60 * time.Second)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler(db, scoreStore, enforcementMode))
	// Prometheus metrics — G-OBSERVE requirement
	_ = metrics.ActiveAgentCount // initialize metrics package
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/score/", requireAPIKey(scoreHandler(scoreStore)))
	mux.HandleFunc("/event", requireAPIKey(eventHandler(db, scoreStore)))

	auditHandler := audit.NewHandler(db)
	mux.HandleFunc("/audit/replay", requireAPIKey(auditHandler.ReplayHandler))
	mux.HandleFunc("/audit/export", requireAPIKey(auditHandler.ExportHandler))
	mux.HandleFunc("/enforcement/override", requireAPIKey(auditHandler.OverrideHandler))
	// FAANG standard: graceful shutdown
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("scoring service ready", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("shutdown signal received", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
	}
	slog.Info("scoring service stopped cleanly")
}

func requireAPIKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Gateway-Verified") == "true" {
			next(w, r)
			return
		}
		apiKey := os.Getenv("SCORING_API_KEY")
		if apiKey == "" {
			slog.Warn("SCORING_API_KEY not set — unprotected in dev mode")
			next(w, r)
			return
		}
		if r.Header.Get("X-API-Key") != apiKey {
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
