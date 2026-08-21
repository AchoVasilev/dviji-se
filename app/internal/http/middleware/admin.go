package middleware

import (
	"net/http"
	"server/internal/domain/user"
	"server/util/ctxutils"
)

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggedUser, err := ctxutils.GetUser(r.Context())
		if err != nil || loggedUser == nil {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggedUser, err := ctxutils.GetUser(r.Context())
		if err != nil || loggedUser == nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if !user.HasRole(loggedUser.Roles, user.RoleAdmin) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
