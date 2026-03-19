package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// SIEMEvent is the payload sent to the SIEM webhook.
// Structured for Splunk HEC, Microsoft Sentinel, and CrowdStrike formats.
type SIEMEvent struct {
	// Splunk HEC envelope
	Time       int64       `json:"time"`
	Host       string      `json:"host"`
	Source     string      `json:"source"`
	Sourcetype string      `json:"sourcetype"`
	Index      string      `json:"index,omitempty"`
	Event      SIEMPayload `json:"event"`
}

// SIEMPayload is the inner event data.
type SIEMPayload struct {
	EventType    string          `json:"event_type"`
	AgentDID     string          `json:"agent_did"`
	Decision     string          `json:"decision"`
	Score        int             `json:"score"`
	ScoreDelta   int             `json:"score_delta"`
	PolicyFired  string          `json:"policy_fired"`
	ReasonObject json.RawMessage `json:"reason_object"`
	Severity     string          `json:"severity"`
	Timestamp    time.Time       `json:"timestamp"`
	ProductName  string          `json:"product"`
	Version      string          `json:"version"`
}

// SIEMWebhook sends enforcement decisions to a configured SIEM.
type SIEMWebhook struct {
	webhookURL string
	authToken  string
	hostname   string
	client     *http.Client
	enabled    bool
}

// NewSIEMWebhook creates a SIEM webhook from environment config.
// SIEM_WEBHOOK_URL — Splunk HEC URL or generic webhook endpoint
// SIEM_AUTH_TOKEN  — Splunk HEC token or Bearer token
func NewSIEMWebhook() *SIEMWebhook {
	webhookURL := os.Getenv("SIEM_WEBHOOK_URL")
	authToken := os.Getenv("SIEM_AUTH_TOKEN")
	hostname, _ := os.Hostname()

	enabled := webhookURL != ""
	if !enabled {
		slog.Warn("SIEM_WEBHOOK_URL not set — SIEM integration disabled")
	}

	return &SIEMWebhook{
		webhookURL: webhookURL,
		authToken:  authToken,
		hostname:   hostname,
		enabled:    enabled,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// SendBlocked fires a SIEM event for every BLOCKED decision.
// Non-blocking — runs in goroutine, never delays enforcement path.
func (s *SIEMWebhook) SendBlocked(
	agentDID string,
	score int,
	scoreDelta int,
	policyFired string,
	reasonJSON []byte,
) {
	if !s.enabled {
		return
	}
	// Fire and forget — SIEM delivery must never block enforcement
	go s.send(agentDID, "BLOCKED", score, scoreDelta, policyFired, reasonJSON, "HIGH")
}

// SendRestricted fires a SIEM event for RESTRICTED decisions.
func (s *SIEMWebhook) SendRestricted(
	agentDID string,
	score int,
	scoreDelta int,
	policyFired string,
	reasonJSON []byte,
) {
	if !s.enabled {
		return
	}
	go s.send(agentDID, "RESTRICTED", score, scoreDelta, policyFired, reasonJSON, "MEDIUM")
}

func (s *SIEMWebhook) send(
	agentDID, decision string,
	score, scoreDelta int,
	policyFired string,
	reasonJSON []byte,
	severity string,
) {
	event := SIEMEvent{
		Time:       time.Now().Unix(),
		Host:       s.hostname,
		Source:     "agentrepengine",
		Sourcetype: "agent:enforcement",
		Event: SIEMPayload{
			EventType:    "agent_enforcement_decision",
			AgentDID:     agentDID,
			Decision:     decision,
			Score:        score,
			ScoreDelta:   scoreDelta,
			PolicyFired:  policyFired,
			ReasonObject: reasonJSON,
			Severity:     severity,
			Timestamp:    time.Now(),
			ProductName:  "AgentRepEngine",
			Version:      "1.0.0",
		},
	}

	payload, err := json.Marshal(event)
	if err != nil {
		slog.Error("siem_marshal_failed", "error", err)
		return
	}

	req, err := http.NewRequest("POST", s.webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		slog.Error("siem_request_failed", "error", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if s.authToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Splunk %s", s.authToken))
	}

	resp, err := s.client.Do(req)
	if err != nil {
		slog.Error("siem_delivery_failed",
			"agent_did", agentDID,
			"decision", decision,
			"error", err,
		)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		slog.Error("siem_rejected",
			"agent_did", agentDID,
			"status", resp.StatusCode,
		)
		return
	}

	slog.Info("siem_delivered",
		"agent_did", agentDID,
		"decision", decision,
		"severity", severity,
		"status", resp.StatusCode,
	)
}
