package identity

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// TokenTTL is the JWT expiry. 3600s = 1 hour max per APEX spec.
	TokenTTL = 3600 * time.Second

	// KeySize is RSA key size. 2048 minimum for RS256.
	KeySize = 2048

	// PrivateKeyPath is where the RS256 private key is persisted.
	PrivateKeyPath = "keys/private_key.pem"
	PublicKeyPath  = "keys/public_key.pem"

	// replayKeyPrefix is the Redis key prefix for used JTIs.
	// used_token:{jti} → "1" with TTL = token exp
	replayKeyPrefix = "used_token:"
)

// KeyPair holds the RSA key pair used for JWT signing.
type KeyPair struct {
	Private *rsa.PrivateKey
	Public  *rsa.PublicKey
}

// LoadOrGenerateKeys loads the RS256 key pair from disk.
// If keys do not exist, generates and persists them.
func LoadOrGenerateKeys() (*KeyPair, error) {
	if _, err := os.Stat(PrivateKeyPath); err == nil {
		return loadKeysFromDisk()
	}
	return generateAndPersistKeys()
}

func loadKeysFromDisk() (*KeyPair, error) {
	privBytes, err := os.ReadFile(PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	block, _ := pem.Decode(privBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block for private key")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	return &KeyPair{
		Private: privateKey,
		Public:  &privateKey.PublicKey,
	}, nil
}

func generateAndPersistKeys() (*KeyPair, error) {
	if err := os.MkdirAll("keys", 0700); err != nil {
		return nil, fmt.Errorf("create keys dir: %w", err)
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, KeySize)
	if err != nil {
		return nil, fmt.Errorf("generate RSA key: %w", err)
	}
	privFile, err := os.OpenFile(PrivateKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, fmt.Errorf("open private key file: %w", err)
	}
	defer privFile.Close()
	if err := pem.Encode(privFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}); err != nil {
		return nil, fmt.Errorf("encode private key: %w", err)
	}
	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("marshal public key: %w", err)
	}
	pubFile, err := os.OpenFile(PublicKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("open public key file: %w", err)
	}
	defer pubFile.Close()
	if err := pem.Encode(pubFile, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}); err != nil {
		return nil, fmt.Errorf("encode public key: %w", err)
	}
	return &KeyPair{
		Private: privateKey,
		Public:  &privateKey.PublicKey,
	}, nil
}

// SignToken creates a signed RS256 JWT for an agent.
// NEVER uses HS256 — symmetric key is a shared secret risk.
func SignToken(claims *AgentClaims, keys *KeyPair) (string, error) {
	if err := validateClaims(claims); err != nil {
		return "", fmt.Errorf("invalid claims: %w", err)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(keys.Private)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// VerifyToken parses and validates a JWT string.
// L23 FIX: checks jti against Redis used-token cache to prevent replay.
// Fail-open on Redis unavailability — infrastructure failure never blocks agents.
func VerifyToken(tokenString string, keys *KeyPair, rdb *redis.Client) (*AgentClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&AgentClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return keys.Public, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}

	claims, ok := token.Claims.(*AgentClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// L23: JWT replay detection
	// jti must be present — tokens without jti are rejected
	jti := claims.RegisteredClaims.ID
	if jti == "" {
		return nil, errors.New("token missing jti claim — replay detection requires jti")
	}

	ctx := context.Background()
	replayKey := replayKeyPrefix + jti

	// Check if this jti has been used before
	// Fail-open: if Redis is down, log and allow (infrastructure failure)
	used, err := rdb.Exists(ctx, replayKey).Result()
	if err != nil {
		// Redis unavailable — fail-open per APEX Law L2
		// Log the failure but do not block the agent
		fmt.Printf("WARN: replay check unavailable (Redis error: %v) — failing open\n", err)
	} else if used > 0 {
		return nil, fmt.Errorf("token replay detected: jti %s already used", jti)
	} else {
		// Mark this jti as used with TTL matching token expiry
		ttl := TokenTTL
		if claims.ExpiresAt != nil {
			remaining := time.Until(claims.ExpiresAt.Time)
			if remaining > 0 {
				ttl = remaining
			}
		}
		if setErr := rdb.SetEx(ctx, replayKey, "1", ttl).Err(); setErr != nil {
			// Cannot store — fail-open, log warning
			fmt.Printf("WARN: cannot store jti in Redis (error: %v) — failing open\n", setErr)
		}
	}

	return claims, nil
}

// NewAgentClaims builds a complete AgentClaims struct.
// L23 FIX: populates jti (RegisteredClaims.ID) with a UUID.
func NewAgentClaims(agentDID, instanceID, lineageHash, orgID string, depth int) *AgentClaims {
	now := time.Now()
	return &AgentClaims{
		AgentDID:     agentDID,
		InstanceID:   instanceID,
		LineageHash:  lineageHash,
		OrgID:        orgID,
		LineageDepth: depth,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), // jti — unique per token, required for replay detection
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(TokenTTL)),
		},
	}
}

// ComputeLineageHash computes sha256(parentDID + "|" + spawnedAtUnix).
func ComputeLineageHash(parentDID string, spawnedAt time.Time) string {
	input := fmt.Sprintf("%s|%d", parentDID, spawnedAt.Unix())
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// validateClaims checks all required fields are non-empty.
func validateClaims(c *AgentClaims) error {
	if c.AgentDID == "" {
		return errors.New("agent_did is required")
	}
	if c.InstanceID == "" {
		return errors.New("instance_id is required")
	}
	if c.LineageHash == "" {
		return errors.New("lineage_hash is required")
	}
	if c.OrgID == "" {
		return errors.New("org_id is required")
	}
	if c.RegisteredClaims.ID == "" {
		return errors.New("jti is required — use NewAgentClaims to generate tokens")
	}
	return nil
}
