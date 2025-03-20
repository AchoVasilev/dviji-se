package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"server/util/securityutil"
)

type contextKey string

var NonceKey contextKey = "nonces"

type Nonces struct {
	Htmx        string
	Tw          string
	HtmxCssHash string
}

func ContentSecurityPolicy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonces, found := getNonces(r.Context())
		if !found {
			nonces = Nonces{
				Htmx:        securityutil.GenerateRandomString(16),
				Tw:          securityutil.GenerateRandomString(16),
				HtmxCssHash: "sha-256-bsV5JivYxvGywDAZ22EZJKBFip65Ng9xoJVLbBg7bdo=",
			}
		}

		ctx := context.WithValue(r.Context(), NonceKey, nonces)
		cspHeader := fmt.Sprintf(`
		  default-src 'self'; 
		  script-src 'nonce-%s'; 
		  style-src 'nonce-%s' '%s'; 
		  img-src 'self' https:;
		  connect-src 'self'; 
		  frame-ancestors 'none'; 
		  form-action 'self'; 
		  object-src 'none';`, nonces.Htmx, nonces.Tw, nonces.HtmxCssHash)

		w.Header().Set("Content-Security-Policy", cspHeader)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getNonces(ctx context.Context) (Nonces, bool) {
	nonces, ok := ctx.Value(NonceKey).(Nonces)

	return nonces, ok
}

func GetHtmxNonce(ctx context.Context) string {
	nonces, ok := getNonces(ctx)
	if !ok {
		slog.Error("No HTMX nonce set!")
		panic(errors.New("No HTMX nonce set!"))
	}

	return nonces.Htmx
}

func GetTwNonce(ctx context.Context) string {
	nonces, ok := getNonces(ctx)
	if !ok {
		slog.Error("No TailwindCSS nonce set!")
		panic(errors.New("No TailwindCSS nonce set!"))
	}

	return nonces.Tw
}

func ContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ext := filepath.Ext(r.URL.Path)
		mimeType := mime.TypeByExtension(ext)

		if mimeType == "" {
			mimeType = "text/html; charset=utf-8"
		}

		w.Header().Set("Content-Type", mimeType)
		next.ServeHTTP(w, r)
	})
}
