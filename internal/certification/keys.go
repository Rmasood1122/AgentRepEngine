package certification

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

const (
	// CertKeyPath reuses the same RS256 key pair as identity signing.
	// Phase 2: separate operational key with annual rotation.
	CertKeyPath = "keys/private_key.pem"
)

// LoadOperationalKey loads the RS256 private key from disk.
// Reuses identity key pair — same pattern as internal/identity/jwt.go.
func LoadOperationalKey() (*rsa.PrivateKey, error) {
	privBytes, err := os.ReadFile(CertKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	block, _ := pem.Decode(privBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	return key, nil
}

// SignReport signs the report's CertHash with RS256.
// CertHash must be computed first via HashReport().
func SignReport(report *CertificationReport, key *rsa.PrivateKey) error {
	if report.CertHash == "" {
		return fmt.Errorf("cert_hash empty — call HashReport first")
	}
	digest := sha256.Sum256([]byte(report.CertHash))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		return fmt.Errorf("sign report: %w", err)
	}
	report.Signature = sig
	report.SignatureKeyID = "are-operational-v1"
	return nil
}
