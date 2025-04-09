package httputils

import (
	"net/http"
	"time"
)

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
