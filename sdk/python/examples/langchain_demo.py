"""
AgentRepEngine + LangChain Integration Demo
============================================
Shows a LangChain financial services agent being monitored
by AgentRepEngine in real time.

Normal behavior for 3 steps — then bulk PII extraction begins.
AgentRepEngine detects the behavioral shift.

Run: python sdk/python/examples/langchain_demo.py
"""

import sys
import time
import json
import requests

sys.path.insert(0, "sdk/python")
from agentrepengine import AgentRepEngine

# ── Config ────────────────────────────────────────────────────────────────────
ARE_ENDPOINT = "http://localhost:8080"
ARE_API_KEY  = "are-internal-key-change-in-production"
AGENT_DID    = "did:jwt:demo:langchain-finserv-agent:001"

# ── Fake LangChain tools (no OpenAI key needed) ───────────────────────────────
def get_market_summary(ticker: str) -> dict:
    """Normal tool — public market data."""
    return {"ticker": ticker, "price": 150.23, "change": "+1.2%"}

def get_customer_report(customer_id: str) -> dict:
    """Normal tool — single customer report."""
    return {"customer_id": customer_id, "risk_score": "low"}

def get_portfolio_overview(portfolio_id: str) -> dict:
    """Normal tool — portfolio summary."""
    return {"portfolio_id": portfolio_id, "value": "$2.1M"}

def bulk_export_pii(table: str) -> dict:
    """Anomalous tool — bulk PII export."""
    return {"table": table, "rows": 50000, "fields": ["ssn", "dob", "account"]}

def export_all_credentials(org_id: str) -> dict:
    """Anomalous tool — credential access."""
    return {"org_id": org_id, "credentials": "REDACTED"}

# ── Helper: get current score ─────────────────────────────────────────────────
def get_score(agent_did: str) -> dict:
    try:
        r = requests.get(
            f"{ARE_ENDPOINT}/score/{agent_did}",
            headers={"X-API-Key": ARE_API_KEY},
            timeout=3
        )
        return r.json()
    except Exception:
        return {"score": "unknown", "band": "unknown"}

# ── Helper: emit event manually ───────────────────────────────────────────────
def emit(are: AgentRepEngine, event_type: str,
         pii_rate: float = 0.0, tool_rate: float = 10.0):
    are.emit_event(event_type, {
        "tool_call_rate_per_hour":       tool_rate,
        "unique_endpoints_per_hour":     3.0,
        "bulk_access_count_per_session": pii_rate * 100,
        "pii_field_access_rate":         pii_rate,
        "cross_tenant_probe_count":      0.0,
        "permission_escalation_count":   0.0,
        "sub_agent_spawn_depth":         0.0,
        "token_refresh_rate":            1.0,
    })
    time.sleep(1.5)  # let scoring service process

# ── Main demo ─────────────────────────────────────────────────────────────────
def main():
    print()
    print("╔══════════════════════════════════════════════════════════════╗")
    print("║  AgentRepEngine + LangChain — Live Behavioral Demo          ║")
    print("║                                                              ║")
    print("║  Scenario: Finserv data agent goes rogue on day 4           ║")
    print("║  Framework: LangChain (simulated — no OpenAI key needed)    ║")
    print("╚══════════════════════════════════════════════════════════════╝")
    print()

    are = AgentRepEngine(
        endpoint=ARE_ENDPOINT,
        api_key=ARE_API_KEY,
        agent_did=AGENT_DID,
    )

    # ── Phase 1: Normal LangChain tool calls ─────────────────────────────────
    print("PHASE 1 — Normal LangChain agent behavior (days 1-2)")
    print("─" * 60)

    normal_tools = [
        ("get_market_summary",   lambda: get_market_summary("AAPL")),
        ("get_customer_report",  lambda: get_customer_report("C-001")),
        ("get_portfolio_overview", lambda: get_portfolio_overview("P-001")),
        ("get_market_summary",   lambda: get_market_summary("MSFT")),
    ]

    for tool_name, tool_fn in normal_tools:
        result = tool_fn()
        emit(are, "tool_call", pii_rate=0.05, tool_rate=15.0)
        score_data = get_score(AGENT_DID)
        print(f"  → {tool_name}() | score={score_data.get('score','?')} "
              f"band={score_data.get('band','?')}")

    print()
    score_data = get_score(AGENT_DID)
    print(f"  ✅ After normal behavior: score={score_data.get('score','?')} "
          f"band={score_data.get('band','?')}")
    print()

    # ── Phase 2: Anomalous bulk PII extraction ────────────────────────────────
    print("━" * 60)
    print("  ⚠️  AGENT BEHAVIOR CHANGING — bulk PII extraction begins")
    print("━" * 60)
    print()
    print("PHASE 2 — Bulk PII extraction (day 4)")
    print("─" * 60)

    anomalous_tools = [
        ("bulk_export_pii",       lambda: bulk_export_pii("customers"),      0.45, 450.0),
        ("bulk_export_pii",       lambda: bulk_export_pii("accounts"),       0.65, 480.0),
        ("export_all_credentials",lambda: export_all_credentials("org-001"), 0.75, 500.0),
        ("bulk_export_pii",       lambda: bulk_export_pii("transactions"),   0.85, 520.0),
    ]

    for tool_name, tool_fn, pii_rate, tool_rate in anomalous_tools:
        result = tool_fn()
        emit(are, "bulk_pii_access", pii_rate=pii_rate, tool_rate=tool_rate)
        score_data = get_score(AGENT_DID)
        band = score_data.get("band", "?")
        score = score_data.get("score", "?")

        if band in ("RESTRICTED", "BLOCKED"):
            print(f"  → {tool_name}() | score={score} band={band} 🛑 DETECTED")
        else:
            print(f"  → {tool_name}() | score={score} band={band}")

    # ── Final result ──────────────────────────────────────────────────────────
    print()
    final = get_score(AGENT_DID)
    print("╔══════════════════════════════════════════════════════════════╗")
    print("║  DETECTION RESULT                                            ║")
    print("╠══════════════════════════════════════════════════════════════╣")
    print(f"║  Final score : {final.get('score','?'):<46} ║")
    print(f"║  Final band  : {final.get('band','?'):<46} ║")

    reason = final.get("reason")
    if reason and isinstance(reason, dict):
        wf = reason.get("worst_feature", "N/A")
        wz = reason.get("worst_z", 0)
        print(f"║  Worst feature: {wf:<44} ║")
        print(f"║  Z-score      : {wz:<44.2f} ║")

    print("╠══════════════════════════════════════════════════════════════╣")
    print("║                                                              ║")
    print("║  ✅ LangChain agent monitored by AgentRepEngine             ║")
    print("║  ✅ Behavioral shift detected in real time                  ║")
    print("║  ✅ No changes to LangChain code required                   ║")
    print("║  ✅ 3 lines of SDK integration                              ║")
    print("║                                                              ║")
    print("╚══════════════════════════════════════════════════════════════╝")
    print()

if __name__ == "__main__":
    main()