package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/agentrepengine/are/internal/audit"
)

func dashboardHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/dashboard.html")
	}
}

func regulatoryPackageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID := r.URL.Query().Get("org_id")
		framework := r.URL.Query().Get("framework")
		if orgID == "" {
			orgID = "default"
		}
		if framework == "" {
			framework = "HIPAA"
		}

		pkg, err := audit.GenerateRegulatoryPackage(db, orgID, 30, framework)
		if err != nil {
			http.Error(w, fmt.Sprintf("package generation failed: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pkg)
	}
}
