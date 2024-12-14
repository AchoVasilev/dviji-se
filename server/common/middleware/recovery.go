package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"server/common/api"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("Caught panic: %v. Stack trace: %s", err, string(debug.Stack()))

				api.SendInternalServerResponse(writer, req)
			}
		}()

		next.ServeHTTP(writer, req)
		return
	})
}
