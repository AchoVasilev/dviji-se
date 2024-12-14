package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"server/common/api"

	"github.com/google/uuid"
)

type CtxKeyRequestId string

const (
	requestIdHeader                     = "X-REQUEST-ID"
	contextKeyRequestId CtxKeyRequestId = "requestId"
)

func PopulateRequestId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		if fromClient := req.Header.Get(requestIdHeader); fromClient != "" {
			slog.Info(fmt.Sprintf("RequestId from client is present: %s", fromClient))
			id, err := newRequestId()
			if err != nil {
				api.SendInternalServerResponse(writer)
				return
			}

			ctx = withRequestId(ctx, id)
			req = req.WithContext(ctx)
		} else if existing := RequestIdFromContext(ctx); existing == "" {
			id, err := newRequestId()
			if err != nil {
				api.SendInternalServerResponse(writer)
				return
			}

			slog.Info(fmt.Sprintf("Appending new request id: %s", id))
			ctx = withRequestId(ctx, id)
			req = req.WithContext(ctx)
		}

		next.ServeHTTP(writer, req)
	})
}

func RequestIdFromContext(ctx context.Context) string {
	value := ctx.Value(contextKeyRequestId)
	if value == nil {
		return ""
	}

	id, ok := value.(string)
	if !ok {
		return ""
	}

	return id
}

func withRequestId(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKeyRequestId, id)
}

func newRequestId() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		slog.Error(fmt.Sprintf("Error generating new request id: Error: %v", err))
		return "", err
	}

	return id.String(), nil
}
