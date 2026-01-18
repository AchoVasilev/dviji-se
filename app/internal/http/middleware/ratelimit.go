package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	requests map[string]*clientRequests
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

type clientRequests struct {
	count     int
	firstSeen time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*clientRequests),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, client := range rl.requests {
			if now.Sub(client.firstSeen) > rl.window {
				delete(rl.requests, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) isAllowed(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.requests[ip]

	if !exists || now.Sub(client.firstSeen) > rl.window {
		rl.requests[ip] = &clientRequests{
			count:     1,
			firstSeen: now,
		}
		return true
	}

	if client.count >= rl.limit {
		return false
	}

	client.count++
	return true
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		if !rl.isAllowed(ip) {
			slog.Warn("Rate limit exceeded", "ip", ip, "path", r.URL.Path)
			http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the list
		if ip := net.ParseIP(xff); ip != nil {
			return ip.String()
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		if ip := net.ParseIP(xri); ip != nil {
			return ip.String()
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// Pre-configured rate limiters for different use cases

// AuthRateLimiter - strict rate limiting for auth endpoints (5 requests per minute)
func AuthRateLimiter() *RateLimiter {
	return NewRateLimiter(5, time.Minute)
}

// PasswordResetRateLimiter - very strict for password reset (3 requests per 5 minutes)
func PasswordResetRateLimiter() *RateLimiter {
	return NewRateLimiter(3, 5*time.Minute)
}

// APIRateLimiter - more permissive for general API (100 requests per minute)
func APIRateLimiter() *RateLimiter {
	return NewRateLimiter(100, time.Minute)
}
