package handlers

import (
	"testing"
	"time"
)

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := NewRateLimiter(100, 1*time.Minute)
	for i := 0; i < 100; i++ {
		allowed, _ := rl.Allow("did:jwt:test:agent:001")
		if !allowed {
			t.Fatalf("request %d should be allowed, got denied", i+1)
		}
	}
	t.Log("A2: 100 requests within limit all allowed")
}

func TestRateLimiter_DeniesOverLimit(t *testing.T) {
	rl := NewRateLimiter(100, 1*time.Minute)
	for i := 0; i < 100; i++ {
		rl.Allow("did:jwt:test:agent:001")
	}
	allowed, retryAfter := rl.Allow("did:jwt:test:agent:001")
	if allowed {
		t.Fatal("101st request should be denied, got allowed")
	}
	if retryAfter < 1*time.Second {
		t.Errorf("retry-after should be >= 1s, got %v", retryAfter)
	}
	t.Logf("A2: 101st request denied, retry-after=%v", retryAfter)
}

func TestRateLimiter_PerAgentIsolation(t *testing.T) {
	rl := NewRateLimiter(100, 1*time.Minute)
	for i := 0; i < 100; i++ {
		rl.Allow("did:jwt:test:agent:A")
	}
	allowed, _ := rl.Allow("did:jwt:test:agent:B")
	if !allowed {
		t.Fatal("agent B should not be affected by agent A limit")
	}
	t.Log("A2: Per-agent isolation confirmed")
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl := NewRateLimiter(5, 50*time.Millisecond)
	for i := 0; i < 5; i++ {
		rl.Allow("did:jwt:test:agent:001")
	}
	allowed, _ := rl.Allow("did:jwt:test:agent:001")
	if allowed {
		t.Fatal("should be denied at limit")
	}
	time.Sleep(60 * time.Millisecond)
	allowed, _ = rl.Allow("did:jwt:test:agent:001")
	if !allowed {
		t.Fatal("should be allowed after window expiry")
	}
	t.Log("A2: Window expiry resets rate limit")
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter(100, 1*time.Minute)
	done := make(chan bool, 50)
	for i := 0; i < 50; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				rl.Allow("did:jwt:test:concurrent")
			}
			done <- true
		}()
	}
	for i := 0; i < 50; i++ {
		<-done
	}
	t.Log("A2: 50 goroutines x 10 requests -- no race, no panic")
}
