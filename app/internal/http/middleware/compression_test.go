package middleware

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
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
	require.NoError(t, err, "client GET failed with unexpected error")
	defer resp.Body.Close()

	contents, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "unexpected error while reading response body")

	require.Equal(t, expect, string(contents), "unexpected response content")
	require.Empty(t, resp.Header.Get(contentEncodingHeader), "unexpected header: Content-Encoding")
}

func TestCompressionMiddleware_Gzip(t *testing.T) {
	teardown := setup()
	defer teardown()

	client := getClient()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	req.Header.Set(acceptEncodingHeader, gzipEncoding)

	resp, err := client.Do(req)
	require.NoError(t, err, "client GET failed with unexpected error")
	defer resp.Body.Close()

	require.Equal(t, gzipEncoding, resp.Header.Get(contentEncodingHeader), "invalid Content-Encoding for gzip response")

	var buf bytes.Buffer
	reader, err := gzip.NewReader(resp.Body)
	defer reader.Close()
	require.NoError(t, err, "failed to create gzip reader")

	_, err = buf.ReadFrom(reader)
	require.NoError(t, err, "unexpected error while reading gzip response body")

	require.Equal(t, expect, buf.String(), "unexpected gzip response content")
}

func TestCompressionMiddleware_Deflate(t *testing.T) {
	teardown := setup()
	defer teardown()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	req.Header.Set(acceptEncodingHeader, deflateEncoding)

	client := getClient()
	resp, err := client.Do(req)
	require.NoError(t, err, "client GET failed with unexpected error")
	defer resp.Body.Close()

	require.Equal(t, deflateEncoding, resp.Header.Get(contentEncodingHeader), "invalid Content-Encoding for deflate response")

	var buf bytes.Buffer
	reader := flate.NewReader(resp.Body)
	defer reader.Close()

	_, err = buf.ReadFrom(reader)
	require.NoError(t, err, "unexpected error while reading deflate response body")

	require.Equal(t, expect, buf.String(), "unexpected deflate response content")
}
