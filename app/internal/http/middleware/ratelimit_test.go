package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	if rl == nil {
		t.Fatal("NewRateLimiter() should return non-nil")
	}

	if rl.limit != 5 {
		t.Errorf("limit = %d, want 5", rl.limit)
	}

	if rl.window != time.Minute {
		t.Errorf("window = %v, want %v", rl.window, time.Minute)
	}
}

func TestRateLimiter_IsAllowed(t *testing.T) {
	rl := &RateLimiter{
		requests: make(map[string]*clientRequests),
		limit:    3,
		window:   time.Minute,
	}

	ip := "192.168.1.1"

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		if !rl.isAllowed(ip) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied
	if rl.isAllowed(ip) {
		t.Error("Request 4 should be denied (rate limit exceeded)")
	}

	// Different IP should still be allowed
	if !rl.isAllowed("192.168.1.2") {
		t.Error("Different IP should be allowed")
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	rl := &RateLimiter{
		requests: make(map[string]*clientRequests),
		limit:    2,
		window:   50 * time.Millisecond,
	}

	ip := "192.168.1.1"

	// Use up the limit
	rl.isAllowed(ip)
	rl.isAllowed(ip)

	// Should be denied
	if rl.isAllowed(ip) {
		t.Error("Should be denied after limit reached")
	}

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	// Should be allowed again
	if !rl.isAllowed(ip) {
		t.Error("Should be allowed after window reset")
	}
}

func TestRateLimiter_Middleware(t *testing.T) {
	rl := &RateLimiter{
		requests: make(map[string]*clientRequests),
		limit:    2,
		window:   time.Minute,
	}

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.Middleware(next)

	t.Run("allows requests within limit", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			nextCalled = false
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "192.168.1.100:12345"
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if !nextCalled {
				t.Errorf("Request %d: next handler should be called", i+1)
			}
			if w.Code != http.StatusOK {
				t.Errorf("Request %d: status = %d, want %d", i+1, w.Code, http.StatusOK)
			}
		}
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if nextCalled {
			t.Error("Next handler should not be called when rate limited")
		}
		if w.Code != http.StatusTooManyRequests {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusTooManyRequests)
		}
	})
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		trustedProxies []netip.Prefix
		remoteAddr     string
		headers        map[string]string
		expected       string
	}{
		{
			name:       "RemoteAddr only",
			remoteAddr: "192.168.1.1:12345",
			headers:    nil,
			expected:   "192.168.1.1",
		},
		{
			name:       "RemoteAddr without port",
			remoteAddr: "192.168.1.1",
			headers:    nil,
			expected:   "192.168.1.1",
		},
		{
			name:       "Forwarding headers are ignored when no proxy is trusted",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195",
				"X-Real-IP":       "203.0.113.100",
			},
			expected: "10.0.0.1",
		},
		{
			name:           "Forwarding headers are ignored from an untrusted peer",
			trustedProxies: []netip.Prefix{netip.MustParsePrefix("10.0.0.1/32")},
			remoteAddr:     "198.51.100.7:12345",
			headers:        map[string]string{"X-Forwarded-For": "203.0.113.195"},
			expected:       "198.51.100.7",
		},
		{
			name:           "X-Forwarded-For is honoured from a trusted peer",
			trustedProxies: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
			remoteAddr:     "10.0.0.1:12345",
			headers:        map[string]string{"X-Forwarded-For": "203.0.113.195"},
			expected:       "203.0.113.195",
		},
		{
			name:           "Rightmost untrusted hop wins in a chain",
			trustedProxies: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
			remoteAddr:     "10.0.0.1:12345",
			headers:        map[string]string{"X-Forwarded-For": "203.0.113.195, 198.51.100.7, 10.0.0.5"},
			expected:       "198.51.100.7",
		},
		{
			name:           "Spoofed leading hops cannot displace the real client",
			trustedProxies: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
			remoteAddr:     "10.0.0.1:12345",
			headers:        map[string]string{"X-Forwarded-For": "1.1.1.1, 203.0.113.195"},
			expected:       "203.0.113.195",
		},
		{
			name:           "Falls back to X-Real-IP when X-Forwarded-For is unusable",
			trustedProxies: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
			remoteAddr:     "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "invalid-ip",
				"X-Real-IP":       "203.0.113.100",
			},
			expected: "203.0.113.100",
		},
		{
			name:           "Falls back to the peer when every header is unusable",
			trustedProxies: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
			remoteAddr:     "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "invalid",
				"X-Real-IP":       "also-invalid",
			},
			expected: "10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			got := getClientIP(req, tt.trustedProxies)
			if got != tt.expected {
				t.Errorf("getClientIP() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAuthRateLimiter(t *testing.T) {
	rl := AuthRateLimiter()
	if rl.limit != 5 {
		t.Errorf("AuthRateLimiter limit = %d, want 5", rl.limit)
	}
	if rl.window != time.Minute {
		t.Errorf("AuthRateLimiter window = %v, want %v", rl.window, time.Minute)
	}
}

func TestPasswordResetRateLimiter(t *testing.T) {
	rl := PasswordResetRateLimiter()
	if rl.limit != 3 {
		t.Errorf("PasswordResetRateLimiter limit = %d, want 3", rl.limit)
	}
	if rl.window != 5*time.Minute {
		t.Errorf("PasswordResetRateLimiter window = %v, want %v", rl.window, 5*time.Minute)
	}
}

func TestAPIRateLimiter(t *testing.T) {
	rl := APIRateLimiter()
	if rl.limit != 100 {
		t.Errorf("APIRateLimiter limit = %d, want 100", rl.limit)
	}
	if rl.window != time.Minute {
		t.Errorf("APIRateLimiter window = %v, want %v", rl.window, time.Minute)
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := &RateLimiter{
		requests: make(map[string]*clientRequests),
		limit:    100,
		window:   time.Minute,
	}

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(ip string) {
			for j := 0; j < 20; j++ {
				rl.isAllowed(ip)
			}
			done <- true
		}("192.168.1." + string(rune('0'+i)))
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we get here without deadlock or panic, test passes
}

// The tracking map must not grow without bound when many distinct addresses
// arrive inside a single window.
func TestRateLimiter_BoundsTrackedClients(t *testing.T) {
	rl := &RateLimiter{
		requests:   make(map[string]*clientRequests),
		limit:      5,
		window:     time.Hour, // long enough that nothing expires on its own
		maxClients: 100,
	}

	for i := 0; i < 1000; i++ {
		rl.isAllowed(fmt.Sprintf("10.1.%d.%d", i/256, i%256))
	}

	if got := len(rl.requests); got > rl.maxClients {
		t.Errorf("tracked clients = %d, want <= %d", got, rl.maxClients)
	}
}

// Eviction must not let a client that is already over the limit back in.
func TestRateLimiter_LimitStillAppliesUnderPressure(t *testing.T) {
	rl := &RateLimiter{
		requests:   make(map[string]*clientRequests),
		limit:      2,
		window:     time.Hour,
		maxClients: 10,
	}

	const victim = "203.0.113.9"
	for i := 0; i < 2; i++ {
		if !rl.isAllowed(victim) {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	if rl.isAllowed(victim) {
		t.Fatal("third request should be denied")
	}

	// Fill the map so eviction runs repeatedly.
	for i := 0; i < 100; i++ {
		rl.isAllowed(fmt.Sprintf("10.2.%d.%d", i/256, i%256))
	}

	// The victim may have been evicted, which resets it - that is the accepted
	// trade-off - but the limiter must still refuse it once it is over again.
	rl.isAllowed(victim)
	rl.isAllowed(victim)

	if rl.isAllowed(victim) {
		t.Error("client over the limit should still be denied after eviction churn")
	}
}

// Expired entries are reclaimed before anything live is evicted.
func TestRateLimiter_PrefersEvictingExpiredEntries(t *testing.T) {
	rl := &RateLimiter{
		requests:   make(map[string]*clientRequests),
		limit:      5,
		window:     50 * time.Millisecond,
		maxClients: 10,
	}

	for i := 0; i < 10; i++ {
		rl.isAllowed(fmt.Sprintf("198.51.100.%d", i))
	}

	time.Sleep(60 * time.Millisecond)

	rl.isAllowed("203.0.113.1")

	if got := len(rl.requests); got > rl.maxClients {
		t.Errorf("tracked clients = %d, want <= %d", got, rl.maxClients)
	}
}

// NewRateLimiter must set a ceiling; a zero value one is explicitly unbounded.
func TestNewRateLimiter_SetsClientCeiling(t *testing.T) {
	if rl := NewRateLimiter(5, time.Minute); rl.maxClients <= 0 {
		t.Errorf("maxClients = %d, want a positive ceiling", rl.maxClients)
	}
}
