# ==========================================
# Multi-stage build for SysMonitorBot
# ==========================================

# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o sysmonitorbot ./cmd

# ==========================================
# Runtime stage
# ==========================================
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Set timezone to Asia/Tokyo
ENV TZ=Asia/Tokyo

# Copy binary from builder
COPY --from=builder /app/sysmonitorbot .

# Default interval (can be overridden)
ENV INTERVAL=1800

# Expose health check port
EXPOSE 8000

# Run the bot
ENTRYPOINT ["./sysmonitorbot"]
CMD ["-interval", "1800"]