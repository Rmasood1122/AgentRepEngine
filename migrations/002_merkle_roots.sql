-- Migration 002: Merkle audit tree root persistence
-- DORA Article 17 / SOC2 CC7.2

CREATE TABLE IF NOT EXISTS merkle_roots (
    id              BIGSERIAL PRIMARY KEY,
    root_hash       VARCHAR(64)  NOT NULL,
    decision_count  INTEGER      NOT NULL DEFAULT 0,
    window_start    TIMESTAMPTZ  NOT NULL,
    window_end      TIMESTAMPTZ  NOT NULL,
    generated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_merkle_window UNIQUE (window_start, window_end)
);

CREATE INDEX IF NOT EXISTS idx_merkle_roots_window_end ON merkle_roots (window_end DESC);

-- INSERT-only: auditors must not be able to delete roots
REVOKE DELETE ON merkle_roots FROM are;
REVOKE UPDATE ON merkle_roots FROM are;
COMMENT ON TABLE merkle_roots IS 'Tamper-evident Merkle root log — DORA Art.17 / SOC2 CC7.2';