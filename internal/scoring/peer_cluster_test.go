package scoring

import "testing"

func TestPeerClusterScoreWithSmallCluster(t *testing.T) {
	t.Skip("TODO: implement — confidence weight should be 0.1 for <3 enterprises")
}

func TestPeerClusterScoreWithLargeCluster(t *testing.T) {
	t.Skip("TODO: implement — confidence weight should be 0.7 for >10 enterprises")
}

func TestPeerClusterDeviationDetectsBaselineEvasion(t *testing.T) {
	t.Skip("TODO: implement — attacker calibrated to individual baseline should fail peer check")
}
