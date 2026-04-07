package consensus

import (
	"fmt"
	"testing"
)

func makeCluster(n int) []*RaftNode {
	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("node-%d", i+1)
	}
	nodes := make([]*RaftNode, n)
	for i, id := range ids {
		nodes[i] = NewRaftNode(id, ids)
	}
	return nodes
}

func TestNewRaftNodeDefaults(t *testing.T) {
	node := NewRaftNode("node-1", []string{"node-1", "node-2", "node-3"})
	if node.Role != RoleFollower {
		t.Fatalf("expected FOLLOWER, got %s", node.Role)
	}
	if node.CurrentTerm != 0 {
		t.Fatalf("expected term 0, got %d", node.CurrentTerm)
	}
	if node.QuorumSize() != 2 {
		t.Fatalf("expected quorum 2 for 3 nodes, got %d", node.QuorumSize())
	}
}

func TestQuorumSize(t *testing.T) {
	cases := []struct{ nodes, quorum int }{
		{1, 1},
		{2, 2},
		{3, 2},
		{5, 3},
		{7, 4},
	}
	for _, c := range cases {
		ids := make([]string, c.nodes)
		for i := range ids {
			ids[i] = fmt.Sprintf("n%d", i)
		}
		node := NewRaftNode("n0", ids)
		if node.QuorumSize() != c.quorum {
			t.Errorf("nodes=%d: expected quorum %d, got %d", c.nodes, c.quorum, node.QuorumSize())
		}
	}
}

func TestRequestVoteStaleTerm(t *testing.T) {
	node := NewRaftNode("node-1", []string{"node-1", "node-2"})
	node.CurrentTerm = 5
	resp := node.RequestVote(VoteRequest{Term: 3, CandidateID: "node-2"})
	if resp.VoteGranted {
		t.Fatal("should not grant vote for stale term")
	}
}

func TestRequestVoteGranted(t *testing.T) {
	node := NewRaftNode("node-1", []string{"node-1", "node-2"})
	resp := node.RequestVote(VoteRequest{
		Term:         1,
		CandidateID:  "node-2",
		LastLogIndex: 0,
		LastLogTerm:  0,
	})
	if !resp.VoteGranted {
		t.Fatal("should grant vote for higher term with empty log")
	}
}

func TestRequestVoteAlreadyVoted(t *testing.T) {
	node := NewRaftNode("node-1", []string{"node-1", "node-2", "node-3"})
	node.RequestVote(VoteRequest{Term: 1, CandidateID: "node-2"})
	resp := node.RequestVote(VoteRequest{Term: 1, CandidateID: "node-3"})
	if resp.VoteGranted {
		t.Fatal("should not grant vote to second candidate in same term")
	}
}

func TestAppendEntriesStaleTerm(t *testing.T) {
	node := NewRaftNode("node-1", []string{"node-1", "node-2"})
	node.CurrentTerm = 5
	resp := node.AppendEntries(AppendEntriesRequest{Term: 3, LeaderID: "node-2"})
	if resp.Success {
		t.Fatal("should reject AppendEntries with stale term")
	}
}

func TestProposeCeilingCommandQuorum(t *testing.T) {
	nodes := make([]*RaftNode, 3)
	ids := []string{"node-1", "node-2", "node-3"}
	for i, id := range ids {
		nodes[i] = NewRaftNode(id, ids)
	}

	// Elect node-0 as leader
	nodes[0].CurrentTerm = 1
	nodes[0].BecomeLeader()

	cmd := CeilingCommand{
		Type:     "SET_CEILING",
		AgentID:  "agent-suspicious-001",
		Value:    400,
		AuthorID: "operator-rehan",
	}

	result, err := nodes[0].ProposeCeilingCommand(cmd, nodes[1:])
	if err != nil {
		t.Fatalf("ProposeCeilingCommand error: %v", err)
	}
	if !result.Approved {
		t.Fatalf("expected quorum approval, got %+v", result)
	}
	if result.VotesFor < result.QuorumSize {
		t.Fatalf("votes %d < quorum %d", result.VotesFor, result.QuorumSize)
	}
}

func TestProposeCeilingCommandApplied(t *testing.T) {
	nodes := make([]*RaftNode, 3)
	ids := []string{"node-1", "node-2", "node-3"}
	for i, id := range ids {
		nodes[i] = NewRaftNode(id, ids)
	}
	nodes[0].CurrentTerm = 1
	nodes[0].BecomeLeader()

	cmd := CeilingCommand{
		Type:     "SET_CEILING",
		AgentID:  "agent-001",
		Value:    400,
		AuthorID: "op-1",
	}
	nodes[0].ProposeCeilingCommand(cmd, nodes[1:])

	val, active := nodes[0].GetCeilingState("agent-001")
	if !active {
		t.Fatal("ceiling should be active after SET_CEILING commit")
	}
	if val != 400 {
		t.Fatalf("expected ceiling 400, got %d", val)
	}
}

func TestClearCeilingRequiresCommit(t *testing.T) {
	nodes := make([]*RaftNode, 3)
	ids := []string{"node-1", "node-2", "node-3"}
	for i, id := range ids {
		nodes[i] = NewRaftNode(id, ids)
	}
	nodes[0].CurrentTerm = 1
	nodes[0].BecomeLeader()

	// Set ceiling first
	nodes[0].ProposeCeilingCommand(CeilingCommand{
		Type: "SET_CEILING", AgentID: "agent-001", Value: 400, AuthorID: "op-1",
	}, nodes[1:])

	// Clear ceiling via quorum
	nodes[0].ProposeCeilingCommand(CeilingCommand{
		Type: "CLEAR_CEILING", AgentID: "agent-001", Value: 0, AuthorID: "op-1",
	}, nodes[1:])

	_, active := nodes[0].GetCeilingState("agent-001")
	if active {
		t.Fatal("ceiling should be cleared after CLEAR_CEILING commit")
	}
}

func TestNonLeaderCannotPropose(t *testing.T) {
	node := NewRaftNode("node-1", []string{"node-1", "node-2"})
	// node is FOLLOWER by default
	_, err := node.ProposeCeilingCommand(CeilingCommand{
		Type: "SET_CEILING", AgentID: "agent-001", Value: 400,
	}, nil)
	if err == nil {
		t.Fatal("follower should not be able to propose commands")
	}
}

func TestAuditLogPopulated(t *testing.T) {
	nodes := make([]*RaftNode, 3)
	ids := []string{"node-1", "node-2", "node-3"}
	for i, id := range ids {
		nodes[i] = NewRaftNode(id, ids)
	}
	nodes[0].CurrentTerm = 1
	nodes[0].BecomeLeader()

	nodes[0].ProposeCeilingCommand(CeilingCommand{
		Type: "SET_CEILING", AgentID: "agent-001", Value: 400, AuthorID: "op-1",
	}, nodes[1:])

	log := nodes[0].AuditLog()
	if len(log) == 0 {
		t.Fatal("audit log should contain committed commands")
	}
	if log[0].Type != "SET_CEILING" {
		t.Fatalf("expected SET_CEILING in audit log, got %s", log[0].Type)
	}
}

func TestClusterStatus(t *testing.T) {
	node := NewRaftNode("node-1", []string{"node-1", "node-2", "node-3"})
	status := node.ClusterStatus()
	if status["node_id"] != "node-1" {
		t.Fatal("wrong node_id in status")
	}
	if status["role"] != "FOLLOWER" {
		t.Fatal("expected FOLLOWER role in status")
	}
}

func TestRegulatoryNote(t *testing.T) {
	note := RegulatoryNote()
	if len(note) == 0 {
		t.Fatal("RegulatoryNote must not be empty")
	}
}
