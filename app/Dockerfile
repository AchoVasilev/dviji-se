# ============================================
# Stage 1: Build CSS with Tailwind
# ============================================
FROM alpine:3.21 AS tailwind-builder

WORKDIR /build

# Download the tailwindcss CLI. The musl builds are the ones that run on
# alpine, and the arch is taken from the build platform rather than hardcoded.
ARG TARGETARCH
RUN apk add --no-cache curl libstdc++ libgcc && \
    case "$TARGETARCH" in \
      amd64) TW_ASSET=tailwindcss-linux-x64-musl ;; \
      arm64) TW_ASSET=tailwindcss-linux-arm64-musl ;; \
      *) echo "unsupported TARGETARCH: $TARGETARCH" >&2; exit 1 ;; \
    esac && \
    curl -fsSLO "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/${TW_ASSET}" && \
    chmod +x "${TW_ASSET}" && \
    mv "${TW_ASSET}" /usr/local/bin/tailwindcss

# Copy only files needed for CSS build
COPY tailwind.config.js ./
COPY web/static/css/main.css ./web/static/css/
COPY web/templates ./web/templates

# Build minified CSS
RUN tailwindcss -i ./web/static/css/main.css -o ./web/static/css/styles.css --minify

# ============================================
# Stage 2: Build Go binary
# ============================================
FROM golang:1.25-alpine AS go-builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Copy generated CSS from tailwind stage
COPY --from=tailwind-builder /build/web/static/css/styles.css ./web/static/css/styles.css

# Generate templ files (version pinned by the tool directive in go.mod)
RUN go tool templ generate

# Build static binary. Architecture is left to the builder so this image also
# builds natively on arm64 hosts.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /build/server \
    ./cmd/main.go

# ============================================
# Stage 3: Runtime image
# ============================================
FROM alpine:3.21 AS runtime

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# Copy binary
COPY --from=go-builder /build/server ./server

# Copy static assets
COPY --from=go-builder /build/web/static ./web/static

# Copy migrations (if running migrations from the app)
COPY --from=go-builder /build/cmd/db/migrations ./cmd/db/migrations

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
ENTRYPOINT ["./server"]
