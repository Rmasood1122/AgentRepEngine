package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agentrepengine/are/internal/audit"
	"github.com/agentrepengine/are/internal/enforcement"
	"github.com/agentrepengine/are/internal/identity"
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

	// V6 FIX: ModeController — runtime enforcement mode + auto-rollback
	// Stores mode in Redis — changes take effect without restart
	// Auto-rolls back to observe if FP rate exceeds 2%
	// Never auto-escalates to enforce — human decision only
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

	// Background metrics collector
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
	// Health endpoint returns live enforcement mode from Redis
	mux.HandleFunc("/health", healthHandler(db, scoreStore, modeCtrl))
	mux.HandleFunc("/jwks", jwksHandler())
	_ = metrics.ActiveAgentCount
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/score/", requireAPIKey(scoreHandler(scoreStore)))
	mux.HandleFunc("/event", requireAPIKey(eventHandler(db, scoreStore)))

	auditHandler := audit.NewHandler(db)
	mux.HandleFunc("/audit/replay", requireAPIKey(auditHandler.ReplayHandler))
	mux.HandleFunc("/audit/export", requireAPIKey(auditHandler.ExportHandler))
	mux.HandleFunc("/enforcement/override", requireAPIKey(auditHandler.OverrideHandler))

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

// healthHandler returns live enforcement mode from Redis via ModeController.
// Mode reflects auto-rollback state instantly — no restart needed.
func healthHandler(db *sql.DB, s *store.ScoreStore, mc *enforcement.ModeController) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "ok"
		if err := db.Ping(); err != nil {
			dbStatus = "error"
		}
		redisStatus := "ok"
		if err := s.Ping(); err != nil {
			redisStatus = "error"
		}
		// Live mode from Redis — reflects auto-rollback instantly
		currentMode := mc.GetMode()
		// H10 FIX: verify hash chain on every health check
		hashChainValid := "true"
		var chainResult bool
		var chainErr error
		chainErr = db.QueryRowContext(r.Context(),
			`SELECT verify_hash_chain('enforcement_decisions') AS valid`,
		).Scan(&chainResult)
		if chainErr != nil || !chainResult {
			hashChainValid = "false"
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
  "status": "ok",
  "redis": "%s",
  "postgres": "%s",
  "enforcement_mode": "%s",
  "hash_chain_valid": %s,
  "log_format": "json"
}`, redisStatus, dbStatus, currentMode, hashChainValid)
	}
}

// jwksHandler returns the RS256 public key in JWKS format.
// Used by Kong plugin to verify JWT signatures at gateway layer.
// Public key — no authentication required.
func jwksHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keys, err := identity.LoadOrGenerateKeys()
		if err != nil {
			slog.Error("jwks_load_failed", "error", err)
			http.Error(w, "key load failed", http.StatusInternalServerError)
			return
		}

		pubKey := keys.Public
		n := base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes())
		e := base64.RawURLEncoding.EncodeToString(
			[]byte{byte(pubKey.E >> 16), byte(pubKey.E >> 8), byte(pubKey.E)})

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "alg": "RS256",
      "kid": "are-v1",
      "n": "%s",
      "e": "%s"
    }
  ]
}`, n, e)
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

		var body struct {
			AgentDID    string `json:"agent_did"`
			OrgID       string `json:"org_id"`
			EventType   string `json:"event_type"`
			PrivacyTier int    `json:"privacy_tier"`
			Payload     struct {
				Method         string  `json:"method"`
				Path           string  `json:"path"`
				StatusCode     int     `json:"status_code"`
				ScoreAtRequest float64 `json:"score_at_request"`
				BandAtRequest  string  `json:"band_at_request"`
			} `json:"payload"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if body.AgentDID == "" || body.EventType == "" {
			http.Error(w, "agent_did and event_type required", http.StatusBadRequest)
			return
		}

		privacyTier := body.PrivacyTier
		if privacyTier == 0 {
			privacyTier = 1
		}

		vector := scoring.FeatureVector{
			ToolCallRatePerHour:       1.0,
			UniqueEndpointsPerHour:    1.0,
			BulkAccessCountPerSession: 0,
			PIIFieldAccessRate:        0,
			CrossTenantProbeCount:     0,
			PermissionEscalationCount: 0,
			SubAgentSpawnDepth:        0,
			TokenRefreshRate:          0,
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
