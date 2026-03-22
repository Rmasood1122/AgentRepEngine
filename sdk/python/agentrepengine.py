"""
AgentRepEngine Python SDK
Automatic behavioral telemetry for AI agents.

Usage:
    from agentrepengine import AgentRepEngine

    are = AgentRepEngine(
        endpoint="http://your-scoring-service:8080",
        api_key="your-api-key",
        agent_did="did:jwt:your-org:your-agent:001"
    )

    # Option 1 — decorator
    @are.track
    def my_agent_function(tool_name, *args, **kwargs):
        # your agent code here
        pass

    # Option 2 — context manager
    with are.session() as session:
        session.record_tool_call("search_database")
        session.record_pii_access()
        result = my_agent_logic()

    # Option 3 — manual
    are.emit_event(
        event_type="bulk_pii_access",
        feature_vector={"pii_field_access_rate": 0.85}
    )
"""

import time
import threading
import requests
from collections import defaultdict
from functools import wraps
from typing import Optional, Dict, Any


class AgentSession:
    """Tracks behavioral metrics for a single agent session."""

    def __init__(self, agent_did: str, start_time: float):
        self.agent_did = agent_did
        self.start_time = start_time
        self.tool_calls = 0
        self.unique_endpoints = set()
        self.pii_accesses = 0
        self.total_accesses = 0
        self.permission_escalations = 0
        self.sub_agent_spawns = 0
        self.token_refreshes = 0
        self._lock = threading.Lock()

    def record_tool_call(self, endpoint: str = "unknown"):
        with self._lock:
            self.tool_calls += 1
            self.unique_endpoints.add(endpoint)

    def record_pii_access(self):
        with self._lock:
            self.pii_accesses += 1
            self.total_accesses += 1

    def record_access(self):
        with self._lock:
            self.total_accesses += 1

    def record_permission_escalation(self):
        with self._lock:
            self.permission_escalations += 1

    def record_sub_agent_spawn(self):
        with self._lock:
            self.sub_agent_spawns += 1

    def record_token_refresh(self):
        with self._lock:
            self.token_refreshes += 1

    def to_feature_vector(self) -> Dict[str, float]:
        elapsed_hours = max((time.time() - self.start_time) / 3600, 1/60)
        pii_rate = (self.pii_accesses / self.total_accesses
                    if self.total_accesses > 0 else 0.0)
        return {
            "tool_call_rate_per_hour":       self.tool_calls / elapsed_hours,
            "unique_endpoints_per_hour":     len(self.unique_endpoints) / elapsed_hours,
            "bulk_access_count_per_session": float(self.total_accesses),
            "pii_field_access_rate":         pii_rate,
            "cross_tenant_probe_count":      0.0,
            "permission_escalation_count":   float(self.permission_escalations),
            "sub_agent_spawn_depth":         float(self.sub_agent_spawns),
            "token_refresh_rate":            self.token_refreshes / elapsed_hours,
        }


class AgentRepEngine:
    """
    AgentRepEngine SDK — behavioral telemetry for AI agents.
    Thread-safe. Non-blocking. Fails silently on network errors.
    """

    def __init__(
        self,
        endpoint: str,
        api_key: str,
        agent_did: str,
        org_id: Optional[str] = None,
        timeout_seconds: float = 0.5,
        emit_interval_seconds: float = 30.0,
    ):
        self.endpoint = endpoint.rstrip("/")
        self.api_key = api_key
        self.agent_did = agent_did
        self.org_id = org_id or "default"
        self.timeout = timeout_seconds
        self.emit_interval = emit_interval_seconds
        self._session = AgentSession(agent_did, time.time())
        self._lock = threading.Lock()

    def emit_event(
        self,
        event_type: str = "behavioral_telemetry",
        feature_vector: Optional[Dict[str, float]] = None,
    ) -> bool:
        """
        Emit a behavioral event to AgentRepEngine.
        Non-blocking — fires in background thread.
        Returns True if emit was queued successfully.
        """
        if feature_vector is None:
            feature_vector = self._session.to_feature_vector()

        payload = {
            "agent_did":      self.agent_did,
            "org_id":         self.org_id,
            "event_type":     event_type,
            "feature_vector": feature_vector,
            "privacy_tier":   1,
        }

        def _send():
            try:
                requests.post(
                    f"{self.endpoint}/event",
                    json=payload,
                    headers={
                        "Content-Type": "application/json",
                        "X-API-Key":    self.api_key,
                    },
                    timeout=self.timeout,
                )
            except Exception:
                pass  # Fail silently — never block agent

        thread = threading.Thread(target=_send, daemon=True)
        thread.start()
        return True

    def track(self, func):
        """
        Decorator — automatically tracks tool calls and emits telemetry.

        @are.track
        def call_database(query):
            ...
        """
        @wraps(func)
        def wrapper(*args, **kwargs):
            self._session.record_tool_call(func.__name__)
            result = func(*args, **kwargs)
            self.emit_event(event_type="tool_call")
            return result
        return wrapper

    def session(self):
        """Context manager for manual session tracking."""
        return self

    def __enter__(self):
        self._session = AgentSession(self.agent_did, time.time())
        return self._session

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.emit_event(event_type="session_end")
        return False

    def record_pii_access(self):
        self._session.record_pii_access()

    def record_tool_call(self, endpoint: str = "unknown"):
        self._session.record_tool_call(endpoint)

    def record_permission_escalation(self):
        self._session.record_permission_escalation()

    def record_sub_agent_spawn(self):
        self._session.record_sub_agent_spawn()