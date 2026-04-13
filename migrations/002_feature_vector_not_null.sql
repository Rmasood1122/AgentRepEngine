-- C-6: NOT NULL on agent_events.feature_vector
-- Applied directly April 13, 2026
BEGIN;
UPDATE agent_events SET feature_vector = '{}'::jsonb WHERE feature_vector IS NULL;
ALTER TABLE agent_events ALTER COLUMN feature_vector SET NOT NULL;
ALTER TABLE agent_events ALTER COLUMN feature_vector SET DEFAULT '{}'::jsonb;
COMMIT;
