package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type TrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Score     float64   `json:"score"`
	Decision  string    `json:"decision"`
}

type TrendResponse struct {
	AgentDID   string       `json:"agent_did"`
	WindowDays int          `json:"window_days"`
	Points     []TrendPoint `json:"points"`
	Count      int          `json:"count"`
	MinScore   float64      `json:"min_score"`
	MaxScore   float64      `json:"max_score"`
	AvgScore   float64      `json:"avg_score"`
}

func HandleAgentTrend(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
		if len(parts) < 3 || parts[0] != "agent" || parts[2] != "trend" {
			http.Error(w, `{"error":"invalid path, expected /agent/{did}/trend"}`, http.StatusBadRequest)
			return
		}
		agentDID := parts[1]
		if agentDID == "" {
			http.Error(w, `{"error":"agent_did required"}`, http.StatusBadRequest)
			return
		}

		days := 30
		if d := r.URL.Query().Get("days"); d != "" {
			parsed, err := strconv.Atoi(d)
			if err != nil || parsed < 1 || parsed > 365 {
				http.Error(w, `{"error":"days must be 1-365"}`, http.StatusBadRequest)
				return
			}
			days = parsed
		}

		query := `
			SELECT score, decision, created_at
			FROM enforcement_decisions
			WHERE agent_did = $1
			  AND created_at >= NOW() - ($2 || ' days')::INTERVAL
			ORDER BY created_at ASC
			LIMIT 1000
		`
		rows, err := db.Query(query, agentDID, strconv.Itoa(days))
		if err != nil {
			http.Error(w, `{"error":"database query failed"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var points []TrendPoint
		var minScore, maxScore, sumScore float64
		first := true

		for rows.Next() {
			var p TrendPoint
			if err := rows.Scan(&p.Score, &p.Decision, &p.Timestamp); err != nil {
				http.Error(w, `{"error":"scan failed"}`, http.StatusInternalServerError)
				return
			}
			points = append(points, p)
			sumScore += p.Score
			if first {
				minScore = p.Score
				maxScore = p.Score
				first = false
			} else {
				if p.Score < minScore {
					minScore = p.Score
				}
				if p.Score > maxScore {
					maxScore = p.Score
				}
			}
		}

		if err := rows.Err(); err != nil {
			http.Error(w, `{"error":"row iteration failed"}`, http.StatusInternalServerError)
			return
		}

		avgScore := 0.0
		if len(points) > 0 {
			avgScore = sumScore / float64(len(points))
		}

		resp := TrendResponse{
			AgentDID:   agentDID,
			WindowDays: days,
			Points:     points,
			Count:      len(points),
			MinScore:   minScore,
			MaxScore:   maxScore,
			AvgScore:   avgScore,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
