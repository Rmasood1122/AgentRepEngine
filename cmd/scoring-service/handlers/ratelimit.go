package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// A2 Hardening Sprint: per-agent-DID rate limit on /event.
// Prevents queue flood → OOM → enforcement on stale data.
// 100 events/min per agent-DID. Returns 429 + Retry-After.
// 16x headroom over highest legitimate rate in corpus (~6/min).

const (
	// RateLimitPerMinute is the max events per agent-DID per minute.
	RateLimitPerMinute = 100

	// RateLimitWindow is the sliding window duration.
	RateLimitWindow = 1 * time.Minute

	// RateLimitCleanupInterval controls how often expired entries are pruned.
	RateLimitCleanupInterval = 5 * time.Minute
)

// rateBucket tracks event timestamps for one agent-DID.
type rateBucket struct {
	mu         sync.Mutex
	timestamps []time.Time
}

// RateLimiter provides per-agent-DID rate limiting.
type RateLimiter struct {
	mu      sync.RWMutex
	buckets map[string]*rateBucket
	limit   int
	window  time.Duration
}

// NewRateLimiter creates a rate limiter with cleanup goroutine.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*rateBucket),
		limit:   limit,
		window:  window,
	}
	go rl.cleanup()
	return rl
}

// Allow checks if the agent-DID is within rate limit.
// Returns (allowed bool, retryAfter time.Duration).
func (rl *RateLimiter) Allow(agentDID string) (bool, time.Duration) {
	now := time.Now()
	cutoff := now.Add(-rl.window)

	rl.mu.RLock()
	bucket, exists := rl.buckets[agentDID]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		bucket, exists = rl.buckets[agentDID]
		if !exists {
			bucket = &rateBucket{}
			rl.buckets[agentDID] = bucket
		}
		rl.mu.Unlock()
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Prune expired timestamps
	valid := bucket.timestamps[:0]
	for _, ts := range bucket.timestamps {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	bucket.timestamps = valid

	if len(bucket.timestamps) >= rl.limit {
		// Calculate retry-after from oldest event in window
		oldest := bucket.timestamps[0]
		retryAfter := oldest.Add(rl.window).Sub(now)
		if retryAfter < time.Second {
			retryAfter = time.Second
		}
		return false, retryAfter
	}

	bucket.timestamps = append(bucket.timestamps, now)
	return true, 0
}

// cleanup periodically removes empty buckets to prevent memory growth.
func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(RateLimitCleanupInterval)
		now := time.Now()
		cutoff := now.Add(-rl.window)

		rl.mu.Lock()
		for did, bucket := range rl.buckets {
			bucket.mu.Lock()
			hasRecent := false
			for _, ts := range bucket.timestamps {
				if ts.After(cutoff) {
					hasRecent = true
					break
				}
			}
			if !hasRecent {
				delete(rl.buckets, did)
			}
			bucket.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware wraps an http.HandlerFunc with per-agent-DID rate limiting.
// Reads agent_did from JSON body without consuming it (peeks via buffered copy).
// Returns 429 with Retry-After header when limit exceeded.
func RateLimitMiddleware(rl *RateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract agent_did from request body without consuming it
		// We need to peek at the body, so buffer it
		var peek struct {
			AgentDID string `json:"agent_did"`
		}

		// Use a TeeReader to preserve the body for the next handler
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = ReadBody(r)
			r.Body = NewBodyReader(bodyBytes)
		}

		if err := json.Unmarshal(bodyBytes, &peek); err != nil || peek.AgentDID == "" {
			// Can't extract agent_did — let the event handler return 400
			next(w, r)
			return
		}

		allowed, retryAfter := rl.Allow(peek.AgentDID)
		if !allowed {
			slog.Warn("rate_limit_exceeded",
				"agent_did", peek.AgentDID,
				"limit", rl.limit,
				"window", rl.window.String(),
				"retry_after_seconds", int(retryAfter.Seconds()),
			)
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(retryAfter.Seconds())))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintf(w, `{"error":"rate_limit_exceeded","agent_did":%q,"limit":%d,"window":"1m","retry_after_seconds":%d}`,
				peek.AgentDID, rl.limit, int(retryAfter.Seconds()))
			return
		}

		// Restore body for the next handler
		r.Body = NewBodyReader(bodyBytes)
		next(w, r)
	}
}
