-- migrations/003_baseline_observations.sql
-- PURPOSE: Append-only baseline store — eliminates Welford race condition
-- SAFETY:  New table only. No existing tables modified.
--          INSERT has no read-modify-write dependency = zero race condition.
--          Redis still caches computed baseline (mean, stddev).
--          Only the write pattern changes — hot path is unchanged.
-- DATE:    2026-04-16

CREATE TABLE IF NOT EXISTS baseline_observations (
  id           BIGSERIAL PRIMARY KEY,
  org_id       TEXT             NOT NULL,
  agent_did    TEXT             NOT NULL,
  dimension    TEXT             NOT NULL,
  observed_val DOUBLE PRECISION NOT NULL,
  event_ts     TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_baseline_obs_lookup
  ON baseline_observations(org_id, agent_did, dimension, event_ts DESC);

CREATE INDEX IF NOT EXISTS idx_baseline_obs_retention
  ON baseline_observations(event_ts);

GRANT INSERT, SELECT ON baseline_observations TO are;
GRANT USAGE, SELECT ON SEQUENCE baseline_observations_id_seq TO are;
