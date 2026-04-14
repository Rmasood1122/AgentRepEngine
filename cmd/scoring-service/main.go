package main

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/agentrepengine/are/internal/api"
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

	// C-5: Fail-closed on missing API key.
	// The original code warned and passed through — that is the vulnerability.
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

	// Start retention job — purges processed events older than 90 days every 24h
	store.StartRetentionJob(db)
	slog.Info("retention job started", "retention_days", store.RetentionDays,
		"queue_depth_warning_threshold", store.QueueDepthWarning)

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

	// auth is the closure-bound authentication middleware.
	// Inlined here (not a separate package) to avoid import cycle with internal packages.
	// Uses the apiKey captured from os.Getenv above — validated non-empty at startup.
	auth := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			provided := extractBearer(r.Header.Get("Authorization"))

			// C-5 FIX 1: Removed X-Gateway-Verified bypass.
			// Original code allowed any caller to set X-Gateway-Verified: true
			// and skip authentication entirely. This is an unauthenticated bypass
			// that any internal attacker could exploit.
			// Kong sends the real SCORING_API_KEY — it does not need a bypass header.

			// C-5 FIX 2: Also accept X-API-Key for Kong plugin backward compatibility.
			// Kong plugin currently sends X-API-Key. Both headers are accepted during
			// the transition period. Remove X-API-Key support after Kong plugin is
			// updated to send Bearer token (see kong/plugins/agent-reputation/handler.lua).
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
	// /health — Docker Compose health checks + Kong upstream probes must reach this
	//           without credentials. Contains no behavioral data.
	// /metrics — Prometheus scraper; restrict at network level (firewall/Kong route),
	//            not at application level. No behavioral agent data exposed.
	// /jwks    — Public key endpoint. RS256 public keys are not secret by definition.
	//            Must be reachable by Kong for JWT verification.
	mux.HandleFunc("/health", healthHandler(db, scoreStore, modeCtrl))
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/jwks", jwksHandler())

	// AUTHENTICATED — all endpoints that expose behavioral data or accept agent events:
	mux.HandleFunc("/verify", auth(verifyHandler(scoreStore)))
	mux.HandleFunc("/score/", auth(scoreHandler(scoreStore)))
	mux.HandleFunc("/event", auth(eventHandler(db, scoreStore)))
	mux.HandleFunc("/audit/replay", auth(audit.NewHandler(db).ReplayHandler))
	mux.HandleFunc("/audit/export", auth(audit.NewHandler(db).ExportHandler))
	mux.HandleFunc("/dashboard", auth(dashboardHandler(db)))
	mux.HandleFunc("/api/regulatory-package", auth(regulatoryPackageHandler(db)))
	mux.HandleFunc("/agent/", auth(apiHandler.HandleAgentClear))
	mux.HandleFunc("/enforcement/override", auth(overrideHandler(db, modeCtrl)))

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

func coalesceF(val, fallback float64) float64 {
	if val != 0 {
		return val
	}
	return fallback
}

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
		currentMode := mc.GetMode()
		hashChainValid := "true"
		var chainResult bool
		var chainErr error
		chainErr = db.QueryRowContext(r.Context(),
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

func verifyHandler(s *store.ScoreStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"valid":false,"error":"invalid request body"}`)
			return
		}
		if body.Token == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"valid":false,"error":"token required"}`)
			return
		}
		keys, err := identity.LoadOrGenerateKeys()
		if err != nil {
			slog.Error("verify_keys_load_failed", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"valid":false,"error":"key load failed"}`)
			return
		}
		claims, err := identity.VerifyToken(body.Token, keys, s.GetRedisClient())
		if err != nil {
			slog.Warn("verify_token_rejected", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, `{"valid":false,"error":%q}`, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"valid":true,"agent_did":%q,"org_id":%q,"instance_id":%q,"lineage_hash":%q}`,
			claims.AgentDID, claims.OrgID, claims.InstanceID, claims.LineageHash)
	}
}

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

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
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
			ToolCallRatePerHour:       coalesceF(fv.ToolCallRatePerHour, 1.0),
			UniqueEndpointsPerHour:    coalesceF(fv.UniqueEndpointsPerHour, 1.0),
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

// overrideHandler handles human override of enforcement decisions.
// Authenticated — requires SCORING_API_KEY.
func overrideHandler(db *sql.DB, mc *enforcement.ModeController) http.HandlerFunc {
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
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
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

// dashboardHandler and regulatoryPackageHandler are defined in dashboard_handler.go.
// C-5 NOTE: dashboard is now wrapped with auth() in the mux registration above —
// previously it was unauthenticated. The implementation stays in dashboard_handler.go.
