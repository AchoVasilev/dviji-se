package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"server/internal/domain/user"
	"server/util/ctxutils"
	"server/util/httputils"
	"server/util/securityutil"
	"testing"

	"github.com/google/uuid"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_KEY", "test-jwt-secret-key-for-testing-only")
	os.Setenv("JWT_REFRESH_KEY", "test-jwt-refresh-secret-key-for-testing")
	os.Setenv("XSRF", "test-xsrf-key-for-testing")

	os.Exit(m.Run())
}

// checkAuthResult reports whether CheckAuth attached a user to the context.
func checkAuthResult(t *testing.T, req *http.Request) *securityutil.LoggedInUser {
	t.Helper()

	var attached *securityutil.LoggedInUser
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, err := ctxutils.GetUser(r.Context()); err == nil {
			attached = u
		}

		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	CheckAuth(next).ServeHTTP(w, req)

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
