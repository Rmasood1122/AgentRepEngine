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

// SnapshotBaselines copies current std_dev values from agent_baselines
// into agent_baseline_snapshots (one row per org/agent/feature per day).
// Called daily by the retention job.
//
// This snapshot is the historical reference used by CheckVarianceGrowthRate()
// in internal/scoring/policy.go to detect slow-walk baseline poisoning.
// Without daily snapshots, variance growth detection is silent — the LEFT JOIN
// in CheckVarianceGrowthRate always returns NULL and all checks are skipped.
//
// INSERT ... ON CONFLICT DO NOTHING ensures idempotency:
// running twice in one day produces exactly one snapshot row per feature.
func SnapshotBaselines(ctx context.Context, db *sql.DB) (int64, error) {
	result, err := db.ExecContext(ctx, `
		INSERT INTO agent_baseline_snapshots
			(org_id, agent_did, feature_name, std_dev, sample_count, snapshot_date)
		SELECT
			org_id,
			agent_did,
			feature_name,
			std_dev,
			sample_count,
			NOW()
		FROM agent_baselines
		WHERE sample_count >= 100
		ON CONFLICT (org_id, agent_did, feature_name, DATE(snapshot_date))
		DO NOTHING
	`)
	if err != nil {
		return 0, fmt.Errorf("snapshot baselines: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows, nil
}

// StartRetentionJob runs PurgeOldEvents and SnapshotBaselines every 24 hours.
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

	// Snapshot current baselines for variance growth rate detection.
	// Required by CheckVarianceGrowthRate() in internal/scoring/policy.go.
	snapped, err := SnapshotBaselines(ctx, db)
	if err != nil {
		log.Printf("snapshot_baselines_error: %v", err)
	} else {
		log.Printf("snapshot_baselines_complete: inserted=%d new baseline snapshots", snapped)
	}
}
