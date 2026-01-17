package ctxutils

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"server/util/securityutil"

	"github.com/google/uuid"
)

type contextKey string

const (
	contextKeyRequestId contextKey = "requestId"
	xsrfKey             contextKey = "xsrf"
	loggedUser          contextKey = "user"
)

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

func WithRequestId(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKeyRequestId, id)
}

func WithCSRFToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, xsrfKey, token)
}

func WithLoggedUser(ctx context.Context, user *securityutil.LoggedInUser) context.Context {
	return context.WithValue(ctx, loggedUser, user)
}

func NewRequestId() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		slog.Error(fmt.Sprintf("Error generating new request id: Error: %v", err))
		return "", err
	}

	return id.String(), nil
}

func GetCSRF(ctx context.Context) string {
	token, ok := ctx.Value(xsrfKey).(string)
	if !ok {
		return ""
	}

	return token
}

func GetUser(ctx context.Context) (*securityutil.LoggedInUser, error) {
	user, ok := ctx.Value(loggedUser).(*securityutil.LoggedInUser)
	if !ok || user == nil {
		return nil, errors.New("User not logged in")
	}

	return user, nil
}
