package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"server/util/ctxutils"
)

var baseHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	AddSource: false,
})

func AppendLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		reqId := ctxutils.RequestIdFromContext(ctx)

		logger := slog.New(baseHandler).With("requestId", reqId)
		slog.SetDefault(logger)

		next.ServeHTTP(writer, req)
	})
}
