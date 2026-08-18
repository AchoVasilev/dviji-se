package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"server/util/ctxutils"
	"server/util/httputils"
	"server/util/securityutil"
	"strings"
	"time"
)

// SessionValidator reports whether a token that is otherwise valid has been
// revoked - for instance by a password change.
type SessionValidator interface {
	IsSessionValid(ctx context.Context, userId string, issuedAt time.Time) bool
}

// CheckAuth attaches the logged in user to the context when a valid token is
// present. A missing or invalid token is not an error here: the request
// continues unauthenticated and RequireAuth decides whether that is allowed.
//
// The validator is consulted only when a well formed token is found, so
// anonymous traffic costs no lookup.
func CheckAuth(sessions SessionValidator) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Static files are served the same way to everyone, and the auth
			// cookie rides along on every one of them. Authenticating them
			// turned a single page view into one revocation lookup per asset.
			if isStaticAssetPath(r.URL.Path) {
				h.ServeHTTP(w, r)
				return
			}

			for _, token := range authTokens(r) {
				loggedInUser, err := securityutil.UserFromToken(token)
				if err != nil {
					continue
				}

				if !sessions.IsSessionValid(r.Context(), loggedInUser.Id, loggedInUser.IssuedAt) {
					slog.InfoContext(r.Context(), "Rejected a revoked token", "userId", loggedInUser.Id)
					continue
				}

				h.ServeHTTP(w, r.WithContext(ctxutils.WithLoggedUser(r.Context(), loggedInUser)))
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

// isStaticAssetPath reports whether the request is for a file under /static,
// which is public and identical for every visitor.
func isStaticAssetPath(path string) bool {
	return strings.HasPrefix(path, "/static/")
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
