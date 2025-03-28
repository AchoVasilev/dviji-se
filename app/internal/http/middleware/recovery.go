package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"server/util/httputils"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("Caught panic: %v. Stack trace: %s", err, string(debug.Stack()))

				acceptHeader := req.Header.Get("Accept")
				if acceptHeader == "application/json" {
					req.Header.Set("Content-Type", "application/json")
					httputils.SendInternalServerResponse(writer, req)
				} else {
					req.Header.Set("Content-Type", "text/html; charset=utf-8")
					http.Redirect(writer, req, "/error", http.StatusSeeOther)
				}
			}
		}()

		next.ServeHTTP(writer, req)
	})
}
