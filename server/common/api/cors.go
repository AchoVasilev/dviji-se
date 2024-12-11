package api

import (
	"net/http"
	"slices"
	"strings"
)

var origins = []string{
	"localhost:4200",
}

var methods = []string{"GET", "POST", "PUT", "PATCH", "OPTIONS"}

var headers = []string{"Accept", "Accept-Encoding", "Content-Type", "Content-Length", "X-CSRF-TOKEN", "Authorization", "User-Agent"}

func CheckCORS(next http.Handler) http.Handler {
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
