package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"server/util/httputils"
	"strings"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.ErrorContext(req.Context(), "Caught panic", "panic", err, "stack", string(debug.Stack()))

				// Content-Type belongs on the response. Setting it on req did
				// nothing, and SendInternalServerResponse writes its own anyway.
				if strings.Contains(req.Header.Get("Accept"), "application/json") {
					httputils.SendInternalServerResponse(writer, req)
					return
				}

				writer.Header().Set("Content-Type", "text/html; charset=utf-8")
				http.Redirect(writer, req, "/error", http.StatusSeeOther)
			}
		}()

		next.ServeHTTP(writer, req)
	})
}
