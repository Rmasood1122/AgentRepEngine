# AgentRepEngine

**Runtime behavioral enforcement for AI agents in regulated enterprises.**

AgentRepEngine scores every AI agent request in real time, detects anomalous
behavior, and blocks risky actions before data leaves — at the gateway.

## What it does

An AI agent can look clean for days — then start bulk-extracting PII.
AgentRepEngine catches it before the first record leaves.

- **Behavioral scoring** — every agent request scored via velocity + z-score
- **Gateway enforcement** — Kong plugin, RS256 JWT identity, p99 <10ms
- **Tamper-evident audit trail** — hash chain, INSERT-only enforcement log
- **SIEM integration** — structured reason object on every blocked decision
- **Slow-walk detection** — 100% detection rate on 7-day distributed attacks
- **Python SDK** — LangChain integration, 3 lines of code

## Quick demo
```bash
docker compose up -d
bash scripts/synthetic-agent-demo.sh
```

Agent scores 850 (TRUSTED) for 3 days. Bulk PII extraction begins day 4.
Detected at request 15. Score drops to 400 (RESTRICTED). Reason object
shipped to SIEM. Audit trail tamper-evident. Demo runtime: 12 seconds.

## Install
```bash
git clone https://github.com/Rehanrana11/AgentRepEngine
cd AgentRepEngine
docker compose up -d
```

Prerequisites: Docker, Kong gateway, Redis, PostgreSQL.
Install time: under 4 hours.

## Stack

| Component | Technology |
|-----------|-----------|
| Scoring service | Go |
| Gateway plugin | Kong (Lua) |
| Score cache | Redis |
| Event queue | PostgreSQL async |
| Identity | JWT + RS256 + JWKS |
| Anomaly detection | Velocity + z-score |
| SDK | Python |

## Performance

| Metric | Value |
|--------|-------|
| Call-level overhead | 0.25ns (Linux) |
| Memory allocations | 0 B/op |
| FP rate | 0.00% on 100-scenario corpus |
| Slow-walk detection | 100% (10/10 scenarios) |
| Demo runtime | 12 seconds |

## Research foundation

AgentRepEngine implements the
[Agentic Trust Protocol (ATP)](https://doi.org/10.5281/zenodo.17917568)
and
[Agentic Trust Graph (ATG)](https://doi.org/10.5281/zenodo.17917659)
published by Rehan Masood, Naseem A2A Research Lab, December 2025.

## License

© 2026 Rehan Masood. All Rights Reserved.
See LICENSE for details.