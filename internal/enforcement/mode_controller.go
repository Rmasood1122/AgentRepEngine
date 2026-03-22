package enforcement

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	ModeKey         = "config:enforcement_mode"
	ModeObserve     = "observe"
	ModeEnforce     = "enforce"
	FPRateThreshold = 2.0
	FPCheckInterval = 5 * time.Minute
	FPWindowHours   = 1
)

// SIEMNotifier sends alerts when auto-rollback fires.
type SIEMNotifier interface {
	SendAlert(event, message string)
}

// ModeController manages enforcement mode at runtime via Redis.
// Changes take effect immediately — no service restart required.
type ModeController struct {
	rdb  *redis.Client
	db   *sql.DB
	siem SIEMNotifier
	ctx  context.Context
}

// NewModeController creates a ModeController and seeds initial mode.
func NewModeController(rdb *redis.Client, db *sql.DB, initialMode string) *ModeController {
	mc := &ModeController{
		rdb:  rdb,
		db:   db,
		siem: nil,
		ctx:  context.Background(),
	}
	if err := mc.setMode(initialMode); err != nil {
		slog.Error("mode_controller_init_failed", "error", err)
	}
	slog.Info("mode_controller_initialized", "mode", initialMode)
	return mc
}

// GetMode returns current enforcement mode from Redis.
// Falls back to observe if Redis unavailable — fail-safe.
func (mc *ModeController) GetMode() string {
	mode, err := mc.rdb.Get(mc.ctx, ModeKey).Result()
	if err != nil {
		slog.Warn("mode_read_failed_defaulting_observe", "error", err)
		return ModeObserve
	}
	return mode
}

// SetEnforce switches to enforce mode.
// Requires explicit human action — never called automatically.
func (mc *ModeController) SetEnforce() error {
	if err := mc.setMode(ModeEnforce); err != nil {
		return err
	}
	slog.Info("enforcement_mode_changed",
		"from", ModeObserve,
		"to", ModeEnforce,
		"trigger", "manual")
	return nil
}

// IsEnforcing returns true if currently in enforce mode.
func (mc *ModeController) IsEnforcing() bool {
	return mc.GetMode() == ModeEnforce
}

// setMode writes mode to Redis.
func (mc *ModeController) setMode(mode string) error {
	return mc.rdb.Set(mc.ctx, ModeKey, mode, 0).Err()
}

// StartFPMonitor runs the auto-rollback loop in a goroutine.
// Checks FP rate every FPCheckInterval.
// Auto-rolls back to observe if FP rate exceeds threshold.
// Never auto-escalates to enforce — human decision only.
func (mc *ModeController) StartFPMonitor() {
	go func() {
		ticker := time.NewTicker(FPCheckInterval)
		defer ticker.Stop()
		for range ticker.C {
			mc.checkAndRollbackIfNeeded()
		}
	}()
	slog.Info("fp_monitor_started",
		"interval", FPCheckInterval,
		"threshold_pct", FPRateThreshold,
		"window_hours", FPWindowHours)
}

// checkAndRollbackIfNeeded measures FP rate and rolls back if needed.
func (mc *ModeController) checkAndRollbackIfNeeded() {
	currentMode := mc.GetMode()
	if currentMode != ModeEnforce {
		return
	}

	fpRate, err := mc.measureFPRate()
	if err != nil {
		slog.Error("fp_rate_measurement_failed", "error", err)
		return
	}

	if fpRate < 0 {
		return
	}

	slog.Info("fp_rate_check",
		"fp_rate_pct", fpRate,
		"threshold_pct", FPRateThreshold,
		"mode", currentMode)

	if fpRate > FPRateThreshold {
		mc.rollbackToObserve(fpRate)
	}
}

// measureFPRate queries FP rate over the last hour.
// Returns -1 if no enforcement decisions exist in window.
func (mc *ModeController) measureFPRate() (float64, error) {
	var fpRate sql.NullFloat64

	err := mc.db.QueryRowContext(mc.ctx, `
		SELECT
			ROUND(100.0 *
				COUNT(*) FILTER (WHERE override = true)
				/ NULLIF(COUNT(*) FILTER (WHERE decision = 'BLOCKED'), 0),
			2) AS fp_rate
		FROM enforcement_decisions
		WHERE created_at > NOW() - INTERVAL '1 hour'
	`).Scan(&fpRate)

	if err != nil {
		return 0, fmt.Errorf("fp_rate_query: %w", err)
	}

	if !fpRate.Valid {
		return -1, nil
	}

	return fpRate.Float64, nil
}

// rollbackToObserve switches to observe and fires SIEM alert.
func (mc *ModeController) rollbackToObserve(fpRate float64) {
	if err := mc.setMode(ModeObserve); err != nil {
		slog.Error("auto_rollback_failed",
			"error", err,
			"fp_rate_pct", fpRate)
		return
	}

	slog.Warn("auto_rollback_triggered",
		"reason", "fp_rate_exceeded_threshold",
		"fp_rate_pct", fpRate,
		"threshold_pct", FPRateThreshold,
		"action", "rolled_back_to_observe",
		"re_enable", "manual_action_required")

	if mc.siem != nil {
		mc.siem.SendAlert(
			"auto_rollback",
			fmt.Sprintf(
				"AgentRepEngine auto-rollback: FP rate %.2f%% exceeded %.2f%% threshold. "+
					"Enforcement set to OBSERVE. Manual re-enable required.",
				fpRate, FPRateThreshold,
			),
		)
	}
}
