package scoring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// inMemorySIRStore — test-only implementation of SIRStore
type inMemorySIRStore struct {
	records map[string]*SIRRecord
}

func newInMemorySIRStore() *inMemorySIRStore {
	return &inMemorySIRStore{records: make(map[string]*SIRRecord)}
}

func (s *inMemorySIRStore) GetSIRRecord(ctx context.Context, agentDID string) (*SIRRecord, error) {
	r, ok := s.records[agentDID]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return r, nil
}

func (s *inMemorySIRStore) PutSIRRecord(ctx context.Context, record *SIRRecord) error {
	s.records[record.AgentDID] = record
	return nil
}

func newTestMachine() (*SIRMachine, *inMemorySIRStore) {
	store := newInMemorySIRStore()
	thresholds := DefaultSIRThresholds()
	return NewSIRMachine(thresholds, store), store
}

func TestSIR_NewAgentStartsInObserve(t *testing.T) {
	m, _ := newTestMachine()
	record, err := m.Evaluate(context.Background(), "did:jwt:test:001", 0.30)
	require.NoError(t, err)
	assert.Equal(t, SIRObserve, record.CurrentState)
	assert.Equal(t, 0, len(record.Transitions))
}

func TestSIR_ObserveToSuspect(t *testing.T) {
	m, _ := newTestMachine()
	record, err := m.Evaluate(context.Background(), "did:jwt:test:001", 0.70)
	require.NoError(t, err)
	assert.Equal(t, SIRSuspect, record.CurrentState)
	assert.Equal(t, 1, len(record.Transitions))
	assert.Equal(t, "score_threshold_suspect", record.Transitions[0].Reason)
}

func TestSIR_ObserveToIncident(t *testing.T) {
	m, _ := newTestMachine()
	record, err := m.Evaluate(context.Background(), "did:jwt:test:001", 0.90)
	require.NoError(t, err)
	assert.Equal(t, SIRIncident, record.CurrentState)
	assert.Equal(t, 1, record.IncidentCount)
}

func TestSIR_SuspectEscalatestoIncident(t *testing.T) {
	m, _ := newTestMachine()
	ctx := context.Background()
	did := "did:jwt:test:escalate"
	m.Evaluate(ctx, did, 0.70)                // SUSPECT
	record, err := m.Evaluate(ctx, did, 0.85) // INCIDENT
	require.NoError(t, err)
	assert.Equal(t, SIRIncident, record.CurrentState)
	assert.Equal(t, 2, len(record.Transitions))
}

func TestSIR_SuspectRecovery(t *testing.T) {
	m, _ := newTestMachine()
	ctx := context.Background()
	did := "did:jwt:test:recover"
	m.Evaluate(ctx, did, 0.70)                // SUSPECT
	record, err := m.Evaluate(ctx, did, 0.30) // back to OBSERVE
	require.NoError(t, err)
	assert.Equal(t, SIRObserve, record.CurrentState)
}

func TestSIR_IncidentRequiresHumanClear(t *testing.T) {
	m, _ := newTestMachine()
	ctx := context.Background()
	did := "did:jwt:test:block"
	m.Evaluate(ctx, did, 0.90) // INCIDENT
	// Additional score events must not move out of INCIDENT
	record, err := m.Evaluate(ctx, did, 0.20)
	require.NoError(t, err)
	assert.Equal(t, SIRIncident, record.CurrentState, "INCIDENT must not self-recover")
}

func TestSIR_HumanClearStartsProbation(t *testing.T) {
	m, _ := newTestMachine()
	ctx := context.Background()
	did := "did:jwt:test:clear"
	m.Evaluate(ctx, did, 0.90) // INCIDENT
	record, err := m.HumanClear(ctx, did, "soc_analyst:jsmith")
	require.NoError(t, err)
	assert.Equal(t, SIRRemediated, record.CurrentState)
	assert.NotNil(t, record.ProbationEndsAt)
	assert.Equal(t, "soc_analyst:jsmith", record.ClearRequestedBy)
}

func TestSIR_HumanClearRequiresIncidentState(t *testing.T) {
	m, _ := newTestMachine()
	ctx := context.Background()
	did := "did:jwt:test:notincident"
	m.Evaluate(ctx, did, 0.30) // OBSERVE
	_, err := m.HumanClear(ctx, did, "soc_analyst:jsmith")
	assert.Error(t, err, "clear on non-INCIDENT must error")
}

func TestSIR_ProbationRelapse(t *testing.T) {
	m, _ := newTestMachine()
	// Shorten probation for test
	m.thresholds.ProbationWindow = 1 * time.Millisecond
	ctx := context.Background()
	did := "did:jwt:test:relapse"
	m.Evaluate(ctx, did, 0.90)                 // INCIDENT
	m.HumanClear(ctx, did, "soc_analyst:test") // REMEDIATED
	// Relapse during probation
	record, err := m.Evaluate(ctx, did, 0.90)
	require.NoError(t, err)
	assert.Equal(t, SIRIncident, record.CurrentState)
	assert.Equal(t, 2, record.IncidentCount, "relapse increments IncidentCount")
}

func TestSIR_ProbationElapsedTransitionsToCleared(t *testing.T) {
	m, _ := newTestMachine()
	m.thresholds.ProbationWindow = 1 * time.Millisecond
	ctx := context.Background()
	did := "did:jwt:test:cleared"
	m.Evaluate(ctx, did, 0.90)                 // INCIDENT
	m.HumanClear(ctx, did, "soc_analyst:test") // REMEDIATED
	time.Sleep(5 * time.Millisecond)           // probation elapsed
	record, err := m.Evaluate(ctx, did, 0.30)  // trigger check
	require.NoError(t, err)
	assert.Equal(t, SIRCleared, record.CurrentState)
}

func TestSIR_TransitionAuditTrailNeverOverwritten(t *testing.T) {
	m, _ := newTestMachine()
	ctx := context.Background()
	did := "did:jwt:test:audit"
	m.Evaluate(ctx, did, 0.70)                 // SUSPECT (1 transition)
	m.Evaluate(ctx, did, 0.90)                 // INCIDENT (2 transitions)
	m.HumanClear(ctx, did, "soc_analyst:test") // REMEDIATED (3 transitions)
	store := m.store.(*inMemorySIRStore)
	record := store.records[did]
	assert.Equal(t, 3, len(record.Transitions), "all transitions must be preserved")
	assert.Equal(t, SIRSuspect, record.Transitions[0].ToState)
	assert.Equal(t, SIRIncident, record.Transitions[1].ToState)
	assert.Equal(t, SIRRemediated, record.Transitions[2].ToState)
}
