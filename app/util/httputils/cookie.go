package httputils

import (
	"net/http"
	"time"
)

type CookieName string

const (
	AuthCookieName    CookieName = "X-LOGIN-TOKEN"
	RefreshCookieName CookieName = "X-REFRESH-TOKEN"
	XSRFCookieName    CookieName = "csrf_token"
	// CSRFSessionCookieName identifies an anonymous browser so its CSRF token
	// can be bound to something.
	CSRFSessionCookieName CookieName = "csrf_session"
)

// SetHttpOnlyCookie sets a persistent HTTP-only cookie with expiration
func SetHttpOnlyCookie(name CookieName, value string, expirationTime time.Time, writer http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     string(name),
		Value:    value,
		Expires:  expirationTime,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	}

	http.SetCookie(writer, cookie)
}

// RefreshTokenPath scopes the refresh cookie so it is only sent to the refresh
// endpoint rather than riding along on every request.
const RefreshTokenPath = "/refresh-token"

// ClearCookie expires a cookie by setting MaxAge to -1
func ClearCookie(name CookieName, writer http.ResponseWriter) {
	ClearCookieAtPath(name, "/", writer)
}

// ClearCookieAtPath expires a cookie set on a specific path. A cookie is only
// replaced by one with a matching path, so a path scoped cookie cannot be
// cleared through the root path.
func ClearCookieAtPath(name CookieName, path string, writer http.ResponseWriter) {
	http.SetCookie(writer, &http.Cookie{
		Name:     string(name),
		Value:    "",
		Path:     path,
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	})
}

// SetRefreshCookie stores the refresh token, scoped to the refresh endpoint.
func SetRefreshCookie(value string, expirationTime time.Time, writer http.ResponseWriter) {
	http.SetCookie(writer, &http.Cookie{
		Name:     string(RefreshCookieName),
		Value:    value,
		Path:     RefreshTokenPath,
		Expires:  expirationTime,
		MaxAge:   int(time.Until(expirationTime).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	})
}

// SetAuthCookie sets an auth cookie - persistent if rememberMe is true, session cookie otherwise
func SetAuthCookie(name CookieName, value string, expirationTime time.Time, rememberMe bool, writer http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     string(name),
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	}

	// Only set Expires for "Remember Me" - otherwise it's a session cookie
	if rememberMe {
		cookie.Expires = expirationTime
		cookie.MaxAge = int(time.Until(expirationTime).Seconds())
	}

	http.SetCookie(writer, cookie)
}
