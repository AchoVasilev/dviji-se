package middleware

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"server/util/securityutil"
	"strings"
	"time"

	"golang.org/x/net/xsrftoken"
)

type contextKey string

var NonceKey contextKey = "nonces"

type Nonces struct {
	Htmx        string
	Tw          string
	HtmxCssHash string
}

var XSRFKey contextKey = "xsrf"

func CSRFCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			xsrfCookie, err := r.Cookie("csrf_token")
			if err != nil || xsrfCookie == nil {
				key := os.Getenv("XSRF")
				csrfToken := xsrftoken.Generate(key, "", http.MethodPost)
				http.SetCookie(w, &http.Cookie{
					Name:     "csrf_token",
					Value:    csrfToken,
					Expires:  time.Now().Add(24 * time.Hour),
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
				})

				w.Header().Set("X-CSRF-TOKEN", csrfToken)
				ctx := context.WithValue(r.Context(), XSRFKey, csrfToken)
				r = r.WithContext(ctx)
			}

			if xsrfCookie != nil {
				ctx := context.WithValue(r.Context(), XSRFKey, xsrfCookie.Value)
				r = r.WithContext(ctx)
			}
		}

		next.ServeHTTP(w, r)
	})
}

func GetCSRF(ctx context.Context) string {
	token, ok := ctx.Value(XSRFKey).(string)
	if !ok {
		return ""
	}

	return token
}

func CSRFValidate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("csrf_token")
		if err != nil {
			http.Error(w, "Access forbidden", http.StatusForbidden)
			slog.Warn("No CSRF token")
			if !strings.HasPrefix(r.URL.Path, "/api") {
				w.Header().Set("HX-Redirect", "/error")
			}

			return
		}

		key := os.Getenv("XSRF")
		isValid := xsrftoken.Valid(cookie.Value, key, "", http.MethodPost)
		if !isValid {
			http.Error(w, "Access forbidden", http.StatusForbidden)
			slog.Warn("Invalid CSRF token")

			if !strings.HasPrefix(r.URL.Path, "/api") {
				w.Header().Set("HX-Redirect", "/error")
			}

			return
		}

		csrfHeader := r.Header.Get("X-CSRF-Token")
		if csrfHeader == "" {
			http.Error(w, "Access forbidden", http.StatusForbidden)
			slog.Warn("Invalid CSRF token")

			if !strings.HasPrefix(r.URL.Path, "/api") {
				w.Header().Set("HX-Redirect", "/error")
			}

			return
		}

		isValid = xsrftoken.Valid(csrfHeader, key, "", http.MethodPost)
		if !isValid {
			http.Error(w, "Access forbidden", http.StatusForbidden)
			slog.Warn("Invalid CSRF token")

			if !strings.HasPrefix(r.URL.Path, "/api") {
				w.Header().Set("HX-Redirect", "/error")
			}

			return
		}

		next.ServeHTTP(w, r)
	})
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

		if strings.HasPrefix(r.URL.Path, "/api/") && mimeType == "" {
			mimeType = "application/json"
		}

		if mimeType == "" {
			mimeType = "text/html; charset=utf-8"
		}

		w.Header().Set("Content-Type", mimeType)
		next.ServeHTTP(w, r)
	})
}

func CacheStaticAssets(next http.Handler, staticDir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(staticDir, r.URL.Path)

		etag, err := generateETag(filePath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if match := r.Header.Get("If-None-Match"); match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		ext := strings.ToLower(filepath.Ext(r.URL.Path))
		switch ext {
		case ".css", ".js":
			w.Header().Set("Cache-Control", "public, max-age=86400") // 1 day
		case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
			w.Header().Set("Cache-Control", "public, max-age=31556952, immutable") // 1 year
		case ".html":
			w.Header().Set("Cache-Control", "no-cache") // Always request fresh
		default:
			w.Header().Set("Cache-Control", "public, max-age=3600") // Default 1 hour
		}

		w.Header().Set("ETag", etag)
		w.Header().Set("Vary", "Accept-Encoding")

		next.ServeHTTP(w, r)
	})
}

func generateETag(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
