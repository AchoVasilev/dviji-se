package middleware

import (
	"net/http"
	"server/util/ctxutils"
	"server/util/httputils"
	"server/util/securityutil"
	"strings"
)

func CheckAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		authCookie, err := r.Cookie(string(httputils.AuthCookieName))

		hasBearerPrefix := strings.HasPrefix(authHeader, "Bearer ")
		if !hasBearerPrefix && (authCookie == nil || err != nil) {
			h.ServeHTTP(w, r)
			return
		}

		if hasBearerPrefix {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			loggedInUser, err := securityutil.UserFromToken(token)
			if err != nil {
				h.ServeHTTP(w, r)
				return
			}

			h.ServeHTTP(w, r.WithContext(ctxutils.WithLoggedUser(r.Context(), loggedInUser)))
			return
		}

		if authCookie.Value != "" {
			loggedInUser, err := securityutil.UserFromToken(authCookie.Value)
			if err != nil {
				h.ServeHTTP(w, r.WithContext(ctxutils.WithLoggedUser(r.Context(), loggedInUser)))
				return
			}
		}

		h.ServeHTTP(w, r)
	})
}
