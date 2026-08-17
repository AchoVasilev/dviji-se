package middleware

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"server/internal/config"
	"server/util/ctxutils"
	"server/util/httputils"
	"strings"
	"time"

	"golang.org/x/net/xsrftoken"
)

// csrfTokenTTL matches the lifetime baked into the xsrftoken value itself.
const csrfTokenTTL = 24 * time.Hour

// csrfIdentity returns the principal a CSRF token is bound to. Binding matters
// because an attacker who can plant cookies (a subdomain, a MITM on plain HTTP)
// could otherwise submit their own cookie and header pair and pass validation.
// A token minted for one principal is useless to another.
//
// Logged in requests bind to the user id. Anonymous ones bind to a per browser
// id so that flows like login are protected too.
func csrfIdentity(r *http.Request) string {
	if user, err := ctxutils.GetUser(r.Context()); err == nil && user != nil {
		return "user:" + user.Id
	}

	if sessionCookie, err := r.Cookie(string(httputils.CSRFSessionCookieName)); err == nil && sessionCookie.Value != "" {
		return "anon:" + sessionCookie.Value
	}

	return ""
}

func CSRFCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		identity := csrfIdentity(r)
		if identity == "" {
			// No anonymous session yet; start one so the token has something to
			// bind to, and use it for the token minted below.
			sessionId, err := newCSRFSessionId()
			if err != nil {
				slog.ErrorContext(r.Context(), "Could not create CSRF session id", "error", err)
				next.ServeHTTP(w, r)
				return
			}

			httputils.SetHttpOnlyCookie(httputils.CSRFSessionCookieName, sessionId, time.Now().Add(csrfTokenTTL), w)
			identity = "anon:" + sessionId
		}

		csrfToken := ""
		if cookie, err := r.Cookie(string(httputils.XSRFCookieName)); err == nil && cookie.Value != "" {
			csrfToken = cookie.Value
		}

		// Mint a fresh token when there is none, when it expired, or when the
		// identity changed - which is what happens across login and logout.
		if csrfToken == "" || !xsrftoken.Valid(csrfToken, config.XSRFKey(), identity, "") {
			csrfToken = xsrftoken.Generate(config.XSRFKey(), identity, "")
			httputils.SetHttpOnlyCookie(httputils.XSRFCookieName, csrfToken, time.Now().Add(csrfTokenTTL), w)
			w.Header().Set("X-CSRF-TOKEN", csrfToken)
		}

		next.ServeHTTP(w, r.WithContext(ctxutils.WithCSRFToken(r.Context(), csrfToken)))
	})
}

func CSRFValidate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie(string(httputils.XSRFCookieName))
		if err != nil {
			csrfError(w, r, "No CSRF token")
			return
		}

		identity := csrfIdentity(r)
		if identity == "" {
			csrfError(w, r, "No CSRF identity")
			return
		}

		xsrfKey := config.XSRFKey()
		if !xsrftoken.Valid(cookie.Value, xsrfKey, identity, "") {
			csrfError(w, r, "Invalid CSRF cookie")
			return
		}

		csrfHeader := r.Header.Get("X-CSRF-Token")
		if csrfHeader == "" || !xsrftoken.Valid(csrfHeader, xsrfKey, identity, "") {
			csrfError(w, r, "Invalid CSRF header")
			return
		}

		// Both halves of the double submit must be the same token. Without this
		// any token valid for this identity pairs with any other.
		if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(csrfHeader)) != 1 {
			csrfError(w, r, "CSRF cookie and header do not match")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func newCSRFSessionId() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}

func csrfError(w http.ResponseWriter, r *http.Request, msg string) {
	slog.WarnContext(r.Context(), msg, "path", r.URL.Path)

	// Headers must be set before http.Error writes the status line and body.
	if !strings.HasPrefix(r.URL.Path, "/api") {
		w.Header().Set("HX-Redirect", "/error")
	}

	http.Error(w, "Access forbidden", http.StatusForbidden)
}

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Only in production: sending this over plain HTTP in development would
		// pin the browser to HTTPS for localhost.
		if config.IsProduction() {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}

// adminCSP is looser than the public policy: TinyMCE is loaded from its CDN and
// the post form carries inline scripts, so 'unsafe-inline' is still required
// here. Tightening it means moving that inline JS into a static file.
const adminCSP = `
			  default-src 'self';
			  script-src 'self' 'unsafe-inline' https://cdn.tiny.cloud;
			  style-src 'self' 'unsafe-inline' https://cdn.tiny.cloud;
			  font-src 'self';
			  img-src 'self' https: data: blob:;
			  connect-src 'self' https://cdn.tiny.cloud;
			  frame-ancestors 'none';
			  form-action 'self';
			  object-src 'none';`

// publicCSP allows no third party scripts at all: htmx and its json-enc
// extension are served from /static rather than a CDN.
const publicCSP = `
			  default-src 'self';
			  script-src 'self';
			  style-src 'self' 'unsafe-inline';
			  font-src 'self';
			  img-src 'self' https: data:;
			  connect-src 'self' https://res.cloudinary.com;
			  frame-ancestors 'none';
			  form-action 'self';
			  object-src 'none';`

func ContentSecurityPolicy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cspHeader := publicCSP
		if strings.HasPrefix(r.URL.Path, "/admin") {
			cspHeader = adminCSP
		}

		w.Header().Set("Content-Security-Policy", cspHeader)

		next.ServeHTTP(w, r)
	})
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
		absStaticDir, err := filepath.Abs(staticDir)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		absFilePath, err := filepath.Abs(filePath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if !strings.HasPrefix(absFilePath, absStaticDir) {
			http.NotFound(w, r)
			return
		}

		// Prefer the hash computed once at startup. Falling back to hashing the
		// file keeps assets added after boot working, but the common path must
		// not re-read every static file on every request.
		etag, ok := assetHash("/static/" + strings.TrimPrefix(r.URL.Path, "/"))
		if !ok {
			etag, err = generateETag(filePath)
			if err != nil {
				http.NotFound(w, r)
				return
			}
		}

		if match := r.Header.Get("If-None-Match"); match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		ext := strings.ToLower(filepath.Ext(r.URL.Path))
		switch ext {
		case ".css", ".js":
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable") // 1 year
		case ".woff2", ".woff":
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable") // 1 year
		case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable") // 1 year
		case ".html":
			w.Header().Set("Cache-Control", "no-cache") // Always request fresh
		default:
			w.Header().Set("Cache-Control", "public, max-age=86400") // 1 day
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
