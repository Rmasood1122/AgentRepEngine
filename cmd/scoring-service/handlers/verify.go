package handlers

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/agentrepengine/are/internal/identity"
	"github.com/agentrepengine/are/internal/store"
)

// VerifyHandler validates a JWT token against the RS256 public key.
// Authenticated. Called by Kong plugin to verify agent identity.
func VerifyHandler(s *store.ScoreStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Token string `json:"token"`
		}
		if err := decodeJSON(r, &body); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"valid":false,"error":"invalid request body"}`)
			return
		}
		if body.Token == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"valid":false,"error":"token required"}`)
			return
		}
		keys, err := identity.LoadOrGenerateKeys()
		if err != nil {
			slog.Error("verify_keys_load_failed", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"valid":false,"error":"key load failed"}`)
			return
		}
		claims, err := identity.VerifyToken(body.Token, keys, s.GetRedisClient())
		if err != nil {
			slog.Warn("verify_token_rejected", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, `{"valid":false,"error":%q}`, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"valid":true,"agent_did":%q,"org_id":%q,"instance_id":%q,"lineage_hash":%q}`,
			claims.AgentDID, claims.OrgID, claims.InstanceID, claims.LineageHash)
	}
}

// JWKSHandler serves the RS256 public key in JWKS format.
// Unauthenticated — public keys are not secret by definition.
// Required by Kong plugin for JWT verification.
func JWKSHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keys, err := identity.LoadOrGenerateKeys()
		if err != nil {
			slog.Error("jwks_load_failed", "error", err)
			http.Error(w, "key load failed", http.StatusInternalServerError)
			return
		}
		pubKey := keys.Public
		n := base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes())
		e := base64.RawURLEncoding.EncodeToString(
			[]byte{byte(pubKey.E >> 16), byte(pubKey.E >> 8), byte(pubKey.E)})
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "alg": "RS256",
      "kid": "are-v1",
      "n": "%s",
      "e": "%s"
    }
  ]
}`, n, e)
	}
}
