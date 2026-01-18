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
