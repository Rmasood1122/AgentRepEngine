"""
AgentRepEngine — LangChain Callback Handler
sdk/langchain/are_callback.py
"""

from __future__ import annotations

import threading
import time
from typing import Any, Dict, List, Optional, Union
from uuid import UUID

import requests
from langchain_core.callbacks.base import BaseCallbackHandler
from langchain_core.outputs import LLMResult


class ARECallbackHandler(BaseCallbackHandler):
    """LangChain callback handler for AgentRepEngine behavioral scoring."""

    def __init__(
        self,
        agent_did: str,
        org_id: str,
        are_url: str = "http://localhost:8080",
        api_key: Optional[str] = None,
        privacy_tier: int = 1,
        timeout: float = 2.0,
        async_send: bool = True,
    ) -> None:
        super().__init__()
        self.agent_did = agent_did
        self.org_id = org_id
        self.are_url = are_url.rstrip("/")
        self.api_key = api_key
        self.privacy_tier = privacy_tier
        self.timeout = timeout
        self.async_send = async_send

        self._lock = threading.Lock()
        self._tool_calls: List[float] = []
        self._endpoints_seen: set = set()
        self._bulk_access_count: int = 0
        self._pii_field_accesses: int = 0
        self._total_events: int = 0
        self._token_refreshes: int = 0
        self._sub_agent_depth: int = 0
        self._cross_tenant_probes: int = 0
        self._permission_escalations: int = 0
        self._session_start: float = time.time()
        self._active_runs: Dict[str, int] = {}

    def on_tool_start(
        self,
        serialized: Dict[str, Any],
        input_str: str,
        *,
        run_id: UUID,
        parent_run_id: Optional[UUID] = None,
        tags: Optional[List[str]] = None,
        metadata: Optional[Dict[str, Any]] = None,
        **kwargs: Any,
    ) -> None:
        tool_name = serialized.get("name", "unknown_tool")
        with self._lock:
            now = time.time()
            self._tool_calls.append(now)
            self._endpoints_seen.add(tool_name)
            self._total_events += 1
            pii_keywords = {"email", "ssn", "credit", "password",
                            "health", "medical", "phi", "pii", "private"}
            if any(kw in tool_name.lower() for kw in pii_keywords):
                self._pii_field_accesses += 1
            if metadata and metadata.get("cross_tenant"):
                self._cross_tenant_probes += 1
        self._send_event(
            event_type="tool_call",
            payload={
                "method": "TOOL",
                "path": f"/tool/{tool_name}",
                "status_code": 0,
                "score_at_request": 0,
                "band_at_request": "UNKNOWN",
            },
        )

    def on_tool_end(
        self,
        output: str,
        *,
        run_id: UUID,
        parent_run_id: Optional[UUID] = None,
        **kwargs: Any,
    ) -> None:
        with self._lock:
            if len(output) > 10_000:
                self._bulk_access_count += 1

    def on_tool_error(
        self,
        error: Union[Exception, KeyboardInterrupt],
        *,
        run_id: UUID,
        parent_run_id: Optional[UUID] = None,
        **kwargs: Any,
    ) -> None:
        with self._lock:
            self._total_events += 1
        self._send_event(
            event_type="tool_error",
            payload={
                "method": "TOOL_ERROR",
                "path": f"/tool/error/{type(error).__name__}",
                "status_code": 500,
                "score_at_request": 0,
                "band_at_request": "UNKNOWN",
            },
        )

    def on_llm_start(
        self,
        serialized: Dict[str, Any],
        prompts: List[str],
        *,
        run_id: UUID,
        parent_run_id: Optional[UUID] = None,
        **kwargs: Any,
    ) -> None:
        with self._lock:
            self._total_events += 1
            if len(self._tool_calls) > 0:
                self._token_refreshes += 1

    def on_llm_end(
        self,
        response: LLMResult,
        *,
        run_id: UUID,
        parent_run_id: Optional[UUID] = None,
        **kwargs: Any,
    ) -> None:
        self._send_event(
            event_type="llm_call",
            payload={
                "method": "LLM",
                "path": "/llm/completion",
                "status_code": 200,
                "score_at_request": 0,
                "band_at_request": "UNKNOWN",
            },
        )

    def on_agent_action(
        self,
        action: Any,
        *,
        run_id: UUID,
        parent_run_id: Optional[UUID] = None,
        **kwargs: Any,
    ) -> None:
        with self._lock:
            depth = self._active_runs.get(str(parent_run_id), 0) + 1
            self._active_runs[str(run_id)] = depth
            self._sub_agent_depth = max(self._sub_agent_depth, depth)
            tool = getattr(action, "tool", "")
            escalation_keywords = {"admin", "root", "sudo", "elevate",
                                   "privilege", "system", "execute"}
            if any(kw in tool.lower() for kw in escalation_keywords):
                self._permission_escalations += 1

    def on_chain_start(
        self,
        serialized: Dict[str, Any],
        inputs: Dict[str, Any],
        *,
        run_id: UUID,
        parent_run_id: Optional[UUID] = None,
        **kwargs: Any,
    ) -> None:
        with self._lock:
            depth = self._active_runs.get(str(parent_run_id), 0) + 1
            self._active_runs[str(run_id)] = depth

    def on_chain_end(
        self,
        outputs: Dict[str, Any],
        *,
        run_id: UUID,
        parent_run_id: Optional[UUID] = None,
        **kwargs: Any,
    ) -> None:
        with self._lock:
            self._active_runs.pop(str(run_id), None)

    def _compute_feature_vector(self) -> Dict[str, float]:
        now = time.time()
        hour_ago = now - 3600.0
        recent_calls = [t for t in self._tool_calls if t > hour_ago]
        elapsed_hours = max((now - self._session_start) / 3600.0, 1/3600.0)
        tool_call_rate = len(recent_calls) / elapsed_hours
        unique_endpoints = len(self._endpoints_seen) / elapsed_hours
        total = max(self._total_events, 1)
        pii_rate = self._pii_field_accesses / total
        return {
            "tool_call_rate_per_hour": round(tool_call_rate, 4),
            "unique_endpoints_per_hour": round(unique_endpoints, 4),
            "bulk_access_count_per_session": float(self._bulk_access_count),
            "pii_field_access_rate": round(pii_rate, 4),
            "cross_tenant_probe_count": float(self._cross_tenant_probes),
            "permission_escalation_count": float(self._permission_escalations),
            "sub_agent_spawn_depth": float(self._sub_agent_depth),
            "token_refresh_rate": round(
                self._token_refreshes / elapsed_hours, 4),
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

    def reset_session_counters(self) -> None:
        with self._lock:
            self._tool_calls.clear()
            self._endpoints_seen.clear()
            self._bulk_access_count = 0
            self._pii_field_accesses = 0
            self._total_events = 0
            self._token_refreshes = 0
            self._sub_agent_depth = 0
            self._cross_tenant_probes = 0
            self._permission_escalations = 0
            self._session_start = time.time()
            self._active_runs.clear()