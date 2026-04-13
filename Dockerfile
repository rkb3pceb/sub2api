# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go module files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -trimpath \
    -o sub2api \
    ./main.go

# Final stage - minimal runtime image
FROM scratch

# Copy timezone data and CA certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled binary
COPY --from=builder /app/sub2api /sub2api

# Expose the default port
# Note: override at runtime with -e PORT=<port> if needed
EXPOSE 8080

# Run as non-root by default (numeric UID for compatibility with scratch)
# UID 65534 is the conventional 'nobody' user
USER 65534:65534

# Set the entrypoint
ENTRYPOINT ["/sub2api"]
