package middleware

import (
	"net/http"
	"os"
	"slices"
	"strings"
)

var origins = func() []string {
	envOrigins := os.Getenv("CORS_ORIGINS")
	if envOrigins == "" {
		return []string{}
	}
	return strings.Split(envOrigins, ",")
}()

var methods = []string{"GET", "POST", "PUT", "PATCH", "OPTIONS"}

var headers = []string{"Accept", "Accept-Encoding", "Content-Type", "Content-Length", "X-CSRF-TOKEN", "Authorization", "User-Agent"}

func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		if isPreflight(req) {
			origin := req.Header.Get("Origin")
			method := req.Header.Get("Access-Control-Request-Method")

			if slices.Contains(origins, origin) && slices.Contains(methods, method) {
				writer.Header().Add("Access-Control-Allow-Origin", origin)
				writer.Header().Add("Access-Control-Allow-Methods", strings.Join(methods, ", "))
			}
		} else {
			origin := req.Header.Get("Origin")
			if slices.Contains(origins, origin) {
				writer.Header().Add("Access-Control-Allow-Origin", origin)
				writer.Header().Add("Access-Control-Allow-Headers", strings.Join(headers, ", "))
			}
		}

		writer.Header().Add("Vary", "Origin")
		next.ServeHTTP(writer, req)
	})
}

func isPreflight(req *http.Request) bool {
	return req.Method == "OPTIONS" &&
		req.Header.Get("Origin") != "" &&
		req.Header.Get("Access-Control-Request-Method") != ""
}
