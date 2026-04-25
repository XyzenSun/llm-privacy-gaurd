# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Cache go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary (CGO_ENABLED=0 works because glebarez/sqlite is pure Go)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

# Runtime stage
FROM alpine:3.19

# Install ca-certificates for HTTPS calls to LLM APIs
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy binary and web assets
COPY --from=builder /server .
COPY --from=builder /app/web ./web

# Create data directory for SQLite database
RUN mkdir -p /app/data && chown -R appuser:appgroup /app

USER appuser

# Default port
EXPOSE 19999

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:19999/api/v1/health || exit 1

ENTRYPOINT ["/app/server"]