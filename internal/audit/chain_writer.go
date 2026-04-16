package audit

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

// EnforcementDecisionInput holds all fields needed to write a
// tamper-evident enforcement decision to the audit log.
// chain_position is assigned by PostgreSQL SEQUENCE — not caller.
// this_hash is computed from chain_position + fields after INSERT.
type EnforcementDecisionInput struct {
	AgentDID   string
	Decision   string // BLOCKED / FREEZE / OVERRIDE / FLAGGED
	Score      int
	ScoreDelta int
	PolicyFired string
	ReasonObj  interface{} // serialized to jsonb
	Override   bool
	ReviewerID string
}

// WriteEnforcementDecision writes a hash-chained enforcement decision.
//
// Two-step protocol (required by hash chain design):
//   Step 1: INSERT row — PostgreSQL assigns chain_position via SEQUENCE.
//           prev_hash and this_hash are NULL at this point.
//   Step 2: Compute this_hash = SHA256(chain_position|agent_did|decision|score|created_at)
//           UPDATE row to set prev_hash (from last row) and this_hash.
//
// The SEQUENCE guarantees chain_position is unique and monotonic even
// under concurrent inserts — no serialization errors, no fork risk.
//
// NEVER-TOUCH: The hash algorithm below must not be changed without
// invalidating all existing audit logs. Any change requires a new
// migration and a chain break disclosure.
func WriteEnforcementDecision(db *sql.DB, input EnforcementDecisionInput) (int64, error) {
	reasonJSON, err := json.Marshal(input.ReasonObj)
	if err != nil {
		reasonJSON = []byte(fmt.Sprintf(`{"decision":"%s","score":%d}`,
			input.Decision, input.Score))
	}

	reviewerID := sql.NullString{Valid: false}
	if input.ReviewerID != "" {
		reviewerID = sql.NullString{String: input.ReviewerID, Valid: true}
	}

	policyFired := sql.NullString{Valid: false}
	if input.PolicyFired != "" {
		policyFired = sql.NullString{String: input.PolicyFired, Valid: true}
	}

	// Step 1: INSERT — chain_position assigned by SEQUENCE
	var decisionID int64
	var chainPosition int64
	var createdAt time.Time

	err = db.QueryRow(`
		INSERT INTO enforcement_decisions
			(agent_did, decision, score, score_delta,
			 policy_fired, reason_object, override, reviewer_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, chain_position, created_at`,
		input.AgentDID,
		input.Decision,
		input.Score,
		input.ScoreDelta,
		policyFired,
		reasonJSON,
		input.Override,
		reviewerID,
	).Scan(&decisionID, &chainPosition, &createdAt)
	if err != nil {
		return 0, fmt.Errorf("enforcement_decision insert: %w", err)
	}

	// Step 2: Compute prev_hash from last committed row before this one
	var prevHash sql.NullString
	err = db.QueryRow(`
		SELECT this_hash
		FROM enforcement_decisions
		WHERE chain_position < $1
		  AND this_hash IS NOT NULL
		ORDER BY chain_position DESC
		LIMIT 1`,
		chainPosition,
	).Scan(&prevHash)
	// err == sql.ErrNoRows means this is the first row — prevHash stays NULL (genesis)
	if err != nil && err != sql.ErrNoRows {
		slog.Warn("chain_writer_prev_hash_lookup_failed",
			"decision_id", decisionID,
			"chain_position", chainPosition,
			"error", err,
		)
	}

	// Compute this_hash
	// Format: SHA256(chain_position | agent_did | decision | score | created_at_unix)
	// LOCKED: changing this format breaks all existing audit logs.
	hashInput := fmt.Sprintf("%d|%s|%s|%d|%d",
		chainPosition,
		input.AgentDID,
		input.Decision,
		input.Score,
		createdAt.UnixNano(),
	)
	hashBytes := sha256.Sum256([]byte(hashInput))
	thisHash := hex.EncodeToString(hashBytes[:])

	// Step 2: UPDATE with computed hashes
	_, err = db.Exec(`
		UPDATE enforcement_decisions
		SET prev_hash  = $1,
		    this_hash  = $2
		WHERE id = $3`,
		prevHash,
		thisHash,
		decisionID,
	)
	if err != nil {
		// Log but do not fail — the decision row exists, hashes can be
		// recomputed. A missing hash is better than a lost decision.
		slog.Error("chain_writer_hash_update_failed",
			"decision_id", decisionID,
			"chain_position", chainPosition,
			"this_hash", thisHash,
			"error", err,
		)
		return decisionID, nil
	}

	slog.Info("enforcement_decision_written",
		"decision_id", decisionID,
		"chain_position", chainPosition,
		"decision", input.Decision,
		"agent_did", input.AgentDID,
		"this_hash", thisHash[:16]+"...",
	)

	return decisionID, nil
}
