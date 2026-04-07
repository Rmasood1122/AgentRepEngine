# Formal Verification — AgentRepEngine

## Files

| File | Tool | Purpose |
|---|---|---|
| `sir_recovery.smv` | NuSMV 2.6+ | SIR state machine safety + liveness |
| `ceiling_invariant.tla` | TLA+ / TLC | Ceiling score invariant (next) |

## Running NuSMV
```bash
nusmv internal/formal/sir_recovery.smv
```

All `LTLSPEC` lines must return:
specification ... is true
## Properties Verified

### Safety (S1–S5)
- S1: TRUSTED → ISOLATED transition is impossible (must pass SUSPICIOUS)
- S2: ISOLATED requires human_clear to exit — no automatic escape
- S3: FP rate exceeded never escalates enforcement
- S4: RECOVERING cannot jump to ISOLATED without anomaly
- S5: LOW score while TRUSTED always triggers transition

### Liveness (L1–L3)
- L1: ISOLATED always eventually reaches TRUSTED (not a permanent trap)
- L2: SUSPICIOUS always eventually resolves
- L3: RECOVERING always eventually exits

### Compliance (C1–C3)
- C1: DORA Article 17 — isolation always preceded by classification
- C2: SOC2 CC7.2 — human oversight required for isolation exit
- C3: NIST AI RMF — auto-rollback protects TRUSTED agents

## Regulatory Mapping

| Property | Regulation | Article/Control |
|---|---|---|
| S2 (human clear required) | DORA | Article 17 |
| C2 (human oversight) | SOC2 | CC7.2 |
| C3 (auto-rollback) | NIST AI RMF | MS-2.5 |
| L1 (no permanent isolation) | GDPR | Article 22 (automated decisions) |