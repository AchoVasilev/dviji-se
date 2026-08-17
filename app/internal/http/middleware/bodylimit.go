package middleware

import (
	"net/http"
	"strings"
)

const (
	// MaxJSONBodyBytes caps ordinary form and JSON submissions.
	MaxJSONBodyBytes int64 = 1 << 20 // 1 MiB

	// MaxUploadBodyBytes caps the admin upload endpoints, which parse
	// multipart forms up to 10 MiB and need headroom for the encoding.
	MaxUploadBodyBytes int64 = 12 << 20 // 12 MiB
)

// LimitRequestBody caps how much a client may send. Without it a request body
// is read until the client stops, so a single slow upload can consume memory
// indefinitely.
func LimitRequestBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil || r.Method == http.MethodGet || r.Method == http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}

		limit := MaxJSONBodyBytes
		if isUploadPath(r.URL.Path) {
			limit = MaxUploadBodyBytes
		}

		r.Body = http.MaxBytesReader(w, r.Body, limit)

		next.ServeHTTP(w, r)
	})
}

func isUploadPath(path string) bool {
	return strings.HasPrefix(path, "/api/admin/upload")
}
