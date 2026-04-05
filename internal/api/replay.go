package api

// POST /v1/replay
// Accepts: sequence of historical agent events (by event IDs or time range)
// Returns: score ARE would have produced at each step
//
// Proves: scores are deterministic and reproducible.
// Serves: legal/compliance teams reconstructing incident timelines.
// Answers: technical reviewer asking "how do I verify your score is correct?"
//
// This is the technical trust proof. Given the same sequence of events,
// ARE produces the same score every time. The reviewer can verify this
// by running replay twice on the same event sequence.

// TODO: Implement ReplayEvents(eventIDs []string, orgID string) []ReplayResult
// TODO: Load events from PostgreSQL audit table in strict chronological order
// TODO: Reconstruct Welford state at each step
// TODO: Return score + reason object for each event in sequence
