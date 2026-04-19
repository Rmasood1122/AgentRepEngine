-- Migration 008: Zero-argument wrapper for verify_hash_chain()
-- Original function takes table_name TEXT, returns BOOLEAN.
-- This wrapper defaults to 'enforcement_decisions' for C13 demo claim.

CREATE OR REPLACE FUNCTION verify_hash_chain()
RETURNS boolean AS $$
BEGIN
    RETURN verify_hash_chain('enforcement_decisions');
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMENT ON FUNCTION verify_hash_chain() IS
    'Zero-argument wrapper — defaults to enforcement_decisions table. C13 demo artifact.';
