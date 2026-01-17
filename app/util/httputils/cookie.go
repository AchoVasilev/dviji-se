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
