# Enforce Mode Activation Gate — AgentRepEngine

## Rule
Enforce mode MUST NOT be activated until ALL conditions below are confirmed.
This is a hard rule. No exceptions during the pilot period.

## Activation Checklist
[ ] 14 consecutive days of observe mode with zero false positives
[ ] All flagged events reviewed per-action with customer security team present
[ ] Customer written sign-off on enforcement activation (email or signed doc)
[ ] Redis health check passing: `GET /health → redis: "ok"`
[ ] Auto-rollback threshold confirmed in config: `ENFORCEMENT_MODE=observe` default
[ ] Kong plugin fail-open behavior confirmed: scoring service 5xx → pass through

## Why This Rule Exists
One false positive on a critical workflow (payment, compliance report, database
operation) during enforce mode destroys pilot trust permanently. Auto-rollback
requires multiple FPs to trigger — by then the damage is done.

Observe mode is not a limitation. It is the proof that ARE works without risk.

## What Happens If a FP Occurs in Enforce Mode
1. Scoring service auto-rolls back to observe mode if FP rate exceeds 2%
2. Customer notified immediately
3. Incident documented in enforcement_decisions with override=true
4. Root cause identified before enforce mode re-activated
5. Customer sign-off required again before re-activation

## Pilot Agreement Reference
This gate is documented in the pilot scope agreement signed before Day 1.
Lloyd Lemish / NWN security lead sign-off required before enforce activation.

## Command to Check Current Mode
```bash
curl http://localhost:8080/health | grep enforcement_mode
```

## Command to Activate Enforce Mode (after all gates passed)
```bash
# Only run after all checklist items confirmed in writing
curl -X POST http://localhost:8080/enforcement/override \
  -H "X-API-Key: $SCORING_API_KEY" \
  -d '{"mode": "enforce", "reason": "14-day clean observe complete, customer signed off"}'
```