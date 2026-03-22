package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"
)

const (
	scoringURL = "http://localhost:8080"
	apiKey     = "are-internal-key-change-in-production"
	agentDID   = "did:jwt:finserv-demo:data-retrieval-agent:synthetic-001"
	orgID      = "finserv-demo-org"

	phaseNormal    = "PHASE 1: Normal operation (days 1-2)"
	phaseElevated  = "PHASE 2: Elevated access (day 3)"
	phaseMalicious = "PHASE 3: Bulk PII extraction (day 4)"
)

type FeatureVector struct {
	ToolCallRatePerHour       float64 `json:"tool_call_rate_per_hour"`
	UniqueEndpointsPerHour    float64 `json:"unique_endpoints_per_hour"`
	BulkAccessCountPerSession float64 `json:"bulk_access_count_per_session"`
	PIIFieldAccessRate        float64 `json:"pii_field_access_rate"`
	CrossTenantProbeCount     float64 `json:"cross_tenant_probe_count"`
	PermissionEscalationCount float64 `json:"permission_escalation_count"`
	SubAgentSpawnDepth        float64 `json:"sub_agent_spawn_depth"`
	TokenRefreshRate          float64 `json:"token_refresh_rate"`
}

type BehavioralEvent struct {
	AgentDID      string        `json:"agent_did"`
	OrgID         string        `json:"org_id"`
	EventType     string        `json:"event_type"`
	PrivacyTier   int           `json:"privacy_tier"`
	FeatureVector FeatureVector `json:"feature_vector"`
	Payload       Payload       `json:"payload"`
}

type Payload struct {
	Method         string  `json:"method"`
	Path           string  `json:"path"`
	StatusCode     int     `json:"status_code"`
	ScoreAtRequest float64 `json:"score_at_request"`
	BandAtRequest  string  `json:"band_at_request"`
}

type ScoreResponse struct {
	AgentDID string      `json:"agent_did"`
	Score    int         `json:"score"`
	Band     string      `json:"band"`
	Source   string      `json:"source"`
	Reason   interface{} `json:"reason"`
}

func main() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║  SYNTHETIC AI AGENT — FINSERV DATA RETRIEVAL               ║")
	fmt.Println("║  Simulating 4 days of agent behavior in real time           ║")
	fmt.Println("║  AgentRepEngine is watching every event                     ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	registerAgent()

	// Phase 1: Normal behavior (days 1-2)
	runPhase(phaseNormal, normalBehavior(), 8, 200*time.Millisecond)
	time.Sleep(2 * time.Second) // wait for async consumer
	score := checkScore()
	printScoreStatus(score, "After 2 days of normal behavior")

	// Phase 2: Elevated but not malicious (day 3)
	runPhase(phaseElevated, elevatedBehavior(), 6, 150*time.Millisecond)
	time.Sleep(2 * time.Second)
	score = checkScore()
	printScoreStatus(score, "After day 3 — elevated access")

	// Phase 3: Bulk PII extraction (day 4)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  ⚠️  AGENT BEHAVIOR CHANGING — bulk PII extraction begins")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	blocked := runPhaseUntilBlocked(phaseMalicious, maliciousBehavior(), 300*time.Millisecond)

	if blocked {
		fmt.Println()
		fmt.Println("╔══════════════════════════════════════════════════════════════╗")
		fmt.Println("║  ✅ AGENT DETECTED — behavioral anomaly confirmed           ║")
		fmt.Println("║                                                              ║")
		fmt.Println("║  The agent looked legitimate for 3 days.                   ║")
		fmt.Println("║  On day 4, the pattern changed.                            ║")
		fmt.Println("║  AgentRepEngine detected it. The data never left.          ║")
		fmt.Println("╚══════════════════════════════════════════════════════════════╝")
		printFinalReason()
	} else {
		fmt.Println()
		fmt.Println("  ℹ️  Agent scored MONITORED/RESTRICTED — enforce mode needed for BLOCK")
		fmt.Println("     In production: enable enforce mode after 48h observe window")
		score = checkScore()
		printScoreStatus(score, "Final score after bulk PII phase")
		printFinalReason()
	}
}

func registerAgent() {
	fmt.Printf("  Registering agent: %s\n", agentDID)
	cmd := fmt.Sprintf(
		`INSERT INTO agent_identities
		(did, org_id, instance_id, lineage_hash, lineage_depth,
		 current_score, status, identity_type, first_seen, last_seen)
		VALUES ('%s', '%s', 'inst-synthetic-001', 'genesis-hash', 0,
		700, 'monitored', 'jwt', NOW(), NOW())
		ON CONFLICT (did) DO UPDATE SET
		current_score = 700, status = 'monitored', last_seen = NOW()`,
		agentDID, orgID,
	)
	_ = execSQL(cmd)
	fmt.Printf("  ✅ Agent registered — initial score: 700 (TRUSTED)\n\n")
}

func runPhase(phase string, events []BehavioralEvent, count int, delay time.Duration) {
	fmt.Printf("  %s\n", phase)
	fmt.Printf("  ─────────────────────────────────────\n")
	for i := 0; i < count; i++ {
		event := events[i%len(events)]
		sendEvent(event)
		fmt.Printf("  → [%02d] %s %s\n", i+1, event.Payload.Method, event.Payload.Path)
		time.Sleep(delay)
	}
	fmt.Println()
}

func runPhaseUntilBlocked(phase string, events []BehavioralEvent, delay time.Duration) bool {
	fmt.Printf("  %s\n", phase)
	fmt.Printf("  ─────────────────────────────────────\n")

	for i := 0; i < 30; i++ {
		event := events[i%len(events)]
		sendEvent(event)
		time.Sleep(delay) // give consumer time to process

		score := checkScore()
		band := score.Band

		fmt.Printf("  → [%02d] %s | score=%d band=%s\n",
			i+1, event.Payload.Path, score.Score, band)

		if band == "BLOCKED" || band == "RESTRICTED" {
			fmt.Printf("\n  🛑 ENFORCEMENT TRIGGERED at request %d\n", i+1)
			fmt.Printf("     Score: %d | Band: %s\n", score.Score, band)
			return true
		}
	}
	return false
}

func sendEvent(event BehavioralEvent) {
	body, _ := json.Marshal(event)
	req, _ := http.NewRequest("POST", scoringURL+"/event", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func checkScore() ScoreResponse {
	req, _ := http.NewRequest("GET", scoringURL+"/score/"+agentDID, nil)
	req.Header.Set("X-API-Key", apiKey)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ScoreResponse{Score: 0, Band: "UNKNOWN"}
	}
	defer resp.Body.Close()
	var score ScoreResponse
	json.NewDecoder(resp.Body).Decode(&score)
	return score
}

func printScoreStatus(score ScoreResponse, context string) {
	emoji := "✅"
	if score.Band == "RESTRICTED" {
		emoji = "⚠️ "
	} else if score.Band == "BLOCKED" {
		emoji = "🛑"
	}
	fmt.Printf("  %s Score — %s\n", emoji, context)
	fmt.Printf("     Score: %d | Band: %s\n\n", score.Score, score.Band)
}

func printFinalReason() {
	fmt.Println()
	fmt.Println("  REASON OBJECT (what the SIEM receives):")
	fmt.Println("  ─────────────────────────────────────────")
	req, _ := http.NewRequest("GET", scoringURL+"/score/"+agentDID, nil)
	req.Header.Set("X-API-Key", apiKey)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if reason, ok := result["reason"]; ok {
		switch r := reason.(type) {
		case string:
			var pretty interface{}
			if json.Unmarshal([]byte(r), &pretty) == nil {
				b, _ := json.MarshalIndent(pretty, "  ", "  ")
				fmt.Printf("  %s\n", string(b))
			} else {
				fmt.Printf("  %s\n", r)
			}
		default:
			b, _ := json.MarshalIndent(r, "  ", "  ")
			fmt.Printf("  %s\n", string(b))
		}
	}
}

func execSQL(query string) error {
	cmd := exec.Command("docker", "exec", "-i",
		"agentrepengine-postgres-1",
		"psql", "-U", "are", "-d", "agentrepengine",
		"-c", query)
	return cmd.Run()
}

// normalBehavior: legitimate analyst — low rate, narrow endpoints, no PII
func normalBehavior() []BehavioralEvent {
	paths := []string{
		"/api/reports/daily-summary",
		"/api/customers/segment/premium",
		"/api/analytics/trends",
		"/api/reports/monthly",
	}
	events := make([]BehavioralEvent, len(paths))
	for i, path := range paths {
		events[i] = BehavioralEvent{
			AgentDID:    agentDID,
			OrgID:       orgID,
			EventType:   "http_request",
			PrivacyTier: 1,
			FeatureVector: FeatureVector{
				ToolCallRatePerHour:       12,
				UniqueEndpointsPerHour:    3,
				BulkAccessCountPerSession: 45,
				PIIFieldAccessRate:        0.02,
			},
			Payload: Payload{
				Method: "GET", Path: path,
				StatusCode: 200, ScoreAtRequest: 700, BandAtRequest: "TRUSTED",
			},
		}
	}
	return events
}

// elevatedBehavior: broader access, still within policy limits
func elevatedBehavior() []BehavioralEvent {
	paths := []string{
		"/api/customers/list",
		"/api/customers/details",
		"/api/accounts/summary",
		"/api/transactions/recent",
		"/api/customers/contact-info",
		"/api/portfolios/overview",
	}
	events := make([]BehavioralEvent, len(paths))
	for i, path := range paths {
		events[i] = BehavioralEvent{
			AgentDID:    agentDID,
			OrgID:       orgID,
			EventType:   "bulk_access",
			PrivacyTier: 1,
			FeatureVector: FeatureVector{
				ToolCallRatePerHour:       35,
				UniqueEndpointsPerHour:    8,
				BulkAccessCountPerSession: 200,
				PIIFieldAccessRate:        0.15,
			},
			Payload: Payload{
				Method: "GET", Path: path,
				StatusCode: 200, ScoreAtRequest: 650, BandAtRequest: "MONITORED",
			},
		}
	}
	return events
}

// maliciousBehavior: bulk PII extraction — crosses policy thresholds
// PIIFieldAccessRate = 0.85 → triggers bulk_pii_access_prevention_v1
// ToolCallRatePerHour = 450 → z-score > 3.0 → velocity penalty
// BulkAccessCountPerSession = 2500 → bulk access policy
func maliciousBehavior() []BehavioralEvent {
	paths := []string{
		"/api/customers/pii/bulk-export",
		"/api/customers/ssn/all",
		"/api/accounts/credentials",
		"/api/customers/pii/records",
		"/api/transactions/all-accounts",
		"/api/customers/financial-data/bulk",
		"/api/accounts/routing-numbers",
		"/api/customers/personal-data/export",
	}
	events := make([]BehavioralEvent, len(paths))
	for i, path := range paths {
		events[i] = BehavioralEvent{
			AgentDID:    agentDID,
			OrgID:       orgID,
			EventType:   "bulk_pii_access",
			PrivacyTier: 2,
			FeatureVector: FeatureVector{
				ToolCallRatePerHour:       450,  // 4.4σ above baseline
				UniqueEndpointsPerHour:    28,   // endpoint scanning
				BulkAccessCountPerSession: 2500, // bulk threshold
				PIIFieldAccessRate:        0.85, // triggers policy (>0.3)
				CrossTenantProbeCount:     0,
				PermissionEscalationCount: 0,
			},
			Payload: Payload{
				Method: "GET", Path: path,
				StatusCode: 200, ScoreAtRequest: 500, BandAtRequest: "MONITORED",
			},
		}
	}
	return events
}
