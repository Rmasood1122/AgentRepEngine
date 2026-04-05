-- Migration 004: Add computed scoring dimensions as queryable columns
-- Additive only — zero impact on existing data or queries
-- Purpose: Enable SQL analytics on individual scoring dimensions
-- Required for: M2 research note, M7 CISO report, Phase 2 ML training

ALTER TABLE scoring_explanations
  ADD COLUMN IF NOT EXISTS history_score      INTEGER,
  ADD COLUMN IF NOT EXISTS velocity_score     INTEGER,
  ADD COLUMN IF NOT EXISTS z_score            DECIMAL(8,4),
  ADD COLUMN IF NOT EXISTS composite_score    INTEGER,
  ADD COLUMN IF NOT EXISTS band               TEXT,
  ADD COLUMN IF NOT EXISTS z_score_feature    TEXT,
  ADD COLUMN IF NOT EXISTS data_points_count  INTEGER;

-- Index the columns that will be queried most in analytics
CREATE INDEX IF NOT EXISTS idx_scoring_explanations_z_score
  ON scoring_explanations(z_score);

CREATE INDEX IF NOT EXISTS idx_scoring_explanations_band
  ON scoring_explanations(band);

CREATE INDEX IF NOT EXISTS idx_scoring_explanations_composite_score
  ON scoring_explanations(composite_score);

COMMENT ON COLUMN scoring_explanations.history_score IS
  'H component of Score = Clamp(0.5*H + 0.5*V, 0, 1000)';

COMMENT ON COLUMN scoring_explanations.velocity_score IS
  'V component of Score = Clamp(0.5*H + 0.5*V, 0, 1000)';

COMMENT ON COLUMN scoring_explanations.z_score IS
  'z = (observed_rate - agent_baseline) / agent_std_dev at time of decision';

COMMENT ON COLUMN scoring_explanations.z_score_feature IS
  'Which feature dimension produced the worst z_score';

COMMENT ON COLUMN scoring_explanations.data_points_count IS
  'Number of historical data points in baseline at time of scoring';