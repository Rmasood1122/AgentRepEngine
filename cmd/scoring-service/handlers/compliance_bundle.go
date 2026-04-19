package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/agentrepengine/are/internal/audit"
)

type ComplianceBundleResponse struct {
	OrgID      string                              `json:"org_id"`
	WindowDays int                                 `json:"window_days"`
	Frameworks map[string]*audit.RegulatoryPackage `json:"frameworks"`
	Count      int                                 `json:"framework_count"`
}

func HandleComplianceBundle(db *sql.DB) http.HandlerFunc {
	frameworks := []string{"SOX", "HIPAA", "DORA", "FFIEC", "NIST_RMF"}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		orgID := r.URL.Query().Get("org_id")
		if orgID == "" {
			orgID = "test-org"
		}

		windowDays := 30
		if wd := r.URL.Query().Get("window_days"); wd != "" {
			parsed, err := strconv.Atoi(wd)
			if err != nil || parsed < 1 || parsed > 365 {
				http.Error(w, `{"error":"window_days must be 1-365"}`, http.StatusBadRequest)
				return
			}
			windowDays = parsed
		}

		result := ComplianceBundleResponse{
			OrgID:      orgID,
			WindowDays: windowDays,
			Frameworks: make(map[string]*audit.RegulatoryPackage),
			Count:      len(frameworks),
		}

		for _, fw := range frameworks {
			pkg, err := audit.GenerateRegulatoryPackage(db, orgID, windowDays, fw)
			if err != nil {
				result.Frameworks[fw] = nil
				continue
			}
			result.Frameworks[fw] = pkg
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
