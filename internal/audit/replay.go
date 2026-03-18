package audit

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// ReplayEvent is one event in an agent's timeline.
type ReplayEvent struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	CreatedAt     time.Time       `json:"created_at"`
	FeatureVector json.RawMessage `json:"feature_vector,omitempty"`
}

// ReplayDecision is one enforcement decision in the timeline.
type ReplayDecision struct {
	ID           int64           `json:"id"`
	Decision     string          `json:"decision"`
	Score        int             `json:"score"`
	ScoreDelta   *int            `json:"score_delta,omitempty"`
	PolicyFired  *string         `json:"policy_fired,omitempty"`
	ReasonObject json.RawMessage `json:"reason_object,omitempty"`
	Override     bool            `json:"override"`
	CreatedAt    time.Time       `json:"created_at"`
}

// ReplayResponse is the full timeline for one agent.
type ReplayResponse struct {
	AgentDID    string           `json:"agent_did"`
	FromTime    time.Time        `json:"from"`
	ToTime      time.Time        `json:"to"`
	Events      []ReplayEvent    `json:"events"`
	Decisions   []ReplayDecision `json:"decisions"`
	TotalEvents int              `json:"total_events"`
	GeneratedAt time.Time        `json:"generated_at"`
}

// SOC2ExportRow is one row in the SOC2 export.
type SOC2ExportRow struct {
	ID           int64           `json:"id"`
	AgentDID     string          `json:"agent_did"`
	Decision     string          `json:"decision"`
	Score        int             `json:"score"`
	ScoreDelta   *int            `json:"score_delta,omitempty"`
	PolicyFired  *string         `json:"policy_fired,omitempty"`
	ReasonObject json.RawMessage `json:"reason_object,omitempty"`
	Override     bool            `json:"override"`
	ThisHash     *string         `json:"this_hash,omitempty"`
	PrevHash     *string         `json:"prev_hash,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

// SOC2Export is the full compliance export.
type SOC2Export struct {
	ExportedAt    time.Time       `json:"exported_at"`
	FromTime      time.Time       `json:"from"`
	ToTime        time.Time       `json:"to"`
	TotalRows     int             `json:"total_rows"`
	ChainVerified bool            `json:"chain_verified"`
	Decisions     []SOC2ExportRow `json:"decisions"`
}

// Handler handles audit endpoints.
type Handler struct {
	db *sql.DB
}

// NewHandler creates a new audit handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// ReplayHandler handles GET /audit/replay?agent_did=...&from=...&to=...
// Returns full event timeline and enforcement decisions for one agent.
// FM1 prevention: returns metadata only — no payload content.
// FM3 prevention: org_id enforced (Phase 1 single-tenant — skipped for now).
func (h *Handler) ReplayHandler(w http.ResponseWriter, r *http.Request) {
	agentDID := r.URL.Query().Get("agent_did")
	if agentDID == "" {
		http.Error(w, "agent_did is required", http.StatusBadRequest)
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()

	if fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = t
		}
	}
	if toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = t
		}
	}

	requestID := r.Header.Get("X-Request-ID")
	slog.Info("replay_request",
		"request_id", requestID,
		"agent_did", agentDID,
		"from", from,
		"to", to,
	)

	// Fetch events — metadata only (Tier 1 privacy)
	events, err := h.fetchEvents(agentDID, from, to)
	if err != nil {
		slog.Error("fetch events failed",
			"agent_did", agentDID, "error", err)
		http.Error(w, "failed to fetch events", http.StatusInternalServerError)
		return
	}

	// Fetch enforcement decisions
	decisions, err := h.fetchDecisions(agentDID, from, to)
	if err != nil {
		slog.Error("fetch decisions failed",
			"agent_did", agentDID, "error", err)
		http.Error(w, "failed to fetch decisions", http.StatusInternalServerError)
		return
	}

	response := ReplayResponse{
		AgentDID:    agentDID,
		FromTime:    from,
		ToTime:      to,
		Events:      events,
		Decisions:   decisions,
		TotalEvents: len(events),
		GeneratedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExportHandler handles GET /audit/export?from=...&to=...&format=soc2
// Returns SOC2-compatible JSON with hash chain verification.
func (h *Handler) ExportHandler(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from := time.Now().Add(-30 * 24 * time.Hour)
	to := time.Now()

	if fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = t
		}
	}
	if toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			to = t
		}
	}

	// Verify hash chain
	chainValid := h.verifyHashChain()

	// Fetch all decisions in range
	rows, err := h.db.Query(`
		SELECT id, agent_did, decision, score, score_delta,
			policy_fired, reason_object, override,
			this_hash, prev_hash, created_at
		FROM enforcement_decisions
		WHERE created_at BETWEEN $1 AND $2
		ORDER BY id ASC`,
		from, to)
	if err != nil {
		http.Error(w, "export failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var decisions []SOC2ExportRow
	for rows.Next() {
		var row SOC2ExportRow
		var reasonJSON []byte
		err := rows.Scan(
			&row.ID, &row.AgentDID, &row.Decision, &row.Score,
			&row.ScoreDelta, &row.PolicyFired, &reasonJSON,
			&row.Override, &row.ThisHash, &row.PrevHash, &row.CreatedAt,
		)
		if err != nil {
			continue
		}
		row.ReasonObject = reasonJSON
		decisions = append(decisions, row)
	}

	export := SOC2Export{
		ExportedAt:    time.Now(),
		FromTime:      from,
		ToTime:        to,
		TotalRows:     len(decisions),
		ChainVerified: chainValid,
		Decisions:     decisions,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=are-audit-%s.json",
			time.Now().Format("2006-01-02")))
	json.NewEncoder(w).Encode(export)
}

func (h *Handler) fetchEvents(agentDID string,
	from, to time.Time) ([]ReplayEvent, error) {

	rows, err := h.db.Query(`
		SELECT event_id, event_type, created_at
		FROM agent_events
		WHERE agent_did = $1
		AND created_at BETWEEN $2 AND $3
		ORDER BY created_at ASC`,
		agentDID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []ReplayEvent
	for rows.Next() {
		var e ReplayEvent
		if err := rows.Scan(&e.EventID, &e.EventType, &e.CreatedAt); err != nil {
			continue
		}
		events = append(events, e)
	}
	return events, nil
}

func (h *Handler) fetchDecisions(agentDID string,
	from, to time.Time) ([]ReplayDecision, error) {

	rows, err := h.db.Query(`
		SELECT id, decision, score, score_delta, policy_fired,
			reason_object, override, created_at
		FROM enforcement_decisions
		WHERE agent_did = $1
		AND created_at BETWEEN $2 AND $3
		ORDER BY created_at ASC`,
		agentDID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []ReplayDecision
	for rows.Next() {
		var d ReplayDecision
		var reasonJSON []byte
		if err := rows.Scan(
			&d.ID, &d.Decision, &d.Score, &d.ScoreDelta,
			&d.PolicyFired, &reasonJSON, &d.Override, &d.CreatedAt,
		); err != nil {
			continue
		}
		d.ReasonObject = reasonJSON
		decisions = append(decisions, d)
	}
	return decisions, nil
}

func (h *Handler) verifyHashChain() bool {
	var valid bool
	err := h.db.QueryRow(
		`SELECT verify_hash_chain('enforcement_decisions')`).Scan(&valid)
	if err != nil {
		slog.Error("hash chain verify failed", "error", err)
		return false
	}
	return valid
}
