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
