# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build all binaries
RUN mkdir -p /app/bin && \
    for cmd in ./cmd/*/; do \
        if [ -d "$cmd" ] && [ -f "$cmd/main.go" ]; then \
            binary=$(basename "$cmd"); \
            echo "Building $binary..."; \
            CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/"$binary" "$cmd"; \
        fi; \
    done

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /root/

# Copy binaries from builder
COPY --from=builder /app/bin/* ./

# Default command
CMD ["./api"]