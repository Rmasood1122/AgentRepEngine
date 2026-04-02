# AgentRepEngine — Python SDK

Runtime behavioral enforcement for LangChain and LangGraph agents.

## What this does

Every tool call, LLM invocation, and node execution your agent makes
is reported to AgentRepEngine for real-time behavioral scoring.
ARE builds a 30-day behavioral baseline per agent and flags anomalies
before they complete.

**Zero latency impact** — events are sent asynchronously.
**Fail open** — if ARE is unavailable, your agent keeps running.
**No code changes** to your agent logic — just add the callback/middleware.

---

## Installation
```bash
pip install requests langchain langchain-core langgraph
```

---

## LangChain — Quick Start
```python
from sdk.langchain.are_callback import ARECallbackHandler

handler = ARECallbackHandler(
    agent_did="did:are:my-agent-001",
    org_id="my-org",
    are_url="http://localhost:8080",
    api_key="your-api-key",
)

from langchain.agents import initialize_agent, AgentType
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(model="gpt-4o")
agent = initialize_agent(
    tools=tools,
    llm=llm,
    agent=AgentType.OPENAI_FUNCTIONS,
    callbacks=[handler],
)

result = agent.run("Search for the latest AI security papers")

score = handler.get_current_score()
print(f"Agent score: {score['score']} ({score['band']})")
# Agent score: 847 (TRUSTED)
```

---

## LangGraph — Quick Start

### Option A: Wrap individual nodes
```python
from sdk.langgraph.are_middleware import AREMiddleware
from langgraph.graph import StateGraph, END
from typing import TypedDict, List

class AgentState(TypedDict):
    messages: List[str]
    result: str

are = AREMiddleware(
    agent_did="did:are:langgraph-agent-001",
    org_id="my-org",
    are_url="http://localhost:8080",
    api_key="your-api-key",
)

def call_llm(state: AgentState) -> AgentState:
    return {**state, "result": "llm output"}

def call_tool(state: AgentState) -> AgentState:
    return {**state, "result": "tool output"}

builder = StateGraph(AgentState)
builder.add_node("llm", are.wrap_node("llm", call_llm))
builder.add_node("tool", are.wrap_node("tool", call_tool))
builder.set_entry_point("llm")
builder.add_edge("llm", "tool")
builder.add_edge("tool", END)

graph = builder.compile()
result = graph.invoke({"messages": [], "result": ""})

print(are.session_summary())
```

### Option B: Wrap compiled graph
```python
graph = builder.compile()
graph = are.wrap_graph(graph)
result = graph.invoke({"messages": [], "result": ""})
```

---

## Configuration

| Parameter | Default | Description |
|-----------|---------|-------------|
| `agent_did` | required | Unique agent identity |
| `org_id` | required | Organisation ID for baseline isolation |
| `are_url` | `http://localhost:8080` | ARE scoring service URL |
| `api_key` | `None` | X-API-Key header |
| `privacy_tier` | `1` | 1=full, 2=aggregated, 3=anonymised |
| `timeout` | `2.0` | HTTP timeout seconds |
| `async_send` | `True` | Background thread sending |
| `cycle_threshold` | `10` | Node visits before cycle alert (LangGraph) |

---

## Feature vector — what IS scored

| Feature | Source |
|---------|--------|
| `tool_call_rate_per_hour` | Rolling tool/node call count |
| `unique_endpoints_per_hour` | Distinct tools/nodes accessed |
| `bulk_access_count_per_session` | Large output reads >10KB |
| `pii_field_access_rate` | Fraction of events touching PII tools |
| `cross_tenant_probe_count` | Cross-org access attempts |
| `permission_escalation_count` | Admin/root/privilege tool calls |
| `sub_agent_spawn_depth` | Max nested agent/chain depth |
| `token_refresh_rate` | LLM re-invocation rate |

---

## Running tests
```bash
pip install pytest requests --break-system-packages

# LangChain tests
python -m pytest sdk/langchain/are_callback_test.py -v

# LangGraph tests
python -m pytest sdk/langgraph/are_middleware_test.py -v

# All SDK tests
python -m pytest sdk/ -v
```

No running ARE instance needed — tests use a mock HTTP server.

---

## Fail-open guarantee

ARE **never** blocks your agent. Enforcement happens at the
Kong gateway layer — not in the SDK. The SDK is instrumentation only.

---

*AgentRepEngine · Naseem A2A Research Lab*
*DOI: 10.5281/zenodo.19169185*