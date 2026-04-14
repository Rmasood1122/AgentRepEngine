package main

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/agentrepengine/are/cmd/scoring-service/handlers"
	"github.com/agentrepengine/are/internal/api"
	"github.com/agentrepengine/are/internal/audit"
	"github.com/agentrepengine/are/internal/enforcement"
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

	// C-5: Fail-closed on missing API key.
	// Production and pilot deployments must always have SCORING_API_KEY set.
	// Dev mode exception: set SCORING_API_KEY=dev-only-insecure to be explicit.
	apiKey := os.Getenv("SCORING_API_KEY")
	if apiKey == "" {
		slog.Error("SCORING_API_KEY environment variable is not set. " +
			"Refusing to start — all /score, /event, /audit, /verify, /dashboard endpoints " +
			"would be unauthenticated. Set SCORING_API_KEY before starting.")
		os.Exit(1)
	}

	slog.Info("starting scoring service",
		"port", port,
		"enforcement_mode", enforcementMode,
		"api_key_configured", true,
	)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Error("postgres connect failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

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

	modeCtrl := enforcement.NewModeController(
		scoreStore.GetRedisClient(), db, enforcementMode)
	modeCtrl.StartFPMonitor()
	slog.Info("mode_controller_started",
		"initial_mode", enforcementMode,
		"fp_threshold_pct", 2.0,
		"check_interval", "5m")

	consumer := scoring.NewEventConsumer(db, scoreStore)
	go consumer.Start()
	slog.Info("event consumer started")

	store.StartRetentionJob(db)
	slog.Info("retention job started",
		"retention_days", store.RetentionDays,
		"queue_depth_warning_threshold", store.QueueDepthWarning)

	// Background metrics updater — agent count, queue depth, FP rate.
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

	apiHandler := api.NewHandler(scoring.NewSIRMachine(scoring.DefaultSIRThresholds(), scoreStore))

	// auth is the API key authentication middleware.
	// Inlined in main to avoid import cycle — auth captures apiKey from os.Getenv.
	// Kong sends SCORING_API_KEY via X-API-Key header (backward compat) or Bearer token.
	auth := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			provided := extractBearer(r.Header.Get("Authorization"))
			if provided == "" {
				provided = r.Header.Get("X-API-Key")
			}
			if provided == "" {
				slog.Warn("unauthorized_request_missing_token",
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
				)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			// Constant-time comparison prevents timing-based key enumeration.
			if subtle.ConstantTimeCompare([]byte(provided), []byte(apiKey)) != 1 {
				slog.Warn("unauthorized_request_invalid_token",
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
				)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next(w, r)
		}
	}

	mux := http.NewServeMux()

	// UNAUTHENTICATED — intentional, documented:
	// /health  — Docker health checks + Kong upstream probes. No behavioral data.
	// /metrics — Prometheus scraper. Restrict at network level, not application level.
	// /jwks    — RS256 public keys. Not secret by definition.
	mux.HandleFunc("/health", handlers.HealthHandler(db, scoreStore, modeCtrl))
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/jwks", handlers.JWKSHandler())

	// AUTHENTICATED — all endpoints exposing behavioral data or accepting agent events:
	mux.HandleFunc("/verify", auth(handlers.VerifyHandler(scoreStore)))
	mux.HandleFunc("/score/", auth(handlers.ScoreHandler(scoreStore)))
	mux.HandleFunc("/event", auth(handlers.EventHandler(db, scoreStore)))
	mux.HandleFunc("/audit/replay", auth(audit.NewHandler(db).ReplayHandler))
	mux.HandleFunc("/audit/export", auth(audit.NewHandler(db).ExportHandler))
	mux.HandleFunc("/dashboard", auth(handlers.DashboardHandler(db)))
	mux.HandleFunc("/api/regulatory-package", auth(handlers.RegulatoryPackageHandler(db)))
	mux.HandleFunc("/agent/", auth(apiHandler.HandleAgentClear))
	mux.HandleFunc("/enforcement/override", auth(handlers.OverrideHandler(db, modeCtrl)))

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

// extractBearer pulls the token from "Authorization: Bearer <token>".
// Returns empty string if the header is absent or not a Bearer scheme.
func extractBearer(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
