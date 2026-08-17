package middleware

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

const (
	acceptEncodingHeader  = "Accept-Encoding"
	contentEncodingHeader = "Content-Encoding"
	contentLengthHeader   = "Content-Length"
	contentTypeHeader     = "Content-Type"
	gzipEncoding          = "gzip"
	deflateEncoding       = "deflate"
)

// compressedResponseWriter defers the decision to compress until the status and
// Content-Type are known. Compressing eagerly wraps bodies that must stay empty
// (304, 204) and re-compresses formats that are already compressed.
type compressedResponseWriter struct {
	http.ResponseWriter

	encoding    string
	compressor  io.WriteCloser
	wroteHeader bool
}

func (crw *compressedResponseWriter) WriteHeader(status int) {
	if crw.wroteHeader {
		crw.ResponseWriter.WriteHeader(status)
		return
	}

	crw.wroteHeader = true

	if shouldCompress(status, crw.Header().Get(contentTypeHeader)) {
		crw.Header().Set(contentEncodingHeader, crw.encoding)

		// The compressed length is not known ahead of time.
		crw.Header().Del(contentLengthHeader)

		switch crw.encoding {
		case gzipEncoding:
			crw.compressor = gzip.NewWriter(crw.ResponseWriter)
		case deflateEncoding:
			if flateWriter, err := flate.NewWriter(crw.ResponseWriter, flate.BestCompression); err == nil {
				crw.compressor = flateWriter
			}
		}
	}

	crw.ResponseWriter.WriteHeader(status)
}

func (crw *compressedResponseWriter) Write(payload []byte) (int, error) {
	if !crw.wroteHeader {
		crw.WriteHeader(http.StatusOK)
	}

	if crw.compressor != nil {
		return crw.compressor.Write(payload)
	}

	return crw.ResponseWriter.Write(payload)
}

// Flush keeps HTMX streaming responses working through the wrapper.
func (crw *compressedResponseWriter) Flush() {
	if flusher, ok := crw.compressor.(interface{ Flush() error }); ok {
		_ = flusher.Flush()
	}

	if flusher, ok := crw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (crw *compressedResponseWriter) Close() {
	if crw.compressor != nil {
		_ = crw.compressor.Close()
	}
}

// shouldCompress reports whether a body of this type is worth compressing.
// Statuses that carry no body must be left alone entirely.
func shouldCompress(status int, contentType string) bool {
	if status == http.StatusNoContent || status == http.StatusNotModified {
		return false
	}

	mediaType, _, _ := strings.Cut(contentType, ";")
	mediaType = strings.TrimSpace(strings.ToLower(mediaType))

	switch {
	case mediaType == "":
		return false
	case strings.HasPrefix(mediaType, "text/"):
		return true
	case mediaType == "application/json",
		mediaType == "application/xml",
		mediaType == "application/rss+xml",
		mediaType == "application/javascript",
		mediaType == "image/svg+xml":
		return true
	default:
		// Images, fonts and archives are already compressed; running them
		// through gzip costs CPU and typically grows the payload.
		return false
	}
}

func negotiateEncoding(req *http.Request) string {
	for encoding := range strings.SplitSeq(req.Header.Get(acceptEncodingHeader), ",") {
		switch strings.TrimSpace(strings.ToLower(encoding)) {
		case gzipEncoding:
			return gzipEncoding
		case deflateEncoding:
			return deflateEncoding
		}
	}

	return ""
}

func EnableCompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		// Responses differ by Accept-Encoding whether or not this one ends up
		// compressed, so caches need this on every response.
		writer.Header().Add("Vary", acceptEncodingHeader)

		encoding := negotiateEncoding(req)
		if encoding == "" {
			next.ServeHTTP(writer, req)
			return
		}

		compressionWriter := &compressedResponseWriter{
			ResponseWriter: writer,
			encoding:       encoding,
		}
		defer compressionWriter.Close()

		next.ServeHTTP(compressionWriter, req)
	})
}
