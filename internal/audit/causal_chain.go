package audit

import (
	"database/sql"
	"fmt"
	"time"
)

// CausalLink represents one hop in a multi-agent causal chain.
type CausalLink struct {
	AgentID       string    `json:"agent_id"`
	EventID       string    `json:"event_id"`
	Action        string    `json:"action"`
	AnomalyScore  float64   `json:"anomaly_score"`
	CreatedAt     time.Time `json:"created_at"`
	ParentAgentID string    `json:"parent_agent_id,omitempty"`
}

// CausalChain is the full attribution result for one enforcement event.
type CausalChain struct {
	EnforcementEventID         string       `json:"enforcement_event_id"`
	RootAgentID                string       `json:"root_agent_id"`
	EarliestEnforceableAgentID string       `json:"earliest_enforceable_agent_id"`
	Chain                      []CausalLink `json:"chain"`
	ChainLength                int          `json:"chain_length"`
}

// BuildCausalChain reconstructs the multi-agent causal chain for a given
// enforcement event ID. Returns root agent, full chain, and earliest
// enforceable agent for HIPAA breach attribution.
func BuildCausalChain(db *sql.DB, enforcementEventID string) (*CausalChain, error) {
	// Load the trigger event
	var rootAgentID, action string
	var anomalyScore float64
	var createdAt time.Time

	err := db.QueryRow(`
		SELECT agent_id, action, anomaly_score, created_at
		FROM enforcement_decisions
		WHERE event_id = $1
	`, enforcementEventID).Scan(&rootAgentID, &action, &anomalyScore, &createdAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("enforcement event not found: %s", enforcementEventID)
	}
	if err != nil {
		return nil, fmt.Errorf("load enforcement event: %w", err)
	}

	// Walk the causal chain via agent_calls table
	rows, err := db.Query(`
		WITH RECURSIVE chain AS (
			SELECT
				ac.caller_agent_id,
				ac.callee_agent_id,
				ee.event_id,
				ee.action,
				ee.anomaly_score,
				ee.created_at,
				0 AS depth
			FROM agent_calls ac
			JOIN enforcement_decisions ee ON ee.agent_id = ac.callee_agent_id
			WHERE ac.callee_agent_id = $1

			UNION ALL

			SELECT
				ac2.caller_agent_id,
				ac2.callee_agent_id,
				ee2.event_id,
				ee2.action,
				ee2.anomaly_score,
				ee2.created_at,
				chain.depth + 1
			FROM agent_calls ac2
			JOIN enforcement_decisions ee2 ON ee2.agent_id = ac2.callee_agent_id
			JOIN chain ON chain.caller_agent_id = ac2.callee_agent_id
			WHERE chain.depth < 10
		)
		SELECT callee_agent_id, event_id, action, anomaly_score, created_at, caller_agent_id
		FROM chain
		ORDER BY depth ASC
	`, rootAgentID)

	if err != nil {
		return nil, fmt.Errorf("walk causal chain: %w", err)
	}
	defer rows.Close()

	links := []CausalLink{
		{
			AgentID:      rootAgentID,
			EventID:      enforcementEventID,
			Action:       action,
			AnomalyScore: anomalyScore,
			CreatedAt:    createdAt,
		},
	}

	earliestEnforceable := rootAgentID

	for rows.Next() {
		var link CausalLink
		if err := rows.Scan(
			&link.AgentID,
			&link.EventID,
			&link.Action,
			&link.AnomalyScore,
			&link.CreatedAt,
			&link.ParentAgentID,
		); err != nil {
			return nil, fmt.Errorf("scan causal link: %w", err)
		}
		links = append(links, link)
		if link.AnomalyScore > 0.5 {
			earliestEnforceable = link.AgentID
		}
	}

	return &CausalChain{
		EnforcementEventID:         enforcementEventID,
		RootAgentID:                rootAgentID,
		EarliestEnforceableAgentID: earliestEnforceable,
		Chain:                      links,
		ChainLength:                len(links),
	}, nil
}
