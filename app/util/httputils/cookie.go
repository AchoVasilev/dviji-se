package httputils

import (
	"net/http"
	"time"
)

type AuthCookie string

var AuthCookieName AuthCookie = "dviji_se_login"

func SetHttpOnlyCookie(name string, value string, expirationTime time.Time, writer http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  expirationTime,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	}

	http.SetCookie(writer, cookie)
}
