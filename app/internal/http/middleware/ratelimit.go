package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"server/internal/config"
	"strings"
	"sync"
	"time"
)

// maxTrackedClients bounds the tracking map. Entries are only released when
// their window lapses, so without a ceiling a flood of distinct addresses grows
// the map until the next sweep - or without bound if they keep arriving.
const maxTrackedClients = 50_000

type RateLimiter struct {
	requests   map[string]*clientRequests
	mu         sync.RWMutex
	limit      int
	window     time.Duration
	maxClients int
}

type clientRequests struct {
	count     int
	firstSeen time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests:   make(map[string]*clientRequests),
		limit:      limit,
		window:     window,
		maxClients: maxTrackedClients,
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
		if !exists {
			rl.makeRoomLocked(now)
		}

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

// makeRoomLocked keeps the map under its ceiling before a new client is added.
// Expired entries go first; if that is not enough, the oldest surviving entry
// is dropped. Evicting beats refusing to track, which would let an attacker
// disable the limiter by flooding it with addresses.
//
// Callers must hold rl.mu.
func (rl *RateLimiter) makeRoomLocked(now time.Time) {
	// A non-positive ceiling means unbounded; NewRateLimiter always sets one.
	if rl.maxClients <= 0 || len(rl.requests) < rl.maxClients {
		return
	}

	for ip, client := range rl.requests {
		if now.Sub(client.firstSeen) > rl.window {
			delete(rl.requests, ip)
		}
	}

	if len(rl.requests) < rl.maxClients {
		return
	}

	oldestIP := ""
	var oldestSeen time.Time
	for ip, client := range rl.requests {
		if oldestIP == "" || client.firstSeen.Before(oldestSeen) {
			oldestIP, oldestSeen = ip, client.firstSeen
		}
	}

	if oldestIP != "" {
		delete(rl.requests, oldestIP)
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r, config.TrustedProxies())

		if !rl.isAllowed(ip) {
			slog.WarnContext(r.Context(), "Rate limit exceeded", "ip", ip, "path", r.URL.Path)
			http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientIP resolves the address to rate limit on. Forwarding headers are
// only honoured when the immediate peer is a configured trusted proxy: any
// client can set them, so trusting them unconditionally would let a caller
// rotate X-Forwarded-For and bypass the limit entirely.
func getClientIP(r *http.Request, trusted []netip.Prefix) string {
	peer := remoteAddrHost(r)

	if len(trusted) == 0 || !isTrustedProxy(peer, trusted) {
		return peer
	}

	// X-Forwarded-For is a chain, "client, proxy1, proxy2". Walk it from the
	// right and take the first address that is not itself a trusted proxy;
	// anything further left is attacker controlled.
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		hops := strings.Split(forwarded, ",")
		for i := len(hops) - 1; i >= 0; i-- {
			addr, err := netip.ParseAddr(strings.TrimSpace(hops[i]))
			if err != nil {
				continue
			}

			if !isTrustedProxy(addr.String(), trusted) {
				return addr.String()
			}
		}
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		if addr, err := netip.ParseAddr(realIP); err == nil {
			return addr.String()
		}
	}

	return peer
}

func remoteAddrHost(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

func isTrustedProxy(ip string, trusted []netip.Prefix) bool {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return false
	}

	for _, prefix := range trusted {
		if prefix.Contains(addr) {
			return true
		}
	}

	return false
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
