import re

# ── CHANGE 1: siem.go — add SendQueueBacklog method ──────────────────────────
with open('internal/audit/siem.go', 'r', encoding='utf-8') as f:
    siem = f.read()

# Insert SendQueueBacklog after SendRestricted closing brace
old_siem = """func (s *SIEMWebhook) SendRestricted(
        agentDID string,
        score, scoreDelta int,
        policyFired string,
        reasonJSON []byte,
) {
        go s.send(agentDID, "RESTRICTED", score, scoreDelta, policyFired, reasonJSON, "MEDIUM")
}"""

new_siem = """func (s *SIEMWebhook) SendRestricted(
        agentDID string,
        score, scoreDelta int,
        policyFired string,
        reasonJSON []byte,
) {
        go s.send(agentDID, "RESTRICTED", score, scoreDelta, policyFired, reasonJSON, "MEDIUM")
}

// SendQueueBacklog fires a SIEM alert when the async event queue exceeds
// the backlog threshold. Indicates consumer is falling behind — scores
// may be stale. Operators must investigate before agents are mis-scored.
// Wired from EventConsumer.updateQueueDepth() in consumer.go.
func (s *SIEMWebhook) SendQueueBacklog(depth int) {
        reason := []byte(`{"alert":"event_queue_backlog","action":"investigate_consumer"}`)
        go s.send(
                "system",
                "QUEUE_BACKLOG",
                depth, 0,
                "event_queue_depth_exceeded",
                reason,
                "HIGH",
        )
}"""

if old_siem in siem:
    siem = siem.replace(old_siem, new_siem)
    print("SUCCESS siem.go: SendQueueBacklog added")
else:
    print("ERROR siem.go: SendRestricted block not found — check whitespace")

with open('internal/audit/siem.go', 'w', encoding='utf-8') as f:
    f.write(siem)

# ── CHANGE 2: consumer.go — wire EventQueueDepth + backlog alert ─────────────
with open('internal/scoring/consumer.go', 'r', encoding='utf-8') as f:
    consumer = f.read()

# 2a: Add metrics + audit imports
old_imports = '''import (
        "database/sql"
        "encoding/json"
        "log/slog"
        "math"
        "time"
)'''

new_imports = '''import (
        "database/sql"
        "encoding/json"
        "log/slog"
        "math"
        "time"

        "github.com/Rehanrana11/AgentRepEngine/internal/audit"
        "github.com/Rehanrana11/AgentRepEngine/internal/metrics"
)'''

if old_imports in consumer:
    consumer = consumer.replace(old_imports, new_imports)
    print("SUCCESS consumer.go: imports updated")
else:
    print("ERROR consumer.go: imports block not found")

# 2b: Add siem field to EventConsumer struct
old_struct = '''// EventConsumer reads from agent_event_queue and updates scores.
type EventConsumer struct {
        db          *sql.DB
        scoreWriter ScoreWriter
        batchSize   int
        interval    time.Duration
}'''

new_struct = '''// EventConsumer reads from agent_event_queue and updates scores.
type EventConsumer struct {
        db          *sql.DB
        scoreWriter ScoreWriter
        siem        *audit.SIEMWebhook
        batchSize   int
        interval    time.Duration
}

// QueueBacklogThreshold is the unprocessed event count that triggers
// a SIEM alert. At 1000 unprocessed events the consumer is materially
// behind — agent scores may not reflect recent behavioral events.
const QueueBacklogThreshold = 1000'''

if old_struct in consumer:
    consumer = consumer.replace(old_struct, new_struct)
    print("SUCCESS consumer.go: struct updated with siem field")
else:
    print("ERROR consumer.go: struct block not found")

# 2c: Wire siem in constructor
old_constructor = '''func NewEventConsumer(db *sql.DB, sw ScoreWriter) *EventConsumer {
        return &EventConsumer{
                db:          db,
                scoreWriter: sw,
                batchSize:   100,
                interval:    5 * time.Second,
        }
}'''

new_constructor = '''func NewEventConsumer(db *sql.DB, sw ScoreWriter) *EventConsumer {
        return &EventConsumer{
                db:          db,
                scoreWriter: sw,
                siem:        audit.NewSIEMWebhook(),
                batchSize:   100,
                interval:    5 * time.Second,
        }
}'''

if old_constructor in consumer:
    consumer = consumer.replace(old_constructor, new_constructor)
    print("SUCCESS consumer.go: constructor wired with siem")
else:
    print("ERROR consumer.go: constructor block not found")

# 2d: Add updateQueueDepth call after processBatch in Start()
old_start_loop = '''        for {
                processed, err := c.processBatch()
                if err != nil {
                        slog.Error("consumer batch error", "error", err)
                } else if processed > 0 {
                        slog.Info("consumer batch processed", "count", processed)
                }
                time.Sleep(c.interval)
        }'''

new_start_loop = '''        for {
                processed, err := c.processBatch()
                if err != nil {
                        slog.Error("consumer batch error", "error", err)
                } else if processed > 0 {
                        slog.Info("consumer batch processed", "count", processed)
                }
                c.updateQueueDepth()
                time.Sleep(c.interval)
        }'''

if old_start_loop in consumer:
    consumer = consumer.replace(old_start_loop, new_start_loop)
    print("SUCCESS consumer.go: updateQueueDepth call wired in Start()")
else:
    print("ERROR consumer.go: Start() loop not found")

# 2e: Add updateQueueDepth method before processBatch
old_processbatch = '''func (c *EventConsumer) processBatch() (int, error) {'''

new_processbatch = '''// updateQueueDepth queries the unprocessed event count, updates the
// Prometheus gauge, and fires a SIEM alert if the backlog exceeds
// QueueBacklogThreshold. Called after every consumer batch cycle.
// Addresses FMEA RPN-210: async pipeline backlog → stale agent scores.
func (c *EventConsumer) updateQueueDepth() {
        var depth int
        err := c.db.QueryRow(`
                SELECT COUNT(*)
                FROM agent_event_queue
                WHERE processed = false`).Scan(&depth)
        if err != nil {
                slog.Warn("queue_depth_check_failed", "error", err)
                return
        }
        metrics.EventQueueDepth.Set(float64(depth))
        slog.Debug("queue_depth_updated", "depth", depth)
        if depth > QueueBacklogThreshold {
                slog.Warn("event_queue_backlog_alert",
                        "depth", depth,
                        "threshold", QueueBacklogThreshold,
                        "action", "investigate_consumer_lag",
                )
                c.siem.SendQueueBacklog(depth)
        }
}

func (c *EventConsumer) processBatch() (int, error) {'''

if old_processbatch in consumer:
    consumer = consumer.replace(old_processbatch, new_processbatch)
    print("SUCCESS consumer.go: updateQueueDepth method added")
else:
    print("ERROR consumer.go: processBatch func not found")

with open('internal/scoring/consumer.go', 'w', encoding='utf-8') as f:
    f.write(consumer)

print("\nAll changes applied. Run: go build ./... to verify.")
