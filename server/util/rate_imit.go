package util

import (
	"net/http"
	"sync"
	"time"
)

var rateLimiter = struct {
	sync.Mutex
	lastRequests map[string]time.Time
}{lastRequests: make(map[string]time.Time)}

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.RemoteAddr + r.UserAgent()
		rateLimiter.Lock()

		last, exists := rateLimiter.lastRequests[key]
		now := time.Now()

		if exists && now.Sub(last) < 2*time.Second {
			rateLimiter.Unlock()
			WriteJSONError(w, "Too many requests", http.StatusTooManyRequests)

			return
		}

		rateLimiter.lastRequests[key] = now
		rateLimiter.Unlock()

		next.ServeHTTP(w, r)
	})
}
