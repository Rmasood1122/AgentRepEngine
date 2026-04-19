package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ClearanceVerdict struct {
	AgentDID       string      `json:"agent_did"`
	Cleared        bool        `json:"cleared"`
	Verdict        string      `json:"verdict"`
	Timestamp      time.Time   `json:"timestamp"`
	ScoreCheck     CheckResult `json:"score_check"`
	ViolationCheck CheckResult `json:"violation_check"`
	MaturityCheck  CheckResult `json:"maturity_check"`
	DenialReasons  []string    `json:"denial_reasons,omitempty"`
}

type CheckResult struct {
	Passed    bool    `json:"passed"`
	Label     string  `json:"label"`
	Actual    float64 `json:"actual"`
	Threshold float64 `json:"threshold"`
}

func HandleAgentClearance(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
		if len(parts) < 3 || parts[0] != "agent" || parts[2] != "clearance" {
			http.Error(w, `{"error":"invalid path, expected /agent/{did}/clearance"}`, http.StatusBadRequest)
			return
		}
		agentDID := parts[1]
		if agentDID == "" {
			http.Error(w, `{"error":"agent_did required"}`, http.StatusBadRequest)
			return
		}

		scoreThreshold := 700.0
		if st := r.URL.Query().Get("score_threshold"); st != "" {
			if parsed, err := strconv.ParseFloat(st, 64); err == nil && parsed >= 0 && parsed <= 1000 {
				scoreThreshold = parsed
			}
		}

		violationWindowHours := 24
		if vw := r.URL.Query().Get("violation_window_hours"); vw != "" {
			if parsed, err := strconv.Atoi(vw); err == nil && parsed >= 1 && parsed <= 8760 {
				violationWindowHours = parsed
			}
		}

		minSamples := 10.0
		if ms := r.URL.Query().Get("min_samples"); ms != "" {
			if parsed, err := strconv.ParseFloat(ms, 64); err == nil && parsed >= 1 {
				minSamples = parsed
			}
		}

		var denialReasons []string

		var currentScore float64
		var agentStatus string
		err := db.QueryRow(
			`SELECT current_score, status FROM agent_identities WHERE did = $1`,
			agentDID,
		).Scan(&currentScore, &agentStatus)

		scoreCheck := CheckResult{
			Label:     "current_score >= threshold",
			Threshold: scoreThreshold,
		}
		if err == sql.ErrNoRows {
			scoreCheck.Passed = false
			scoreCheck.Actual = 0
			denialReasons = append(denialReasons, "agent not found in identity registry")
		} else if err != nil {
			http.Error(w, `{"error":"database query failed"}`, http.StatusInternalServerError)
			return
		} else {
			scoreCheck.Actual = currentScore
			scoreCheck.Passed = currentScore >= scoreThreshold
			if !scoreCheck.Passed {
				denialReasons = append(denialReasons,
					fmt.Sprintf("score %.1f below threshold %.1f", currentScore, scoreThreshold))
			}
			if agentStatus == "suspended" || agentStatus == "revoked" {
				scoreCheck.Passed = false
				denialReasons = append(denialReasons,
					fmt.Sprintf("agent status is %s", agentStatus))
			}
		}

		var blockCount int64
		err = db.QueryRow(
			`SELECT COUNT(*) FROM enforcement_decisions
			 WHERE agent_did = $1
			   AND decision = 'BLOCK'
			   AND created_at >= NOW() - ($2 || ' hours')::INTERVAL`,
			agentDID, strconv.Itoa(violationWindowHours),
		).Scan(&blockCount)

		violationCheck := CheckResult{
			Label:     "zero BLOCK decisions in window",
			Threshold: 0,
		}
		if err != nil {
			http.Error(w, `{"error":"violation check failed"}`, http.StatusInternalServerError)
			return
		}
		violationCheck.Actual = float64(blockCount)
		violationCheck.Passed = blockCount == 0
		if !violationCheck.Passed {
			denialReasons = append(denialReasons,
				fmt.Sprintf("%d BLOCK decision(s) in last %d hours", blockCount, violationWindowHours))
		}

		var sampleCount float64
		err = db.QueryRow(
			`SELECT COALESCE(sample_count, 0) FROM agent_baselines WHERE agent_did = $1`,
			agentDID,
		).Scan(&sampleCount)

		maturityCheck := CheckResult{
			Label:     "baseline sample_count >= minimum",
			Threshold: minSamples,
		}
		if err == sql.ErrNoRows {
			maturityCheck.Passed = false
			maturityCheck.Actual = 0
			denialReasons = append(denialReasons, "no baseline established")
		} else if err != nil {
			http.Error(w, `{"error":"maturity check failed"}`, http.StatusInternalServerError)
			return
		} else {
			maturityCheck.Actual = sampleCount
			maturityCheck.Passed = sampleCount >= minSamples
			if !maturityCheck.Passed {
				denialReasons = append(denialReasons,
					fmt.Sprintf("baseline maturity %.0f below minimum %.0f", sampleCount, minSamples))
			}
		}

		cleared := scoreCheck.Passed && violationCheck.Passed && maturityCheck.Passed
		verdict := "DENIED"
		if cleared {
			verdict = "CLEARED"
		}

		resp := ClearanceVerdict{
			AgentDID:       agentDID,
			Cleared:        cleared,
			Verdict:        verdict,
			Timestamp:      time.Now().UTC(),
			ScoreCheck:     scoreCheck,
			ViolationCheck: violationCheck,
			MaturityCheck:  maturityCheck,
			DenialReasons:  denialReasons,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
