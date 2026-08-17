package middleware

import (
	"context"
	"log/slog"
	"os"
	"server/util/ctxutils"
)

var baseHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	AddSource: false,
})

// requestIdHandler stamps the request id onto every record that carries a
// context holding one. Resolving it per record - rather than binding it to a
// logger - is what makes request correlation safe under concurrency.
type requestIdHandler struct {
	slog.Handler
}

func (h requestIdHandler) Handle(ctx context.Context, record slog.Record) error {
	if reqId := ctxutils.RequestIdFromContext(ctx); reqId != "" {
		record.AddAttrs(slog.String("requestId", reqId))
	}

	return h.Handler.Handle(ctx, record)
}

func (h requestIdHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return requestIdHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h requestIdHandler) WithGroup(name string) slog.Handler {
	return requestIdHandler{Handler: h.Handler.WithGroup(name)}
}

// InitLogger installs the default logger once, at startup. It must not be
// called per request: slog.SetDefault mutates global state, so doing that in a
// middleware both races and attributes ids to the wrong requests.
func InitLogger() {
	slog.SetDefault(slog.New(requestIdHandler{Handler: baseHandler}))
}
