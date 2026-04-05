package attacks

// SlowWalkMicrosoftBypassDemo demonstrates the attack class that
// passes Microsoft's Agent Governance Toolkit (stateless, 0.1ms)
// but is detected and blocked by ARE's behavioral baseline scoring.
//
// Attack profile:
// - Agent conducts PII access rate escalation over 50 calls
// - Each individual call: within normal rate limits (no rule violated)
// - Microsoft AGT result: all 50 calls pass (stateless policy check)
// - ARE result: flagged at call 43 (z_score 4.2 above 30-day baseline)
//              blocked at call 44
//
// Demo runtime: ~30 seconds on Windows Docker
// Output: side-by-side comparison Microsoft result vs ARE result

// TODO: Implement RunDemo() function
// TODO: Simulate stateless policy check (Microsoft behavior)
// TODO: Run against ARE scoring service on localhost:8080
// TODO: Print side-by-side results with reason objects

func RunDemo() {
	// TODO
}
