package httpdelivery

import (
	"net/http"
	"sync"
	"time"

	apperrors "food_delivery_platform/shared/errors"

	"golang.org/x/time/rate"
)

const codeTooManyRequests = "TOO_MANY_REQUESTS"

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// IPRateLimiter controla a taxa de requisições por endereço IP usando token bucket.
// Entradas inativas por mais de ttl são removidas periodicamente pelo cleanupLoop.
type IPRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*limiterEntry
	r       rate.Limit
	burst   int
	ttl     time.Duration
}

func newIPRateLimiter(r rate.Limit, burst int) *IPRateLimiter {
	rl := &IPRateLimiter{
		entries: make(map[string]*limiterEntry),
		r:       r,
		burst:   burst,
		ttl:     10 * time.Minute,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	e, ok := rl.entries[ip]
	if !ok {
		e = &limiterEntry{limiter: rate.NewLimiter(rl.r, rl.burst)}
		rl.entries[ip] = e
	}
	e.lastSeen = time.Now()
	return e.limiter
}

func (rl *IPRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.evictStale()
	}
}

func (rl *IPRateLimiter) evictStale() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	cutoff := time.Now().Add(-rl.ttl)
	for ip, e := range rl.entries {
		if e.lastSeen.Before(cutoff) {
			delete(rl.entries, ip)
		}
	}
}

// Middleware retorna um http.Handler que rejeita com 429 quando o IP esgota o limite.
func (rl *IPRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractClientIP(r.RemoteAddr)
		if !rl.getLimiter(ip).Allow() {
			writeAppError(w, r, apperrors.New(
				codeTooManyRequests,
				"rate limit exceeded, please try again later",
				http.StatusTooManyRequests,
				nil,
			))
			return
		}
		next.ServeHTTP(w, r)
	})
}
