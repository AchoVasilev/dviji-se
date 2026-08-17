package middleware

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	mux    *http.ServeMux
	server *httptest.Server
	expect = `{"message": "success"}`
)

func setup() func() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	return func() {
		server.Close()
	}
}

func getClient() *http.Client {
	mux.Handle("/test", getTestHandler())
	return &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}
}

func getTestHandler() http.Handler {
	return EnableCompression(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The ContentType middleware runs inside EnableCompression in the real
		// chain, so the type is always set before a handler writes.
		w.Header().Set(contentTypeHeader, "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expect))
	}))
}

func TestCompressionMiddleware_PlainText(t *testing.T) {
	teardown := setup()
	defer teardown()

	client := getClient()

	resp, err := client.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("client GET failed with unexpected error: %v", err)
	}
	defer resp.Body.Close()

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error while reading response body: %v", err)
	}

	if string(contents) != expect {
		t.Errorf("unexpected response content: got %q, want %q", string(contents), expect)
	}

	if resp.Header.Get(contentEncodingHeader) != "" {
		t.Errorf("unexpected header Content-Encoding: got %q, want empty", resp.Header.Get(contentEncodingHeader))
	}
}

func TestCompressionMiddleware_Gzip(t *testing.T) {
	teardown := setup()
	defer teardown()

	client := getClient()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	req.Header.Set(acceptEncodingHeader, gzipEncoding)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client GET failed with unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get(contentEncodingHeader) != gzipEncoding {
		t.Errorf("invalid Content-Encoding for gzip response: got %q, want %q", resp.Header.Get(contentEncodingHeader), gzipEncoding)
	}

	var buf bytes.Buffer
	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	_, err = buf.ReadFrom(reader)
	if err != nil {
		t.Fatalf("unexpected error while reading gzip response body: %v", err)
	}

	if buf.String() != expect {
		t.Errorf("unexpected gzip response content: got %q, want %q", buf.String(), expect)
	}
}

func TestCompressionMiddleware_Deflate(t *testing.T) {
	teardown := setup()
	defer teardown()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	req.Header.Set(acceptEncodingHeader, deflateEncoding)

	client := getClient()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("client GET failed with unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get(contentEncodingHeader) != deflateEncoding {
		t.Errorf("invalid Content-Encoding for deflate response: got %q, want %q", resp.Header.Get(contentEncodingHeader), deflateEncoding)
	}

	var buf bytes.Buffer
	reader := flate.NewReader(resp.Body)
	defer reader.Close()

	_, err = buf.ReadFrom(reader)
	if err != nil {
		t.Fatalf("unexpected error while reading deflate response body: %v", err)
	}

	if buf.String() != expect {
		t.Errorf("unexpected deflate response content: got %q, want %q", buf.String(), expect)
	}
}

func TestShouldCompress(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		contentType string
		want        bool
	}{
		{"html", http.StatusOK, "text/html; charset=utf-8", true},
		{"json", http.StatusOK, "application/json", true},
		{"rss", http.StatusOK, "application/rss+xml; charset=utf-8", true},
		{"svg", http.StatusOK, "image/svg+xml", true},
		{"png is already compressed", http.StatusOK, "image/png", false},
		{"webp is already compressed", http.StatusOK, "image/webp", false},
		{"woff2 is already compressed", http.StatusOK, "font/woff2", false},
		{"not modified carries no body", http.StatusNotModified, "text/html", false},
		{"no content carries no body", http.StatusNoContent, "text/html", false},
		{"unknown type", http.StatusOK, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldCompress(tt.status, tt.contentType); got != tt.want {
				t.Errorf("shouldCompress(%d, %q) = %v, want %v", tt.status, tt.contentType, got, tt.want)
			}
		})
	}
}

// A 304 must carry no body at all; the previous implementation appended a gzip
// header and footer to it.
func TestCompression_NotModifiedHasEmptyBody(t *testing.T) {
	handler := EnableCompression(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentTypeHeader, "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotModified)
	}))

	req := httptest.NewRequest(http.MethodGet, "/asset.css", nil)
	req.Header.Set(acceptEncodingHeader, gzipEncoding)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotModified {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotModified)
	}

	if encoding := resp.Header.Get(contentEncodingHeader); encoding != "" {
		t.Errorf("Content-Encoding = %q, want empty on a 304", encoding)
	}

	body, _ := io.ReadAll(resp.Body)
	if len(body) != 0 {
		t.Errorf("304 response body = %q, want empty", body)
	}
}

func TestCompression_VaryIsAlwaysSet(t *testing.T) {
	handler := EnableCompression(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentTypeHeader, "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}))

	// No Accept-Encoding at all: the response still varies by it.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if vary := w.Result().Header.Get("Vary"); vary != acceptEncodingHeader {
		t.Errorf("Vary = %q, want %q", vary, acceptEncodingHeader)
	}
}
