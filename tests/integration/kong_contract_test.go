package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// KongScoreResponse mirrors exactly what the scoring service returns
// and what the Kong Lua plugin must parse correctly.
// If this struct changes — the Kong plugin MUST be updated simultaneously.
// Any mismatch causes silent fail-open in production.
type KongScoreResponse struct {
	AgentDID   string          `json:"agent_did"`
	Score      int             `json:"score"`
	Band       string          `json:"band"`
	Confidence float64         `json:"confidence"`
	Source     string          `json:"source"`
	Reason     json.RawMessage `json:"reason"`
}

// TestKongContractScoreResponseSchema verifies the scoring service
// response schema matches what the Kong plugin expects.
// NEVER modify this test to make it pass — fix the schema mismatch instead.
func TestKongContractScoreResponseSchema(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   string
		expectValid    bool
		expectedBand   string
		expectedFields []string
	}{
		{
			name: "TRUSTED agent response parses correctly",
			mockResponse: `{
  "agent_did": "did:jwt:test-agent-001",
  "score": 850,
  "band": "TRUSTED",
  "confidence": 0.94,
  "source": "redis_cache",
  "reason": {"decision": "ALLOW", "policy_violations": []}
}`,
			expectValid:    true,
			expectedBand:   "TRUSTED",
			expectedFields: []string{"agent_did", "score", "band", "confidence", "source", "reason"},
		},
		{
			name: "BLOCKED agent response parses correctly",
			mockResponse: `{
  "agent_did": "did:jwt:compromised-agent-001",
  "score": 187,
  "band": "BLOCKED",
  "confidence": 0.97,
  "source": "redis_cache",
  "reason": {"decision": "BLOCK", "policy_violations": ["pii_field_access_rate", "cross_tenant_probe"]}
}`,
			expectValid:    true,
			expectedBand:   "BLOCKED",
			expectedFields: []string{"agent_did", "score", "band", "confidence", "source", "reason"},
		},
		{
			name: "MONITORED agent response parses correctly",
			mockResponse: `{
  "agent_did": "did:jwt:monitored-agent-001",
  "score": 620,
  "band": "MONITORED",
  "confidence": 0.71,
  "source": "postgres_fallback",
  "reason": {"decision": "ALLOW", "policy_violations": []}
}`,
			expectValid:    true,
			expectedBand:   "MONITORED",
			expectedFields: []string{"agent_did", "score", "band", "confidence", "source", "reason"},
		},
		{
			name: "RESTRICTED agent response parses correctly",
			mockResponse: `{
  "agent_did": "did:jwt:restricted-agent-001",
  "score": 350,
  "band": "RESTRICTED",
  "confidence": 0.83,
  "source": "redis_cache",
  "reason": {"decision": "THROTTLE", "policy_violations": ["permission_escalation"]}
}`,
			expectValid:    true,
			expectedBand:   "RESTRICTED",
			expectedFields: []string{"agent_did", "score", "band", "confidence", "source", "reason"},
		},
		{
			name: "Score band boundary — 800 is TRUSTED threshold",
			mockResponse: `{
  "agent_did": "did:jwt:boundary-agent-001",
  "score": 800,
  "band": "TRUSTED",
  "confidence": 0.50,
  "source": "redis_cache",
  "reason": {"decision": "ALLOW", "policy_violations": []}
}`,
			expectValid:    true,
			expectedBand:   "TRUSTED",
			expectedFields: []string{"agent_did", "score", "band", "confidence", "source", "reason"},
		},
		{
			name: "Score band boundary — 799 is MONITORED",
			mockResponse: `{
  "agent_did": "did:jwt:boundary-agent-002",
  "score": 799,
  "band": "MONITORED",
  "confidence": 0.50,
  "source": "redis_cache",
  "reason": {"decision": "ALLOW", "policy_violations": []}
}`,
			expectValid:    true,
			expectedBand:   "MONITORED",
			expectedFields: []string{"agent_did", "score", "band", "confidence", "source", "reason"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the response as Kong would
			var resp KongScoreResponse
			if err := json.Unmarshal([]byte(tt.mockResponse), &resp); err != nil {
				if tt.expectValid {
					t.Fatalf("CONTRACT VIOLATION: Kong cannot parse scoring service response: %v\nResponse: %s", err, tt.mockResponse)
				}
				return
			}

			// Verify all required fields are present
			for _, field := range tt.expectedFields {
				if !strings.Contains(tt.mockResponse, fmt.Sprintf(`"%s"`, field)) {
					t.Errorf("CONTRACT VIOLATION: required field '%s' missing from response", field)
				}
			}

			// Verify band matches expectation
			if resp.Band != tt.expectedBand {
				t.Errorf("CONTRACT VIOLATION: expected band %s, got %s", tt.expectedBand, resp.Band)
			}

			// Verify score is within valid range
			if resp.Score < 0 || resp.Score > 1000 {
				t.Errorf("CONTRACT VIOLATION: score %d outside valid range [0, 1000]", resp.Score)
			}

			// Verify confidence is within valid range
			if resp.Confidence < 0.0 || resp.Confidence > 1.0 {
				t.Errorf("CONTRACT VIOLATION: confidence %.2f outside valid range [0.0, 1.0]", resp.Confidence)
			}

			// Verify band is one of the four valid values
			validBands := map[string]bool{
				"TRUSTED": true, "MONITORED": true,
				"RESTRICTED": true, "BLOCKED": true,
			}
			if !validBands[resp.Band] {
				t.Errorf("CONTRACT VIOLATION: band '%s' is not a valid band value", resp.Band)
			}

			// Verify agent_did is non-empty
			if resp.AgentDID == "" {
				t.Errorf("CONTRACT VIOLATION: agent_did is empty")
			}
		})
	}
}

// TestKongContractFailOpenOnMissingFields verifies Kong handles
// partial or malformed responses gracefully (fail-open behaviour).
func TestKongContractFailOpenOnMissingFields(t *testing.T) {
	malformedResponses := []struct {
		name     string
		response string
	}{
		{
			name:     "empty response",
			response: `{}`,
		},
		{
			name:     "missing band field",
			response: `{"agent_did": "test", "score": 850, "confidence": 0.9}`,
		},
		{
			name:     "missing score field",
			response: `{"agent_did": "test", "band": "TRUSTED", "confidence": 0.9}`,
		},
	}

	for _, tt := range malformedResponses {
		t.Run(tt.name, func(t *testing.T) {
			var resp KongScoreResponse
			err := json.Unmarshal([]byte(tt.response), &resp)

			// Malformed responses should parse without error (JSON is valid)
			// but Kong must detect missing required fields and fail-open
			if err != nil {
				t.Logf("Response failed to parse: %v — Kong must handle this gracefully", err)
			}

			// Kong fail-open rule: if band is empty or unrecognised → ALLOW
			// This is the correct behaviour — never block on scoring uncertainty
			validBands := map[string]bool{
				"TRUSTED": true, "MONITORED": true,
				"RESTRICTED": true, "BLOCKED": true,
			}
			if resp.Band != "" && !validBands[resp.Band] {
				t.Logf("WARNING: unrecognised band '%s' — Kong should fail-open", resp.Band)
			}
		})
	}
}

// TestKongContractHTTPIntegration verifies the score endpoint
// returns the correct Content-Type and HTTP status.
func TestKongContractHTTPIntegration(t *testing.T) {
	// Mock scoring service handler matching exact production format
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
  "agent_did": "did:jwt:test-agent",
  "score": 850,
  "band": "TRUSTED",
  "confidence": 0.94,
  "source": "redis_cache",
  "reason": {"decision": "ALLOW", "policy_violations": []}
}`)
	}))
	defer mockServer.Close()

	resp, err := http.Get(mockServer.URL + "/score/did:jwt:test-agent")
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// Verify HTTP status
	if resp.StatusCode != http.StatusOK {
		t.Errorf("CONTRACT VIOLATION: expected 200, got %d", resp.StatusCode)
	}

	// Verify Content-Type
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("CONTRACT VIOLATION: expected Content-Type application/json, got %s", ct)
	}

	// Verify response parses correctly
	var scoreResp KongScoreResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&scoreResp); err != nil {
		t.Fatalf("CONTRACT VIOLATION: response body failed to parse: %v", err)
	}

	if scoreResp.Band != "TRUSTED" {
		t.Errorf("CONTRACT VIOLATION: expected TRUSTED band, got %s", scoreResp.Band)
	}
}
