package middleware

import (
	"net/http"
	"strings"
)

const (
	// MaxJSONBodyBytes caps ordinary form and JSON submissions.
	MaxJSONBodyBytes int64 = 1 << 20 // 1 MiB

	// MaxUploadFileBytes is the size a single uploaded file may have. This is
	// the number the browser checks before sending, the handler checks on
	// arrival and the error message quotes, so the three cannot disagree.
	MaxUploadFileBytes int64 = 20 << 20 // 20 MiB

	// MaxUploadBodyBytes caps the whole request on the upload endpoints. It is
	// larger than the file because multipart adds boundaries, headers and the
	// remaining form fields; without the slack a file of exactly the permitted
	// size would be truncated by the body limit and reported as a broken form
	// rather than a clear "too large".
	MaxUploadBodyBytes int64 = MaxUploadFileBytes + (1 << 20)
)

// MaxUploadFileMB is the limit in whole megabytes, for messages and for the
// data attribute the upload page hands to the browser.
func MaxUploadFileMB() int64 {
	return MaxUploadFileBytes / (1 << 20)
}

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
