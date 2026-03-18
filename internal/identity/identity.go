package identity

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

const (
	// InitialScore for all new agents — probation band ceiling is 750.
	InitialScore = 700

	// OrphanScore for agents with no registered identity.
	OrphanScore = 500

	// ProbationDuration is the 48h window before full trust is possible.
	ProbationDuration = 48 * time.Hour

	// IdentityCyclingThreshold — >5 new identities per org per hour = alert.
	IdentityCyclingThreshold = 5
)

// AgentIdentity represents a row in the agent_identities table.
type AgentIdentity struct {
	DID                string
	OrgID              string
	InstanceID         string
	LineageHash        string
	LineageDepth       int
	CurrentScore       int
	Status             string
	IdentityType       string
	FirstSeen          time.Time
	LastSeen           time.Time
	ProbationExpiresAt *time.Time
}

// Store handles all database operations for agent identities.
type Store struct {
	db *sql.DB
}

// NewStore creates a new identity store.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// RegisterAgent creates a new agent identity with probation mode.
// FM3 prevention: probation_expires_at always set on INSERT.
// New agents start at score 700, status probation, 48h window.
func (s *Store) RegisterAgent(did, orgID, instanceID, lineageHash string, depth int) (*AgentIdentity, error) {
	now := time.Now()
	probationExpires := now.Add(ProbationDuration)

	query := `
		INSERT INTO agent_identities
			(did, org_id, instance_id, lineage_hash, lineage_depth,
			 current_score, status, identity_type,
			 first_seen, last_seen, probation_expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'probation', 'jwt', $7, $7, $8)
		ON CONFLICT (did) DO UPDATE SET last_seen = $7
		RETURNING did, org_id, instance_id, lineage_hash, lineage_depth,
			current_score, status, identity_type,
			first_seen, last_seen, probation_expires_at`

	agent := &AgentIdentity{}
	err := s.db.QueryRow(query,
		did, orgID, instanceID, lineageHash, depth,
		InitialScore, now, probationExpires,
	).Scan(
		&agent.DID, &agent.OrgID, &agent.InstanceID,
		&agent.LineageHash, &agent.LineageDepth,
		&agent.CurrentScore, &agent.Status, &agent.IdentityType,
		&agent.FirstSeen, &agent.LastSeen, &agent.ProbationExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("register agent: %w", err)
	}
	return agent, nil
}

// GetAgent retrieves an agent identity by DID.
// Returns nil if not found — caller treats as orphan (score 500).
func (s *Store) GetAgent(did string) (*AgentIdentity, error) {
	query := `
		SELECT did, org_id, instance_id, lineage_hash, lineage_depth,
			current_score, status, identity_type,
			first_seen, last_seen, probation_expires_at
		FROM agent_identities WHERE did = $1`

	agent := &AgentIdentity{}
	err := s.db.QueryRow(query, did).Scan(
		&agent.DID, &agent.OrgID, &agent.InstanceID,
		&agent.LineageHash, &agent.LineageDepth,
		&agent.CurrentScore, &agent.Status, &agent.IdentityType,
		&agent.FirstSeen, &agent.LastSeen, &agent.ProbationExpiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // orphan — caller assigns score 500
	}
	if err != nil {
		return nil, fmt.Errorf("get agent: %w", err)
	}
	return agent, nil
}

// SubAgentScore returns the starting score for a sub-agent.
// Rule: min(parent_score, 700) — sub-agents cannot inherit high trust.
func SubAgentScore(parentScore int) int {
	if parentScore < InitialScore {
		return parentScore
	}
	return InitialScore
}

// IsInProbation returns true if the agent is still in the 48h window.
func (a *AgentIdentity) IsInProbation() bool {
	if a.ProbationExpiresAt == nil {
		return false
	}
	return time.Now().Before(*a.ProbationExpiresAt)
}

// CheckIdentityCycling returns true if an org has created
// more than 5 new agent identities in the last hour.
// Indicates possible identity cycling attack.
func (s *Store) CheckIdentityCycling(orgID string) (bool, int, error) {
	query := `
		SELECT COUNT(*) FROM agent_identities
		WHERE org_id = $1
		AND first_seen > NOW() - INTERVAL '1 hour'`

	var count int
	err := s.db.QueryRow(query, orgID).Scan(&count)
	if err != nil {
		return false, 0, fmt.Errorf("check identity cycling: %w", err)
	}
	return count > IdentityCyclingThreshold, count, nil
}

// ExitProbation moves an agent from probation to monitored status
// if all exit criteria are met.
func (s *Store) ExitProbation(did string) error {
	query := `
		UPDATE agent_identities
		SET status = 'monitored'
		WHERE did = $1
		AND status = 'probation'
		AND probation_expires_at < NOW()`

	_, err := s.db.Exec(query, did)
	if err != nil {
		return fmt.Errorf("exit probation: %w", err)
	}
	return nil
}
