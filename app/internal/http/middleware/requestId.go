package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"server/util/ctxutils"
	"server/util/httputils"
)

type CtxKeyRequestId string

const (
	requestIdHeader = "X-Request-Id"
)

func PopulateRequestId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		if fromClient := req.Header.Get(requestIdHeader); fromClient != "" {
			slog.Info(fmt.Sprintf("RequestId from client is present: %s", fromClient))
			id, err := ctxutils.NewRequestId()
			if err != nil {
				httputils.SendInternalServerResponse(writer, req)
				return
			}

			ctx = ctxutils.WithRequestId(ctx, id)
			req = req.WithContext(ctx)
		} else if existing := ctxutils.RequestIdFromContext(ctx); existing == "" {
			id, err := ctxutils.NewRequestId()
			if err != nil {
				httputils.SendInternalServerResponse(writer, req)
				return
			}

			slog.Info(fmt.Sprintf("Appending new request id: %s", id))
			ctx = ctxutils.WithRequestId(ctx, id)
			req = req.WithContext(ctx)
		}

		next.ServeHTTP(writer, req)
	})
}
