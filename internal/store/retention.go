package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

const (
	// RetentionDays is how long processed events are kept in the queue.
	// Unprocessed events are never deleted automatically.
	RetentionDays = 90

	// QueueDepthWarning triggers a log warning when queue exceeds this size.
	QueueDepthWarning = 100_000
)

// QueueDepth returns the current number of rows in agent_event_queue.
func QueueDepth(ctx context.Context, db *sql.DB) (int64, error) {
	var count int64
	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM agent_event_queue`,
	).Scan(&count)
	return count, err
}

// PurgeOldEvents deletes processed events older than RetentionDays.
// Unprocessed events are never purged — they must be handled by the DLQ.
// Returns the number of rows deleted.
func PurgeOldEvents(ctx context.Context, db *sql.DB) (int64, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -RetentionDays)
	result, err := db.ExecContext(ctx, `
		DELETE FROM agent_event_queue
		WHERE processed = true
		AND created_at < $1
	`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("purge failed: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}

// StartRetentionJob runs PurgeOldEvents every 24 hours in the background.
// Call once at startup. Logs results and queue depth on each run.
func StartRetentionJob(db *sql.DB) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		// Run once immediately at startup
		runRetention(db)
		for range ticker.C {
			runRetention(db)
		}
	}()
}

func runRetention(db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	deleted, err := PurgeOldEvents(ctx, db)
	if err != nil {
		log.Printf("retention_error: %v", err)
	} else {
		log.Printf("retention_complete: deleted=%d processed events older than %d days",
			deleted, RetentionDays)
	}

	depth, err := QueueDepth(ctx, db)
	if err != nil {
		log.Printf("queue_depth_error: %v", err)
		return
	}

	if depth >= QueueDepthWarning {
		log.Printf("QUEUE_DEPTH_WARNING: %d rows in agent_event_queue — exceeds %d threshold. "+
			"Check processing pipeline or increase retention purge frequency.",
			depth, QueueDepthWarning)
	} else {
		log.Printf("queue_depth_ok: %d rows", depth)
	}
}
