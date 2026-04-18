package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strings"

	"github.com/agentrepengine/are/internal/certification"
)

// CertificationHandler returns the HTTP handler for POST /certificate/issue/{agent_did}
func CertificationHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract agent_did from path: /certificate/issue/{agent_did}
		path := strings.TrimPrefix(r.URL.Path, "/certificate/issue/")
		agentDID := strings.TrimSpace(path)
		if agentDID == "" || agentDID == r.URL.Path {
			http.Error(w, `{"error":"agent_did required in path"}`, http.StatusBadRequest)
			return
		}

		orgID := r.URL.Query().Get("org_id")
		if orgID == "" {
			orgID = "default"
		}

		// Generate report from enforcement_decisions + agent_baselines
		report, err := certification.Generate(db, agentDID, orgID, 30)
		if err != nil {
			slog.Error("certification_generate_failed",
				"agent_did", agentDID, "error", err)
			http.Error(w, `{"error":"report generation failed"}`, http.StatusInternalServerError)
			return
		}

		// Hash all fields
		if err := certification.HashReport(report); err != nil {
			slog.Error("certification_hash_failed",
				"agent_did", agentDID, "error", err)
			http.Error(w, `{"error":"report hash failed"}`, http.StatusInternalServerError)
			return
		}

		// Sign with RS256
		key, err := certification.LoadOperationalKey()
		if err != nil {
			slog.Error("certification_key_load_failed", "error", err)
			http.Error(w, `{"error":"signing key unavailable"}`, http.StatusInternalServerError)
			return
		}

		if err := certification.SignReport(report, key); err != nil {
			slog.Error("certification_sign_failed",
				"agent_did", agentDID, "error", err)
			http.Error(w, `{"error":"report signing failed"}`, http.StatusInternalServerError)
			return
		}

		data, err := certification.ToJSON(report)
		if err != nil {
			http.Error(w, `{"error":"json marshal failed"}`, http.StatusInternalServerError)
			return
		}

		slog.Info("certification_issued",
			"agent_did", agentDID,
			"cert_hash", report.CertHash,
			"enforcement_count", len(report.EnforcementLog),
			"drift_count", len(report.DriftEvents),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}
