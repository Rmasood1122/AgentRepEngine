package kong_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

const kongURL = "http://localhost:8000/test"

type httpbinResponse struct {
	Headers map[string]string `json:"headers"`
}

// TestInvalidJWTDeceptionModel verifies ARE's deception model:
// invalid JWT returns 200 (not 401) — attacker cannot detect enforcement layer.
// Competitors (Lakera, Microsoft AGT, Cisco) return 401.
// ARE returns 200 with internal flags in upstream headers.
// Locks FIX-1 from handler.lua v1.5.0 permanently.
func TestInvalidJWTDeceptionModel(t *testing.T) {
	req, err := http.NewRequest("GET", kongURL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer invalidtoken123")
	req.Header.Set("X-Agent-DID", "test-agent")
	req.Header.Set("X-Agent-Org-ID", "test-org")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed (is Kong running?): %v", err)
	}
	defer resp.Body.Close()

	// Deception model: must return 200 not 401
	if resp.StatusCode != 200 {
		t.Errorf("deception model broken: expected 200, got %d", resp.StatusCode)
	}

	// Read httpbin body to inspect upstream headers set by Kong plugin
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var hb httpbinResponse
	if err := json.Unmarshal(body, &hb); err != nil {
		t.Fatalf("failed to parse httpbin response: %v", err)
	}

	// Kong plugin must flag invalid JWT on upstream request
	if hb.Headers["X-Agent-Invalid-Jwt"] != "true" {
		t.Errorf("expected X-Agent-Invalid-Jwt: true in upstream headers, got %q",
			hb.Headers["X-Agent-Invalid-Jwt"])
	}

	// Must assign degraded score
	if hb.Headers["X-Agent-Score"] != "500" {
		t.Errorf("expected X-Agent-Score: 500, got %q",
			hb.Headers["X-Agent-Score"])
	}

	// Must assign MONITORED band
	if hb.Headers["X-Agent-Band"] != "MONITORED" {
		t.Errorf("expected X-Agent-Band: MONITORED, got %q",
			hb.Headers["X-Agent-Band"])
	}

	t.Logf("PASS deception model: HTTP 200, score=%s band=%s invalid-jwt=true",
		hb.Headers["X-Agent-Score"], hb.Headers["X-Agent-Band"])
}

// TestOrphanAgentFailOpen verifies fail-open behavior:
// no Authorization header passes through as orphan agent at score 500.
func TestOrphanAgentFailOpen(t *testing.T) {
	req, err := http.NewRequest("GET", kongURL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("X-Agent-Org-ID", "test-org")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed (is Kong running?): %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("fail-open broken: expected 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	var hb httpbinResponse
	if err := json.Unmarshal(body, &hb); err != nil {
		t.Fatalf("failed to parse httpbin response: %v", err)
	}

	if hb.Headers["X-Agent-Orphan"] != "true" {
		t.Errorf("expected X-Agent-Orphan: true, got %q",
			hb.Headers["X-Agent-Orphan"])
	}

	if hb.Headers["X-Agent-Score"] != "500" {
		t.Errorf("expected X-Agent-Score: 500, got %q",
			hb.Headers["X-Agent-Score"])
	}

	t.Logf("PASS fail-open: orphan agent score=%s band=%s",
		hb.Headers["X-Agent-Score"], hb.Headers["X-Agent-Band"])
}
