package api

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimitConfig holds rate limiter settings.
type RateLimitConfig struct {
	// Rate is the number of requests per second allowed.
	Rate float64
	// Burst is the maximum burst size.
	Burst int
}

// ipRateLimiter manages per-IP rate limiters.
type ipRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
}

func newIPRateLimiter(r float64, burst int) *ipRateLimiter {
	return &ipRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(r),
		burst:    burst,
	}
}

func (l *ipRateLimiter) getLimiter(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, ok := l.limiters[ip]
	if !ok {
		limiter = rate.NewLimiter(l.rate, l.burst)
		l.limiters[ip] = limiter
	}
	return limiter
}

// RateLimitMiddleware returns middleware that rate-limits requests to the
// specified paths by client IP address.
func RateLimitMiddleware(cfg RateLimitConfig, paths map[string]bool, next http.Handler) http.Handler {
	limiter := newIPRateLimiter(cfg.Rate, cfg.Burst)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !paths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		ip := clientIP(r)
		if !limiter.getLimiter(ip).Allow() {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", 1))
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// clientIP extracts the client IP from the request.
func clientIP(r *http.Request) string {
	// Use RemoteAddr, stripping port.
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
