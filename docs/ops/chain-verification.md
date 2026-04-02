# Chain Verification — AgentRepEngine

## Purpose
Verifies the tamper-evident hash chain in the `enforcement_decisions` table.
Customer-runnable. No ARE involvement required.

## Usage
```bash
# Verify all records
./verify-chain

# Verify a date range
./verify-chain --from 2026-04-01 --to 2026-04-30

# Custom database connection
./verify-chain --dsn "postgres://user:pass@host:5432/agentrepengine?sslmode=disable"
```

## Output
```
Chain intact: 14,823 records verified ✓
```
or
```
CHAIN BROKEN at record ID 8821
  Expected prev_hash: abc123...
  Stored   prev_hash: def456...
```

## Exit Codes
- `0` — chain intact
- `1` — database connection error
- `2` — chain broken or hash mismatch

## What It Verifies
1. Each record's `prev_hash` matches the `this_hash` of the preceding record
2. Each record's `this_hash` matches SHA-256 of `id|agent_did|decision|score|prev_hash`
3. Sequential ordering by `id` is preserved

## If the Chain Is Broken
A broken chain means one of:
- A record was modified after insertion
- A record was deleted
- A database migration reordered rows (VACUUM FULL, PITR restore)

Contact AgentRepEngine immediately with the broken record ID and timestamp.

## Migration Safety
Before any PostgreSQL maintenance (VACUUM FULL, ALTER TABLE, PITR restore):
1. Run `./verify-chain` and save output
2. Perform maintenance
3. Run `./verify-chain` again and confirm output matches
4. Document in the rotation log below

## Verification Log
| Date | Range | Records | Result | Who |
|------|-------|---------|--------|-----|
| 2026-04-02 | all | 0 | intact | Rehan Rana |