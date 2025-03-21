package util

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.RWMutex
	requests map[string]time.Time
	ttl      time.Duration
	window   time.Duration
}

func newRateLimiter(ttl, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]time.Time),
		ttl:      ttl,
		window:   window,
	}

	go rl.cleanupRoutine()
	return rl
}

func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		rl.mu.Lock()
		for key, last := range rl.requests {
			if now.Sub(last) > rl.ttl {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) allow(key string) bool {
	now := time.Now()
	rl.mu.RLock()
	last, exists := rl.requests[key]
	rl.mu.RUnlock()

	if exists && now.Sub(last) < rl.window {
		return false
	}

	rl.mu.Lock()
	rl.requests[key] = now
	rl.mu.Unlock()
	return true
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

var globalRateLimiter = newRateLimiter(5*time.Minute, 2*time.Second)

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := getClientIP(r)
		if !globalRateLimiter.allow(key) {
			WriteJSONError(w, "Too many requests")
			return
		}

		next.ServeHTTP(w, r)
	})
}
