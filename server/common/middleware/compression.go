package middleware

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

const (
	acceptEncodingHeader  = "Accept-Encoding"
	contentEncodingHeader = "Content-Encoding"
	gzipEncoding          = "gzip"
	deflateEncoding       = "deflate"
)

type compressedResponseWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func (crw *compressedResponseWriter) Write(payload []byte) (int, error) {
	return crw.writer.Write(payload)
}

func (crw *compressedResponseWriter) Close() {
	if flateWriter, ok := crw.writer.(*flate.Writer); ok {
		flateWriter.Flush()
		flateWriter.Close()
	}

	if gzipWriter, ok := crw.writer.(*gzip.Writer); ok {
		gzipWriter.Flush()
		gzipWriter.Close()
	}

	if closer, ok := crw.writer.(io.Closer); ok {
		closer.Close()
	}
}

func initCompressionResponseWriter(writer http.ResponseWriter, req *http.Request) *compressedResponseWriter {
	encodings := strings.Split(req.Header.Get(acceptEncodingHeader), ",")
	for _, encoding := range encodings {
		switch strings.TrimSpace(encoding) {
		case gzipEncoding:
			slog.Info("Using gzip compression")
			writer.Header().Set(contentEncodingHeader, gzipEncoding)
			return &compressedResponseWriter{
				ResponseWriter: writer,
				writer:         gzip.NewWriter(writer),
			}

		case deflateEncoding:
			slog.Info("Using deflate encoding")
			writer.Header().Set(contentEncodingHeader, deflateEncoding)
			flateWriter, _ := flate.NewWriter(writer, flate.BestCompression)
			return &compressedResponseWriter{
				ResponseWriter: writer,
				writer:         flateWriter,
			}
		}
	}

	slog.Info("No compression used")
	return &compressedResponseWriter{
		ResponseWriter: writer,
		writer:         writer,
	}
}

func EnableCompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		compressionWriter := initCompressionResponseWriter(writer, req)
		defer compressionWriter.Close()
		next.ServeHTTP(compressionWriter, req)
	})
}
