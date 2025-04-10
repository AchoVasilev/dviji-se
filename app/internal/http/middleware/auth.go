package middleware

import (
	"net/http"
	"server/util/httputils"
	"strings"
)

func CheckAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		authCookie, err := r.Cookie(string(httputils.AuthCookieName))

		if !strings.HasPrefix(authHeader, "Bearer ") && (authCookie == nil || err != nil) {
			h.ServeHTTP(w, r)
			return
		}
	}
}
