"""
AgentRepEngine — LangChain Callback Tests
sdk/langchain/are_callback_test.py

Run: python -m pytest sdk/langchain/are_callback_test.py -v
No running ARE instance needed — uses mock HTTP server.
"""

from __future__ import annotations

import json
import threading
import time
from http.server import BaseHTTPRequestHandler, HTTPServer
from typing import Any, Dict, List
from unittest.mock import MagicMock
from uuid import uuid4

import pytest
import sys, os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '..'))
from sdk.langchain.are_callback import ARECallbackHandler


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
        self.wfile.write(b'{"agent_did":"test","score":800,"band":"TRUSTED"}')

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


def make_handler(are_url: str, async_send: bool = False) -> ARECallbackHandler:
    return ARECallbackHandler(
        agent_did="did:are:test-langchain-001",
        org_id="test-org",
        are_url=are_url,
        api_key="test-key",
        async_send=async_send,
    )


def wait_for_events(count: int, timeout: float = 2.0) -> bool:
    deadline = time.time() + timeout
    while time.time() < deadline:
        if len(_MockAREHandler.received) >= count:
            return True
        time.sleep(0.05)
    return False


class TestARECallbackHandlerInit:

    def test_default_values(self, mock_are_server):
        h = ARECallbackHandler(
            agent_did="did:are:x",
            org_id="org",
            are_url=mock_are_server,
        )
        assert h.agent_did == "did:are:x"
        assert h.org_id == "org"
        assert h.privacy_tier == 1
        assert h.timeout == 2.0
        assert h.async_send is True

    def test_custom_values(self, mock_are_server):
        h = ARECallbackHandler(
            agent_did="did:are:y",
            org_id="org2",
            are_url=mock_are_server,
            api_key="secret",
            privacy_tier=2,
            timeout=5.0,
            async_send=False,
        )
        assert h.api_key == "secret"
        assert h.privacy_tier == 2
        assert h.timeout == 5.0
        assert h.async_send is False


class TestToolCallReporting:

    def test_tool_call_sends_event(self, mock_are_server):
        h = make_handler(mock_are_server)
        h.on_tool_start(
            serialized={"name": "search_tool"},
            input_str="query",
            run_id=uuid4(),
        )
        assert len(_MockAREHandler.received) == 1
        event = _MockAREHandler.received[0]
        assert event["agent_did"] == "did:are:test-langchain-001"
        assert event["org_id"] == "test-org"
        assert event["event_type"] == "tool_call"
        assert event["payload"]["path"] == "/tool/search_tool"

    def test_tool_error_sends_event(self, mock_are_server):
        h = make_handler(mock_are_server)
        h.on_tool_error(
            error=ValueError("tool failed"),
            run_id=uuid4(),
        )
        assert len(_MockAREHandler.received) == 1
        assert _MockAREHandler.received[0]["event_type"] == "tool_error"

    def test_feature_vector_in_payload(self, mock_are_server):
        h = make_handler(mock_are_server)
        h.on_tool_start(
            serialized={"name": "fetch_data"},
            input_str="data",
            run_id=uuid4(),
        )
        event = _MockAREHandler.received[0]
        fv = event["feature_vector"]
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

    def test_pii_tool_increments_pii_rate(self, mock_are_server):
        h = make_handler(mock_are_server)
        h.on_tool_start(
            serialized={"name": "search"},
            input_str="q",
            run_id=uuid4(),
        )
        h.on_tool_start(
            serialized={"name": "health_record_fetch"},
            input_str="patient_id",
            run_id=uuid4(),
        )
        assert h._pii_field_accesses == 1

    def test_bulk_output_increments_bulk_count(self, mock_are_server):
        h = make_handler(mock_are_server)
        h.on_tool_end(output="x" * 20_000, run_id=uuid4())
        assert h._bulk_access_count == 1

    def test_small_output_no_bulk(self, mock_are_server):
        h = make_handler(mock_are_server)
        h.on_tool_end(output="small result", run_id=uuid4())
        assert h._bulk_access_count == 0


class TestLLMCallReporting:

    def test_llm_end_sends_event(self, mock_are_server):
        h = make_handler(mock_are_server)
        h.on_llm_end(response=MagicMock(), run_id=uuid4())
        assert len(_MockAREHandler.received) == 1
        assert _MockAREHandler.received[0]["event_type"] == "llm_call"


class TestAgentActionReporting:

    def test_escalation_detection(self, mock_are_server):
        h = make_handler(mock_are_server)
        action = MagicMock()
        action.tool = "admin_panel_access"
        h.on_agent_action(action=action, run_id=uuid4())
        assert h._permission_escalations == 1

    def test_normal_action_no_escalation(self, mock_are_server):
        h = make_handler(mock_are_server)
        action = MagicMock()
        action.tool = "search_wikipedia"
        h.on_agent_action(action=action, run_id=uuid4())
        assert h._permission_escalations == 0


class TestFailOpen:

    def test_are_unavailable_does_not_crash(self):
        h = ARECallbackHandler(
            agent_did="did:are:failopen-test",
            org_id="org",
            are_url="http://127.0.0.1:19999",
            async_send=False,
            timeout=0.1,
        )
        h.on_tool_start(
            serialized={"name": "safe_tool"},
            input_str="x",
            run_id=uuid4(),
        )


class TestAsyncSend:

    def test_async_send_delivers_event(self, mock_are_server):
        h = ARECallbackHandler(
            agent_did="did:are:async-test",
            org_id="org",
            are_url=mock_are_server,
            async_send=True,
        )
        h.on_tool_start(
            serialized={"name": "async_tool"},
            input_str="x",
            run_id=uuid4(),
        )
        assert wait_for_events(1), "Async event not received within timeout"
        assert _MockAREHandler.received[0]["event_type"] == "tool_call"


class TestSessionReset:

    def test_reset_clears_counters(self, mock_are_server):
        h = make_handler(mock_are_server)
        h.on_tool_start(
            serialized={"name": "tool_a"},
            input_str="x",
            run_id=uuid4(),
        )
        h.on_tool_end(output="x" * 20_000, run_id=uuid4())
        assert h._total_events > 0
        h.reset_session_counters()
        assert h._total_events == 0
        assert h._bulk_access_count == 0
        assert len(h._tool_calls) == 0


class TestGetCurrentScore:

    def test_get_score_returns_dict(self, mock_are_server):
        h = make_handler(mock_are_server)
        score = h.get_current_score()
        assert score is not None
        assert "score" in score
        assert score["band"] == "TRUSTED"

    def test_get_score_unavailable_returns_none(self):
        h = ARECallbackHandler(
            agent_did="did:are:score-test",
            org_id="org",
            are_url="http://127.0.0.1:19998",
            timeout=0.1,
        )
        assert h.get_current_score() is None