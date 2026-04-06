package scoring

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SIRState represents the Suspect-Incident-Remediated lifecycle state of an agent.
type SIRState string

const (
	SIRObserve    SIRState = "OBSERVE"
	SIRSuspect    SIRState = "SUSPECT"
	SIRIncident   SIRState = "INCIDENT"
	SIRRemediated SIRState = "REMEDIATED"
	SIRCleared    SIRState = "CLEARED"
)

// SIRTransition records a state change with full audit context.
type SIRTransition struct {
	AgentDID    string    `json:"agent_did"`
	FromState   SIRState  `json:"from_state"`
	ToState     SIRState  `json:"to_state"`
	Score       float64   `json:"score"`
	Reason      string    `json:"reason"`
	TriggeredBy string    `json:"triggered_by"`
	Timestamp   time.Time `json:"timestamp"`
}

// SIRRecord is the persisted state for a single agent.
type SIRRecord struct {
	AgentDID         string          `json:"agent_did"`
	CurrentState     SIRState        `json:"current_state"`
	EnteredStateAt   time.Time       `json:"entered_state_at"`
	LastScore        float64         `json:"last_score"`
	IncidentCount    int             `json:"incident_count"`
	ClearRequestedBy string          `json:"clear_requested_by,omitempty"`
	ClearRequestedAt *time.Time      `json:"clear_requested_at,omitempty"`
	ProbationEndsAt  *time.Time      `json:"probation_ends_at,omitempty"`
	Transitions      []SIRTransition `json:"transitions"`
}

// SIRThresholds configures score boundaries for state transitions.
type SIRThresholds struct {
	SuspectScore    float64
	IncidentScore   float64
	ProbationWindow time.Duration
	ClearWindow     time.Duration
}

func DefaultSIRThresholds() SIRThresholds {
	return SIRThresholds{
		SuspectScore:    0.65,
		IncidentScore:   0.80,
		ProbationWindow: 7 * 24 * time.Hour,
		ClearWindow:     48 * time.Hour,
	}
}

// SIRStore is the persistence interface — implemented by ScoreStore.
type SIRStore interface {
	GetSIRRecord(ctx context.Context, agentDID string) (*SIRRecord, error)
	PutSIRRecord(ctx context.Context, record *SIRRecord) error
}

// SIRMachine evaluates score events and drives lifecycle transitions.
type SIRMachine struct {
	thresholds SIRThresholds
	store      SIRStore
	mu         sync.Mutex
}

func NewSIRMachine(thresholds SIRThresholds, store SIRStore) *SIRMachine {
	return &SIRMachine{
		thresholds: thresholds,
		store:      store,
	}
}

// Evaluate processes a new score for an agent and returns the resulting SIRRecord.
func (m *SIRMachine) Evaluate(ctx context.Context, agentDID string, score float64) (*SIRRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, err := m.store.GetSIRRecord(ctx, agentDID)
	if err != nil {
		record = &SIRRecord{
			AgentDID:       agentDID,
			CurrentState:   SIRObserve,
			EnteredStateAt: time.Now().UTC(),
			Transitions:    []SIRTransition{},
		}
	}

	record.LastScore = score

	switch record.CurrentState {
	case SIRObserve, SIRCleared:
		if score >= m.thresholds.IncidentScore {
			m.transition(record, SIRIncident, score, "score_threshold_incident", "scoring_engine")
		} else if score >= m.thresholds.SuspectScore {
			m.transition(record, SIRSuspect, score, "score_threshold_suspect", "scoring_engine")
		}

	case SIRSuspect:
		if score >= m.thresholds.IncidentScore {
			m.transition(record, SIRIncident, score, "score_escalation", "scoring_engine")
		} else if score < m.thresholds.SuspectScore {
			m.transition(record, SIRObserve, score, "score_recovery", "scoring_engine")
		}

	case SIRIncident:
		// Only human clear can move out of INCIDENT — score events do not change state

	case SIRRemediated:
		if record.ProbationEndsAt != nil && time.Now().UTC().After(*record.ProbationEndsAt) {
			m.transition(record, SIRCleared, score, "probation_elapsed", "scoring_engine")
		} else if score >= m.thresholds.IncidentScore {
			// Relapse during probation — transition() increments IncidentCount
			m.transition(record, SIRIncident, score, "probation_relapse", "scoring_engine")
		}
	}

	if err := m.store.PutSIRRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("sir_machine: persist failed for %s: %w", agentDID, err)
	}

	return record, nil
}

// HumanClear moves an INCIDENT agent to REMEDIATED and starts probation.
func (m *SIRMachine) HumanClear(ctx context.Context, agentDID string, clearedBy string) (*SIRRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, err := m.store.GetSIRRecord(ctx, agentDID)
	if err != nil {
		return nil, fmt.Errorf("sir_machine: agent not found: %s", agentDID)
	}

	if record.CurrentState != SIRIncident {
		return nil, fmt.Errorf("sir_machine: clear requires INCIDENT state, got %s", record.CurrentState)
	}

	now := time.Now().UTC()
	probationEnd := now.Add(m.thresholds.ProbationWindow)

	record.ClearRequestedBy = clearedBy
	record.ClearRequestedAt = &now
	record.ProbationEndsAt = &probationEnd

	m.transition(record, SIRRemediated, record.LastScore, "human_clear", clearedBy)

	if err := m.store.PutSIRRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("sir_machine: persist failed after human clear: %w", err)
	}

	return record, nil
}

// transition appends a state change — never overwrites.
func (m *SIRMachine) transition(record *SIRRecord, to SIRState, score float64, reason, triggeredBy string) {
	t := SIRTransition{
		AgentDID:    record.AgentDID,
		FromState:   record.CurrentState,
		ToState:     to,
		Score:       score,
		Reason:      reason,
		TriggeredBy: triggeredBy,
		Timestamp:   time.Now().UTC(),
	}
	record.Transitions = append(record.Transitions, t)
	record.CurrentState = to
	record.EnteredStateAt = t.Timestamp

	if to == SIRIncident {
		record.IncidentCount++
	}
}
