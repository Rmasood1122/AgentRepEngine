# AgentRepEngine Python SDK

Runtime behavioral telemetry for AI agents. 3 lines to integrate.

## Install
```bash
pip install requests
```

## Quickstart
```python
from agentrepengine import AgentRepEngine

are = AgentRepEngine(
    endpoint="http://your-scoring-service:8080",
    api_key="your-api-key",
    agent_did="did:jwt:your-org:your-agent:001"
)

# Decorator — automatic tracking
@are.track
def search_database(query):
    return db.execute(query)

# Manual PII access tracking
are.record_pii_access()

# Manual emit
are.emit_event("bulk_pii_access", {
    "pii_field_access_rate": 0.85,
    "tool_call_rate_per_hour": 450
})
```

## What it tracks automatically

| Metric | How |
|--------|-----|
| Tool call rate | `@are.track` decorator |
| Unique endpoints | Per function name |
| PII access rate | `are.record_pii_access()` |
| Permission escalations | `are.record_permission_escalation()` |
| Sub-agent spawns | `are.record_sub_agent_spawn()` |

## Design principles

- **Non-blocking** — all emits are background threads
- **Fail silent** — network errors never block your agent
- **Privacy Tier 1** — only metadata, no content
- **Zero dependencies** — only `requests`