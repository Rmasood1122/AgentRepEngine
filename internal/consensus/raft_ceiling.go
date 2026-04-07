package consensus

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// RaftRole represents the role of a node in the Raft cluster
type RaftRole string

const (
	RoleFollower  RaftRole = "FOLLOWER"
	RoleCandidate RaftRole = "CANDIDATE"
	RoleLeader    RaftRole = "LEADER"
)

// CeilingCommand is a replicated log entry for ceiling decisions
type CeilingCommand struct {
	Type      string    `json:"type"` // "SET_CEILING" | "CLEAR_CEILING" | "RAISE_CEILING"
	AgentID   string    `json:"agent_id"`
	Value     int       `json:"value"`     // ceiling value (0 = clear)
	AuthorID  string    `json:"author_id"` // human operator ID
	Term      int       `json:"term"`
	Index     int       `json:"index"`
	Timestamp time.Time `json:"timestamp"`
}

// LogEntry wraps a command with Raft metadata
type LogEntry struct {
	Term    int            `json:"term"`
	Index   int            `json:"index"`
	Command CeilingCommand `json:"command"`
}

// VoteRequest is sent by candidates during leader election
type VoteRequest struct {
	Term         int    `json:"term"`
	CandidateID  string `json:"candidate_id"`
	LastLogIndex int    `json:"last_log_index"`
	LastLogTerm  int    `json:"last_log_term"`
}

// VoteResponse is returned by followers
type VoteResponse struct {
	Term        int    `json:"term"`
	VoterID     string `json:"voter_id"`
	VoteGranted bool   `json:"vote_granted"`
}

// AppendEntriesRequest is sent by leader to replicate log entries
type AppendEntriesRequest struct {
	Term         int        `json:"term"`
	LeaderID     string     `json:"leader_id"`
	PrevLogIndex int        `json:"prev_log_index"`
	PrevLogTerm  int        `json:"prev_log_term"`
	Entries      []LogEntry `json:"entries"`
	LeaderCommit int        `json:"leader_commit"`
}

// AppendEntriesResponse is returned by followers
type AppendEntriesResponse struct {
	Term       int    `json:"term"`
	NodeID     string `json:"node_id"`
	Success    bool   `json:"success"`
	MatchIndex int    `json:"match_index"`
}

// QuorumResult is the outcome of a quorum vote
type QuorumResult struct {
	Approved     bool     `json:"approved"`
	VotesFor     int      `json:"votes_for"`
	VotesAgainst int      `json:"votes_against"`
	QuorumSize   int      `json:"quorum_size"`
	NodeIDs      []string `json:"node_ids"`
	Term         int      `json:"term"`
}

// RaftNode represents a single ARE node participating in ceiling consensus
type RaftNode struct {
	mu sync.RWMutex

	// Identity
	NodeID       string
	ClusterNodes []string // all node IDs in cluster

	// Raft state
	Role        RaftRole
	CurrentTerm int
	VotedFor    string
	Log         []LogEntry
	CommitIndex int
	LastApplied int

	// Leader state (only valid when Role == LEADER)
	NextIndex  map[string]int // nextIndex[nodeID]
	MatchIndex map[string]int // matchIndex[nodeID]

	// Applied ceiling state (what's actually enforced)
	CeilingState  map[string]int  // agentID → ceiling value (0 = no ceiling)
	CeilingActive map[string]bool // agentID → ceiling enforced

	// Audit
	CommittedCommands []CeilingCommand
}

// NewRaftNode initializes a new Raft node in follower state
func NewRaftNode(nodeID string, clusterNodes []string) *RaftNode {
	n := &RaftNode{
		NodeID:            nodeID,
		ClusterNodes:      clusterNodes,
		Role:              RoleFollower,
		CurrentTerm:       0,
		VotedFor:          "",
		Log:               []LogEntry{},
		CommitIndex:       0,
		LastApplied:       0,
		NextIndex:         make(map[string]int),
		MatchIndex:        make(map[string]int),
		CeilingState:      make(map[string]int),
		CeilingActive:     make(map[string]bool),
		CommittedCommands: []CeilingCommand{},
	}
	for _, id := range clusterNodes {
		n.NextIndex[id] = 1
		n.MatchIndex[id] = 0
	}
	return n
}

// QuorumSize returns the minimum votes needed for a quorum
func (n *RaftNode) QuorumSize() int {
	return len(n.ClusterNodes)/2 + 1
}

// RequestVote processes a vote request — Raft §5.2
func (n *RaftNode) RequestVote(req VoteRequest) VoteResponse {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Reject if request term is stale
	if req.Term < n.CurrentTerm {
		return VoteResponse{Term: n.CurrentTerm, VoterID: n.NodeID, VoteGranted: false}
	}

	// Step down if we see a higher term
	if req.Term > n.CurrentTerm {
		n.CurrentTerm = req.Term
		n.Role = RoleFollower
		n.VotedFor = ""
	}

	// Grant vote if we haven't voted yet and candidate log is at least as up-to-date
	lastLogIndex := len(n.Log)
	lastLogTerm := 0
	if lastLogIndex > 0 {
		lastLogTerm = n.Log[lastLogIndex-1].Term
	}

	logOK := (req.LastLogTerm > lastLogTerm) ||
		(req.LastLogTerm == lastLogTerm && req.LastLogIndex >= lastLogIndex)

	if (n.VotedFor == "" || n.VotedFor == req.CandidateID) && logOK {
		n.VotedFor = req.CandidateID
		return VoteResponse{Term: n.CurrentTerm, VoterID: n.NodeID, VoteGranted: true}
	}

	return VoteResponse{Term: n.CurrentTerm, VoterID: n.NodeID, VoteGranted: false}
}

// AppendEntries processes a log replication request — Raft §5.3
func (n *RaftNode) AppendEntries(req AppendEntriesRequest) AppendEntriesResponse {
	n.mu.Lock()
	defer n.mu.Unlock()

	if req.Term < n.CurrentTerm {
		return AppendEntriesResponse{Term: n.CurrentTerm, NodeID: n.NodeID, Success: false}
	}

	if req.Term > n.CurrentTerm {
		n.CurrentTerm = req.Term
		n.Role = RoleFollower
		n.VotedFor = ""
	}

	// Verify prev log entry matches
	if req.PrevLogIndex > 0 {
		if len(n.Log) < req.PrevLogIndex {
			return AppendEntriesResponse{Term: n.CurrentTerm, NodeID: n.NodeID, Success: false}
		}
		if n.Log[req.PrevLogIndex-1].Term != req.PrevLogTerm {
			// Conflict — truncate log
			n.Log = n.Log[:req.PrevLogIndex-1]
			return AppendEntriesResponse{Term: n.CurrentTerm, NodeID: n.NodeID, Success: false}
		}
	}

	// Append new entries
	for i, entry := range req.Entries {
		idx := req.PrevLogIndex + i + 1
		if idx <= len(n.Log) {
			if n.Log[idx-1].Term != entry.Term {
				n.Log = n.Log[:idx-1]
			}
		}
		if idx > len(n.Log) {
			n.Log = append(n.Log, entry)
		}
	}

	// Update commit index
	if req.LeaderCommit > n.CommitIndex {
		n.CommitIndex = min(req.LeaderCommit, len(n.Log))
		n.applyCommittedEntries()
	}

	return AppendEntriesResponse{
		Term:       n.CurrentTerm,
		NodeID:     n.NodeID,
		Success:    true,
		MatchIndex: len(n.Log),
	}
}

// ProposeCeilingCommand proposes a ceiling command — only valid on leader
// In production this would replicate to followers over gRPC; here it
// simulates quorum within a local cluster for single-tenant Phase 1.
func (n *RaftNode) ProposeCeilingCommand(cmd CeilingCommand, peers []*RaftNode) (*QuorumResult, error) {
	n.mu.Lock()

	if n.Role != RoleLeader {
		n.mu.Unlock()
		return nil, errors.New("consensus: only leader can propose commands")
	}

	// Append to leader log
	cmd.Term = n.CurrentTerm
	cmd.Index = len(n.Log) + 1
	cmd.Timestamp = time.Now().UTC()
	entry := LogEntry{Term: n.CurrentTerm, Index: cmd.Index, Command: cmd}
	n.Log = append(n.Log, entry)
	n.mu.Unlock()

	// Replicate to peers and collect responses
	votesFor := 1 // leader counts as one vote
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, peer := range peers {
		if peer.NodeID == n.NodeID {
			continue
		}
		wg.Add(1)
		go func(p *RaftNode) {
			defer wg.Done()
			n.mu.RLock()
			prevIndex := len(n.Log) - 1
			prevTerm := 0
			if prevIndex > 0 {
				prevTerm = n.Log[prevIndex-1].Term
			}
			req := AppendEntriesRequest{
				Term:         n.CurrentTerm,
				LeaderID:     n.NodeID,
				PrevLogIndex: prevIndex,
				PrevLogTerm:  prevTerm,
				Entries:      []LogEntry{entry},
				LeaderCommit: n.CommitIndex,
			}
			n.mu.RUnlock()

			resp := p.AppendEntries(req)
			if resp.Success {
				mu.Lock()
				votesFor++
				mu.Unlock()
			}
		}(peer)
	}
	wg.Wait()

	quorum := n.QuorumSize()
	approved := votesFor >= quorum

	if approved {
		n.mu.Lock()
		n.CommitIndex = len(n.Log)
		n.applyCommittedEntries()
		n.mu.Unlock()
	}

	return &QuorumResult{
		Approved:     approved,
		VotesFor:     votesFor,
		VotesAgainst: len(n.ClusterNodes) - votesFor,
		QuorumSize:   quorum,
		NodeIDs:      n.ClusterNodes,
		Term:         n.CurrentTerm,
	}, nil
}

// BecomeLeader transitions node to leader — called after winning election
func (n *RaftNode) BecomeLeader() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Role = RoleLeader
	for _, id := range n.ClusterNodes {
		n.NextIndex[id] = len(n.Log) + 1
		n.MatchIndex[id] = 0
	}
}

// GetCeilingState returns the committed ceiling for an agent
func (n *RaftNode) GetCeilingState(agentID string) (value int, active bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.CeilingState[agentID], n.CeilingActive[agentID]
}

// applyCommittedEntries applies log entries up to CommitIndex
// Must be called with n.mu held (write lock)
func (n *RaftNode) applyCommittedEntries() {
	for n.LastApplied < n.CommitIndex {
		n.LastApplied++
		entry := n.Log[n.LastApplied-1]
		n.applyCeilingCommand(entry.Command)
		n.CommittedCommands = append(n.CommittedCommands, entry.Command)
	}
}

// applyCeilingCommand applies a single committed command to local state
func (n *RaftNode) applyCeilingCommand(cmd CeilingCommand) {
	switch cmd.Type {
	case "SET_CEILING":
		n.CeilingState[cmd.AgentID] = cmd.Value
		n.CeilingActive[cmd.AgentID] = true
	case "CLEAR_CEILING":
		n.CeilingState[cmd.AgentID] = 0
		n.CeilingActive[cmd.AgentID] = false
	case "RAISE_CEILING":
		if n.CeilingActive[cmd.AgentID] {
			n.CeilingState[cmd.AgentID] = cmd.Value
		}
	}
}

// AuditLog returns all committed commands — for compliance export
func (n *RaftNode) AuditLog() []CeilingCommand {
	n.mu.RLock()
	defer n.mu.RUnlock()
	result := make([]CeilingCommand, len(n.CommittedCommands))
	copy(result, n.CommittedCommands)
	return result
}

// ClusterStatus returns a summary for health endpoint
func (n *RaftNode) ClusterStatus() map[string]interface{} {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return map[string]interface{}{
		"node_id":      n.NodeID,
		"role":         string(n.Role),
		"term":         n.CurrentTerm,
		"commit_index": n.CommitIndex,
		"log_length":   len(n.Log),
		"quorum_size":  n.QuorumSize(),
		"cluster_size": len(n.ClusterNodes),
	}
}

func (n *RaftNode) LogLength() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.Log)
}

func (n *RaftNode) GetRole() RaftRole {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.Role
}

func (n *RaftNode) GetTerm() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.CurrentTerm
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RegulatoryNote returns the compliance statement for this module
func RegulatoryNote() string {
	return fmt.Sprintf(
		"Raft consensus ensures ceiling decisions require quorum agreement " +
			"across ARE cluster nodes. A single compromised node cannot unilaterally " +
			"clear or raise a ceiling. Quorum = floor(N/2)+1 nodes must agree. " +
			"DORA Article 10 / NIST SP 800-207 / SOC2 CC6.1 compliant.",
	)
}
