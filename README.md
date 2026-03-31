# AgentRepEngine

**Runtime behavioral trust and enforcement infrastructure for AI agents in regulated enterprises.**

AgentRepEngine's behavioral attention engine scores every AI agent request in
real time, detects statistical behavioral deviation and adversarial baseline
poisoning, and stops risky actions before data leaves — at the gateway, with
cryptographic proof.

[![DOI](https://zenodo.org/badge/DOI/10.5281/zenodo.19169185.svg)](https://doi.org/10.5281/zenodo.19169185)

---

## What it does

An AI agent can look clean for days — then start bulk-extracting PII.
AgentRepEngine catches it before the first record leaves.

- **Behavioral attention engine** — 8-dimension weighted feature scoring with
  per-org z-score statistical deviation detection, 0.25ns overhead
- **Gateway-level behavioral enforcement** — Kong plugin, RS256 JWT identity,
  10ms p99 hard ceiling, context-sensitive thresholds by agent type
- **Cryptographically non-repudiable enforcement log** — SHA-256 cryptographically non-repudiable enforcement log,
  INSERT-only at PostgreSQL permission level, tamper-evident by architecture
- **SIEM feed** — structured CEF/JSON reason object on every enforcement
  decision, delivered within 500ms. Splunk and Sentinel compatible.
- **Adversarial baseline poisoning defense** — 100% detection rate on 10
  slow-walk synthetic scenarios. Novel attack class, two-layer defense.
- **Risk-staged deployment** — zero-impact visibility → flagging → enforcement.
  You control the pace. We never advance without your sign-off.
- **Python SDK** — LangChain integration, 3 lines of code

## Live detection proof

```bash
docker compose up -d
bash scripts/synthetic-agent-demo.sh
```

Agent scores 850 TRUSTED for normal behavior. Bulk PII extraction begins.
Detected at request 15. Score drops to 400 RESTRICTED. Structured reason
object shipped to SIEM. Cryptographically non-repudiable enforcement log
verified. Detection proof runtime: ~60 seconds streaming.

## Install

```bash
git clone https://github.com/Rehanrana11/AgentRepEngine
cd AgentRepEngine
docker compose up -d
```

Prerequisites: Docker, Kong Gateway 3.6.x, Redis 7+, PostgreSQL 16+.
Installation time: under 4 hours, customer-hosted, no external dependencies.

## Stack

| Component | Technology |
|-----------|-----------|
| Behavioral attention engine | Go |
| Gateway plugin | Kong (Lua) |
| Score cache | Redis |
| Event queue | PostgreSQL async |
| Identity | JWT + RS256 + JWKS |
| Statistical deviation detection | Velocity + z-score (Welford's online) |
| Adversarial baseline poisoning defense | Variance growth rate monitoring |
| SDK | Python |

## Performance

| Metric | Value |
|--------|-------|
| Call-level overhead | 0.25ns (Linux), 0 B/op allocations |
| Score computation p99 | 10ms hard ceiling |
| False positive rate | 0.00% on 100-scenario held-out validation corpus |
| True positive rate | 86.67% |
| F1 score | 0.9286 |
| Precision | 100% (zero legitimate agents blocked) |
| Adversarial baseline poisoning detection | 100% (10/10 slow-walk scenarios) |
| Detection proof runtime | ~60 seconds streaming |

## Regulatory alignment

| Regulation | ARE Coverage |
|---|---|
| DORA Article 17 | Cryptographically non-repudiable enforcement log |
| DORA Article 28 | AI agent behavioral scoring + policy enforcement |
| SEC AI governance | Hash-chained audit trail + structured reason objects |
| HIPAA | No PHI stored. Behavioral metadata only. |
| GDPR Article 22 | Human-in-the-loop enforcement validation at every stage |
| SOC2 CC7.2 | System monitoring with Prometheus + SIEM integration |

## Ensemble enforcement — why 0.00% false positives

ARE requires two independent detection systems to agree before any agent is
blocked: the behavioral attention engine (score-based) AND OWASP LLM Top 10
enforcement policies (rule-based). A single miscalibrated metric cannot block
a legitimate agent. This is the architectural reason ARE achieves 0.00% false
positive rate on its held-out validation corpus.

## Adversarial baseline poisoning (novel attack class)

A compromised agent that gradually escalates its activity rate over 7-14 days
can shift a behavioral scoring system's baseline, making anomalous behavior
appear normal. ARE defends against this with two independent layers:
1. OWASP enforcement policies catch high-value actions regardless of score
2. Variance growth rate monitoring flags the baseline shift as it begins

100% detection on synthetic finserv slow-walk corpus. No other AI agent
security system addresses this attack class.

## Research foundation

AgentRepEngine implements the
[Agentic Trust Protocol (ATP)](https://doi.org/10.5281/zenodo.17917568)
and
[Agentic Trust Graph (ATG)](https://doi.org/10.5281/zenodo.17917659)
published by Rehan Masood, Naseem A2A Research Lab, December 2025.

**AgentRepEngine DOI:** https://doi.org/10.5281/zenodo.19169185

## Enterprise documents

| Document | Purpose |
|---|---|
| [Operational Safety Architecture](docs/enterprise/operational-safety-architecture.md) | How failure is handled, risk-staged deployment, HITL design |
| [Pilot Letter of Understanding](docs/enterprise/pilot-letter-of-understanding.md) | 30-day zero-risk visibility deployment agreement |
| [Competitive Positioning](docs/enterprise/competitive-positioning.md) | Lakera, Palo Alto, SIEM — what ARE does that they don't |
| [Product Maturity Statement](docs/enterprise/honest-maturity-statement.md) | What is production-ready, what is not |
| [GDPR Position](docs/enterprise/gdpr-position.md) | Data subject rights, tombstone procedure |
| [Prerequisites Checklist](docs/enterprise/prerequisites-checklist.md) | Install requirements, JSON schema, SIEM integration |

## License

© 2026 Rehan Masood. All Rights Reserved.
See LICENSE for details.
