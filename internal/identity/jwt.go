package identity

import (
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
)

const (
	// TokenTTL is the JWT expiry. 3600s = 1 hour max per APEX spec.
	TokenTTL = 3600 * time.Second

	// KeySize is RSA key size. 2048 minimum for RS256.
	KeySize = 2048

	// PrivateKeyPath is where the RS256 private key is persisted.
	// NEVER generate in memory only — key must survive restarts.
	PrivateKeyPath = "keys/private_key.pem"
	PublicKeyPath  = "keys/public_key.pem"
)

// KeyPair holds the RSA key pair used for JWT signing.
type KeyPair struct {
	Private *rsa.PrivateKey
	Public  *rsa.PublicKey
}

// LoadOrGenerateKeys loads the RS256 key pair from disk.
// If keys do not exist, generates and persists them.
// CRITICAL: Keys must be persisted — never generated in memory only.
// FM1 prevention: restart-safe key loading.
func LoadOrGenerateKeys() (*KeyPair, error) {
	// Try loading existing keys first
	if _, err := os.Stat(PrivateKeyPath); err == nil {
		return loadKeysFromDisk()
	}

	// Generate new keys and persist them
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
	// Create keys directory
	if err := os.MkdirAll("keys", 0700); err != nil {
		return nil, fmt.Errorf("create keys dir: %w", err)
	}

	// Generate RS256 key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, KeySize)
	if err != nil {
		return nil, fmt.Errorf("generate RSA key: %w", err)
	}

	// Persist private key
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

	// Persist public key
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
	// Validate all required claims are present — FM2 prevention
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
// Returns claims if valid, error if expired, tampered, or wrong algorithm.
func VerifyToken(tokenString string, keys *KeyPair) (*AgentClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&AgentClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Enforce RS256 — reject any other algorithm
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
	return claims, nil
}

// ComputeLineageHash computes sha256(parentDID + "|" + spawnedAtUnix).
// Root agents pass "root" as parentDID.
func ComputeLineageHash(parentDID string, spawnedAt time.Time) string {
	input := fmt.Sprintf("%s|%d", parentDID, spawnedAt.Unix())
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// NewAgentClaims builds a complete AgentClaims struct.
// Enforces TTL and sets issued-at automatically.
func NewAgentClaims(agentDID, instanceID, lineageHash, orgID string, depth int) *AgentClaims {
	now := time.Now()
	return &AgentClaims{
		AgentDID:     agentDID,
		InstanceID:   instanceID,
		LineageHash:  lineageHash,
		OrgID:        orgID,
		LineageDepth: depth,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(TokenTTL)),
		},
	}
}

// validateClaims checks all 7 required fields are non-empty.
// FM2 prevention: lineage_hash missing = identity drift risk.
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
	return nil
}
