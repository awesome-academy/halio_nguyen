# ==============================================================================
# SUN Booking Tours — Backend API Dockerfile (Multi-stage Build)
# ==============================================================================

# Stage 1: Build Binary
FROM golang:alpine AS builder

WORKDIR /build

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates tzdata

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Compile optimized static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /build/api-server \
    ./cmd/server

# Stage 2: Minimal Runtime Container
FROM alpine:3.20

WORKDIR /app

# Install CA certificates and timezone data for secure HTTPS & correct time
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy compiled binary from builder
COPY --from=builder /build/api-server /app/api-server
COPY --from=builder /build/migrations /app/migrations

# Switch to non-root user
USER appuser:appgroup

EXPOSE 8080

ENTRYPOINT ["/app/api-server"]
