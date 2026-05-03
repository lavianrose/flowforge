# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files for dependency download
COPY go.mod go.sum ./

# Download dependencies (cached separately)
RUN --mount=type=cache,target=/go/pkg-mod \
    go mod download

# Copy source code
COPY . .

# Copy entrypoint script
COPY docker-entrypoint.sh .

# Build all binaries
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s -extldflags '-static'" \
    -trimpath \
    -o api \
    ./cmd/api

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s -extldflags '-static'" \
    -trimpath \
    -o migrate \
    ./cmd/migrate

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s -extldflags '-static'" \
    -trimpath \
    -o seed \
    ./cmd/seed

# Runtime stage - minimal alpine image
FROM alpine:3.19

# Install only runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wget \
    && update-ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

WORKDIR /app

# Copy binaries and migrations from builder
COPY --from=builder --chown=appuser:appuser /app/api .
COPY --from=builder --chown=appuser:appuser /app/migrate .
COPY --from=builder --chown=appuser:appuser /app/seed .
COPY --from=builder --chown=appuser:appuser /app/migrations ./migrations

# Copy entrypoint script and make executable
COPY --from=builder --chown=appuser:appuser /app/docker-entrypoint.sh .
RUN chmod +x /app/docker-entrypoint.sh

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# Run entrypoint script (migrate → seed → api)
ENTRYPOINT ["/app/docker-entrypoint.sh"]
