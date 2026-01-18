package httputils

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCookieNames(t *testing.T) {
	tests := []struct {
		name     CookieName
		expected string
	}{
		{AuthCookieName, "X-LOGIN-TOKEN"},
		{RefreshCookieName, "X-REFRESH-TOKEN"},
		{XSRFCookieName, "csrf_token"},
	}

	for _, tt := range tests {
		if string(tt.name) != tt.expected {
			t.Errorf("%v = %q, want %q", tt.name, string(tt.name), tt.expected)
		}
	}
}

func TestSetHttpOnlyCookie(t *testing.T) {
	w := httptest.NewRecorder()
	expiration := time.Now().Add(24 * time.Hour)

	SetHttpOnlyCookie(AuthCookieName, "test-token", expiration, w)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]

	t.Run("name is correct", func(t *testing.T) {
		if cookie.Name != string(AuthCookieName) {
			t.Errorf("Cookie name = %q, want %q", cookie.Name, AuthCookieName)
		}
	})

	t.Run("value is correct", func(t *testing.T) {
		if cookie.Value != "test-token" {
			t.Errorf("Cookie value = %q, want %q", cookie.Value, "test-token")
		}
	})

	t.Run("path is root", func(t *testing.T) {
		if cookie.Path != "/" {
			t.Errorf("Cookie path = %q, want /", cookie.Path)
		}
	})

	t.Run("httponly is true", func(t *testing.T) {
		if !cookie.HttpOnly {
			t.Error("Cookie should be HttpOnly")
		}
	})

	t.Run("secure is true", func(t *testing.T) {
		if !cookie.Secure {
			t.Error("Cookie should be Secure")
		}
	})

	t.Run("samesite is strict", func(t *testing.T) {
		if cookie.SameSite != http.SameSiteStrictMode {
			t.Errorf("Cookie SameSite = %v, want %v", cookie.SameSite, http.SameSiteStrictMode)
		}
	})

	t.Run("expiration is set", func(t *testing.T) {
		// Allow for small time differences
		diff := cookie.Expires.Sub(expiration)
		if diff > time.Second || diff < -time.Second {
			t.Errorf("Cookie expiration = %v, want approximately %v", cookie.Expires, expiration)
		}
	})
}

func TestSetAuthCookie_WithRememberMe(t *testing.T) {
	w := httptest.NewRecorder()
	expiration := time.Now().Add(7 * 24 * time.Hour)

	SetAuthCookie(AuthCookieName, "test-token", expiration, true, w)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]

	t.Run("name is correct", func(t *testing.T) {
		if cookie.Name != string(AuthCookieName) {
			t.Errorf("Cookie name = %q, want %q", cookie.Name, AuthCookieName)
		}
	})

	t.Run("value is correct", func(t *testing.T) {
		if cookie.Value != "test-token" {
			t.Errorf("Cookie value = %q, want %q", cookie.Value, "test-token")
		}
	})

	t.Run("expiration is set for rememberMe", func(t *testing.T) {
		if cookie.Expires.IsZero() {
			t.Error("Cookie should have expiration for rememberMe=true")
		}
	})

	t.Run("maxAge is set for rememberMe", func(t *testing.T) {
		if cookie.MaxAge <= 0 {
			t.Error("Cookie MaxAge should be positive for rememberMe=true")
		}
	})

	t.Run("security settings", func(t *testing.T) {
		if !cookie.HttpOnly {
			t.Error("Cookie should be HttpOnly")
		}
		if !cookie.Secure {
			t.Error("Cookie should be Secure")
		}
		if cookie.SameSite != http.SameSiteStrictMode {
			t.Errorf("Cookie SameSite = %v, want %v", cookie.SameSite, http.SameSiteStrictMode)
		}
	})
}

func TestSetAuthCookie_WithoutRememberMe(t *testing.T) {
	w := httptest.NewRecorder()
	expiration := time.Now().Add(15 * time.Minute)

	SetAuthCookie(AuthCookieName, "session-token", expiration, false, w)

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]

	t.Run("name is correct", func(t *testing.T) {
		if cookie.Name != string(AuthCookieName) {
			t.Errorf("Cookie name = %q, want %q", cookie.Name, AuthCookieName)
		}
	})

	t.Run("value is correct", func(t *testing.T) {
		if cookie.Value != "session-token" {
			t.Errorf("Cookie value = %q, want %q", cookie.Value, "session-token")
		}
	})

	t.Run("is session cookie (no expiration)", func(t *testing.T) {
		// Session cookies have zero time for Expires
		if !cookie.Expires.IsZero() && cookie.MaxAge != 0 {
			t.Error("Session cookie should not have Expires or MaxAge set")
		}
	})

	t.Run("security settings", func(t *testing.T) {
		if !cookie.HttpOnly {
			t.Error("Cookie should be HttpOnly")
		}
		if !cookie.Secure {
			t.Error("Cookie should be Secure")
		}
		if cookie.SameSite != http.SameSiteStrictMode {
			t.Errorf("Cookie SameSite = %v, want %v", cookie.SameSite, http.SameSiteStrictMode)
		}
	})
}

func TestSetHttpOnlyCookie_DifferentCookieTypes(t *testing.T) {
	tests := []struct {
		name       CookieName
		value      string
		expiration time.Duration
	}{
		{AuthCookieName, "auth-token-value", 15 * time.Minute},
		{RefreshCookieName, "refresh-token-value", 24 * time.Hour},
		{XSRFCookieName, "csrf-token-value", 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			w := httptest.NewRecorder()
			expiration := time.Now().Add(tt.expiration)

			SetHttpOnlyCookie(tt.name, tt.value, expiration, w)

			cookies := w.Result().Cookies()
			if len(cookies) != 1 {
				t.Fatalf("Expected 1 cookie, got %d", len(cookies))
			}

			if cookies[0].Name != string(tt.name) {
				t.Errorf("Cookie name = %q, want %q", cookies[0].Name, tt.name)
			}

			if cookies[0].Value != tt.value {
				t.Errorf("Cookie value = %q, want %q", cookies[0].Value, tt.value)
			}
		})
	}
}
