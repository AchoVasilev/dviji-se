package middleware

import (
	"net/http"
	"server/util/ctxutils"
)

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := ctxutils.GetUser(r.Context())
		if err != nil || user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := ctxutils.GetUser(r.Context())
		if err != nil || user == nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		isAdmin := false
		for _, role := range user.Roles {
			if role.Name == "ADMIN" {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
