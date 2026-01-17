package httputils

import "net/http"

// IsHTMXRequest checks if the request was made by HTMX
func IsHTMXRequest(req *http.Request) bool {
	return req.Header.Get("HX-Request") == "true"
}
