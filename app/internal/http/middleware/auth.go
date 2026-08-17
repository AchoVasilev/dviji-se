package middleware

import (
	"net/http"
	"server/util/ctxutils"
	"server/util/httputils"
	"server/util/securityutil"
	"strings"
)

// CheckAuth attaches the logged in user to the context when a valid token is
// present. A missing or invalid token is not an error here: the request
// continues unauthenticated and RequireAuth decides whether that is allowed.
func CheckAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, token := range authTokens(r) {
			loggedInUser, err := securityutil.UserFromToken(token)
			if err != nil {
				continue
			}

			h.ServeHTTP(w, r.WithContext(ctxutils.WithLoggedUser(r.Context(), loggedInUser)))
			return
		}

		h.ServeHTTP(w, r)
	})
}

// authTokens returns the candidate tokens in precedence order: the bearer
// header first, then the auth cookie. Either may be absent.
func authTokens(r *http.Request) []string {
	var tokens []string

	if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
		if token := strings.TrimPrefix(authHeader, "Bearer "); token != "" {
			tokens = append(tokens, token)
		}
	}

	// A bearer header that fails to parse must not stop the cookie from being
	// tried - and the cookie may simply not be there.
	if authCookie, err := r.Cookie(string(httputils.AuthCookieName)); err == nil && authCookie.Value != "" {
		tokens = append(tokens, authCookie.Value)
	}

	return tokens
}
