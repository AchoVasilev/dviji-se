# ============================================
# Stage 1: Build CSS with Tailwind
# ============================================
FROM alpine:3.21 AS tailwind-builder

WORKDIR /build

# Download tailwindcss CLI
RUN apk add --no-cache curl && \
    curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 && \
    chmod +x tailwindcss-linux-x64 && \
    mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss

# Copy only files needed for CSS build
COPY tailwind.config.js ./
COPY web/static/css/main.css ./web/static/css/
COPY web/templates ./web/templates

# Build minified CSS
RUN tailwindcss -i ./web/static/css/main.css -o ./web/static/css/styles.css --minify

# ============================================
# Stage 2: Build Go binary
# ============================================
FROM golang:1.24-alpine AS go-builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Copy generated CSS from tailwind stage
COPY --from=tailwind-builder /build/web/static/css/styles.css ./web/static/css/styles.css

# Generate templ files
RUN templ generate

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
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
