package pkg

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	mu      sync.Mutex
	limits  map[string]*bucket
	maxReqs int
	window  time.Duration
}

type bucket struct {
	count      int
	lastReset time.Time
}

// NewRateLimiter creates a new rate limiter
// maxReqs: maximum requests allowed per window
// window: time window for rate limiting
func NewRateLimiter(maxReqs int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		limits:  make(map[string]*bucket),
		maxReqs: maxReqs,
		window:  window,
	}

	// Cleanup old buckets every minute
	go func() {
		ticker := time.NewTicker(time.Minute)
		for range ticker.C {
			rl.cleanupBuckets()
		}
	}()

	return rl
}

// Allow checks if a request from the given identifier should be allowed
func (rl *RateLimiter) Allow(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.limits[identifier]

	if !exists || now.Sub(b.lastReset) > rl.window {
		// Create new bucket or reset existing one
		rl.limits[identifier] = &bucket{
			count:      1,
			lastReset: now,
		}
		return true
	}

	if b.count < rl.maxReqs {
		b.count++
		return true
	}

	return false
}

// cleanupBuckets removes old buckets
func (rl *RateLimiter) cleanupBuckets() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, b := range rl.limits {
		if now.Sub(b.lastReset) > rl.window*2 {
			delete(rl.limits, key)
		}
	}
}

// Global rate limiter instances
var (
	loginLimiter = NewRateLimiter(5, time.Minute)   // 5 attempts per minute
	apiLimiter   = NewRateLimiter(100, time.Minute) // 100 requests per minute
)

// RateLimitLogin middleware for login endpoint
func RateLimitLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		if !loginLimiter.Allow(ip) {
			http.Error(w, "Too many login attempts. Please try again later.", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

// RateLimitAPI middleware for API endpoints
func RateLimitAPI(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		identifier := getClientIP(r)
		if !apiLimiter.Allow(identifier) {
			http.Error(w, "Rate limit exceeded. Maximum 100 requests per minute.", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}
