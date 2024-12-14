package middleware

import (
	"log/slog"
	"net/http"
	"os"
)

func AppendLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			AddSource: false,
		}))

		ctx := req.Context()
		reqId := RequestIdFromContext(ctx)

		logger = logger.With("requestId", reqId)

		slog.SetDefault(logger)

		next.ServeHTTP(writer, req)
	})
}
