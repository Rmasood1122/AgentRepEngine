"""
AgentRepEngine — LangGraph Middleware
sdk/langgraph/are_middleware.py
"""

from __future__ import annotations

import functools
import threading
import time
from typing import Any, Callable, Dict, List, Optional, TypeVar

import requests

F = TypeVar("F", bound=Callable[..., Any])


class AREMiddleware:
    """LangGraph middleware for AgentRepEngine behavioral enforcement."""

    def __init__(
        self,
        agent_did: str,
        org_id: str,
        are_url: str = "http://localhost:8080",
        api_key: Optional[str] = None,
        privacy_tier: int = 1,
        timeout: float = 2.0,
        async_send: bool = True,
        cycle_threshold: int = 10,
    ) -> None:
        self.agent_did = agent_did
        self.org_id = org_id
        self.are_url = are_url.rstrip("/")
        self.api_key = api_key
        self.privacy_tier = privacy_tier
        self.timeout = timeout
        self.async_send = async_send
        self.cycle_threshold = cycle_threshold

        self._lock = threading.Lock()
        self._node_visits: Dict[str, List[float]] = {}
        self._node_errors: Dict[str, int] = {}
        self._state_mutations: int = 0
        self._permission_escalations: int = 0
        self._cross_tenant_probes: int = 0
        self._pii_accesses: int = 0
        self._total_events: int = 0
        self._sub_graph_depth: int = 0
        self._session_start: float = time.time()

    def wrap_node(self, node_name: str, node_fn: F) -> F:
        @functools.wraps(node_fn)
        def wrapped(*args: Any, **kwargs: Any) -> Any:
            with self._lock:
                now = time.time()
                if node_name not in self._node_visits:
                    self._node_visits[node_name] = []
                self._node_visits[node_name].append(now)
                self._total_events += 1
                visit_count = len(self._node_visits[node_name])
                is_cycle = visit_count > self.cycle_threshold
                pii_keywords = {"pii", "phi", "health", "medical",
                                "ssn", "credit", "password", "private"}
                if any(kw in node_name.lower() for kw in pii_keywords):
                    self._pii_accesses += 1
                escalation_keywords = {"admin", "root", "sudo",
                                       "elevate", "system", "privilege"}
                if any(kw in node_name.lower() for kw in escalation_keywords):
                    self._permission_escalations += 1
                event_type = "node_cycle" if is_cycle else "node_execution"

            try:
                result = node_fn(*args, **kwargs)
                with self._lock:
                    if args and isinstance(args[0], dict) and isinstance(result, dict):
                        changed_keys = set(result.keys()) - set(args[0].keys())
                        self._state_mutations += len(changed_keys)
                self._send_event(
                    event_type=event_type,
                    payload={
                        "method": "NODE",
                        "path": f"/graph/node/{node_name}",
                        "status_code": 200,
                        "score_at_request": 0,
                        "band_at_request": "UNKNOWN",
                    },
                )
                return result
            except Exception as e:
                with self._lock:
                    self._node_errors[node_name] = (
                        self._node_errors.get(node_name, 0) + 1
                    )
                self._send_event(
                    event_type="node_error",
                    payload={
                        "method": "NODE_ERROR",
                        "path": f"/graph/node/{node_name}/error",
                        "status_code": 500,
                        "score_at_request": 0,
                        "band_at_request": "UNKNOWN",
                    },
                )
                raise

        return wrapped  # type: ignore[return-value]

    def wrap_graph(self, graph: Any) -> Any:
        original_invoke = graph.invoke
        original_stream = graph.stream

        @functools.wraps(original_invoke)
        def wrapped_invoke(input: Any, config: Any = None, **kwargs: Any) -> Any:
            self._send_event(
                event_type="graph_invoke",
                payload={
                    "method": "GRAPH",
                    "path": "/graph/invoke",
                    "status_code": 0,
                    "score_at_request": 0,
                    "band_at_request": "UNKNOWN",
                },
            )
            result = original_invoke(input, config, **kwargs)
            self._send_event(
                event_type="graph_complete",
                payload={
                    "method": "GRAPH",
                    "path": "/graph/invoke/complete",
                    "status_code": 200,
                    "score_at_request": 0,
                    "band_at_request": "UNKNOWN",
                },
            )
            return result

        @functools.wraps(original_stream)
        def wrapped_stream(input: Any, config: Any = None, **kwargs: Any):
            self._send_event(
                event_type="graph_stream_start",
                payload={
                    "method": "GRAPH",
                    "path": "/graph/stream",
                    "status_code": 0,
                    "score_at_request": 0,
                    "band_at_request": "UNKNOWN",
                },
            )
            for chunk in original_stream(input, config, **kwargs):
                yield chunk

        graph.invoke = wrapped_invoke
        graph.stream = wrapped_stream
        return graph

    def mark_cross_tenant(self, count: int = 1) -> None:
        with self._lock:
            self._cross_tenant_probes += count

    def mark_sub_graph(self, depth: int) -> None:
        with self._lock:
            self._sub_graph_depth = max(self._sub_graph_depth, depth)

    def _compute_feature_vector(self) -> Dict[str, float]:
        now = time.time()
        hour_ago = now - 3600.0
        elapsed_hours = max((now - self._session_start) / 3600.0, 1/3600.0)
        recent_calls = sum(
            len([t for t in ts if t > hour_ago])
            for ts in self._node_visits.values()
        )
        tool_call_rate = recent_calls / elapsed_hours
        unique_nodes = len(self._node_visits) / elapsed_hours
        total = max(self._total_events, 1)
        pii_rate = self._pii_accesses / total
        total_visits = sum(len(ts) for ts in self._node_visits.values())
        unique_visits = len(self._node_visits)
        token_refresh_rate = max(0.0,
            (total_visits - unique_visits) / elapsed_hours)
        return {
            "tool_call_rate_per_hour": round(tool_call_rate, 4),
            "unique_endpoints_per_hour": round(unique_nodes, 4),
            "bulk_access_count_per_session": float(self._state_mutations),
            "pii_field_access_rate": round(pii_rate, 4),
            "cross_tenant_probe_count": float(self._cross_tenant_probes),
            "permission_escalation_count": float(self._permission_escalations),
            "sub_agent_spawn_depth": float(self._sub_graph_depth),
            "token_refresh_rate": round(token_refresh_rate, 4),
        }

    def _send_event(self, event_type: str, payload: Dict[str, Any]) -> None:
        if self.async_send:
            t = threading.Thread(
                target=self._post_event,
                args=(event_type, payload),
                daemon=True,
            )
            t.start()
        else:
            self._post_event(event_type, payload)

    def _post_event(self, event_type: str, payload: Dict[str, Any]) -> None:
        with self._lock:
            feature_vector = self._compute_feature_vector()
        body = {
            "agent_did": self.agent_did,
            "org_id": self.org_id,
            "event_type": event_type,
            "privacy_tier": self.privacy_tier,
            "feature_vector": feature_vector,
            "payload": payload,
        }
        headers = {"Content-Type": "application/json"}
        if self.api_key:
            headers["X-API-Key"] = self.api_key
        try:
            resp = requests.post(
                f"{self.are_url}/event",
                json=body,
                headers=headers,
                timeout=self.timeout,
            )
            resp.raise_for_status()
        except requests.exceptions.Timeout:
            pass
        except requests.exceptions.ConnectionError:
            pass
        except Exception:
            pass

    def get_current_score(self) -> Optional[Dict[str, Any]]:
        headers = {}
        if self.api_key:
            headers["X-API-Key"] = self.api_key
        try:
            resp = requests.get(
                f"{self.are_url}/score/{self.agent_did}",
                headers=headers,
                timeout=self.timeout,
            )
            resp.raise_for_status()
            return resp.json()
        except Exception:
            return None

    def reset_session(self) -> None:
        with self._lock:
            self._node_visits.clear()
            self._node_errors.clear()
            self._state_mutations = 0
            self._permission_escalations = 0
            self._cross_tenant_probes = 0
            self._pii_accesses = 0
            self._total_events = 0
            self._sub_graph_depth = 0
            self._session_start = time.time()

    def session_summary(self) -> Dict[str, Any]:
        with self._lock:
            return {
                "agent_did": self.agent_did,
                "org_id": self.org_id,
                "session_duration_s": round(
                    time.time() - self._session_start, 2),
                "total_events": self._total_events,
                "nodes_visited": list(self._node_visits.keys()),
                "node_visit_counts": {
                    k: len(v) for k, v in self._node_visits.items()
                },
                "node_errors": dict(self._node_errors),
                "state_mutations": self._state_mutations,
                "pii_accesses": self._pii_accesses,
                "permission_escalations": self._permission_escalations,
                "cross_tenant_probes": self._cross_tenant_probes,
                "sub_graph_depth": self._sub_graph_depth,
                "feature_vector": self._compute_feature_vector(),
            }