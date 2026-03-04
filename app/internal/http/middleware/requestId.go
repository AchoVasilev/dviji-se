package middleware

import (
	"net/http"
	"server/util/ctxutils"
	"server/util/httputils"
)

func PopulateRequestId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		if existing := ctxutils.RequestIdFromContext(ctx); existing != "" {
			next.ServeHTTP(writer, req)
			return
		}

		id, err := ctxutils.NewRequestId()
		if err != nil {
			httputils.SendInternalServerResponse(writer, req)
			return
		}

		ctx = ctxutils.WithRequestId(ctx, id)
		req = req.WithContext(ctx)

		next.ServeHTTP(writer, req)
	})
}
