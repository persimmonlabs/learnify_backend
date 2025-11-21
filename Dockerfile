# Learnify API Dockerfile
# Multi-stage build for optimal image size

# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

# Copy scripts and migrations for Railway startup
COPY --from=builder /app/scripts ./scripts/
COPY --from=builder /app/migrations ./migrations/

# Copy config example (override with mounted config in production)
COPY --from=builder /app/config/config.yaml.example ./config/

# Expose port
EXPOSE 8080

# Run
CMD ["./main"]
