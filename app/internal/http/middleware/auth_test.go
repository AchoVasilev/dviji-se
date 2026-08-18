package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"server/internal/domain/user"
	"server/util/ctxutils"
	"server/util/httputils"
	"server/util/securityutil"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_KEY", "test-jwt-secret-key-for-testing-only")
	os.Setenv("JWT_REFRESH_KEY", "test-jwt-refresh-secret-key-for-testing")
	os.Setenv("XSRF", "test-xsrf-key-for-testing")

	os.Exit(m.Run())
}

// allowAllSessions accepts every session, isolating these tests from the
// revocation lookup.
type allowAllSessions struct{}

func (allowAllSessions) IsSessionValid(context.Context, string, time.Time) bool { return true }

// denyAllSessions models a user whose tokens have all been revoked.
type denyAllSessions struct{}

func (denyAllSessions) IsSessionValid(context.Context, string, time.Time) bool { return false }

// checkAuthResult reports whether CheckAuth attached a user to the context.
func checkAuthResult(t *testing.T, req *http.Request) *securityutil.LoggedInUser {
	t.Helper()

	return checkAuthResultWith(t, req, allowAllSessions{})
}

func checkAuthResultWith(t *testing.T, req *http.Request, sessions SessionValidator) *securityutil.LoggedInUser {
	t.Helper()

	var attached *securityutil.LoggedInUser
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, err := ctxutils.GetUser(r.Context()); err == nil {
			attached = u
		}

		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	CheckAuth(sessions)(next).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("next handler was not reached, status = %d", w.Code)
	}

	return attached
}

func validAccessToken(t *testing.T) string {
	t.Helper()

	token, _ := securityutil.GenerateAccessToken(user.User{
		Id:    uuid.New(),
		Email: "test@example.com",
	}, false)

	return token
}

// A bearer header that fails to parse used to leave the cookie lookup holding a
// nil *http.Cookie, panicking on authCookie.Value.
func TestCheckAuth_InvalidBearerWithoutCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")

	if attached := checkAuthResult(t, req); attached != nil {
		t.Errorf("expected no user to be attached, got %+v", attached)
	}
}

func TestCheckAuth_NoCredentials(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if attached := checkAuthResult(t, req); attached != nil {
		t.Errorf("expected no user to be attached, got %+v", attached)
	}
}

func TestCheckAuth_ValidBearer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+validAccessToken(t))

	attached := checkAuthResult(t, req)
	if attached == nil {
		t.Fatal("expected a user to be attached")
	}

	if attached.Username != "test@example.com" {
		t.Errorf("Username = %q, want %q", attached.Username, "test@example.com")
	}
}

func TestCheckAuth_ValidCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  string(httputils.AuthCookieName),
		Value: validAccessToken(t),
	})

	if attached := checkAuthResult(t, req); attached == nil {
		t.Fatal("expected a user to be attached")
	}
}

// An unusable bearer header must not prevent a valid cookie from being used.
func TestCheckAuth_InvalidBearerFallsBackToCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	req.AddCookie(&http.Cookie{
		Name:  string(httputils.AuthCookieName),
		Value: validAccessToken(t),
	})

	if attached := checkAuthResult(t, req); attached == nil {
		t.Fatal("expected the cookie to be used when the bearer token is invalid")
	}
}

func TestCheckAuth_EmptyCookieValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  string(httputils.AuthCookieName),
		Value: "",
	})

	if attached := checkAuthResult(t, req); attached != nil {
		t.Errorf("expected no user to be attached, got %+v", attached)
	}
}

// A revoked token must not authenticate, even though its signature and expiry
// are still valid.
func TestCheckAuth_RejectsRevokedToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+validAccessToken(t))

	if attached := checkAuthResultWith(t, req, denyAllSessions{}); attached != nil {
		t.Errorf("expected a revoked token to be rejected, got %+v", attached)
	}
}

// countingSessions records how often the revocation check is consulted.
type countingSessions struct{ calls int }

func (c *countingSessions) IsSessionValid(context.Context, string, time.Time) bool {
	c.calls++
	return true
}

// Static assets are identical for everyone, and the auth cookie is sent with
// each one. Authenticating them made a page view cost one revocation lookup
// per asset.
func TestCheckAuth_SkipsStaticAssets(t *testing.T) {
	sessions := &countingSessions{}
	token := validAccessToken(t)

	for _, path := range []string{
		"/static/scripts/htmx.min.js",
		"/static/css/styles.css",
		"/static/img/logo-sm.png",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.AddCookie(&http.Cookie{Name: string(httputils.AuthCookieName), Value: token})

		checkAuthResultWith(t, req, sessions)
	}

	if sessions.calls != 0 {
		t.Errorf("revocation checked %d times for static assets, want 0", sessions.calls)
	}
}

func TestCheckAuth_StillAuthenticatesPageRequests(t *testing.T) {
	sessions := &countingSessions{}

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: string(httputils.AuthCookieName), Value: validAccessToken(t)})

	if attached := checkAuthResultWith(t, req, sessions); attached == nil {
		t.Fatal("expected a user to be attached for a page request")
	}

	if sessions.calls != 1 {
		t.Errorf("revocation checked %d times, want 1", sessions.calls)
	}
}
