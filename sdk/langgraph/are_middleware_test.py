"""
AgentRepEngine — LangGraph Middleware Tests
sdk/langgraph/are_middleware_test.py

Run: python -m pytest sdk/langgraph/are_middleware_test.py -v
No running ARE instance needed — uses mock HTTP server.
"""

from __future__ import annotations

import json
import threading
import time
from http.server import BaseHTTPRequestHandler, HTTPServer
from typing import Any, Dict, List

import pytest
import sys, os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '..'))
from sdk.langgraph.are_middleware import AREMiddleware


class _MockAREHandler(BaseHTTPRequestHandler):
    received: List[Dict] = []

    def do_POST(self):
        length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(length)
        try:
            _MockAREHandler.received.append(json.loads(body))
        except Exception:
            pass
        self.send_response(202)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(b'{"status":"queued"}')

    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(b'{"agent_did":"test","score":750,"band":"TRUSTED"}')

    def log_message(self, *args):
        pass


@pytest.fixture(scope="module")
def mock_are_server():
    _MockAREHandler.received = []
    server = HTTPServer(("127.0.0.1", 0), _MockAREHandler)
    port = server.server_address[1]
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    yield f"http://127.0.0.1:{port}"
    server.shutdown()


@pytest.fixture(autouse=True)
def clear_received():
    _MockAREHandler.received.clear()
    yield


def make_middleware(are_url: str, async_send: bool = False) -> AREMiddleware:
    return AREMiddleware(
        agent_did="did:are:test-langgraph-001",
        org_id="test-org",
        are_url=are_url,
        api_key="test-key",
        async_send=async_send,
    )


def dummy_node(state: Dict[str, Any]) -> Dict[str, Any]:
    return {**state, "visited": state.get("visited", []) + ["dummy"]}


def error_node(state: Dict[str, Any]) -> Dict[str, Any]:
    raise RuntimeError("node failed intentionally")


def wait_for_events(count: int, timeout: float = 2.0) -> bool:
    deadline = time.time() + timeout
    while time.time() < deadline:
        if len(_MockAREHandler.received) >= count:
            return True
        time.sleep(0.05)
    return False


class TestAREMiddlewareInit:

    def test_default_values(self, mock_are_server):
        m = AREMiddleware(
            agent_did="did:are:x",
            org_id="org",
            are_url=mock_are_server,
        )
        assert m.agent_did == "did:are:x"
        assert m.privacy_tier == 1
        assert m.async_send is True
        assert m.cycle_threshold == 10

    def test_custom_values(self, mock_are_server):
        m = AREMiddleware(
            agent_did="did:are:y",
            org_id="org",
            are_url=mock_are_server,
            privacy_tier=3,
            cycle_threshold=5,
            async_send=False,
        )
        assert m.privacy_tier == 3
        assert m.cycle_threshold == 5


class TestWrapNode:

    def test_wrapped_node_executes_correctly(self, mock_are_server):
        m = make_middleware(mock_are_server)
        wrapped = m.wrap_node("test_node", dummy_node)
        result = wrapped({"key": "value"})
        assert result["visited"] == ["dummy"]
        assert result["key"] == "value"

    def test_wrapped_node_sends_event(self, mock_are_server):
        m = make_middleware(mock_are_server)
        wrapped = m.wrap_node("test_node", dummy_node)
        wrapped({"x": 1})
        assert len(_MockAREHandler.received) == 1
        event = _MockAREHandler.received[0]
        assert event["event_type"] == "node_execution"
        assert event["payload"]["path"] == "/graph/node/test_node"

    def test_wrapped_node_error_sends_event(self, mock_are_server):
        m = make_middleware(mock_are_server)
        wrapped = m.wrap_node("error_node", error_node)
        with pytest.raises(RuntimeError):
            wrapped({})
        assert len(_MockAREHandler.received) == 1
        assert _MockAREHandler.received[0]["event_type"] == "node_error"

    def test_node_visit_tracked(self, mock_are_server):
        m = make_middleware(mock_are_server)
        wrapped = m.wrap_node("tracked_node", dummy_node)
        for _ in range(3):
            wrapped({"x": 1})
        assert len(m._node_visits["tracked_node"]) == 3

    def test_cycle_detection(self, mock_are_server):
        m = make_middleware(mock_are_server, async_send=False)
        m.cycle_threshold = 2
        wrapped = m.wrap_node("cycle_node", dummy_node)
        wrapped({"x": 1})
        wrapped({"x": 1})
        _MockAREHandler.received.clear()
        wrapped({"x": 1})
        assert _MockAREHandler.received[0]["event_type"] == "node_cycle"

    def test_pii_node_increments_pii(self, mock_are_server):
        m = make_middleware(mock_are_server)
        wrapped = m.wrap_node("phi_data_node", dummy_node)
        wrapped({"x": 1})
        assert m._pii_accesses == 1

    def test_escalation_node_detected(self, mock_are_server):
        m = make_middleware(mock_are_server)
        wrapped = m.wrap_node("admin_access_node", dummy_node)
        wrapped({"x": 1})
        assert m._permission_escalations == 1

    def test_state_mutation_tracked(self, mock_are_server):
        m = make_middleware(mock_are_server)
        wrapped = m.wrap_node("mutation_node", dummy_node)
        wrapped({"x": 1})
        assert m._state_mutations >= 1


class TestFeatureVector:

    def test_all_keys_present(self, mock_are_server):
        m = make_middleware(mock_are_server)
        with m._lock:
            fv = m._compute_feature_vector()
        required = {
            "tool_call_rate_per_hour",
            "unique_endpoints_per_hour",
            "bulk_access_count_per_session",
            "pii_field_access_rate",
            "cross_tenant_probe_count",
            "permission_escalation_count",
            "sub_agent_spawn_depth",
            "token_refresh_rate",
        }
        assert required == set(fv.keys())

    def test_feature_vector_in_event(self, mock_are_server):
        m = make_middleware(mock_are_server)
        wrapped = m.wrap_node("fv_test_node", dummy_node)
        wrapped({"x": 1})
        fv = _MockAREHandler.received[0]["feature_vector"]
        assert len(fv) == 8


class TestManualMarkers:

    def test_mark_cross_tenant(self, mock_are_server):
        m = make_middleware(mock_are_server)
        m.mark_cross_tenant(3)
        assert m._cross_tenant_probes == 3

    def test_mark_sub_graph_takes_max(self, mock_are_server):
        m = make_middleware(mock_are_server)
        m.mark_sub_graph(2)
        m.mark_sub_graph(5)
        assert m._sub_graph_depth == 5


class TestFailOpen:

    def test_are_unavailable_node_still_executes(self):
        m = AREMiddleware(
            agent_did="did:are:failopen",
            org_id="org",
            are_url="http://127.0.0.1:19997",
            async_send=False,
            timeout=0.1,
        )
        wrapped = m.wrap_node("safe_node", dummy_node)
        result = wrapped({"x": 1})
        assert result["visited"] == ["dummy"]


class TestSessionReset:

    def test_reset_clears_all_state(self, mock_are_server):
        m = make_middleware(mock_are_server)
        wrapped = m.wrap_node("node_a", dummy_node)
        wrapped({"x": 1})
        m.mark_cross_tenant(2)
        m.mark_sub_graph(3)
        m.reset_session()
        assert len(m._node_visits) == 0
        assert m._cross_tenant_probes == 0
        assert m._sub_graph_depth == 0
        assert m._total_events == 0


class TestSessionSummary:

    def test_summary_contains_required_fields(self, mock_are_server):
        m = make_middleware(mock_are_server)
        wrapped = m.wrap_node("summary_node", dummy_node)
        wrapped({"x": 1})
        summary = m.session_summary()
        assert "agent_did" in summary
        assert "nodes_visited" in summary
        assert "node_visit_counts" in summary
        assert "feature_vector" in summary
        assert "summary_node" in summary["nodes_visited"]


class TestAsyncSend:

    def test_async_node_wrap_delivers_event(self, mock_are_server):
        m = AREMiddleware(
            agent_did="did:are:async-graph",
            org_id="org",
            are_url=mock_are_server,
            async_send=True,
        )
        wrapped = m.wrap_node("async_node", dummy_node)
        wrapped({"x": 1})
        assert wait_for_events(1), "Async event not received"
        assert _MockAREHandler.received[0]["event_type"] == "node_execution"


class TestGetCurrentScore:

    def test_returns_score_dict(self, mock_are_server):
        m = make_middleware(mock_are_server)
        score = m.get_current_score()
        assert score is not None
        assert score["band"] == "TRUSTED"

    def test_unavailable_returns_none(self):
        m = AREMiddleware(
            agent_did="did:are:score-none",
            org_id="org",
            are_url="http://127.0.0.1:19996",
            timeout=0.1,
        )
        assert m.get_current_score() is None