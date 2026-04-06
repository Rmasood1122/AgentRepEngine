package identity

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"time"
)

// ContinuitySeal represents the identity seal stored at enrollment time.
type ContinuitySeal struct {
	AgentID  string
	Seal     string
	SealedAt time.Time
}

// SealAgentIdentity creates a SHA-256 seal over the agent's baseline hash
// and enrollment JTI. Call this at enrollment time.
func SealAgentIdentity(db *sql.DB, agentID, baselineHash, enrollmentJTI string) (*ContinuitySeal, error) {
	seal := computeSeal(baselineHash, enrollmentJTI)
	now := time.Now().UTC()

	_, err := db.Exec(`
		UPDATE agent_baselines
		SET identity_seal = $1, sealed_at = $2
		WHERE agent_id = $3
	`, seal, now, agentID)
	if err != nil {
		return nil, fmt.Errorf("seal agent identity: %w", err)
	}

	return &ContinuitySeal{
		AgentID:  agentID,
		Seal:     seal,
		SealedAt: now,
	}, nil
}

// VerifyContinuity checks whether the agent's current baseline hash and
// enrollment JTI still match the stored seal. Returns true if intact,
// false if post-redeployment erasure or tampering is detected.
func VerifyContinuity(db *sql.DB, agentID, currentBaselineHash, currentJTI string) (bool, error) {
	var storedSeal sql.NullString
	var sealedAt sql.NullTime

	err := db.QueryRow(`
		SELECT identity_seal, sealed_at
		FROM agent_baselines
		WHERE agent_id = $1
	`, agentID).Scan(&storedSeal, &sealedAt)

	if err == sql.ErrNoRows {
		return false, fmt.Errorf("agent not found: %s", agentID)
	}
	if err != nil {
		return false, fmt.Errorf("verify continuity: %w", err)
	}

	if !storedSeal.Valid || storedSeal.String == "" {
		// No seal — agent was never enrolled properly
		return false, nil
	}

	expected := computeSeal(currentBaselineHash, currentJTI)
	return expected == storedSeal.String, nil
}

// computeSeal produces a deterministic SHA-256 hex string over baseline + JTI.
func computeSeal(baselineHash, jti string) string {
	h := sha256.New()
	h.Write([]byte(baselineHash))
	h.Write([]byte(":"))
	h.Write([]byte(jti))
	return fmt.Sprintf("%x", h.Sum(nil))
}
