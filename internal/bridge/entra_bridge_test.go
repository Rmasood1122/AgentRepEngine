// internal/bridge/entra_bridge_test.go
package bridge

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

// buildTestAGTToken constructs a minimal unsigned JWT with AGT-style claims.
func buildTestAGTToken(tid, oid, appid string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claims := map[string]interface{}{
		"tid":      tid,
		"oid":      oid,
		"appid":    appid,
		"agent_id": "test-agent",
		"exp":      9999999999,
		"iat":      1700000000,
	}
	payload, _ := json.Marshal(claims)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	return header + "." + encodedPayload + ".fakesig"
}

func TestEntraToARE_ValidToken(t *testing.T) {
	token := buildTestAGTToken("tenant-abc", "object-xyz", "app-123")
	result := EntraToARE(token)

	if result.FailOpen {
		t.Fatalf("expected success, got fail_open: %v", result.Err)
	}
	if result.Identity == nil {
		t.Fatal("expected identity, got nil")
	}
	if !strings.HasPrefix(result.Identity.AgentDID, "did:jwt:entra:") {
		t.Errorf("AgentDID format wrong: %s", result.Identity.AgentDID)
	}
	if result.Identity.OrgID != "tenant-abc" {
		t.Errorf("OrgID wrong: got %s, want tenant-abc", result.Identity.OrgID)
	}
	if result.Identity.InstanceID != "app-123" {
		t.Errorf("InstanceID wrong: got %s, want app-123", result.Identity.InstanceID)
	}
	expected := "did:jwt:entra:tenant-abc:object-xyz"
	if result.Identity.AgentDID != expected {
		t.Errorf("AgentDID wrong: got %s, want %s", result.Identity.AgentDID, expected)
	}
}

func TestEntraToARE_EmptyToken(t *testing.T) {
	result := EntraToARE("")
	if !result.FailOpen {
		t.Fatal("expected fail_open for empty token")
	}
}

func TestEntraToARE_MissingTID(t *testing.T) {
	// Token with oid but no tid
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claims := map[string]interface{}{"oid": "object-xyz", "appid": "app-123"}
	payload, _ := json.Marshal(claims)
	token := header + "." + base64.RawURLEncoding.EncodeToString(payload) + ".fakesig"
	result := EntraToARE(token)
	if !result.FailOpen {
		t.Fatal("expected fail_open for missing tid")
	}
}

func TestEntraToARE_MalformedJWT(t *testing.T) {
	result := EntraToARE("not.a.valid.jwt.format.extra")
	if !result.FailOpen {
		t.Fatal("expected fail_open for malformed JWT")
	}
}

func TestIsAGTToken_Valid(t *testing.T) {
	token := buildTestAGTToken("tenant-abc", "object-xyz", "app-123")
	if !IsAGTToken(token) {
		t.Fatal("expected true for valid AGT token")
	}
}

func TestIsAGTToken_Empty(t *testing.T) {
	if IsAGTToken("") {
		t.Fatal("expected false for empty token")
	}
}

func TestIsAGTToken_NonAGT(t *testing.T) {
	// ARE-native token without tid/oid — should return false
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claims := map[string]interface{}{"agent_did": "did:jwt:org:svc:uuid", "org_id": "org1"}
	payload, _ := json.Marshal(claims)
	token := header + "." + base64.RawURLEncoding.EncodeToString(payload) + ".fakesig"
	if IsAGTToken(token) {
		t.Fatal("expected false for non-AGT token")
	}
}
