FROM golang:1.23-alpine AS build

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-w -s -extldflags '-static'" \
    -a -installsuffix cgo \
    -o /pcclub-server ./cmd/server

FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -g 1001 -S pcclub && \
    adduser -u 1001 -S pcclub -G pcclub

WORKDIR /app

# Copy binary and set ownership
COPY --from=build /pcclub-server /app/pcclub-server
RUN chown pcclub:pcclub /app/pcclub-server && \
    chmod +x /app/pcclub-server

# Create data directory
RUN mkdir -p /app/data && \
    chown pcclub:pcclub /app/data

# Switch to non-root user
USER pcclub

# Environment variables with production defaults
ENV PORT=8080
ENV DATABASE_PATH=/app/data/pcclub.db
ENV READ_TIMEOUT=30s
ENV WRITE_TIMEOUT=30s
ENV IDLE_TIMEOUT=120s
ENV DB_MAX_OPEN_CONNS=50
ENV DB_MAX_IDLE_CONNS=10
ENV SEED_DEMO_DATA=false

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/health || exit 1

EXPOSE 8080

CMD ["/app/pcclub-server"]
