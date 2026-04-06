// internal/bridge/entra_bridge.go
// L116 — Microsoft Entra / AGT identity bridge
// Maps Microsoft AGT agent identity claims → ARE JWT format.
// Fail-open: if AGT token is malformed or bridge is unavailable,
// caller falls back to orphan agent path (score=500, MONITORED).
// No hard dependency on Microsoft API — bridge is optional enrichment.

package bridge

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// AGTClaims represents the subset of Microsoft AGT token claims ARE needs.
// AGT tokens are standard JWTs — no Microsoft SDK required.
type AGTClaims struct {
	AgentID  string `json:"agent_id"` // AGT agent identifier
	TenantID string `json:"tid"`      // Microsoft tenant ID
	ObjectID string `json:"oid"`      // Agent object ID in Entra
	AppID    string `json:"appid"`    // Application ID
	Expiry   int64  `json:"exp"`      // Unix expiry
	IssuedAt int64  `json:"iat"`      // Unix issued at
}

// AREBridgeIdentity is the normalized ARE identity derived from AGT claims.
// Caller uses this to construct an ARE JWT via identity.SignToken().
type AREBridgeIdentity struct {
	AgentDID    string // are: did:jwt:entra:<tenant_id>:<object_id>
	OrgID       string // are: <tenant_id>
	InstanceID  string // are: <app_id>
	LineageHash string // are: sha256-stub (real lineage requires ARE-native spawn)
}

// BridgeResult is returned by EntraToARE.
type BridgeResult struct {
	Identity *AREBridgeIdentity
	Err      error
	FailOpen bool // true = bridge failed, caller should use orphan path
}

// EntraToARE extracts AGT claims from a raw JWT token string (no signature
// verification — ARE's /verify endpoint handles RS256 validation downstream).
// This bridge only performs claims extraction and format mapping.
// Signature verification of the AGT token is out of scope — enterprise
// deployments validate AGT tokens at their Entra tenant boundary.
func EntraToARE(agtToken string) BridgeResult {
	if agtToken == "" {
		return BridgeResult{FailOpen: true, Err: fmt.Errorf("empty AGT token")}
	}

	claims, err := extractAGTClaims(agtToken)
	if err != nil {
		return BridgeResult{FailOpen: true, Err: fmt.Errorf("AGT claims extraction failed: %w", err)}
	}

	if claims.TenantID == "" || claims.ObjectID == "" {
		return BridgeResult{FailOpen: true, Err: fmt.Errorf("AGT token missing tid or oid claims")}
	}

	// Map AGT identity → ARE identity format
	identity := &AREBridgeIdentity{
		AgentDID:    fmt.Sprintf("did:jwt:entra:%s:%s", claims.TenantID, claims.ObjectID),
		OrgID:       claims.TenantID,
		InstanceID:  claims.AppID,
		LineageHash: fmt.Sprintf("entra-bridge:%s:%s", claims.TenantID, claims.ObjectID),
	}

	return BridgeResult{Identity: identity, FailOpen: false}
}

// EntraBridgeHandler is an HTTP handler for POST /bridge/entra
// Accepts: {"agt_token": "<raw AGT JWT>"}
// Returns: {"agent_did": "...", "org_id": "...", "instance_id": "...", "lineage_hash": "..."}
// On failure: returns 200 with fail_open=true — never blocks the request path.
func EntraBridgeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			AGTToken string `json:"agt_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"fail_open":true,"error":"invalid request body"}`)
			return
		}

		result := EntraToARE(body.AGTToken)
		w.Header().Set("Content-Type", "application/json")

		if result.FailOpen {
			// Fail-open: return orphan identity, never block
			fmt.Fprintf(w, `{"fail_open":true,"error":%q,"agent_did":"","org_id":"","instance_id":"","lineage_hash":""}`,
				result.Err.Error())
			return
		}

		fmt.Fprintf(w, `{"fail_open":false,"agent_did":%q,"org_id":%q,"instance_id":%q,"lineage_hash":%q}`,
			result.Identity.AgentDID,
			result.Identity.OrgID,
			result.Identity.InstanceID,
			result.Identity.LineageHash,
		)
	}
}

// extractAGTClaims decodes the payload section of a JWT without verifying signature.
func extractAGTClaims(token string) (*AGTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	payload := parts[1]
	// Pad base64 if needed
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}
	payload = strings.ReplaceAll(payload, "-", "+")
	payload = strings.ReplaceAll(payload, "_", "/")

	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %w", err)
	}

	var claims AGTClaims
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, fmt.Errorf("JSON unmarshal failed: %w", err)
	}

	return &claims, nil
}

// IsAGTToken returns true if the token looks like a Microsoft AGT token.
// Heuristic only — used by Kong plugin to decide whether to call bridge.
func IsAGTToken(token string) bool {
	if token == "" {
		return false
	}
	claims, err := extractAGTClaims(token)
	if err != nil {
		return false
	}
	// AGT tokens have tid (tenant ID) and oid (object ID) claims
	return claims.TenantID != "" && claims.ObjectID != ""
}

// Ensure time import is used (expiry validation placeholder for Phase 2)
var _ = time.Now
