package audit

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestSIEM_RetryOn5xx(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	s := &SIEMWebhook{
		webhookURL: server.URL,
		authToken:  "test",
		hostname:   "test",
		enabled:    true,
		client:     server.Client(),
	}

	// Use short backoff for test speed — call doSend directly
	// to avoid waiting 30+60+120 seconds
	success, permanent := s.doSend([]byte(`{"test":true}`), "did:test", "BLOCKED", "HIGH")
	if success {
		t.Fatal("expected failure on 503, got success")
	}
	if permanent {
		t.Fatal("503 should be transient (retryable), not permanent")
	}
	t.Logf("A4: 503 correctly classified as transient failure, attempts=%d", atomic.LoadInt32(&attempts))
}

func TestSIEM_NoRetryOn4xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	s := &SIEMWebhook{
		webhookURL: server.URL,
		authToken:  "test",
		hostname:   "test",
		enabled:    true,
		client:     server.Client(),
	}

	success, permanent := s.doSend([]byte(`{"test":true}`), "did:test", "BLOCKED", "HIGH")
	if success {
		t.Fatal("expected failure on 400, got success")
	}
	if !permanent {
		t.Fatal("400 should be permanent (no retry)")
	}
	t.Log("A4: 400 correctly classified as permanent failure")
}

func TestSIEM_SuccessOn2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := &SIEMWebhook{
		webhookURL: server.URL,
		authToken:  "test",
		hostname:   "test",
		enabled:    true,
		client:     server.Client(),
	}

	success, _ := s.doSend([]byte(`{"test":true}`), "did:test", "BLOCKED", "HIGH")
	if !success {
		t.Fatal("expected success on 200, got failure")
	}
	t.Log("A4: 200 correctly classified as success")
}

func TestSIEM_DisabledSkips(t *testing.T) {
	s := &SIEMWebhook{enabled: false}
	// Should not panic or attempt delivery
	s.SendBlocked("did:test", 100, 0, "test", []byte(`{}`))
	t.Log("A4: Disabled SIEM correctly skips delivery")
}
